// Package consumer 消费 Kafka 上 Canal 投递的 binlog 事件并落库。
// 采用 at-least-once + dedup_key 唯一键, 实现最终恰好一次。
package consumer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"mh-audit-consumer/internal/canal"
	"mh-audit-consumer/internal/config"
	"mh-audit-consumer/internal/mapper"
	"mh-audit-consumer/internal/store"
)

// Consumer Kafka 消费器。
type Consumer struct {
	cfg    *config.Config
	reader messageReader
	mapper *mapper.Mapper
	store  auditStore
	log    *zap.Logger
	dead   deadLetterWriter
}

// messageReader 是 Consumer 需要的 Kafka 最小能力集，隔离具体客户端实现。
type messageReader interface {
	FetchMessage(context.Context) (kafka.Message, error)
	CommitMessages(context.Context, ...kafka.Message) error
	Close() error
}

// auditStore 是 Consumer 需要的审计持久化能力，便于替换与单测。
type auditStore interface {
	BatchInsert(context.Context, []*mapper.Record) error
}

// deadLetterWriter 隔离死信存储；死信未可靠写入时，消息不能被确认。
type deadLetterWriter interface {
	Write(kafka.Message, error) error
}

// New 构造 Consumer。
func New(cfg *config.Config, st *store.Store, mp *mapper.Mapper, log *zap.Logger) *Consumer {
	return newConsumer(cfg, newKafkaReader(cfg), st, mp, log)
}

func newKafkaReader(cfg *config.Config) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		GroupID:        cfg.Kafka.GroupID,
		Topic:          cfg.Kafka.Topic,
		MinBytes:       1,
		MaxBytes:       10 * 1024 * 1024, // 10MB
		MaxWait:        time.Duration(cfg.Kafka.FetchMaxWaitMs) * time.Millisecond,
		SessionTimeout: time.Duration(cfg.Kafka.SessionTimeoutMs) * time.Millisecond,
		CommitInterval: 0, // 关闭自动提交, 入库成功后再手动 CommitMessages
		StartOffset:    kafka.FirstOffset,
	})
}

func newConsumer(cfg *config.Config, reader messageReader, st auditStore, mp *mapper.Mapper, log *zap.Logger) *Consumer {
	return &Consumer{
		cfg:    cfg,
		reader: reader,
		mapper: mp,
		store:  st,
		log:    log,
		dead:   NewDeadLetter(cfg.DeadLetter, log),
	}
}

// Close 释放消费连接。
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// Run 主循环: 攒批拉取 -> 处理入库 -> 成功才提交 offset。
func (c *Consumer) Run(ctx context.Context) error {
	c.logStartup()

	backoff := time.Second
	var pending []kafka.Message
	processed := false
	for {
		if isCancelled(ctx) {
			return nil
		}

		if len(pending) == 0 {
			batch, err := c.fetchBatch(ctx)
			if len(batch) > 0 {
				// 已经取得的消息不能因后续拉取失败而丢弃；处理完成前不再拉新消息。
				pending = batch
				processed = false
				backoff = time.Second
				if err != nil {
					c.log.Warn("攒批时拉取Kafka消息失败，将先处理已获取消息",
						zap.Int("count", len(pending)), zap.Error(err))
				}
			} else if err != nil {
				if isCancelled(ctx) {
					return nil
				}
				if !c.retryAfterFetchFailure(ctx, err, backoff) {
					return nil
				}
				backoff = nextBackoff(backoff, c.cfg.Kafka.RetryMaxBackoffMs)
				continue
			} else {
				continue
			}
		}

		if !processed {
			if err := c.process(ctx, pending); err != nil {
				c.logProcessFailure(len(pending), err)
				if !c.sleep(ctx, backoff) {
					return nil
				}
				backoff = nextBackoff(backoff, c.cfg.Kafka.RetryMaxBackoffMs)
				continue
			}
			processed = true
		}

		if !c.commit(ctx, pending) {
			if !c.sleep(ctx, backoff) {
				return nil
			}
			backoff = nextBackoff(backoff, c.cfg.Kafka.RetryMaxBackoffMs)
			continue
		}
		c.logBatchCommitted(len(pending))
		pending = nil
		processed = false
		backoff = time.Second
	}
}

