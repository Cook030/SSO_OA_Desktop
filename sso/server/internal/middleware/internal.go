package middleware

import (
	"crypto/subtle"

	"mh-sso-svc/internal/utils"

	"github.com/gin-gonic/gin"
)

// InternalAuthMiddleware 内部服务鉴权中间件（用于 /introspect、/revoke-user-sessions）。
// 校验 X-Internal-Token 与配置值，使用常量时间比较防止时序攻击
func InternalAuthMiddleware(internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(utils.InternalTokenHeaderName)
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(internalToken)) != 1 {
			utils.Unauthorized(c, "unauthorized internal access")
			c.Abort()
			return
		}
		c.Next()
	}
}
