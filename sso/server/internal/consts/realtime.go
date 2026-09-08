package consts

// 本文件定义 SSO 与 WebSocket Gateway 之间的实时事件契约（会话下线通知）。
//
// WebSocket Gateway 侧（ws_gateway/internal/protocol/contract.go）保存同一套常量。
// 两侧是独立部署的服务，不共享 Go 包，因此常量以副本形式存在；修改任一处必须同步另一处。

// Redis Key / Stream 契约
const (
	// SessionEventStreamKey 会话领域事件流：SSO 在 Lua 中 XADD，Gateway 以消费组消费
	SessionEventStreamKey = "sso:stream:session-events"

	// DeliveryKeyPrefix 待确认消息存储：sso:delivery:{sessionId} -> Hash(eventId -> payload JSON)
	DeliveryKeyPrefix = "sso:delivery:"

	// ConnectionKeyPrefix 连接在线状态（路由辅助数据，非安全状态）：ws:connections:{sessionId}
	ConnectionKeyPrefix = "ws:connections:"

	// StreamMaxLen 事件流近似裁剪长度，避免无限增长
	StreamMaxLen = 100000
)

// 事件类型与原因（写入 Stream 的 type / reason 字段，同时透传给浏览器）
const (
	EventTypeSessionReplaced   = "session_replaced"   // 被新登录置换
	EventTypeSessionTerminated = "session_terminated" // 被登出/改密/管理员撤销
)

const (
	EventReasonReplacedByNewLogin = "replaced_by_new_login"
	EventReasonPasswordChanged    = "password_changed"
	EventReasonRevokedByAdmin     = "revoked_by_admin"
	EventReasonLoggedOut          = "logged_out"
	EventReasonRefreshReplay      = "refresh_token_replay"
)

// 对外错误原因码：写入 HTTP 响应体的 reason 字段，供前端区分处理
const (
	ReasonSessionReplaced        = "SESSION_REPLACED"        // 会话已被新登录置换
	ReasonSessionRevoked         = "SESSION_REVOKED"         // 会话已被撤销
	ReasonSessionLoggedOut       = "SESSION_LOGGED_OUT"      // 会话已登出
	ReasonSessionExpired         = "SESSION_EXPIRED"         // 会话已过期
	ReasonSessionNotFound        = "SESSION_NOT_FOUND"       // 会话不存在
	ReasonSessionActiveElsewhere = "SESSION_ACTIVE_ELSEWHERE" // 已有有效会话（reject 模式拒绝新登录）
	ReasonTokenExpired           = "TOKEN_EXPIRED"
	ReasonTokenInvalid           = "TOKEN_INVALID"
	ReasonPasswordChanged        = "PASSWORD_CHANGED"
)

// LoginMode 重复登录处理策略
type LoginMode string

const (
	// LoginModeReplace 踢出旧会话：新登录成功，旧会话立即失效并被事件通知下线（单设备登录默认策略）
	LoginModeReplace LoginMode = "replace"
	// LoginModeReject 拒绝新登录：已存在有效会话时，新登录请求直接失败
	LoginModeReject LoginMode = "reject"
)

// ParseLoginMode 解析 login_mode 配置；空值或非法值回退为 replace
func ParseLoginMode(s string) LoginMode {
	switch LoginMode(s) {
	case LoginModeReject:
		return LoginModeReject
	default:
		return LoginModeReplace
	}
}
