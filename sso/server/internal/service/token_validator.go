package service

import (
	"errors"
	"time"

	"mh-sso-svc/internal/cache"
	"mh-sso-svc/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ValidateAccessToken 完整校验 access token
func (s *AuthService) ValidateAccessToken(accessToken string) (*AccessClaims, error) {
	claims, err := s.tokenSvc.ParseAccessToken(accessToken)
	if err != nil {
		if errors.Is(err, ErrTokenExpired) {
			return nil, utils.NewBizError(utils.CodeUnauthorized, "token expired")
		}
		return nil, utils.NewBizError(utils.CodeUnauthorized, "token invalid")
	}

	// 版本缓存快速否决：改密后旧 token 立即失效，无需回源
	if version, ok := s.cache.GetPasswordVersion(claims.UserID); ok && version != claims.PasswordVersion {
		return nil, utils.NewBizError(utils.CodeUnauthorized, "password changed, please login again")
	}

	// session 校验（缓存优先，未命中回源 MySQL 并回写）
	if !s.isSessionActive(claims) {
		return nil, utils.NewBizError(utils.CodeUnauthorized, "session invalid")
	}

	// 用户校验（密码版本以数据库为准）
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewBizError(utils.CodeUnauthorized, "user not found")
		}
		s.log.Error("查询用户失败", zap.Uint64("user_id", claims.UserID), zap.Error(err))
		return nil, utils.NewBizError(utils.CodeServerError, "服务器内部错误")
	}
	if user.PasswordVersion != int32(claims.PasswordVersion) {
		// 回写最新版本，加速后续旧 token 否决
		s.cache.SetPasswordVersion(user.ID, int(user.PasswordVersion), passwordVersionCacheTTL)
		return nil, utils.NewBizError(utils.CodeUnauthorized, "password changed, please login again")
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

// isSessionActive 校验会话是否有效（会话记录以 Redis 为权威存储）
func (s *AuthService) isSessionActive(claims *AccessClaims) bool {
	session, err := s.cache.GetSession(claims.SessionID)
	if err != nil {
		return false
	}
	if !session.IsActive(time.Now()) {
		return false
	}
	// 会话内记录登录时刻密码版本，不一致直接否决
	if session.PasswordVersion != 0 && session.PasswordVersion != claims.PasswordVersion {
		return false
	}
	return true
}
