package middleware

import (
	"errors"
	"strings"

	"mh-sso-svc/internal/service"
	"mh-sso-svc/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// gin.Context 上下文键（AuthMiddleware 写入，handler 读取）
const (
	CtxUserID    = "userId"
	CtxSessionID = "sessionId"
	CtxAccount   = "account"
)

// GetUserID 从上下文读取当前用户 ID；未认证时返回 0
func GetUserID(c *gin.Context) uint64 {
	v, exists := c.Get(CtxUserID)
	if !exists {
		return 0
	}
	id, ok := v.(uint64)
	if !ok {
		return 0
	}
	return id
}

// ExtractAccessTokenCookieFirst 从请求提取 access token：
// 优先 Cookie mh_sso_access_token，其次 Authorization: Bearer
func ExtractAccessTokenCookieFirst(c *gin.Context) (string, string) {
	if token, err := c.Cookie(utils.AccessTokenCookieName); err == nil && token != "" {
		return token, ""
	}
	return extractBearerToken(c)
}

// ExtractAccessTokenBearerFirst 从请求提取 access token：
// 优先 Authorization: Bearer，其次 Cookie mh_sso_access_token
func ExtractAccessTokenBearerFirst(c *gin.Context) (string, string) {
	if token, errMsg := extractBearerToken(c); errMsg == "" {
		return token, ""
	}
	if token, err := c.Cookie(utils.AccessTokenCookieName); err == nil && token != "" {
		return token, ""
	}
	return "", "missing access token"
}

// extractBearerToken 从 Authorization Header 提取 Bearer token
func extractBearerToken(c *gin.Context) (string, string) {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if token != "" {
			return token, ""
		}
	}
	return "", "missing access token"
}

// AuthMiddleware access token 认证中间件（用于 /me、/change-password）。
// 校验链：JWT（签名/issuer/exp）→ session active → user active → passwordVersion 一致；
// 通过后将 userId / sessionId / account 写入上下文
func AuthMiddleware(svc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, errMsg := ExtractAccessTokenCookieFirst(c)
		if errMsg != "" {
			utils.Unauthorized(c, errMsg)
			c.Abort()
			return
		}

		claims, err := svc.ValidateAccessToken(accessToken)
		if err != nil {
			var biz *service.BizError
			if errors.As(err, &biz) {
				utils.Unauthorized(c, biz.Msg)
			} else {
				utils.GetLogger().Error("access token 校验异常",
					zap.String("path", c.Request.URL.Path), zap.Error(err))
				utils.ServerError(c, "服务器内部错误")
			}
			c.Abort()
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxSessionID, claims.SessionID)
		c.Set(CtxAccount, claims.Account)
		c.Next()
	}
}
