package middleware

import (
	"net/http"

	"permission-system/internal/service/rbac"
	"permission-system/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequirePermission 权限校验中间件（Casbin 风格的 subject / object / action 判定）。
//
// object 与 action 由路由注册处显式声明（如 RequirePermission(enforcer, "platform", "create")），
// 中间件负责取出当前用户、加载其角色，并交由 Enforcer 判定是否放行。
func RequirePermission(enforcer rbac.Enforcer, object, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := utils.GetLogger().With(zap.String("request_id", GetRequestID(c)))

		userID := CurrentUserID(c)
		if userID == 0 {
			utils.Unauthorized(c, "未认证")
			c.AbortWithStatus(http.StatusOK)
			return
		}

		roles, err := enforcer.LoadRoles(userID)
		if err != nil {
			log.Error("查询用户角色失败",
				zap.Int64("user_id", userID),
				zap.Error(err),
			)
			utils.ServerError(c, "权限校验失败")
			c.AbortWithStatus(http.StatusOK)
			return
		}

		subject := rbac.Subject{UserID: userID, Roles: roles}
		if !enforcer.Enforce(subject, object, action) {
			log.Info("权限校验未通过",
				zap.Int64("user_id", userID),
				zap.Strings("roles", roles),
				zap.String("object", object),
				zap.String("action", action),
			)
			utils.Forbidden(c, "无权限访问")
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
