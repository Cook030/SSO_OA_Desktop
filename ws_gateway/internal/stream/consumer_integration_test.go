package stream

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"mh-ws-gateway/internal/delivery"
	"mh-ws-gateway/internal/hub"
	"mh-ws-gateway/internal/protocol"
	"mh-ws-gateway/internal/rtest"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// 本文件用真实 Redis 验证会话事件消费的可靠性语义：
//   - 多实例广播投递：每个 Gateway 实例独立消费组消费同一事件流，
//     只有持有目标连接的实例真正推送，不会重复投递也不会漏；
//   - 落盘先行：事件必须先写入 sso:delivery 再 XACK（未确认事件在重连时补发）；
//   - PEL 恢复：实例崩溃后遗留的未确认消息可被重新接管处理（XAUTOCLAIM，Redis ≥6.2）；
//   - 毒丸消息不阻塞消费组；terminate 事件投递后强关连接。

// memConn 测试用连接：记录收到的消息与关闭码
type memConn struct {
	sessionID string
	send      chan []byte
	closed    chan struct{}
	code      atomic.Int32
	once      sync.Once
}

func newMemConn(sessionID string) *memConn {
	return &memConn{
		sessionID: sessionID,
		send:      make(chan []byte, 32),
		closed:    make(chan struct{}),
	}
}

func (m *memConn) SessionID() string { return m.sessionID }
func (m *memConn) Send(payload []byte) error {
	select {
	case m.send <- payload:
		return nil
	default:
		return protocol.ErrQueueFull
	}
}
func (m *memConn) Close(code int, _ string) {
	m.code.CompareAndSwap(0, int32(code))
	m.once.Do(func() { close(m.closed) })
}

func (m *memConn) closeCode() int { return int(m.code.Load()) }

type eventFixture struct {
	EventID         string
	Type            string
	UserID          uint64
	TargetSessionID string
	Reason          string
}

func xadd(t *testing.T, rdb *redis.Client, ev eventFixture) string {
	t.Helper()
	id, err := rdb.XAdd(context.Background(), &redis.XAddArgs{
		Stream: protocol.SessionEventStreamKey,
		Values: map[string]interface{}{
			"eventId":         ev.EventID,
			"type":            ev.Type,
			"userId":          strconv.FormatUint(ev.UserID, 10),
			"targetSessionId": ev.TargetSessionID,
			"reason":          ev.Reason,
			"createdAt":       strconv.FormatInt(time.Now().UnixMilli(), 10),
		},
	}).Result()
	if err != nil {
		t.Fatalf("XADD 失败: %v", err)
	}
	return id
}

func makeOpts(group, consumer string) Options {
	return Options{
		StreamKey:       protocol.SessionEventStreamKey,
		Group:           group,
		ConsumerID:      consumer,
		BatchSize:       16,
		Block:           100 * time.Millisecond,
		PendingIdle:     200 * time.Millisecond,
		PendingInterval: 150 * time.Millisecond,
		TerminateGrace:  300 * time.Millisecond,
	}
}

// waitEvent 从连接上等待一条事件（带超时），超时即测试失败
func waitEvent(t *testing.T, c *memConn, desc string) protocol.ServerMessage {
	t.Helper()
	select {
	case payload := <-c.send:
		var msg protocol.ServerMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			t.Fatalf("%s: 消息非法 JSON: %v", desc, err)
		}
		return msg
	case <-time.After(5 * time.Second):
		t.Fatalf("%s: 等待事件超时", desc)
		return protocol.ServerMessage{}
	}
}

// xpendingCount 消费组内未确认消息数
func xpendingCount(t *testing.T, rdb *redis.Client, group string) int64 {
	t.Helper()
	p, err := rdb.XPending(context.Background(), protocol.SessionEventStreamKey, group).Result()
	if err != nil {
		t.Fatalf("XPENDING 失败: %v", err)
	}
	return p.Count
}

