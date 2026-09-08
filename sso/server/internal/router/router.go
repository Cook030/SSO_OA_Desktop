package router

import (
	"net/http"
	"time"

	"mh-sso-svc/internal/cache"
	"mh-sso-svc/internal/handler"
	"mh-sso-svc/internal/middleware"
	"mh-sso-svc/internal/service"
	"mh-sso-svc/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter 初始化各层依赖并设置路由
func SetupRouter(db *gorm.DB, rdb *cache.Cache, cfg *utils.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	// 全局 CORS：allow_credentials=true 时 Origin 必须显式配置，禁止 *
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		ExposeHeaders:    cfg.CORS.ExposeHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           time.Duration(cfg.CORS.MaxAge) * time.Second,
	}))

	// 全局 request_id 中间件：为整个 HTTP 请求生成唯一 ID 并记录访问日志
	r.Use(middleware.RequestIDMiddleware())

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// 依赖组装：Repository → Service → Handler
	tokenSvc := service.NewTokenService(&cfg.Auth)
	authSvc := service.NewAuthService(db, rdb, tokenSvc, &cfg.Auth)
	authHandler := handler.NewAuthHandler(authSvc, &cfg.Auth)
	internalHandler := handler.NewInternalHandler(authSvc)

	auth := r.Group("/api/v1/auth")
	{
		// 公开接口
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)

		// access token 认证接口
		authed := auth.Group("")
		authed.Use(middleware.AuthMiddleware(authSvc))
		{
			authed.GET("/me", authHandler.Me)
			authed.PUT("/me", authHandler.UpdateProfile)
			authed.POST("/change-password", authHandler.ChangePassword)
		}

		// 供业务后端调用的服务接口。
		// 注意：按 dev.md §4，本接口不应作为公开接口对外暴露；
		// 内部调用方应改用下方 /internal/v1 下带服务间鉴权的接口，此处保留仅为兼容既有调用方。
		auth.GET("/introspect", authHandler.Introspect)
		auth.POST("/revoke-user-sessions", authHandler.RevokeUserSessions)
	}

	// 服务间内部接口：仅供内网服务（当前为 WebSocket Gateway）调用，
	// 由 ServiceAuthMiddleware 校验共享密钥与来源网段，禁止公网访问。
	internal := r.Group("/internal/v1")
	internal.Use(middleware.ServiceAuthMiddleware(&cfg.Internal))
	{
		internal.POST("/sessions/introspect", internalHandler.Introspect)
	}

	return r
}
