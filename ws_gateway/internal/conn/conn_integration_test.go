package conn

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mh-ws-gateway/internal/delivery"
	"mh-ws-gateway/internal/hub"
	"mh-ws-gateway/internal/presence"
	"mh-ws-gateway/internal/protocol"
	"mh-ws-gateway/internal/rtest"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// 本文件验证单条 WebSocket 连接在真实 Redis 上的闭环语义：
//   - 服务端推送事件 → 客户端回 ack → 待投递集合删除（"客户端确认才算送达"）；
//   - 心跳保活 / 静默客户端被断开；
//   - 连接关闭时清理 hub 注册与 Redis 在线状态。

// wsPair 建立一对内存 WebSocket 连接（client ↔ server）
func wsPair(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()
	up := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	serverCh := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := up.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("服务端升级失败: %v", err)
			return
		}
		serverCh <- c
	}))
	t.Cleanup(srv.Close)

	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatalf("客户端拨号失败: %v", err)
	}
	server := <-serverCh
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})
	return client, server
}

func testConnParams(sessionID string, server *websocket.Conn) Params {
	return Params{
		SessionID:  sessionID,
		UserID:     1,
		InstanceID: "gw-test",
		WS:         server,
		Opt: Options{
			PingInterval:   30 * time.Millisecond,
			PongWait:       2 * time.Second,
			WriteQueueSize: 16,
			ReadLimitBytes: 4096,
		},
		Log: zap.NewNop(),
	}
}

// 启动客户端读循环：处理控制帧（自动回 Pong），并把业务消息回传到 channel
func startClientLoop(t *testing.T, client *websocket.Conn) chan []byte {
	t.Helper()
	client.SetPingHandler(func(appData string) error {
		return client.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
	})
	got := make(chan []byte, 16)
	go func() {
		for {
			_, data, err := client.ReadMessage()
			if err != nil {
				return
			}
			got <- data
		}
	}()
	return got
}

func isClosed(c *Connection) bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func TestAckDrivesAtLeastOnceDelivery(t *testing.T) {
	rdb := rtest.Start(t)
	rtest.Flush(t, rdb)
	ctx := context.Background()
	sessionID := "sess-ack"

	store := delivery.NewStore(rdb, time.Hour)
	ps := presence.NewStore(rdb, time.Minute)
	evPayload, err := protocol.MarshalEvent("ev-1", protocol.EventTypeSessionTerminated, "logged_out")
	if err != nil {
		t.Fatalf("构造事件失败: %v", err)
	}
	// 事件已落盘但尚未投递（模拟消费端落盘后、连接建立前的窗口）
	if err := store.Save(ctx, sessionID, "ev-1", evPayload); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}

	client, server := wsPair(t)
	h := hub.NewHub()
	c := New(testConnParams(sessionID, server))
	c.hub = h
	c.delivery = store
	c.presence = ps
	h.Register(c)
	if _, err := ps.Register(ctx, presence.Record{
		InstanceID: "gw-test", ConnectionID: c.ID(), SessionID: sessionID,
	}); err != nil {
		t.Fatalf("presence.Register 失败: %v", err)
	}

	go c.Run()
	got := startClientLoop(t, client)

	// 服务端把已落盘事件推给客户端
	if err := c.Send(evPayload); err != nil {
		t.Fatalf("Send 失败: %v", err)
	}
	select {
	case data := <-got:
		if !bytes.Equal(data, evPayload) {
			t.Fatalf("客户端收到的事件与推送不一致: %s", data)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("客户端等待事件超时")
	}

	// 客户端回 ack：待投递集合才真正删除
	if err := client.WriteJSON(protocol.Ack("ev-1")); err != nil {
		t.Fatalf("发送 ack 失败: %v", err)
	}
	rtest.Eventually(t, 5*time.Second, "ack 后待投递 key 应被删除", func() bool {
		n, _ := rdb.Exists(ctx, protocol.DeliveryKeyPrefix+sessionID).Result()
		return n == 0
	})

	// 连接关闭：释放在线状态并注销 hub
	c.Close(protocol.CloseNormal, "done")
	rtest.Eventually(t, 5*time.Second, "连接关闭应清理 presence 与 hub", func() bool {
		n, _ := rdb.Exists(ctx, protocol.ConnectionKeyPrefix+sessionID).Result()
		_, ok := h.Get(sessionID)
		return n == 0 && !ok
	})
}

func TestSilentClientEvictedByHeartbeatTimeout(t *testing.T) {
	rdb := rtest.Start(t)
	rtest.Flush(t, rdb)

	client, server := wsPair(t)
	c := New(testConnParams("sess-silent", server))
	// 缩短超时：PongWait 200ms，静默客户端应在约 200ms 后被断开
	c.opt.PongWait = 200 * time.Millisecond
	c.opt.PingInterval = 30 * time.Millisecond
	go c.Run()

	// 客户端不回任何消息（不处理 Ping、不读、不写）
	_ = client

	rtest.Eventually(t, 3*time.Second, "静默客户端应被心跳超时断开", func() bool {
		return isClosed(c)
	})
}

func TestHeartbeatKeepsAliveAndRefreshesPresence(t *testing.T) {
	rdb := rtest.Start(t)
	rtest.Flush(t, rdb)
	ctx := context.Background()
	sessionID := "sess-alive"

	ps := presence.NewStore(rdb, 500*time.Millisecond) // 短 TTL：若心跳失效，key 在 500ms 后消失
	client, server := wsPair(t)
	c := New(testConnParams(sessionID, server))
	c.presence = ps
	c.opt.PongWait = 300 * time.Millisecond
	c.opt.PingInterval = 30 * time.Millisecond
	if _, err := ps.Register(ctx, presence.Record{
		InstanceID: "gw-test", ConnectionID: c.ID(), SessionID: sessionID,
	}); err != nil {
		t.Fatalf("presence.Register 失败: %v", err)
	}

	go c.Run()
	startClientLoop(t, client)

	// 客户端持续回 Pong：超过 PongWait 与在线状态 TTL，连接仍存活、路由仍被心跳续期
	time.Sleep(800 * time.Millisecond)
	if isClosed(c) {
		t.Fatalf("正常回 Pong 的连接不应被心跳超时断开")
	}
	n, err := rdb.Exists(ctx, protocol.ConnectionKeyPrefix+sessionID).Result()
	if err != nil || n != 1 {
		t.Fatalf("在线状态应被心跳续期保留，n=%d err=%v", n, err)
	}
	// 心跳刚续期后剩余 TTL 应接近完整 TTL（用毫秒精度 PTTL 断言）
	pttl, err := rdb.PTTL(ctx, protocol.ConnectionKeyPrefix+sessionID).Result()
	if err != nil || pttl < 200*time.Millisecond {
		t.Fatalf("在线状态 TTL 应被心跳持续刷新（剩余应接近 500ms），实际 %v err=%v", pttl, err)
	}

	c.Close(protocol.CloseServerShutdown, "test done")
	rtest.Eventually(t, 3*time.Second, "主动关闭后连接应退出", func() bool {
		return isClosed(c)
	})
}