func (c *Consumer) logStartup() {
	c.log.Info("audit-consumer 启动",
		zap.Strings("brokers", c.cfg.Kafka.Brokers),
		zap.String("topic", c.cfg.Kafka.Topic),
		zap.String("group", c.cfg.Kafka.GroupID))
}

func isCancelled(ctx context.Context) bool { return ctx.Err() != nil }

func (c *Consumer) retryAfterFetchFailure(ctx context.Context, err error, backoff time.Duration) bool {
	c.log.Warn("拉取Kafka消息失败", zap.Error(err))
	return c.sleep(ctx, backoff)
}

func (c *Consumer) logProcessFailure(batchSize int, err error) {
	c.log.Error("处理批次失败, 暂不提交offset并重试", zap.Int("count", batchSize), zap.Error(err))
}

// commit 仅在审计记录已成功持久化后提交 offset；提交失败由 dedup_key 支持安全重放。
func (c *Consumer) commit(ctx context.Context, batch []kafka.Message) bool {
	if err := c.reader.CommitMessages(ctx, batch...); err != nil {
		c.log.Error("提交offset失败", zap.Error(err))
		return false
	}
	return true
}

func (c *Consumer) logBatchCommitted(batchSize int) {
	c.log.Info("批次处理完成并已提交", zap.Int("count", batchSize))
}

// fetchBatch 拉取一批消息: 首条阻塞等待, 之后按 flush_interval 攒批。
func (c *Consumer) fetchBatch(ctx context.Context) ([]kafka.Message, error) {
	first, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}
	batch := []kafka.Message{first}

	deadline := time.Now().Add(c.cfg.Kafka.FlushInterval())
	for len(batch) < c.cfg.Kafka.BatchSize {
		remain := time.Until(deadline)
		if remain <= 0 {
			break
		}
		if remain > 100*time.Millisecond {
			remain = 100 * time.Millisecond
		}
		pollCtx, cancel := context.WithTimeout(ctx, remain)
		msg, err := c.reader.FetchMessage(pollCtx)
		cancel()
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if errors.Is(err, context.DeadlineExceeded) {
				break
			}
			return batch, fmt.Errorf("拉取后续消息失败: %w", err)
		}
		batch = append(batch, msg)
	}
	return batch, nil
}

// process 将批次解析为审计记录并批量入库。
// 单条消息解析失败只走死信, 不影响整批提交; 入库失败才整体返回错误。
func (c *Consumer) process(ctx context.Context, batch []kafka.Message) error {
	records, err := c.buildRecords(batch)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}
	return c.persistRecords(ctx, records)
}

func (c *Consumer) buildRecords(batch []kafka.Message) ([]*mapper.Record, error) {
	records := make([]*mapper.Record, 0)
	for _, msg := range batch {
		rs, err := c.buildMessageRecords(msg)
		if err != nil {
			return nil, err
		}
		records = append(records, rs...)
	}
	return records, nil
}

func (c *Consumer) buildMessageRecords(msg kafka.Message) ([]*mapper.Record, error) {
	flats, err := canal.DecodeValue(msg.Value)
	if err != nil {
		if deadErr := c.dead.Write(msg, err); deadErr != nil {
			return nil, fmt.Errorf("写入解析失败死信失败: %w", deadErr)
		}
		return nil, nil
	}

	records := make([]*mapper.Record, 0)
	for _, flat := range flats {
		if !isAuditable(flat) {
			continue
		}
		rs, err := c.mapper.Build(flat)
		if err != nil {
			if deadErr := c.dead.Write(msg, err); deadErr != nil {
				return nil, fmt.Errorf("写入映射失败死信失败: %w", deadErr)
			}
			continue
		}
		records = append(records, rs...)
	}
	return records, nil
}

func isAuditable(flat *canal.FlatMessage) bool {
	return flat != nil && flat.IsDML() && !flat.IsDDLEvent()
}

func (c *Consumer) persistRecords(ctx context.Context, records []*mapper.Record) error {
	if err := c.store.BatchInsert(ctx, records); err != nil {
		return err
	}
	c.log.Info("审计记录入库", zap.Int("batch", len(records)))
	return nil
}

func nextBackoff(cur time.Duration, maxMs int) time.Duration {
	next := cur * 2
	if cap := time.Duration(maxMs) * time.Millisecond; next > cap {
		return cap
	}
	return next
}

func (c *Consumer) sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
