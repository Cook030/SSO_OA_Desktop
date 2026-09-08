// Package protocol 定义 WebSocket Gateway 与浏览器、以及与 SSO 之间的通信契约。
//
// SSO 侧（sso/server/internal/consts/realtime.go）保存同一套 Redis 常量。
// Gateway 与 SSO 是独立部署的服务，不共享 Go 包，因此常量以副本形式存在，
// 修改时必须同步两侧。
package protocol

import "errors"

// Redis Key / Stream 契约（与 SSO 一致）
const (
	// SessionEventStreamKey 会话事件流：SSO 在 Lua 中 XADD，Gateway 以消费组消费
	SessionEventStreamKey = "sso:stream:session-events"

	// DeliveryKeyPrefix 待确认消息：sso:delivery:{sessionId} -> Hash(eventId -> payload JSON)
	DeliveryKeyPrefix = "sso:delivery:"

	// ConnectionKeyPrefix 连接在线状态：ws:connections:{sessionId}
	// 仅作路由辅助数据（判断同一会话是否已在别的 Gateway 实例上建连），不是安全状态
	ConnectionKeyPrefix = "ws:connections:"
)

// 事件类型（Stream 的 type 字段，原样透传给浏览器）
const (
	EventTypeSessionReplaced   = "session_replaced"
	EventTypeSessionTerminated = "session_terminated"
)

// 消息类型：服务端 → 客户端 / 客户端 → 服务端
const (
	ServerMsgTypeEvent = "event"
	ClientMsgTypeAck   = "ack"
)

// WebSocket 应用级关闭码（4000-4999 为 RFC 6455 保留给应用的区间）。
// 客户端据此决定"能不能重连"：4xxx 中只有 CloseServerShutdown 允许立即重连。
const (
	CloseNormal           = 1000 // 正常关闭
	CloseSessionInvalid   = 4001 // 会话无效 / 已被顶下线：客户端不得用旧凭证重连
	CloseDuplicateSession = 4002 // 同一会话建立了新连接，旧连接被取代
	CloseDeviceMismatch   = 4003 // 设备标识与会话绑定的设备不一致
	CloseServerShutdown   = 4004 // 网关优雅关闭，客户端应重连
	CloseSlowConsumer     = 4005 // 写队列积压，关键事件无法可靠投递
	CloseHeartbeatTimeout = 4006 // 心跳超时
)

// ErrQueueFull 连接写队列已满。
// 普通通知可以丢弃或合并，但 session_replaced 这类关键事件不能静默丢弃，
// 调用方收到本错误时应关闭该连接，避免关键事件被静默丢弃。
var ErrQueueFull = errors.New("realtime: 写队列已满")