func TestMultiInstanceFanoutDelivery(t *testing.T) {
	rdb := rtest.Start(t)
	rtest.Flush(t, rdb)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := delivery.NewStore(rdb, time.Hour)

	// 两个"Gateway 实例"，各自独立的消费组与连接注册表
	hubA := hub.NewHub()
	hubB := hub.NewHub()
	connA := newMemConn("session-A")
	connB := newMemConn("session-B")
	hubA.Register(connA)
	hubB.Register(connB)

	consA := NewConsumer(rdb, makeOpts("ws-gateway:gw-1", "gw-1"), hubA, store, zap.NewNop())
	consB := NewConsumer(rdb, makeOpts("ws-gateway:gw-2", "gw-2"), hubB, store, zap.NewNop())
	for _, c := range []*Consumer{consA, consB} {
		if err := c.EnsureGroup(ctx); err != nil {
			t.Fatalf("创建消费组失败: %v", err)
		}
	}
	go consA.Run(ctx)
	go consB.Run(ctx)

	// 两个用户的事件同时入流
	xadd(t, rdb, eventFixture{EventID: "eA-1", Type: protocol.EventTypeSessionReplaced, UserID: 1, TargetSessionID: "session-A", Reason: "replaced_by_new_login"})
	xadd(t, rdb, eventFixture{EventID: "eB-1", Type: protocol.EventTypeSessionTerminated, UserID: 2, TargetSessionID: "session-B", Reason: "logged_out"})

	// 每个事件恰好被持有目标连接的实例推送给对应连接
	msgA := waitEvent(t, connA, "实例 A 投递 session-A")
	if msgA.EventID != "eA-1" || msgA.EventType != protocol.EventTypeSessionReplaced {
		t.Fatalf("连接 A 收到错误事件: %+v", msgA)
	}
	msgB := waitEvent(t, connB, "实例 B 投递 session-B")
	if msgB.EventID != "eB-1" || msgB.EventType != protocol.EventTypeSessionTerminated {
		t.Fatalf("连接 B 收到错误事件: %+v", msgB)
	}

	// 短暂停顿后确认没有串投 / 重复投递
	time.Sleep(500 * time.Millisecond)
	if len(connA.send) != 0 {
		t.Fatalf("连接 A 不应收到 session-B 的事件（串投/重复投递），队列残留 %d", len(connA.send))
	}
	if len(connB.send) != 0 {
		t.Fatalf("连接 B 不应收到 session-A 的事件（串投/重复投递），队列残留 %d", len(connB.send))
	}

	// 事件已 XACK：两个消费组都不留未确认消息
	rtest.Eventually(t, 5*time.Second, "两实例消费组均应 ACK 全部消息", func() bool {
		return xpendingCount(t, rdb, "ws-gateway:gw-1") == 0 &&
			xpendingCount(t, rdb, "ws-gateway:gw-2") == 0
	})

	// 先落盘后确认：客户端尚未回 ack，待投递集合中应保留两份事件（重连补发依据）
	pendingA, _ := store.Pending(ctx, "session-A")
	pendingB, _ := store.Pending(ctx, "session-B")
	if len(pendingA) != 1 || len(pendingB) != 1 {
		t.Fatalf("未 ack 事件应保留在 delivery：A=%d B=%d", len(pendingA), len(pendingB))
	}
}

func TestTerminateEventForcesCloseAfterGrace(t *testing.T) {
	rdb := rtest.Start(t)
	rtest.Flush(t, rdb)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := delivery.NewStore(rdb, time.Hour)

	h := hub.NewHub()
	conn := newMemConn("session-C")
	h.Register(conn)

	cons := NewConsumer(rdb, makeOpts("ws-gateway:term", "term-1"), h, store, zap.NewNop())
	if err := cons.EnsureGroup(ctx); err != nil {
		t.Fatalf("创建消费组失败: %v", err)
	}
	go cons.Run(ctx)

	xadd(t, rdb, eventFixture{EventID: "eTerm", Type: protocol.EventTypeSessionReplaced, UserID: 3, TargetSessionID: "session-C", Reason: "replaced_by_new_login"})

	msg := waitEvent(t, conn, "terminate 事件应推给在线连接")
	if msg.EventID != "eTerm" {
		t.Fatalf("连接收到错误事件: %+v", msg)
	}
	// 宽限期后强制关闭（Close 幂等）；客户端即使忽略事件也不会留下半开连接
	rtest.Eventually(t, 5*time.Second, "terminate 事件宽限期后应强制关闭连接", func() bool {
		select {
		case <-conn.closed:
			return true
		default:
			return false
		}
	})
	if got := conn.closeCode(); got != protocol.CloseSessionInvalid {
		t.Fatalf("应使用 CloseSessionInvalid(%d) 关闭，实际 %d", protocol.CloseSessionInvalid, got)
	}
}

