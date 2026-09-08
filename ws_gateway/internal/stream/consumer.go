// Package stream 消费 SSO 写入 Redis Stream 的会话事件，并投递给本机连接。
//
// 可靠性保证：
//   - 使用 Consumer Group：Gateway 崩溃时未确认的消息留在 PEL，
//     可由本实例恢复，或其他实例通过 XAUTOCLAIM 接管，不会丢消息；
//   - 投递顺序固定为：写 sso:delivery → XACK → 推给连接。
//     先落盘再确认，保证"确认过的事件一定能补发"。
package stream

import (
	"context"
	"errors"
	"strconv"
	"time"

	"mh-ws-gateway/internal/delivery"
	"mh-ws-gateway/internal/hub"
	"mh-ws-gateway/internal/protocol"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Event 一条会话领域事件
type Event struct {
	StreamID        string // Stream 消息 ID（用于 XACK）
	EventID         string // 业务事件 ID，客户端据此幂等
	Type            string // session_replaced / session_terminated
	UserID          uint64
	TargetSessionID string
	Reason          string
	CreatedAt       int64
}

// Options 消费者参数
type Options struct {
	StreamKey       string
	Group           string
	ConsumerID      string
	BatchSize       int64
	Block           time.Duration
	PendingIdle     time.Duration
	PendingInterval time.Duration
	// TerminateGrace 会话终结事件投递后的宽限期：留出 ACK 时间，超时强制关闭连接
	TerminateGrace time.Duration
}

// Consumer 会话事件消费者
type Consumer struct {
	rdb      *redis.Client
	opt      Options
	hub      *hub.Hub
	delivery *delivery.Store
	log      *zap.Logger
}

// NewConsumer 创建消费者
func NewConsumer(rdb *redis.Client, opt Options, h *hub.Hub, store *delivery.Store, log *zap.Logger) *Consumer {
	return &Consumer{rdb: rdb, opt: opt, hub: h, delivery: store, log: log}
}

// EnsureGroup 幂等创建消费组与 Stream（不存在时用 MKSTREAM 创建）
func (c *Consumer) EnsureGroup(ctx context.Context) error {
	err := c.rdb.XGroupCreateMkStream(ctx, c.opt.StreamKey, c.opt.Group, "0").Err()
	if err != nil && !isBusyGroup(err) {
		return err
	}
	return nil
}

// Run 消费循环；ctx 取消后退出
func (c *Consumer) Run(ctx context.Context) {
	// 悬挂消息接管放在独立协程：即使长时间没有新消息，
	// 其他实例崩溃留下的 PEL 也能被本实例接管
	go c.claimLoop(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		streams, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    c.opt.Group,
			Consumer: c.opt.ConsumerID,
			Streams:  []string{c.opt.StreamKey, ">"},
			Count:    c.opt.BatchSize,
			Block:    c.opt.Block,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue // 本轮无新消息
			}
			if ctx.Err() != nil {
				return
			}
			c.log.Error("读取会话事件流失败", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		for _, s := range streams {
			for _, m := range s.Messages {
				c.handle(ctx, m.ID, m.Values)
			}
		}
	}
}

// claimLoop 周期性接管超时未确认的悬挂消息
func (c *Consumer) claimLoop(ctx context.Context) {
	ticker := time.NewTicker(c.opt.PendingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.claimPending(ctx)
		}
	}
}

// claimPending 通过 XAUTOCLAIM 接管他人遗留的悬挂消息
func (c *Consumer) claimPending(ctx context.Context) {
	start := "0-0"
	for {
		msgs, next, err := c.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   c.opt.StreamKey,
			Group:    c.opt.Group,
			Consumer: c.opt.ConsumerID,
			MinIdle:  c.opt.PendingIdle,
			Start:    start,
			Count:    c.opt.BatchSize,
		}).Result()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.log.Error("接管悬挂消息失败", zap.Error(err))
			return
		}
		for _, m := range msgs {
			c.handle(ctx, m.ID, m.Values)
		}
		if len(msgs) == 0 || next == "" || next == "0-0" {
			return
		}
		start = next
	}
}

