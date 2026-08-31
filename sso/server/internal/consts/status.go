// Package consts 存放 SSO 业务状态常量（会话/令牌状态、审计事件类型）。
package consts

// 审计事件类型（对应 sys_audit_log.action）
const (
	AuditEventLoginSuccess   = "login_success"
	AuditEventLoginFailed    = "login_failed"
	AuditEventLogout         = "logout"
	AuditEventRefresh        = "refresh"
	AuditEventRevoke         = "revoke"
	AuditEventChangePassword = "change_password"
)

// 会话状态（SessionRecord.Status）
const (
	SessionStatusInvalid   = 0 // 无效
	SessionStatusActive    = 1 // 有效
	SessionStatusLoggedOut = 2 // 已登出
	SessionStatusRevoked   = 3 // 已撤销
	SessionStatusExpired   = 4 // 已过期
)

// refresh token 状态（RefreshTokenRecord.Status）
const (
	RefreshTokenStatusInvalid = 0 // 无效
	RefreshTokenStatusActive  = 1 // 有效
	RefreshTokenStatusRotated = 2 // 已轮换
	RefreshTokenStatusRevoked = 3 // 已撤销
	RefreshTokenStatusExpired = 4 // 已过期
)