func TestPoisonMessageDoesNotBlockGroup(t *testing.T) {
	rdb := rtest.Start(t)
	rtest.Flush(t, rdb)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := delivery.NewStore(rdb, time.Hour)

	h := hub.NewHub()
	conn := newMemConn("session-D")
	h.Register(conn)

	cons := NewConsumer(rdb, makeOpts("ws-gateway:poison", "poison-1"), h, store, zap.NewNop())
	if err := cons.EnsureGroup(ctx); err != nil {
		t.Fatalf("创建消费组失败: %v", err)
	}
	go cons.Run(ctx)

	// 缺 eventId / targetSessionId 的毒丸消息
	if _, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: protocol.SessionEventStreamKey,
		Values: map[string]interface{}{"type": "session_replaced", "userId": "9"},
	}).Result(); err != nil {
		t.Fatalf("XADD 毒丸消息失败: %v", err)
	}
	// 毒丸之后的正常消息应仍可消费
	xadd(t, rdb, eventFixture{EventID: "eGood", Type: protocol.EventTypeSessionTerminated, UserID: 4, TargetSessionID: "session-D", Reason: "logged_out"})

	msg := waitEvent(t, conn, "毒丸不应阻塞后续正常消息")
	if msg.EventID != "eGood" {
		t.Fatalf("连接收到错误事件: %+v", msg)
	}
	rtest.Eventually(t, 5*time.Second, "毒丸消息应被 ACK 跳过，消费组不卡死", func() bool {
		return xpendingCount(t, rdb, "ws-gateway:poison") == 0
	})
}

// 测试同一消费组内"实例崩溃后遗留的未确认消息（PEL）被重新接管并处理"。
// XAUTOCLAIM 需要 Redis ≥ 6.2；低版本本地实例上跳过（CI 用新版 Redis 跑）。
func TestPELRecoveryAfterCrash(t *testing.T) {
	rdb := rtest.Start(t)
	rtest.Flush(t, rdb)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	info := rdb.Info(ctx, "server").Val()
	if !strings.Contains(info, "redis_version:6.") &&
		!strings.Contains(info, "redis_version:7.") &&
		!strings.Contains(info, "redis_version:8.") {
		t.Skipf("本用例需要 Redis ≥6.2（XAUTOCLAIM），当前 Redis 版本较低，跳过")
	}

	const group = "ws-gateway:recover"
	store := delivery.NewStore(rdb, time.Hour)

	// 崩溃前注册的在线连接（原实例宕机后，客户端还未重连；这里用新实例的 hub）
	h := hub.NewHub()
	conn := newMemConn("session-X")
	h.Register(conn)

	cons := NewConsumer(rdb, makeOpts(group, "alive-1"), h, store, zap.NewNop())
	if err := cons.EnsureGroup(ctx); err != nil {
		t.Fatalf("创建消费组失败: %v", err)
	}

	// 模拟崩溃实例：把 3 条消息拉进 PEL（XREADGROUP 成功但从未 XACK / 处理），然后"宕机"
	xadd(t, rdb, eventFixture{EventID: "crash-1", Type: protocol.EventTypeSessionTerminated, UserID: 7, TargetSessionID: "session-X", Reason: "logged_out"})
	xadd(t, rdb, eventFixture{EventID: "crash-2", Type: protocol.EventTypeSessionTerminated, UserID: 7, TargetSessionID: "session-X", Reason: "logged_out"})
	xadd(t, rdb, eventFixture{EventID: "crash-3", Type: protocol.EventTypeSessionTerminated, UserID: 7, TargetSessionID: "session-X", Reason: "logged_out"})
	if _, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: "dead-1", // 已崩溃的消费者
		Streams:  []string{protocol.SessionEventStreamKey, ">"},
		Count:    8,
	}).Result(); err != nil {
		t.Fatalf("模拟崩溃消费者读取失败: %v", err)
	}
	if got := xpendingCount(t, rdb, group); got != 3 {
		t.Fatalf("崩溃消费者读取后 PEL 应有 3 条未确认消息，实际 %d", got)
	}

	// 让悬挂消息 idle 超过接管阈值（MinIdle）
	time.Sleep(500 * time.Millisecond)

	// "实例重启"：同组新消费者运行，claimLoop 周期性接管悬挂消息
	go cons.Run(ctx)

	// 3 条悬挂消息最终被接管、落盘并 ACK，全部送达在线连接
	var got int
	rtest.Eventually(t, 8*time.Second, "悬挂消息应被接管并投递到连接", func() bool {
		for {
			select {
			case <-conn.send:
				got++
				if got == 3 {
					return true
				}
			default:
				return false
			}
		}
	})
	rtest.Eventually(t, 8*time.Second, "PEL 应被清空（全部 ACK）", func() bool {
		return xpendingCount(t, rdb, group) == 0
	})
	pending, err := store.Pending(ctx, "session-X")
	if err != nil || len(pending) != 3 {
		t.Fatalf("恢复后 delivery 应保留 3 条未确认事件，实际 %d err=%v", len(pending), err)
	}
}
