package service

import (
	"errors"
	"time"

	"mh-sso-svc/internal/cache"
	"mh-sso-svc/internal/db_model"
	"mh-sso-svc/internal/db_model/query"
	"mh-sso-svc/internal/model"
	"mh-sso-svc/internal/repository"
	"mh-sso-svc/internal/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 业务状态码（与接口文档保持一致）
const (
	CodeOK           = 200
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeConflict     = 409
	CodeServerError  = 500
)

const (
	introspectCacheMaxTTL   = 30 * time.Second // introspect 结果缓存上限
	passwordVersionCacheTTL = 24 * time.Hour   // 用户密码版本缓存
	maxAccountInputLength   = 128              // 登录账号最大输入长度
)

// dummyHash 用于账号不存在时的恒定耗时密码比对，防止时序探测账号存在性
const dummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// BizError 业务错误：handler 据此返回对应业务码与提示
type BizError struct {
	Code int
	Msg  string
}

func (e *BizError) Error() string { return e.Msg }

// NewBizError 创建业务错误
func NewBizError(code int, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

// RequestMeta 请求元信息（审计与限流用）
type RequestMeta struct {
	IP        string
	UserAgent string
	RequestID string
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
	Email           string `json:"email" binding:"omitempty,email"`
	Mobile          string `json:"mobile" binding:"omitempty,min=5,max=32"`
	Nickname        string `json:"nickname"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

// RevokeUserSessionsRequest 撤销指定用户全部会话请求
type RevokeUserSessionsRequest struct {
	UserID uint64 `json:"userId"`
}

// UserInfo 用户展示信息（login / me 响应）
type UserInfo struct {
	ID              uint64    `json:"id"`
	Account         string    `json:"account"`
	Name            string    `json:"name"`
	Phone           string    `json:"phone"`
	Email           string    `json:"email"`
	Role            string    `json:"role"`
	Department      string    `json:"department"`
	PasswordChanged int       `json:"passwordChanged"` // 1 表示修改过密码（password_version > 1）
	CreateTime      time.Time `json:"createTime"`
	UpdateTime      time.Time `json:"updateTime"`
}

// PermItem 用户组/角色展示项
type PermItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// RegisterResult 注册响应
type RegisterResult struct {
	UserID   uint64 `json:"userId"`
	Username string `json:"username"`
	Status   string `json:"status"`
}

// LoginResult 登录响应
type LoginResult struct {
	AccessToken      string   `json:"accessToken"`
	RefreshToken     string   `json:"refreshToken"`
	TokenType        string   `json:"tokenType"`
	ExpiresIn        int      `json:"expiresIn"`
	RefreshExpiresIn int      `json:"refreshExpiresIn"`
	User             UserInfo `json:"user"`
}

// RefreshResult 刷新 token 响应
type RefreshResult struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
}

// MeResult 当前用户信息响应（MVP：基础角色 + 空权限数组）
type MeResult struct {
	User    UserInfo   `json:"user"`
	Groups  []PermItem `json:"groups"`
	Roles   []PermItem `json:"roles"`
	Apps    []string   `json:"apps"`
	Pages   []string   `json:"pages"`
	Apis    []string   `json:"apis"`
	Menus   []string   `json:"menus"`
	Buttons []string   `json:"buttons"`
}

// IntrospectResult token 校验响应
type IntrospectResult struct {
	UserID          uint64 `json:"userId"`
	SessionID       string `json:"sessionId"`
	PasswordVersion int    `json:"passwordVersion"`
	Valid           bool   `json:"valid"`
}

// AuthService SSO 认证核心服务
type AuthService struct {
	db          *gorm.DB
	q           *query.Query
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	rtRepo      *repository.RefreshTokenRepository
	auditRepo   *repository.AuditRepository
	cache       *cache.Cache
	tokenSvc    *TokenService
	cfg         *utils.AuthConfig
	log         *zap.Logger
}

// NewAuthService 创建认证服务（Repository 层在内部组装，事务场景可复用 db）
func NewAuthService(db *gorm.DB, rdb *cache.Cache, tokenSvc *TokenService, cfg *utils.AuthConfig) *AuthService {
	q := query.Use(db)
	return &AuthService{
		db:          db,
		q:           q,
		userRepo:    repository.NewUserRepository(q),
		sessionRepo: repository.NewSessionRepository(db),
		rtRepo:      repository.NewRefreshTokenRepository(db),
		auditRepo:   repository.NewAuditRepository(q),
		cache:       rdb,
		tokenSvc:    tokenSvc,
		cfg:         cfg,
		log:         utils.GetLogger(),
	}
}

func (s *AuthService) refreshTTL() time.Duration {
	return time.Duration(s.cfg.RefreshTokenTTLSecond) * time.Second
}

func (s *AuthService) sessionTTL() time.Duration {
	return time.Duration(s.cfg.SessionTTLSecond) * time.Second
}

// ---------- 注册 ----------

// Register 注册新用户（注册后不自动登录）
func (s *AuthService) Register(req RegisterRequest) (*RegisterResult, error) {
	if req.Password != req.ConfirmPassword {
		return nil, NewBizError(CodeBadRequest, "两次输入的密码不一致")
	}

	// 唯一性预检查（username 必查；email/mobile 非空时查询）
	type uniqueCheck struct {
		field  string
		value  string
		exists func(string) (bool, error)
	}
	for _, check := range []uniqueCheck{
		{"username", req.Username, s.userRepo.ExistsByAccount},
		{"email", req.Email, s.userRepo.ExistsByEmail},
		{"mobile", req.Mobile, s.userRepo.ExistsByPhone},
	} {
		if check.value == "" {
			continue
		}
		exists, err := check.exists(check.value)
		if err != nil {
			s.log.Error("查询用户唯一性失败", zap.String("field", check.field), zap.Error(err))
			return nil, NewBizError(CodeServerError, "服务器内部错误")
		}
		if exists {
			return nil, NewBizError(CodeConflict, "用户已存在")
		}
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		s.log.Error("密码加密失败", zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}

	user := &db_model.SysUser{
		Account:         req.Username,
		Password:        passwordHash,
		Name:            req.Nickname,
		Email:           emptyToNil(req.Email),
		Phone:           emptyToNil(req.Mobile),
		PasswordVersion: 1,
	}
	if user.Name == "" {
		user.Name = req.Username
	}
	if err := s.userRepo.Create(user); err != nil {
		// 唯一索引兜底：并发注册同一账号时返回 409
		if repository.IsDuplicateEntryError(err) {
			return nil, NewBizError(CodeConflict, "用户已存在")
		}
		s.log.Error("创建用户失败", zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}

	s.log.Info("用户注册成功", zap.Uint64("user_id", user.ID), zap.String("username", user.Account))
	return &RegisterResult{
		UserID:   user.ID,
		Username: user.Account,
		Status:   "active",
	}, nil
}

// ---------- 登录 ----------

// Login 账号密码登录：校验用户与密码、创建会话、签发双 token、写缓存、记审计
func (s *AuthService) Login(account, password string, meta RequestMeta) (*LoginResult, error) {
	account = truncate(account, maxAccountInputLength)

	// 1. 登录失败限流（账号 / IP）
	if s.cache.IsLoginRateLimited(account, meta.IP) {
		s.recordAudit(nil, account, model.AuditEventLoginFailed, false, "rate_limited", meta)
		return nil, NewBizError(CodeUnauthorized, "登录失败次数过多，请5分钟后再试")
	}

	// 2. 查询用户（account / email / phone）
	user, err := s.userRepo.FindByAccount(account)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.log.Error("查询用户失败", zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}

	// 3. 校验密码（账号不存在时做恒定耗时比对，防时序探测）
	if user == nil {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
		return s.loginFailed(account, "account_or_password_wrong", meta)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return s.loginFailed(account, "account_or_password_wrong", meta)
	}

	now := time.Now()

	// 4. 创建会话
	sessionID := utils.GenerateOpaqueToken("session_")
	session := &model.SsoSession{
		SessionID:      sessionID,
		UserID:         user.ID,
		LoginIP:        emptyToNil(meta.IP),
		LoginUserAgent: emptyToNil(meta.UserAgent),
		Status:         model.SessionStatusActive,
		LastActiveAt:   now,
		ExpiredAt:      now.Add(s.sessionTTL()),
	}
	if err := s.sessionRepo.Create(session); err != nil {
		s.log.Error("创建会话失败", zap.Uint64("user_id", user.ID), zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}

	// 5. 创建 refresh token（数据库只存 SHA-256 哈希）
	refreshToken := utils.GenerateOpaqueToken("rt_")
	refreshTokenRecord := &model.SsoRefreshToken{
		SessionID: sessionID,
		UserID:    user.ID,
		TokenHash: utils.SHA256Hex(refreshToken),
		Status:    model.RefreshTokenStatusActive,
		ExpiredAt: now.Add(s.refreshTTL()),
	}
	if err := s.rtRepo.Create(refreshTokenRecord); err != nil {
		s.log.Error("创建 refresh token 失败", zap.Uint64("user_id", user.ID), zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}

	// 6. 签发 access token
	accessToken, _, _, err := s.tokenSvc.GenerateAccessToken(user.ID, sessionID, user.Account, int(user.PasswordVersion), now)
	if err != nil {
		s.log.Error("签发 access token 失败", zap.Uint64("user_id", user.ID), zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}

	// 7. 写 session 缓存
	s.cache.SetSessionCache(sessionID, cache.SessionCacheData{
		UserID:          user.ID,
		Status:          model.SessionStatusActive,
		PasswordVersion: int(user.PasswordVersion),
	}, s.sessionTTL())

	// 8. 登录成功：清理失败计数 + 记审计
	s.cache.ClearLoginFailures(account)
	s.recordAudit(&user.ID, user.Account, model.AuditEventLoginSuccess, true, "", meta)

	s.log.Info("登录成功",
		zap.Uint64("user_id", user.ID),
		zap.String("session_id", sessionID),
		zap.String("ip", meta.IP))

	return &LoginResult{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        s.cfg.AccessTokenTTLSecond,
		RefreshExpiresIn: s.cfg.RefreshTokenTTLSecond,
		User:             buildUserInfo(user),
	}, nil
}

// loginFailed 登录失败统一处理：累计限流计数 + 记审计
func (s *AuthService) loginFailed(account, reason string, meta RequestMeta) (*LoginResult, error) {
	s.cache.RecordLoginFailure(account, meta.IP)
	s.recordAudit(nil, account, model.AuditEventLoginFailed, false, reason, meta)
	return nil, NewBizError(CodeUnauthorized, "账号或密码错误")
}

// ---------- 刷新 Token ----------

// Refresh 使用 refresh token 换取新 token
func (s *AuthService) Refresh(refreshToken string, meta RequestMeta) (*RefreshResult, error) {
	if refreshToken == "" {
		return nil, NewBizError(CodeUnauthorized, "missing refresh token")
	}

	tokenHash := utils.SHA256Hex(refreshToken)
	now := time.Now()

	// 1. 查询令牌
	rt, err := s.rtRepo.FindByTokenHash(tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewBizError(CodeUnauthorized, "invalid refresh token")
		}
		s.log.Error("查询 refresh token 失败", zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}

	// 2. 重放检测：已轮换 token 再次使用 -> 撤销整个 family 与会话
	if rt.Status == model.RefreshTokenStatusRotated {
		s.handleRefreshReplay(rt, meta)
		return nil, NewBizError(CodeUnauthorized, "refresh token 已被使用，会话已撤销")
	}
	if rt.Status != model.RefreshTokenStatusActive {
		return nil, NewBizError(CodeUnauthorized, "invalid refresh token")
	}
	if !rt.ExpiredAt.After(now) {
		_ = s.rtRepo.UpdateStatus(rt.ID, model.RefreshTokenStatusExpired)
		return nil, NewBizError(CodeUnauthorized, "refresh token 已过期")
	}

	// 3. 校验会话
	session, err := s.sessionRepo.FindBySessionID(rt.SessionID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error("查询会话失败", zap.Error(err))
			return nil, NewBizError(CodeServerError, "服务器内部错误")
		}
		return nil, NewBizError(CodeUnauthorized, "session invalid")
	}
	if !session.IsActive(now) {
		return nil, NewBizError(CodeUnauthorized, "session invalid")
	}

	// 4. 校验用户
	user, err := s.userRepo.FindByID(rt.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewBizError(CodeUnauthorized, "user not found")
		}
		s.log.Error("查询用户失败", zap.Uint64("user_id", rt.UserID), zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}

	// 5. 并发锁：防止同一 refresh token 并发刷新
	if !s.cache.AcquireRefreshLock(tokenHash, meta.RequestID) {
		return nil, NewBizError(CodeUnauthorized, "刷新请求处理中，请稍后重试")
	}
	defer s.cache.ReleaseRefreshLock(tokenHash, meta.RequestID)

	// 6. 二次确认状态（拿到锁后 token 可能已被并发请求旋转）
	rtLatest, err := s.rtRepo.FindByTokenHash(tokenHash)
	if err != nil {
		s.log.Error("二次查询 refresh token 失败", zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}
	if rtLatest.Status == model.RefreshTokenStatusRotated {
		s.handleRefreshReplay(rtLatest, meta)
		return nil, NewBizError(CodeUnauthorized, "refresh token 已被使用，会话已撤销")
	}
	if rtLatest.Status != model.RefreshTokenStatusActive {
		return nil, NewBizError(CodeUnauthorized, "invalid refresh token")
	}

	// 7. 旋转令牌并续期会话（事务）
	newRefreshToken := utils.GenerateOpaqueToken("rt_")
	refreshExpiredAt := now.Add(s.refreshTTL())
	sessionExpiredAt := now.Add(s.sessionTTL())
	err = s.db.Transaction(func(tx *gorm.DB) error {
		rtTx := repository.NewRefreshTokenRepository(tx)
		sessionTx := repository.NewSessionRepository(tx)

		// 旧 token 标记已轮换（updated_at 即 used_at）
		if err := rtTx.UpdateStatus(rt.ID, model.RefreshTokenStatusRotated); err != nil {
			return err
		}
		// 新 token：family（session_id）与旧 token 一致
		newRecord := &model.SsoRefreshToken{
			SessionID:   rt.SessionID,
			UserID:      rt.UserID,
			TokenHash:   utils.SHA256Hex(newRefreshToken),
			Status:      model.RefreshTokenStatusActive,
			ExpiredAt:   refreshExpiredAt,
			RotatedFrom: &rt.ID,
		}
		if err := rtTx.Create(newRecord); err != nil {
			return err
		}
		// 会话滑动续期
		return sessionTx.Touch(rt.SessionID, now, sessionExpiredAt)
	})
	if err != nil {
		s.log.Error("refresh token 轮换失败", zap.Uint64("user_id", rt.UserID), zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}

	// 8. 签发新 access token
	accessToken, _, _, err := s.tokenSvc.GenerateAccessToken(user.ID, rt.SessionID, user.Account, int(user.PasswordVersion), now)
	if err != nil {
		s.log.Error("签发 access token 失败", zap.Uint64("user_id", user.ID), zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}

	// 9. 更新 session 缓存
	s.cache.SetSessionCache(rt.SessionID, cache.SessionCacheData{
		UserID:          user.ID,
		Status:          model.SessionStatusActive,
		PasswordVersion: int(user.PasswordVersion),
	}, s.sessionTTL())

	s.recordAudit(&user.ID, user.Account, model.AuditEventRefresh, true, "", meta)
	s.log.Info("refresh token 轮换成功",
		zap.Uint64("user_id", user.ID),
		zap.String("session_id", rt.SessionID))

	return &RefreshResult{
		AccessToken:      accessToken,
		RefreshToken:     newRefreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        s.cfg.AccessTokenTTLSecond,
		RefreshExpiresIn: s.cfg.RefreshTokenTTLSecond,
	}, nil
}

// handleRefreshReplay 处理 refresh token 重放
func (s *AuthService) handleRefreshReplay(rt *model.SsoRefreshToken, meta RequestMeta) {
	s.log.Warn("检测到 refresh token 重放，撤销整个会话",
		zap.Uint64("user_id", rt.UserID),
		zap.String("session_id", rt.SessionID),
		zap.String("request_id", meta.RequestID))

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := repository.NewRefreshTokenRepository(tx).RevokeBySessionID(rt.SessionID, model.RefreshTokenStatusRevoked); err != nil {
			return err
		}
		return repository.NewSessionRepository(tx).UpdateStatusBySessionID(rt.SessionID, model.SessionStatusRevoked)
	})
	if err != nil {
		s.log.Error("撤销重放会话失败", zap.Uint64("user_id", rt.UserID), zap.Error(err))
	}

	s.cache.DeleteSessionCache(rt.SessionID)

	userID := rt.UserID
	s.recordAudit(&userID, "", model.AuditEventRefresh, false, "refresh_token_replay", meta)
}

// ---------- 退出登录 ----------

// Logout 尽力而为地撤销当前会话与 refresh token，并清理缓存
func (s *AuthService) Logout(accessToken, refreshToken string, meta RequestMeta) {
	var (
		userID  *uint64
		account string
	)

	// 1. access token 可解析 -> 按 sid 撤销当前会话及其全部 refresh token
	if accessToken != "" {
		if claims, err := s.tokenSvc.ParseAccessToken(accessToken); err == nil && claims.UserID > 0 {
			userID = &claims.UserID
			account = claims.Account
			s.revokeSession(claims.SessionID, model.SessionStatusLoggedOut)
		}
	}

	// 2. refresh token 可用 -> 撤销该 token
	if refreshToken != "" {
		if rt, err := s.rtRepo.FindByTokenHash(utils.SHA256Hex(refreshToken)); err == nil {
			if rt.Status == model.RefreshTokenStatusActive {
				if err := s.rtRepo.UpdateStatus(rt.ID, model.RefreshTokenStatusRevoked); err != nil {
					s.log.Error("登出撤销 refresh token 失败", zap.Uint64("user_id", rt.UserID), zap.Error(err))
				}
				if userID == nil {
					s.revokeSession(rt.SessionID, model.SessionStatusLoggedOut)
					userID = &rt.UserID
				}
			}
		}
	}

	s.recordAudit(userID, account, model.AuditEventLogout, true, "", meta)
}

// revokeSession 撤销指定会话及其全部 refresh token，并清理缓存
func (s *AuthService) revokeSession(sessionID string, status int) {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := repository.NewRefreshTokenRepository(tx).RevokeBySessionID(sessionID, model.RefreshTokenStatusRevoked); err != nil {
			return err
		}
		return repository.NewSessionRepository(tx).UpdateStatusBySessionID(sessionID, status)
	})
	if err != nil {
		s.log.Error("撤销会话失败", zap.String("session_id", sessionID), zap.Error(err))
	}
	s.cache.DeleteSessionCache(sessionID)
}

// ---------- 修改密码 ----------

// ChangePassword 修改密码：更新密码并递增版本，随后撤销该用户全部会话与令牌
func (s *AuthService) ChangePassword(userID uint64, password, confirmPassword string, meta RequestMeta) error {
	if password != confirmPassword {
		return NewBizError(CodeBadRequest, "两次输入的密码不一致")
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		s.log.Error("密码加密失败", zap.Uint64("user_id", userID), zap.Error(err))
		return NewBizError(CodeServerError, "服务器内部错误")
	}

	// 更新密码（password_version + 1 由 SQL 保证原子递增）
	if err := s.userRepo.UpdatePassword(userID, passwordHash); err != nil {
		s.log.Error("更新密码失败", zap.Uint64("user_id", userID), zap.Error(err))
		return NewBizError(CodeServerError, "服务器内部错误")
	}

	// 撤销该用户全部会话与 refresh token，修改后必须重新登录
	s.revokeAllUserSessions(userID, model.AuditEventChangePassword, meta)
	return nil
}

// ---------- 当前用户信息 ----------

// Me 查询当前用户信息
func (s *AuthService) Me(userID uint64) (*MeResult, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewBizError(CodeUnauthorized, "用户不存在")
		}
		s.log.Error("查询用户失败", zap.Uint64("user_id", userID), zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}

	return &MeResult{
		User:    buildUserInfo(user),
		Groups:  []PermItem{},
		Roles:   []PermItem{{Code: "user", Name: "普通用户"}},
		Apps:    []string{},
		Pages:   []string{},
		Apis:    []string{},
		Menus:   []string{},
		Buttons: []string{},
	}, nil
}

// ---------- Token 校验 ----------

// ValidateAccessToken 完整校验 access token
func (s *AuthService) ValidateAccessToken(accessToken string) (*AccessClaims, error) {
	claims, err := s.tokenSvc.ParseAccessToken(accessToken)
	if err != nil {
		if errors.Is(err, ErrTokenExpired) {
			return nil, NewBizError(CodeUnauthorized, "token expired")
		}
		return nil, NewBizError(CodeUnauthorized, "token invalid")
	}

	// 版本缓存快速否决：改密后旧 token 立即失效，无需回源
	if version, ok := s.cache.GetPasswordVersion(claims.UserID); ok && version != claims.PasswordVersion {
		return nil, NewBizError(CodeUnauthorized, "password changed, please login again")
	}

	// session 校验（缓存优先，未命中回源 MySQL 并回写）
	if !s.isSessionActive(claims) {
		return nil, NewBizError(CodeUnauthorized, "session invalid")
	}

	// 用户校验（密码版本以数据库为准）
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewBizError(CodeUnauthorized, "user not found")
		}
		s.log.Error("查询用户失败", zap.Uint64("user_id", claims.UserID), zap.Error(err))
		return nil, NewBizError(CodeServerError, "服务器内部错误")
	}
	if user.PasswordVersion != int32(claims.PasswordVersion) {
		// 回写最新版本，加速后续旧 token 否决
		s.cache.SetPasswordVersion(user.ID, int(user.PasswordVersion), passwordVersionCacheTTL)
		return nil, NewBizError(CodeUnauthorized, "password changed, please login again")
	}
	s.cache.SetPasswordVersion(user.ID, int(user.PasswordVersion), passwordVersionCacheTTL)

	return claims, nil
}

// Introspect 校验 access token 有效性（带 Redis 缓存，TTL 不超过 30 秒）
func (s *AuthService) Introspect(accessToken string) (*IntrospectResult, error) {
	tokenHash := utils.SHA256Hex(accessToken)

	// 缓存命中直接返回
	if data, ok := s.cache.GetIntrospectCache(tokenHash); ok && data.Valid {
		return &IntrospectResult{
			UserID:          data.UserID,
			SessionID:       data.SessionID,
			PasswordVersion: data.PasswordVersion,
			Valid:           true,
		}, nil
	}

	claims, err := s.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, err
	}

	result := &IntrospectResult{
		UserID:          claims.UserID,
		SessionID:       claims.SessionID,
		PasswordVersion: claims.PasswordVersion,
		Valid:           true,
	}

	// 缓存 TTL = min(token 剩余时间, 30s)
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl > introspectCacheMaxTTL {
		ttl = introspectCacheMaxTTL
	}
	s.cache.SetIntrospectCache(tokenHash, cache.IntrospectCacheData{
		UserID:          claims.UserID,
		SessionID:       claims.SessionID,
		PasswordVersion: claims.PasswordVersion,
		Valid:           true,
	}, ttl)

	return result, nil
}

// ---------- 撤销用户会话 ----------

// RevokeUserSessions 撤销指定用户全部会话与 refresh token（内部接口）
func (s *AuthService) RevokeUserSessions(userID uint64, meta RequestMeta) error {
	if userID <= 0 {
		return NewBizError(CodeBadRequest, "userId is required")
	}
	s.revokeAllUserSessions(userID, model.AuditEventRevoke, meta)
	return nil
}

// revokeAllUserSessions 撤销用户全部 active 会话、refresh token 与相关缓存，并记录审计
func (s *AuthService) revokeAllUserSessions(userID uint64, event string, meta RequestMeta) {
	// 先取 active 会话 ID 列表，用于清理缓存
	sessionIDs, err := s.sessionRepo.ListActiveSessionIDs(userID)
	if err != nil {
		s.log.Error("查询用户会话列表失败", zap.Uint64("user_id", userID), zap.Error(err))
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := repository.NewSessionRepository(tx).RevokeActiveByUserID(userID, model.SessionStatusRevoked); err != nil {
			return err
		}
		return repository.NewRefreshTokenRepository(tx).RevokeActiveByUserID(userID, model.RefreshTokenStatusRevoked)
	})
	if err != nil {
		s.log.Error("撤销用户全部会话失败", zap.Uint64("user_id", userID), zap.Error(err))
	}

	// 清理缓存
	for _, sessionID := range sessionIDs {
		s.cache.DeleteSessionCache(sessionID)
	}
	s.cache.DeletePasswordVersion(userID)

	s.recordAudit(&userID, "", event, true, "", meta)
	s.log.Info("已撤销用户全部会话", zap.Uint64("user_id", userID), zap.Int("session_count", len(sessionIDs)))
}

// ---------- 内部辅助 ----------

// isSessionActive 校验会话是否有效（Redis 缓存优先，未命中回源 MySQL 并回写）
func (s *AuthService) isSessionActive(claims *AccessClaims) bool {
	if data, ok := s.cache.GetSessionCache(claims.SessionID); ok {
		if data.Status != model.SessionStatusActive {
			return false
		}
		// 缓存内含登录时刻密码版本时做防御性校验
		if data.PasswordVersion != 0 && data.PasswordVersion != claims.PasswordVersion {
			return false
		}
		return true
	}

	session, err := s.sessionRepo.FindBySessionID(claims.SessionID)
	if err != nil {
		return false
	}
	now := time.Now()
	if !session.IsActive(now) {
		return false
	}

	// 回写缓存，TTL 与会话过期时间对齐（passwordVersion 未知时填 0）
	ttl := time.Until(session.ExpiredAt)
	s.cache.SetSessionCache(session.SessionID, cache.SessionCacheData{
		UserID:          session.UserID,
		Status:          session.Status,
		PasswordVersion: 0,
	}, ttl)
	return true
}

// recordAudit 写入审计日志（失败只记错误日志，不阻断主流程）
func (s *AuthService) recordAudit(userID *uint64, account, eventType string, success bool, failReason string, meta RequestMeta) {
	record := &repository.AuditRecord{
		UserID:     userID,
		Account:    emptyToNil(truncate(account, 64)),
		Action:     eventType,
		Success:    success,
		FailReason: emptyToNil(failReason),
		IP:         emptyToNil(meta.IP),
		UserAgent:  emptyToNil(meta.UserAgent),
		RequestID:  emptyToNil(meta.RequestID),
	}
	if err := s.auditRepo.Create(record); err != nil {
		s.log.Error("写入审计日志失败", zap.String("event_type", eventType), zap.Error(err))
	}
}

// buildUserInfo 组装用户展示信息（字段与接口文档对齐）
func buildUserInfo(user *db_model.SysUser) UserInfo {
	passwordChanged := 0
	if user.PasswordVersion > 1 {
		passwordChanged = 1
	}
	return UserInfo{
		ID:              user.ID,
		Account:         user.Account,
		Name:            user.Name,
		Phone:           nilToEmpty(user.Phone),
		Email:           nilToEmpty(user.Email),
		Role:            "user", // MVP 固定基础角色，接入角色表后扩展
		Department:      nilToEmpty(user.Department),
		PasswordChanged: passwordChanged,
		CreateTime:      user.CreateTime,
		UpdateTime:      user.UpdateTime,
	}
}

// emptyToNil 空字符串转 nil（可空列存 NULL）
func emptyToNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nilToEmpty 指针为 nil 时返回空字符串
func nilToEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// truncate 截断字符串到指定字节长度（防超长输入写库失败）
func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
