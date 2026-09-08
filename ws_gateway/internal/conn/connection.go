// Package conn 实现单条 WebSocket 连接的完整生命周期：
// 建立后的鉴权绑定、读协程、写协程、Ping/Pong 心跳、写队列背压与关闭清理。
//
// 每条连接固定两个 goroutine：
//   - 读协程：只负责读客户端帧（ack 与控制帧），退出即触发关闭；
//   - 写协程：只负责写，独占 socket 写锁，并驱动 Ping 与心跳超时判定。
//
// 这样读写互不阻塞，慢客户端只会堵住自己的写队列，不会拖垮其他连接。
package conn

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"mh-ws-gateway/internal/delivery"
	"mh-ws-gateway/internal/hub"
	"mh-ws-gateway/internal/presence"
	"mh-ws-gateway/internal/protocol"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// writeWait 控制帧 / 关闭帧的写入超时
const writeWait = 5 * time.Second

// Options 连接行为参数（由配置派生，便于单测）
type Options struct {
	PingInterval   time.Duration
	PongWait       time.Duration
	WriteQueueSize int
	ReadLimitBytes int64
}

// Params 建立连接所需的全部依赖
type Params struct {
	SessionID  string
	UserID     uint64
	DeviceID   string
	InstanceID string

	WS       *websocket.Conn
	Hub      *hub.Hub
	Delivery *delivery.Store
	Presence *presence.Store

	Opt Options
	Log *zap.Logger
}

// Connection 一条已通过鉴权的 WebSocket 连接
type Connection struct {
	id         string
	sessionID  string
	userID     uint64
	deviceID   string
	instanceID string

	ws       *websocket.Conn
	hub      *hub.Hub
	delivery *delivery.Store
	presence *presence.Store

	opt Options
	log *zap.Logger

	send      chan []byte
	writeMu   sync.Mutex
	closeOnce sync.Once
	closed    chan struct{}
	lastPong  atomic.Int64
}

// New 创建连接对象（尚未开始读写）
func New(p Params) *Connection {
	queue := p.Opt.WriteQueueSize
	if queue <= 0 {
		queue = 128
	}
	c := &Connection{
		id:         uuid.NewString(),
		sessionID:  p.SessionID,
		userID:     p.UserID,
		deviceID:   p.DeviceID,
		instanceID: p.InstanceID,
		ws:         p.WS,
		hub:        p.Hub,
		delivery:   p.Delivery,
		presence:   p.Presence,
		opt:        p.Opt,
		log:        p.Log,
		send:       make(chan []byte, queue),
		closed:     make(chan struct{}),
	}
	c.lastPong.Store(time.Now().UnixNano())
	return c
}

// ID 连接唯一 ID
func (c *Connection) ID() string { return c.id }

// SessionID 连接绑定的会话 ID
func (c *Connection) SessionID() string { return c.sessionID }

// Run 启动读写协程，直到连接关闭后返回（在写协程中退出）
func (c *Connection) Run() {
	go c.readLoop()
	c.writeLoop()
}

// Send 非阻塞投递消息；队列已满返回 protocol.ErrQueueFull，绝不阻塞调用方
func (c *Connection) Send(payload []byte) error {
	select {
	case <-c.closed:
		return errors.New("realtime: 连接已关闭")
	default:
	}
	select {
	case c.send <- payload:
		return nil
	default:
		return protocol.ErrQueueFull
	}
}

// Close 关闭连接并清理注册状态；可重复调用
func (c *Connection) Close(code int, reason string) {
	c.closeOnce.Do(func() {
		close(c.closed)

		c.writeMu.Lock()
		_ = c.ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(code, reason), time.Now().Add(writeWait))
		_ = c.ws.Close()
		c.writeMu.Unlock()

		if c.hub != nil {
			c.hub.Unregister(c.sessionID, c)
		}
		if c.presence != nil {
			_ = c.presence.Release(context.Background(), c.sessionID, c.id)
		}
		c.log.Info("WebSocket 连接已关闭",
			zap.String("session_id", c.sessionID),
			zap.String("connection_id", c.id),
			zap.Int("code", code),
			zap.String("reason", reason))
	})
}

// ---------- 读协程 ----------

func (c *Connection) readLoop() {
	defer c.Close(protocol.CloseNormal, "read loop exit")

	c.ws.SetReadLimit(c.opt.ReadLimitBytes)
	if err := c.ws.SetReadDeadline(time.Now().Add(c.opt.PongWait)); err != nil {
		return
	}
	// 收到 Pong：刷新存活时间并顺延读超时
	c.ws.SetPongHandler(func(string) error {
		c.lastPong.Store(time.Now().UnixNano())
		return c.ws.SetReadDeadline(time.Now().Add(c.opt.PongWait))
	})

	for {
		var msg protocol.ClientMessage
		if err := c.ws.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				c.log.Warn("读取客户端消息异常",
					zap.String("session_id", c.sessionID), zap.Error(err))
			}
			return
		}
		if msg.Type != protocol.ClientMsgTypeAck || msg.EventID == "" {
			continue
		}
		// 客户端确认后才从待投递集合删除；未确认的事件会在重连后重发
		if err := c.delivery.Ack(context.Background(), c.sessionID, msg.EventID); err != nil {
			c.log.Error("确认事件失败",
				zap.String("session_id", c.sessionID),
				zap.String("event_id", msg.EventID),
				zap.Error(err))
		}
	}
}

// ---------- 写协程 ----------

func (c *Connection) writeLoop() {
	ticker := time.NewTicker(c.opt.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.closed:
			return

		case payload, ok := <-c.send:
			if !ok {
				return
			}
			c.writeMu.Lock()
			_ = c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.ws.WriteMessage(websocket.TextMessage, payload)
			c.writeMu.Unlock()
			if err != nil {
				c.log.Warn("写入消息失败",
					zap.String("session_id", c.sessionID), zap.Error(err))
				c.Close(protocol.CloseNormal, "write failed")
				return
			}

		case <-ticker.C:
			// 心跳超时：半开连接（对端已断但 socket 未关闭）在这里被发现
			if time.Since(time.Unix(0, c.lastPong.Load())) > c.opt.PongWait {
				c.Close(protocol.CloseHeartbeatTimeout, "pong timeout")
				return
			}
			c.writeMu.Lock()
			_ = c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.ws.WriteMessage(websocket.PingMessage, nil)
			c.writeMu.Unlock()
			if err != nil {
				c.Close(protocol.CloseNormal, "ping failed")
				return
			}
			// 续期在线状态：TTL 短于心跳间隔的两倍，连接断开后自然消失
			if c.presence != nil {
				_ = c.presence.Heartbeat(context.Background(), presence.Record{
					InstanceID:    c.instanceID,
					ConnectionID:  c.id,
					SessionID:     c.sessionID,
					LastHeartbeat: time.Now().UnixMilli(),
				})
			}
		}
	}
}