// handle 处理单条 Stream 消息：落盘 → 确认 → 投递
func (c *Consumer) handle(ctx context.Context, id string, values map[string]interface{}) {
	ev := parseEvent(id, values)
	if ev.EventID == "" || ev.TargetSessionID == "" {
		// 格式非法的毒丸消息：记录后确认，避免永久阻塞消费组
		c.log.Error("会话事件格式非法，已跳过",
			zap.String("stream_id", id), zap.Any("values", values))
		_ = c.rdb.XAck(ctx, c.opt.StreamKey, c.opt.Group, id).Err()
		return
	}

	payload, err := protocol.MarshalEvent(ev.EventID, ev.Type, ev.Reason)
	if err != nil {
		c.log.Error("序列化事件失败", zap.String("event_id", ev.EventID), zap.Error(err))
		_ = c.rdb.XAck(ctx, c.opt.StreamKey, c.opt.Group, id).Err()
		return
	}

	// 1. 先落盘，保证"至少一次"
	if err := c.delivery.Save(ctx, ev.TargetSessionID, ev.EventID, payload); err != nil {
		c.log.Error("保存待投递事件失败，稍后重试",
			zap.String("event_id", ev.EventID), zap.Error(err))
		return // 不 XACK：留在 PEL 等待重试
	}
	// 2. 落盘成功才确认
	if err := c.rdb.XAck(ctx, c.opt.StreamKey, c.opt.Group, id).Err(); err != nil {
		c.log.Error("确认 Stream 消息失败", zap.String("stream_id", id), zap.Error(err))
	}
	// 3. 投递给本机连接；不在线则等重连补发
	c.deliver(ev, payload)
}

// deliver 把事件推给目标会话的连接
func (c *Consumer) deliver(ev *Event, payload []byte) {
	target, ok := c.hub.Get(ev.TargetSessionID)
	if !ok {
		c.log.Debug("目标会话不在线，事件留待重连补发",
			zap.String("session_id", ev.TargetSessionID),
			zap.String("event_id", ev.EventID))
		return
	}
	if err := target.Send(payload); err != nil {
		if errors.Is(err, protocol.ErrQueueFull) {
			// 关键事件不能静默丢弃：关闭慢连接。
			// 客户端重连时会因旧会话已失效被拒，并从 HTTP 401 得到下线原因。
			c.log.Warn("写队列已满，关闭慢连接",
				zap.String("session_id", ev.TargetSessionID),
				zap.String("event_id", ev.EventID))
			target.Close(protocol.CloseSlowConsumer, "write queue full")
			return
		}
		c.log.Warn("投递事件失败",
			zap.String("session_id", ev.TargetSessionID),
			zap.String("event_id", ev.EventID),
			zap.Error(err))
		return
	}

	// 会话终结事件：给客户端留出 ACK 与自我清理的时间，随后强制关闭。
	// 即使客户端忽略事件，也不会留下一条"看起来还活着"的长连接。
	if isTerminateEvent(ev.Type) {
		go c.closeAfterGrace(target, ev.TargetSessionID)
	}
}

// isTerminateEvent 是否为会话终结类事件
func isTerminateEvent(eventType string) bool {
	return eventType == protocol.EventTypeSessionReplaced || eventType == protocol.EventTypeSessionTerminated
}

// closeAfterGrace 宽限期后关闭连接（Close 幂等，客户端已自行关闭时无副作用）
func (c *Consumer) closeAfterGrace(target hub.Connection, sessionID string) {
	time.Sleep(c.opt.TerminateGrace)
	c.log.Info("会话终结事件已投递，关闭连接", zap.String("session_id", sessionID))
	target.Close(protocol.CloseSessionInvalid, "session terminated")
}

// parseEvent 解析 Stream 字段；字段缺失时保留零值由调用方判定
func parseEvent(id string, values map[string]interface{}) *Event {
	ev := &Event{
		StreamID:        id,
		EventID:         fieldString(values, "eventId"),
		Type:            fieldString(values, "type"),
		TargetSessionID: fieldString(values, "targetSessionId"),
		Reason:          fieldString(values, "reason"),
	}
	if raw := fieldString(values, "userId"); raw != "" {
		if uid, err := strconv.ParseUint(raw, 10, 64); err == nil {
			ev.UserID = uid
		}
	}
	if raw := fieldString(values, "createdAt"); raw != "" {
		if ts, err := strconv.ParseInt(raw, 10, 64); err == nil {
			ev.CreatedAt = ts
		}
	}
	return ev
}

func fieldString(values map[string]interface{}, key string) string {
	if v, ok := values[key].(string); ok {
		return v
	}
	return ""
}

func isBusyGroup(err error) bool {
	return err != nil && err.Error() == "BUSYGROUP Consumer Group name already exists"
}
