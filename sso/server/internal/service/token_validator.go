package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mh-sso-svc/internal/cache"
	"mh-sso-svc/internal/consts"
	"mh-sso-svc/internal/utils"

	"gorm.io/gorm"
)

// ValidateAccessToken 完整校验 access token
//
// 单设备登录的校验要求：
//
//	JWT 签名与过期时间合法
//	AND JWT.sessionId == sso:user:{userId}:current_session
//	AND sso:session:{sessionId}.status == active
//
// 三条同时满足才算有效。因此即使旧客户端没有收到任何 WebSocket 下线消息，
// 它的 access token 也会在会话被置换的瞬间失效。
func (s *AuthService) ValidateAccessToken(accessToken string) (*AccessClaims, error) {
	claims, err := s.tokenSvc.ParseAccessToken(accessToken)
	if err != nil {
		if errors.Is(err, ErrTokenExpired) {
			return nil, utils.NewBizErrorWithReason(utils.CodeUnauthorized, "登录状态已失效，请重新登录", consts.ReasonTokenExpired)
		}
		return nil, utils.NewBizErrorWithReason(utils.CodeUnauthorized, "登录状态已失效，请重新登录", consts.ReasonTokenInvalid)
	}

	// 版本缓存快速否决：改密后旧 token 立即失效，无需回源
	if version, ok := s.cache.GetPasswordVersion(claims.UserID); ok && version != claims.PasswordVersion {
		return nil, utils.NewBizErrorWithReason(utils.CodeUnauthorized, "密码已修改，请重新登录", consts.ReasonPasswordChanged)
	}

	// 会话校验（Redis 为权威状态，无 MySQL 兜底）
	if err := s.verifySession(claims); err != nil {
		return nil, err
	}

	// 用户校验（密码版本以数据库为准）
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewBizErrorWithReason(utils.CodeUnauthorized, "用户不存在", consts.ReasonTokenInvalid)
		}
		return nil, fmt.Errorf("校验 access token 查询用户失败(uid=%d): %w", claims.UserID, err)
	}
	if user.PasswordVersion != int32(claims.PasswordVersion) {
		// 回写最新版本，加速后续旧 token 否决
		s.cache.SetPasswordVersion(user.ID, int(user.PasswordVersion), passwordVersionCacheTTL)
		return nil, utils.NewBizErrorWithReason(utils.CodeUnauthorized, "密码已修改，请重新登录", consts.ReasonPasswordChanged)
	}
	s.cache.SetPasswordVersion(user.ID, int(user.PasswordVersion), passwordVersionCacheTTL)

	return claims, nil
}

// verifySession 校验会话是否仍为该用户唯一有效会话
func (s *AuthService) verifySession(claims *AccessClaims) error {
	current, err := s.cache.GetCurrentSession(context.Background(), claims.UserID)
	if err != nil {
		// Redis 不可用时必须拒绝，不能退化为"放行但不校验会话"
		return fmt.Errorf("查询当前会话失败(uid=%d): %w", claims.UserID, err)
	}
	// current_session 缺失或不等：该会话已被置换 / 撤销 / 登出
	// 注：首次启用单设备登录时，历史会话没有 current_session，需要重新登录，这是预期行为
	if current == "" || current != claims.SessionID {
		return utils.NewBizErrorWithReason(utils.CodeUnauthorized, "账号已在其他设备登录", consts.ReasonSessionReplaced)
	}

	session, err := s.cache.GetSession(claims.SessionID)
	if err != nil {
		if errors.Is(err, cache.ErrRecordNotFound) {
			return utils.NewBizErrorWithReason(utils.CodeUnauthorized, "登录状态已失效，请重新登录", consts.ReasonSessionNotFound)
		}
		return fmt.Errorf("校验会话查询失败(sid=%s): %w", claims.SessionID, err)
	}
	if !session.IsActive(time.Now()) {
		return sessionInactiveError(session.Status)
	}
	// 会话内记录登录时刻密码版本，不一致直接否决
	if session.PasswordVersion != 0 && session.PasswordVersion != claims.PasswordVersion {
		return utils.NewBizErrorWithReason(utils.CodeUnauthorized, "密码已修改，请重新登录", consts.ReasonPasswordChanged)
	}
	return nil
}

// sessionInactiveError 按会话最终状态生成带原因码的业务错误
func sessionInactiveError(status int) *utils.BizError {
	switch status {
	case consts.SessionStatusReplaced:
		return utils.NewBizErrorWithReason(utils.CodeUnauthorized, "账号已在其他设备登录", consts.ReasonSessionReplaced)
	case consts.SessionStatusLoggedOut:
		return utils.NewBizErrorWithReason(utils.CodeUnauthorized, "登录状态已失效，请重新登录", consts.ReasonSessionLoggedOut)
	case consts.SessionStatusRevoked:
		return utils.NewBizErrorWithReason(utils.CodeUnauthorized, "登录状态已失效，请重新登录", consts.ReasonSessionRevoked)
	default:
		return utils.NewBizErrorWithReason(utils.CodeUnauthorized, "登录状态已过期，请重新登录", consts.ReasonSessionExpired)
	}
}

// Introspect 校验 access token 有效性（带 Redis 缓存，TTL 不超过 30 秒）
func (s *AuthService) Introspect(accessToken string) (*IntrospectResult, error) {
	tokenHash := utils.SHA256Hex(accessToken)

	// 缓存命中时仍需实时确认"该会话仍是当前会话"：
	// 单设备登录要求被顶下线的 access token 立即失效，不能等 30 秒缓存自然过期
	if data, ok := s.cache.GetIntrospectCache(tokenHash); ok && data.Valid {
		current, err := s.cache.GetCurrentSession(context.Background(), data.UserID)
		if err == nil && current != "" && current == data.SessionID {
			return &IntrospectResult{
				UserID:          data.UserID,
				SessionID:       data.SessionID,
				PasswordVersion: data.PasswordVersion,
				Valid:           true,
			}, nil
		}
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

// IntrospectSession 供 WebSocket Gateway 调用的会话校验。
//
// 与 Introspect 的区别：不返回 error 表达"无效"，而是返回 active=false 与机器原因码，
// 让 Gateway 能据此决定关闭连接的关闭码（见 ws_gateway 的关闭码定义）。
// 只有 active=true 的连接才会被 Gateway 接受。
func (s *AuthService) IntrospectSession(accessToken string) (*GatewayIntrospectResult, error) {
	if accessToken == "" {
		return &GatewayIntrospectResult{Active: false, Reason: consts.ReasonTokenInvalid}, nil
	}
	claims, err := s.ValidateAccessToken(accessToken)
	if err != nil {
		return &GatewayIntrospectResult{Active: false, Reason: reasonOf(err)}, nil
	}
	session, err := s.cache.GetSession(claims.SessionID)
	if err != nil {
		// 会话记录异常：按无效处理，但原因码保留，便于 Gateway 侧排查
		return &GatewayIntrospectResult{Active: false, Reason: consts.ReasonSessionNotFound}, nil
	}

	result := &GatewayIntrospectResult{
		Active:    true,
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		DeviceID:  utils.NilToEmpty(session.DeviceID),
	}
	if claims.ExpiresAt != nil {
		result.ExpiresAt = claims.ExpiresAt.Time
	}
	return result, nil
}

// reasonOf 从错误中提取机器原因码；非业务错误返回空串
func reasonOf(err error) string {
	var biz *utils.BizError
	if errors.As(err, &biz) {
		return biz.Reason
	}
	return ""
}
