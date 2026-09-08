package service

import (
	"context"
	"fmt"
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
	cache       *cache.Cache
	log         *zap.Logger
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

// EstablishSessionInput 建立登录会话的入参
type EstablishSessionInput struct {
	User *model.SysUser
	// SessionID 由调用方预先生成，先签发 token 后才提交会话置换。
	SessionID  string
	DeviceID   string // 已规范化的设备唯一标识
	DeviceType string
	LoginIP    string
	UserAgent  string
	SessionTTL time.Duration
	RefreshTTL time.Duration
	Now        time.Time
	// Mode 重复登录策略：replace 踢出旧会话，reject 拒绝本次登录
	Mode consts.LoginMode
}

// EstablishSessionResult 建立登录会话的结果
type EstablishSessionResult struct {
	SessionID          string
	RefreshToken       string
	ReplacedSessionIDs []string // 本次被顶下线的旧会话
	Rejected           bool     // reject 模式下已存在有效会话，本次登录被拒绝
}

// EstablishLoginSession 单设备登录的会话建立入口。
//
// 与旧版 CreateLoginSession 的区别：不再"先建新会话再撤销旧会话"，
// 而是把"撤销旧会话 + 撤销旧 refresh token + 写入新会话 + 写入新 token +
// 重置用户会话索引 + 写 current_session + 为每个旧会话写下线事件"
// 收敛为一次 Redis Lua 调用（cache.ReplaceLoginSession）。
//
// 这样旧客户端即使收不到任何 WebSocket 消息，其 access token 也会在同一次
// 原子操作后立刻失效；也不会出现"状态变了但事件没写"的半成功状态。
func (m *SessionService) EstablishLoginSession(ctx context.Context, in EstablishSessionInput) (*EstablishSessionResult, error) {
	if in.SessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}
	sessionID := in.SessionID
	refreshToken := utils.GenerateOpaqueToken("rt_")

	session := &cache.SessionRecord{
		SessionID:       sessionID,
		UserID:          in.User.ID,
		DeviceID:        utils.EmptyToNil(in.DeviceID),
		DeviceType:      utils.EmptyToNil(in.DeviceType),
		LoginIP:         utils.EmptyToNil(in.LoginIP),
		LoginUserAgent:  utils.EmptyToNil(in.UserAgent),
		Status:          consts.SessionStatusActive,
		PasswordVersion: int(in.User.PasswordVersion),
		LastActiveAt:    in.Now,
		ExpiredAt:       in.Now.Add(in.SessionTTL),
		CreatedAt:       in.Now,
		UpdatedAt:       in.Now,
	}
	rt := &cache.RefreshTokenRecord{
		TokenHash: utils.SHA256Hex(refreshToken),
		SessionID: sessionID,
		UserID:    in.User.ID,
		Status:    consts.RefreshTokenStatusActive,
		ExpiredAt: in.Now.Add(in.RefreshTTL),
		CreatedAt: in.Now,
		UpdatedAt: in.Now,
	}

	res, err := m.cache.ReplaceLoginSession(ctx, cache.LoginReplaceInput{
		UserID:       in.User.ID,
		Mode:         in.Mode,
		Session:      session,
		SessionTTL:   in.SessionTTL,
		RefreshToken: rt,
		RefreshTTL:   in.RefreshTTL,
		Event: cache.SessionEventInput{
			EventType: consts.EventTypeSessionReplaced,
			Reason:    consts.EventReasonReplacedByNewLogin,
		},
	})
	if err != nil {
		return nil, err
	}
	if res.Rejected {
		return &EstablishSessionResult{Rejected: true}, nil
	}

	return &EstablishSessionResult{
		SessionID:          res.SessionID,
		RefreshToken:       refreshToken,
		ReplacedSessionIDs: res.ReplacedSessionIDs,
	}, nil
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

// handleRefreshReplay 处理 refresh token 重放：原子撤销会话并写终止事件。
func (m *SessionService) handleRefreshReplay(rt *cache.RefreshTokenRecord, meta RequestMeta) {
	m.log.Warn("检测到 refresh token 重放，撤销整个会话",
		zap.Uint64("user_id", rt.UserID),
		zap.String("session_id", rt.SessionID),
		zap.String("request_id", meta.RequestID))

	if err := m.revokeSession(context.Background(), rt.UserID, rt.SessionID, consts.SessionStatusRevoked, cache.SessionEventInput{
		EventType: consts.EventTypeSessionTerminated,
		Reason:    consts.EventReasonRefreshReplay,
	}); err != nil {
		m.log.Error("撤销重放会话失败", zap.Uint64("user_id", rt.UserID), zap.Error(err))
	}

	userID := rt.UserID
	m.audit(&userID, "", consts.AuditEventRefresh, false, "refresh_token_replay", meta)
}

// revokeSession 原子撤销指定会话并写事件。旧 session 的请求不会删除新 session 的
// current_session；这一点不能由按 userId 全量撤销替代。
func (m *SessionService) revokeSession(ctx context.Context, userID uint64, sessionID string, status int, ev cache.SessionEventInput) error {
	_, err := m.cache.RevokeSessionWithEvent(ctx, userID, sessionID, status, ev)
	return err
}

// RevokeAllUserSessions 撤销用户全部 active 会话与 refresh token，并向事件流广播下线事件。
//
// 与登录置换一样，状态变更与事件写入在同一次 Lua 中完成：
// Gateway 消费事件后即可通知在线客户端；离线客户端则因 current_session 被清空
// 而在下一次请求时拿到 401。
func (m *SessionService) RevokeAllUserSessions(ctx context.Context, userID uint64, ev cache.SessionEventInput, auditEvent string, meta RequestMeta) ([]string, error) {
	revoked, err := m.cache.RevokeUserSessionsWithEvents(ctx, userID, consts.SessionStatusRevoked, ev)
	if err != nil {
		m.log.Error("撤销用户会话失败", zap.Uint64("user_id", userID), zap.Error(err))
		return nil, err
	}

	// 清理密码版本缓存，强制下次回源
	m.cache.DeletePasswordVersion(userID)

	m.audit(&userID, "", auditEvent, true, "", meta)
	m.log.Info("已撤销用户全部会话并广播下线事件",
		zap.Uint64("user_id", userID),
		zap.String("reason", ev.Reason),
		zap.Int("session_count", len(revoked)))
	return revoked, nil
}
