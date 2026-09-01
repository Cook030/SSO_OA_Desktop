package service

import (
	"time"

	"mh-sso-svc/internal/cache"
	"mh-sso-svc/internal/consts"
	"mh-sso-svc/internal/model"
	"mh-sso-svc/internal/utils"

	"go.uber.org/zap"
)

// 常量（会话/令牌 TTL 派生）
const (
	introspectCacheMaxTTL   = 30 * time.Second // introspect 结果缓存上限
	passwordVersionCacheTTL = 24 * time.Hour   // 用户密码版本缓存
	maxAccountInputLength   = 128              // 登录账号最大输入长度
)

// SessionService 负责会话与 refresh token 的创建、校验、轮换与撤销。
// 核心保证：同一登录/刷新用例内的 Redis 写入要么全部成功、要么整体失败（任一失败返回错误，不产生部分状态）。
type SessionService struct {
	cache      *cache.Cache
	log        *zap.Logger
	recordAudit func(userID *uint64, account, eventType string, success bool, failReason string, meta RequestMeta)
}

// NewSessionService 创建会话服务
func NewSessionService(rdb *cache.Cache, log *zap.Logger) *SessionService {
	return &SessionService{cache: rdb, log: log}
}

// SetAuditRecorder 注入审计写入器（由 AuthService 提供，避免 SessionService 直接依赖审计存储）
func (m *SessionService) SetAuditRecorder(fn func(userID *uint64, account, eventType string, success bool, failReason string, meta RequestMeta)) {
	m.recordAudit = fn
}

// audit 内部封装的审计调用，未注入时静默跳过
func (m *SessionService) audit(userID *uint64, account, eventType string, success bool, failReason string, meta RequestMeta) {
	if m.recordAudit != nil {
		m.recordAudit(userID, account, eventType, success, failReason, meta)
	}
}

// CreateLoginSession 创建一次性会话与 refresh token 记录，返回会话 ID 与 refresh token 明文。
// 两步 Redis 写入任一失败即整体失败，保证登录不产生"有会话无令牌"的残缺状态。
func (m *SessionService) CreateLoginSession(user *model.SysUser, loginIP, userAgent string, sessionTTL, refreshTTL time.Duration, now time.Time) (sessionID, refreshToken string, err error) {
	sessionID = utils.GenerateOpaqueToken("session_")
	refreshToken = utils.GenerateOpaqueToken("rt_")

	session := &cache.SessionRecord{
		SessionID:       sessionID,
		UserID:          user.ID,
		LoginIP:         utils.EmptyToNil(loginIP),
		LoginUserAgent:  utils.EmptyToNil(userAgent),
		Status:          consts.SessionStatusActive,
		PasswordVersion: int(user.PasswordVersion),
		LastActiveAt:    now,
		ExpiredAt:       now.Add(sessionTTL),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err = m.cache.SaveSession(session, sessionTTL); err != nil {
		return "", "", err
	}

	rt := &cache.RefreshTokenRecord{
		TokenHash: utils.SHA256Hex(refreshToken),
		SessionID: sessionID,
		UserID:    user.ID,
		Status:    consts.RefreshTokenStatusActive,
		ExpiredAt: now.Add(refreshTTL),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err = m.cache.SaveRefreshToken(rt, refreshTTL); err != nil {
		return "", "", err
	}
	return sessionID, refreshToken, nil
}

// RotateRefreshToken 轮换 refresh token 并滑动续期会话。
// 顺序：标记旧 token 已轮换 -> 写入新 token -> 续期会话；
// 任一失败立即返回，不产生新 token，保证轮换的原子性。
func (m *SessionService) RotateRefreshToken(oldHash string, userID uint64, sessionID string, sessionTTL, refreshTTL time.Duration, now time.Time) (newRefreshToken string, err error) {
	newRefreshToken = utils.GenerateOpaqueToken("rt_")

	if err = m.cache.UpdateRefreshTokenStatus(oldHash, consts.RefreshTokenStatusRotated); err != nil {
		return "", err
	}

	newRecord := &cache.RefreshTokenRecord{
		TokenHash:   utils.SHA256Hex(newRefreshToken),
		SessionID:   sessionID,
		UserID:      userID,
		Status:      consts.RefreshTokenStatusActive,
		ExpiredAt:   now.Add(refreshTTL),
		RotatedFrom: oldHash,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err = m.cache.SaveRefreshToken(newRecord, refreshTTL); err != nil {
		return "", err
	}

	if err = m.cache.TouchSession(sessionID, now, now.Add(sessionTTL)); err != nil {
		return "", err
	}
	return newRefreshToken, nil
}

// handleRefreshReplay 处理 refresh token 重放：撤销整个 token family 与会话
func (m *SessionService) handleRefreshReplay(rt *cache.RefreshTokenRecord, meta RequestMeta) {
	m.log.Warn("检测到 refresh token 重放，撤销整个会话",
		zap.Uint64("user_id", rt.UserID),
		zap.String("session_id", rt.SessionID),
		zap.String("request_id", meta.RequestID))

	if err := m.cache.RevokeSessionTokens(rt.SessionID, consts.RefreshTokenStatusRevoked); err != nil {
		m.log.Error("撤销重放会话令牌失败", zap.String("session_id", rt.SessionID), zap.Error(err))
	}
	if err := m.cache.UpdateSessionStatus(rt.SessionID, consts.SessionStatusRevoked); err != nil {
		m.log.Error("撤销重放会话失败", zap.Uint64("user_id", rt.UserID), zap.Error(err))
	}

	userID := rt.UserID
	m.audit(&userID, "", consts.AuditEventRefresh, false, "refresh_token_replay", meta)
}

// revokeSession 撤销指定会话及其全部 refresh token（尽力而为）
func (m *SessionService) revokeSession(sessionID string, status int) {
	if err := m.cache.RevokeSessionTokens(sessionID, consts.RefreshTokenStatusRevoked); err != nil {
		m.log.Error("撤销会话令牌失败", zap.String("session_id", sessionID), zap.Error(err))
	}
	if err := m.cache.UpdateSessionStatus(sessionID, status); err != nil {
		m.log.Error("撤销会话失败", zap.String("session_id", sessionID), zap.Error(err))
	}
}

// revokeAllUserSessions 撤销用户全部 active 会话与 refresh token，清理密码版本缓存，并记录审计
func (m *SessionService) revokeAllUserSessions(userID uint64, event string, meta RequestMeta) {
	revoked := 0
	if n, err := m.cache.RevokeUserSessions(userID, consts.SessionStatusRevoked); err != nil {
		m.log.Error("撤销用户会话失败", zap.Uint64("user_id", userID), zap.Error(err))
	} else {
		revoked = n
	}
	if err := m.cache.RevokeUserTokens(userID, consts.RefreshTokenStatusRevoked); err != nil {
		m.log.Error("撤销用户 refresh token 失败", zap.Uint64("user_id", userID), zap.Error(err))
	}

	// 清理密码版本缓存，强制下次回源
	m.cache.DeletePasswordVersion(userID)

	m.audit(&userID, "", event, true, "", meta)
	m.log.Info("已撤销用户全部会话", zap.Uint64("user_id", userID), zap.Int("session_count", revoked))
}
