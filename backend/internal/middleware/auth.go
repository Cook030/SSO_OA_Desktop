package middleware

import (
	"errors"
	"net/http"
	"strconv"

	"permission-system/internal/client"
	"permission-system/internal/repository"
	"permission-system/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// gin.Context 中认证信息的键
const (
	// ContextKeyUserID 当前登录用户ID
	ContextKeyUserID = "userId"
	// ContextKeyAccount 当前登录账号
	ContextKeyAccount = "account"
)

// CurrentUserID 从 gin.Context 读取当前登录用户ID；未认证时为 0
func CurrentUserID(c *gin.Context) int64 {
	return c.GetInt64(ContextKeyUserID)
}

// CurrentAccount 从 gin.Context 读取当前登录账号
func CurrentAccount(c *gin.Context) string {
	account, _ := c.Get(ContextKeyAccount)
	if v, ok := account.(string); ok {
		return v
	}
	return ""
}

// 后端仅认 Cookie 通道，不再解析 Authorization: Bearer Header。
// errMsg 为空表示提取成功；否则为缺少认证信息。
func extractAccessToken(c *gin.Context) (accessToken, errMsg string) {
	token, err := c.Cookie(client.AccessTokenCookieName)
	if err != nil {
		return "", "读取认证信息失败"
	}
	if token == "" {
		return "", "token为空"
	}
	return token, ""
}

// AuthMiddleware SSO Token 认证中间件
// 每个请求携带 SSO accessToken，中间件转发 SSO /api/v1/auth/introspect 校验，
// introspect 通过后用返回的 userId 查本地用户，将用户信息写入上下文
func AuthMiddleware(ssoClient *client.SSOClient, userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := utils.GetLogger().With(zap.String("request_id", GetRequestID(c)))

		accessToken, errMsg := extractAccessToken(c)
		if errMsg != "" {
			utils.Unauthorized(c, errMsg)
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// 转发 SSO introspect
		ssoResp, err := ssoClient.IntrospectToken(accessToken, GetRequestID(c))
		if err != nil {
			var expiredErr *client.SSOTokenExpiredError
			if errors.As(err, &expiredErr) {
				log.Info("SSO introspect 失败",
					zap.String("reason", "token_invalid"),
				)
				// 透传 SSO 返回的 401 消息给前端，由前端决定 refresh 或跳转登录页
				utils.Unauthorized(c, expiredErr.Message)
			} else {
				// SSO 服务本身异常（网络、500等），不应让前端误以为是 Token 过期
				log.Error("SSO introspect 请求失败",
					zap.Error(err),
				)
				utils.ServerError(c, "认证服务异常")
			}
			c.AbortWithStatus(http.StatusOK)
			return
		}
		if ssoResp.Code != 200 {
			log.Info("SSO introspect 失败",
				zap.String("reason", "token_invalid"),
			)
			utils.Unauthorized(c, "Token 校验失败")
			c.AbortWithStatus(http.StatusOK)
			return
		}
		if ssoResp.Data.UserID == "" {
			log.Info("SSO introspect 失败",
				zap.String("reason", "user_id_empty"),
			)
			utils.Unauthorized(c, "UserID为空")
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// 用 SSO 返回的 userId 查本地用户（id 与 SSO userId 对齐）
		userID, err := strconv.ParseInt(ssoResp.Data.UserID, 10, 64)
		if err != nil {
			log.Info("SSO introspect 失败",
				zap.String("reason", "user_id_empty"),
			)
			utils.Unauthorized(c, "用户不存在")
			c.AbortWithStatus(http.StatusOK)
			return
		}
		user, err := userRepo.FindByID(userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Info("SSO introspect 失败",
					zap.String("reason", "user_not_found"),
				)
				utils.Unauthorized(c, "用户不存在")
			} else {
				log.Info("SSO introspect 失败",
					zap.String("reason", "user_query_failed"),
				)
				utils.Unauthorized(c, "用户查询失败")
			}
			c.AbortWithStatus(http.StatusOK)
			return
		}

		log.Info("SSO introspect 成功",
			zap.Int64("user_id", user.ID),
		)

		// 将用户信息存入上下文
		c.Set(ContextKeyUserID, user.ID)
		c.Set(ContextKeyAccount, user.Account)
		c.Next()
	}
}
