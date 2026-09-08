// Package hub 维护本机连接的注册表：sessionId -> 连接。
//
// Hub 只做"按会话定位连接"这一件事，不关心连接怎么读写、消息从哪来。
// 它依赖的是 Connection 接口而不是具体的 conn.Connection，
// 因此不会与连接实现产生循环依赖。
package hub

import (
	"sync"
)

// Connection Hub 管理的连接抽象（由 conn.Connection 实现）
type Connection interface {
	// SessionID 连接绑定的会话 ID，也是注册表的主键
	SessionID() string
	// Send 非阻塞投递；队列满时返回 protocol.ErrQueueFull
	Send(payload []byte) error
	// Close 以指定关闭码关闭连接，可重复调用
	Close(code int, reason string)
}

// Hub 本机连接注册表。同一 sessionId 只保留最后建立的连接。
type Hub struct {
	mu    sync.RWMutex
	conns map[string]Connection
}

// NewHub 创建连接注册表
func NewHub() *Hub {
	return &Hub{conns: make(map[string]Connection)}
}

// Register 登记连接；若同会话已存在连接，返回被取代的旧连接（由调用方决定如何关闭）
func (h *Hub) Register(c Connection) Connection {
	h.mu.Lock()
	defer h.mu.Unlock()
	previous, ok := h.conns[c.SessionID()]
	h.conns[c.SessionID()] = c
	if !ok {
		return nil
	}
	return previous
}

// Unregister 注销连接；只有仍是当前连接时才删除，避免误删新连接
func (h *Hub) Unregister(sessionID string, c Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cur, ok := h.conns[sessionID]; ok && cur == c {
		delete(h.conns, sessionID)
	}
}

// Get 按会话查找连接
func (h *Hub) Get(sessionID string) (Connection, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	c, ok := h.conns[sessionID]
	return c, ok
}

// Len 当前连接数
func (h *Hub) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns)
}

// CloseAll 关闭全部连接（优雅关闭时使用）
func (h *Hub) CloseAll(code int, reason string) {
	h.mu.Lock()
	conns := make([]Connection, 0, len(h.conns))
	for _, c := range h.conns {
		conns = append(conns, c)
	}
	h.conns = make(map[string]Connection)
	h.mu.Unlock()

	for _, c := range conns {
		c.Close(code, reason)
	}
}
