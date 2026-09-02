package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"mh-sso-svc/internal/cache"
	"mh-sso-svc/internal/consts"
	"mh-sso-svc/internal/model/query"
	"mh-sso-svc/internal/repository"
	"mh-sso-svc/internal/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// dummyHash 用于账号不存在时的恒定耗时密码比对，防止时序探测账号存在性
const dummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// AuthService SSO 认证核心服务编排层。
// 仅负责用例编排（Login/Refresh/Logout/ChangePassword/Me/UpdateProfile/RevokeUserSessions）与错误翻译；
// 会话/令牌生命周期委托给 SessionService，token 校验委托给 TokenService，审计写入委托给 AuditRepository。
type AuthService struct {
	q          *query.Query
	userRepo   *repository.UserRepository
	auditRepo  *repository.AuditRepository
	cache      *cache.Cache
	tokenSvc   *TokenService
	sessionSvc *SessionService
	cfg        *utils.AuthConfig
	log        *zap.Logger
}

// NewAuthService 创建认证服务（Repository 层在内部组装；db 仅用于构建 gen query）
func NewAuthService(db *gorm.DB, rdb *cache.Cache, tokenSvc *TokenService, cfg *utils.AuthConfig) *AuthService {
	q := query.Use(db)
	s := &AuthService{
		q:         q,
		userRepo:  repository.NewUserRepository(q),
		auditRepo: repository.NewAuditRepository(q),
		cache:     rdb,
		tokenSvc:  tokenSvc,
		cfg:       cfg,
		log:       utils.GetLogger(),
	}
	// 会话管理器注入审计回调，实现职责分离
	s.sessionSvc = NewSessionService(rdb, s.log)
	s.sessionSvc.SetAuditRecorder(s.recordAudit)
	return s
}

func (s *AuthService) refreshTTL() time.Duration {
	return time.Duration(s.cfg.RefreshTokenTTLSecond) * time.Second
}

func (s *AuthService) sessionTTL() time.Duration {
	return time.Duration(s.cfg.SessionTTLSecond) * time.Second
}

// ---------- 登录 ----------

// Login 账号密码登录：校验用户与密码、创建会话、签发双 token、记审计
func (s *AuthService) Login(account, password string, meta RequestMeta) (*LoginResult, error) {
	account = utils.Truncate(account, maxAccountInputLength)

	// 1. 登录失败限流（账号 / IP）
	if s.cache.IsLoginRateLimited(account, meta.IP) {
		s.recordAudit(nil, account, consts.AuditEventLoginFailed, false, "rate_limited", meta)
		return nil, utils.NewBizError(utils.CodeUnauthorized, "登录失败次数过多，请5分钟后再试")
	}

	// 2. 查询用户（account / email / phone）
	user, err := s.userRepo.FindByAccount(account)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("登录查询用户失败(account=%s): %w", account, err)
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

	// 4. 创建会话与 refresh token（原子：任一写入失败则整体失败，不产生残缺状态）
	sessionID, refreshToken, err := s.sessionSvc.CreateLoginSession(
		user, meta.IP, meta.UserAgent, s.sessionTTL(), s.refreshTTL(), now)
	if err != nil {
		return nil, fmt.Errorf("登录创建会话失败(uid=%d): %w", user.ID, err)
	}

	// 5. 签发 access token
	accessToken, _, _, err := s.tokenSvc.GenerateAccessToken(user.ID, sessionID, user.Account, int(user.PasswordVersion), now)
	if err != nil {
		return nil, fmt.Errorf("登录签发 access token 失败(uid=%d): %w", user.ID, err)
	}

	// 6. 登录成功：清理失败计数 + 记审计
	s.cache.ClearLoginFailures(account)
	s.recordAudit(&user.ID, user.Account, consts.AuditEventLoginSuccess, true, "", meta)

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
	s.recordAudit(nil, account, consts.AuditEventLoginFailed, false, reason, meta)
	return nil, utils.NewBizError(utils.CodeUnauthorized, "账号或密码错误")
}

// ---------- 刷新 Token ----------

