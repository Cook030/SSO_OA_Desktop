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
	reader *kafka.Reader
	mapper *mapper.Mapper
	store  *store.Store
	log    *zap.Logger
	dead   *DeadLetter
}

// New 构造 Consumer。
func New(cfg *config.Config, st *store.Store, mp *mapper.Mapper, log *zap.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
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
	return &Consumer{
		cfg:    cfg,
		reader: reader,
		mapper: mp,
		store:  st,
		log:    log,
		dead:   NewDeadLetter("dead_letter", log),
	}
}

// Close 释放消费连接。
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// Run 主循环: 攒批拉取 -> 处理入库 -> 成功才提交 offset。
func (c *Consumer) Run(ctx context.Context) error {
	c.log.Info("audit-consumer 启动",
		zap.Strings("brokers", c.cfg.Kafka.Brokers),
		zap.String("topic", c.cfg.Kafka.Topic),
		zap.String("group", c.cfg.Kafka.GroupID))

	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return nil
		}

		batch, err := c.fetchBatch(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.log.Warn("拉取Kafka消息失败", zap.Error(err))
			if !c.sleep(ctx, backoff) {
				return nil
			}
			backoff = nextBackoff(backoff, c.cfg.Kafka.RetryMaxBackoffMs)
			continue
		}
		if len(batch) == 0 {
			continue
		}

		backoff = time.Second
		if err := c.process(ctx, batch); err != nil {
			c.log.Error("处理批次失败, 暂不提交offset并重试",
				zap.Int("count", len(batch)), zap.Error(err))
			if !c.sleep(ctx, backoff) {
				return nil
			}
			continue
		}

		if err := c.reader.CommitMessages(ctx, batch...); err != nil {
			// 提交失败: 消息会再次被拉到, 由 dedup_key 幂等兜底。
			c.log.Error("提交offset失败", zap.Error(err))
			continue
		}
		c.log.Info("批次处理完成并已提交",
			zap.Int("count", len(batch)))
	}
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
			return nil, fmt.Errorf("拉取后续消息失败: %w", err)
		}
		batch = append(batch, msg)
	}
	return batch, nil
}

// process 将批次解析为审计记录并批量入库。
// 单条消息解析失败只走死信, 不影响整批提交; 入库失败才整体返回错误。
func (c *Consumer) process(ctx context.Context, batch []kafka.Message) error {
	records := make([]*mapper.Record, 0)
	for _, msg := range batch {
		flats, err := canal.DecodeValue(msg.Value)
		if err != nil {
			c.dead.Write(msg, err)
			continue
		}
		for _, flat := range flats {
			if flat == nil || !flat.IsDML() || flat.IsDDLEvent() {
				continue
			}
			rs, err := c.mapper.Build(flat)
			if err != nil {
				c.dead.Write(msg, err)
				continue
			}
			records = append(records, rs...)
		}
	}
	if len(records) == 0 {
		return nil
	}

	skipped, err := c.store.BatchInsert(ctx, records)
	if err != nil {
		return err
	}
	if skipped > 0 {
		c.log.Warn("部分审计记录因外键被跳过(操作人已删除)",
			zap.Int("skipped", skipped))
	}
	c.log.Info("审计记录入库",
		zap.Int("batch", len(records)), zap.Int("skipped", skipped))
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