// Refresh 使用 refresh token 换取新 token
func (s *AuthService) Refresh(refreshToken string, meta RequestMeta) (*RefreshResult, error) {
	if refreshToken == "" {
		return nil, utils.NewBizError(utils.CodeUnauthorized, "missing refresh token")
	}

	tokenHash := utils.SHA256Hex(refreshToken)
	now := time.Now()

	// 1. 查询令牌
	rt, err := s.cache.GetRefreshToken(tokenHash)
	if err != nil {
		if errors.Is(err, cache.ErrRecordNotFound) {
			return nil, utils.NewBizError(utils.CodeUnauthorized, "invalid refresh token")
		}
		return nil, fmt.Errorf("查询 refresh token 失败: %w", err)
	}

	// 2. 重放检测：已轮换 token 再次使用 -> 撤销整个 family 与会话
	if rt.Status == consts.RefreshTokenStatusRotated {
		s.sessionSvc.handleRefreshReplay(rt, meta)
		return nil, utils.NewBizError(utils.CodeUnauthorized, "refresh token 已被使用，会话已撤销")
	}
	if rt.Status != consts.RefreshTokenStatusActive {
		return nil, utils.NewBizError(utils.CodeUnauthorized, "invalid refresh token")
	}
	if !rt.ExpiredAt.After(now) {
		_ = s.cache.UpdateRefreshTokenStatus(tokenHash, consts.RefreshTokenStatusExpired)
		return nil, utils.NewBizError(utils.CodeUnauthorized, "refresh token 已过期")
	}

	// 3. 校验会话
	session, err := s.cache.GetSession(rt.SessionID)
	if err != nil {
		if !errors.Is(err, cache.ErrRecordNotFound) {
			return nil, fmt.Errorf("刷新查询会话失败(sid=%s): %w", rt.SessionID, err)
		}
		return nil, utils.NewBizError(utils.CodeUnauthorized, "session invalid")
	}
	if !session.IsActive(now) {
		return nil, utils.NewBizError(utils.CodeUnauthorized, "session invalid")
	}

	// 4. 校验用户
	user, err := s.userRepo.FindByID(rt.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewBizError(utils.CodeUnauthorized, "user not found")
		}
		return nil, fmt.Errorf("刷新查询用户失败(uid=%d): %w", rt.UserID, err)
	}

	// 5. 并发锁：防止同一 refresh token 并发刷新
	if !s.cache.AcquireRefreshLock(tokenHash, meta.RequestID) {
		return nil, utils.NewBizError(utils.CodeUnauthorized, "刷新请求处理中，请稍后重试")
	}
	defer s.cache.ReleaseRefreshLock(tokenHash, meta.RequestID)

	// 6. 二次确认状态（拿到锁后 token 可能已被并发请求旋转）
	rtLatest, err := s.cache.GetRefreshToken(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("二次查询 refresh token 失败: %w", err)
	}
	if rtLatest.Status == consts.RefreshTokenStatusRotated {
		s.sessionSvc.handleRefreshReplay(rtLatest, meta)
		return nil, utils.NewBizError(utils.CodeUnauthorized, "refresh token 已被使用，会话已撤销")
	}
	if rtLatest.Status != consts.RefreshTokenStatusActive {
		return nil, utils.NewBizError(utils.CodeUnauthorized, "invalid refresh token")
	}

	// 7. 旋转令牌并续期会话（原子：标记旧 token -> 写新 token -> 续期会话）
	newRefreshToken, err := s.sessionSvc.RotateRefreshToken(
		tokenHash, rt.UserID, rt.SessionID, s.sessionTTL(), s.refreshTTL(), now)
	if err != nil {
		return nil, fmt.Errorf("轮换 refresh token 失败(uid=%d): %w", rt.UserID, err)
	}

	// 8. 签发新 access token
	accessToken, _, _, err := s.tokenSvc.GenerateAccessToken(user.ID, rt.SessionID, user.Account, int(user.PasswordVersion), now)
	if err != nil {
		return nil, fmt.Errorf("刷新签发 access token 失败(uid=%d): %w", user.ID, err)
	}

	// 9. 记审计
	s.recordAudit(&user.ID, user.Account, consts.AuditEventRefresh, true, "", meta)
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
			s.sessionSvc.revokeSession(claims.SessionID, consts.SessionStatusLoggedOut)
		}
	}

	// 2. refresh token 可用 -> 撤销该 token
	if refreshToken != "" {
		if rt, err := s.cache.GetRefreshToken(utils.SHA256Hex(refreshToken)); err == nil {
			if rt.Status == consts.RefreshTokenStatusActive {
				if err := s.cache.UpdateRefreshTokenStatus(rt.TokenHash, consts.RefreshTokenStatusRevoked); err != nil {
					s.log.Error("登出撤销 refresh token 失败", zap.Uint64("user_id", rt.UserID), zap.Error(err))
				}
				if userID == nil {
					s.sessionSvc.revokeSession(rt.SessionID, consts.SessionStatusLoggedOut)
					userID = &rt.UserID
				}
			}
		}
	}

	s.recordAudit(userID, account, consts.AuditEventLogout, true, "", meta)
}

// ---------- 修改密码 ----------

// ChangePassword 修改密码：更新密码并递增版本，随后撤销该用户全部会话与令牌
func (s *AuthService) ChangePassword(userID uint64, password, confirmPassword string, meta RequestMeta) error {
	if password != confirmPassword {
		return utils.NewBizError(utils.CodeBadRequest, "两次输入的密码不一致")
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("修改密码加密失败(uid=%d): %w", userID, err)
	}

	// 更新密码（password_version + 1 由 SQL 保证原子递增）
	if err := s.userRepo.UpdatePassword(userID, passwordHash); err != nil {
		return fmt.Errorf("更新密码失败(uid=%d): %w", userID, err)
	}

	// 撤销该用户全部会话与 refresh token，修改后必须重新登录
	s.sessionSvc.revokeAllUserSessions(userID, consts.AuditEventChangePassword, meta)
	return nil
}

// ---------- 当前用户信息 ----------

// UpdateProfile 更新当前用户资料（姓名/邮箱/手机号），返回更新后的用户信息
func (s *AuthService) UpdateProfile(userID uint64, req UpdateProfileRequest, meta RequestMeta) (*UserInfo, error) {
	name := strings.TrimSpace(req.Nickname)
	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Mobile)

	if name == "" {
		return nil, utils.NewBizError(utils.CodeBadRequest, "姓名不能为空")
	}

	// 唯一性校验（排除自身：保留原值也算合法）
	if email != "" {
		exists, err := s.userRepo.ExistsByEmailExclude(email, userID)
		if err != nil {
			return nil, fmt.Errorf("资料更新邮箱校验失败(uid=%d): %w", userID, err)
		}
		if exists {
			return nil, utils.NewBizError(utils.CodeConflict, "邮箱已被其他账户使用")
		}
	}
	if phone != "" {
		exists, err := s.userRepo.ExistsByPhoneExclude(phone, userID)
		if err != nil {
			return nil, fmt.Errorf("资料更新手机号校验失败(uid=%d): %w", userID, err)
		}
		if exists {
			return nil, utils.NewBizError(utils.CodeConflict, "手机号已被其他账户使用")
		}
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewBizError(utils.CodeUnauthorized, "用户不存在")
		}
		return nil, fmt.Errorf("资料更新查询用户失败(uid=%d): %w", userID, err)
	}
	account := user.Account

	if err := s.userRepo.UpdateProfile(userID, name, utils.EmptyToNil(email), utils.EmptyToNil(phone)); err != nil {
		// 唯一索引兜底：并发情况下仍可能冲突
		if repository.IsDuplicateEntryError(err) {
			return nil, utils.NewBizError(utils.CodeConflict, "邮箱或手机号已被其他账户使用")
		}
		return nil, fmt.Errorf("更新资料失败(uid=%d): %w", userID, err)
	}

	s.recordAudit(&userID, account, consts.AuditEventUpdateProfile, true, "", meta)
	s.log.Info("用户资料更新成功",
		zap.Uint64("user_id", userID),
		zap.String("ip", meta.IP))

	updated, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("资料更新后查询用户失败(uid=%d): %w", userID, err)
	}
	info := buildUserInfo(updated)
	return &info, nil
}

// Me 查询当前用户信息
func (s *AuthService) Me(userID uint64) (*MeResult, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewBizError(utils.CodeUnauthorized, "用户不存在")
		}
		return nil, fmt.Errorf("查询当前用户失败(uid=%d): %w", userID, err)
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

// ---------- 撤销用户会话 ----------

// RevokeUserSessions 撤销指定用户全部会话与 refresh token（内部接口）
func (s *AuthService) RevokeUserSessions(userID uint64, meta RequestMeta) error {
	if userID <= 0 {
		return utils.NewBizError(utils.CodeBadRequest, "userId is required")
	}
	s.sessionSvc.revokeAllUserSessions(userID, consts.AuditEventRevoke, meta)
	return nil
}
