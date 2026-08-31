package router

import (
	"net/http"
	"time"

	"permission-system/internal/client"
	"permission-system/internal/db_model/query"
	"permission-system/internal/handler"
	"permission-system/internal/middleware"
	"permission-system/internal/repository"
	"permission-system/internal/service"
	"permission-system/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// SetupRouter 初始化各层依赖并设置路由
func SetupRouter(db *gorm.DB, cfg *utils.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	// 全局 CORS 中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		ExposeHeaders:    cfg.CORS.ExposeHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           time.Duration(cfg.CORS.MaxAge) * time.Second,
	}))

	// 全局 request_id 中间件：为整个 HTTP 请求生成唯一 ID 并记录请求开始/完成日志
	r.Use(middleware.RequestIDMiddleware())

	// 创建 gen query，用于类型安全的数据库操作
	q := query.Use(db)

	// 组装 Repository → Service → Handler
	userRepo := repository.NewUserRepository(q)
	platformRepo := repository.NewPlatformRepository(q)
	userPlatformRepo := repository.NewUserPlatformRepository(q)

	// PermissionService 最先创建（被 Employee/Platform 两个 Service 依赖）
	permissionService := service.NewPermissionService(q, userPlatformRepo, platformRepo)

	ssoClient := client.NewSSOClient(cfg.SSO)
	platformService := service.NewPlatformService(platformRepo)
	employeeService := service.NewEmployeeService(q, userRepo, ssoClient)

	platformHandler := handler.NewPlatformHandler(platformService, permissionService, q)
	employeeHandler := handler.NewEmployeeHandler(employeeService, permissionService, q)
	permissionHandler := handler.NewPermissionHandler(permissionService, q)

	testGroup := r.Group("/test")
	{
		testGroup.GET(
			"/alive",
			func(ctx *gin.Context) {
				ctx.String(http.StatusOK, "pong1")
			},
		)
	}

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")

	// 需要认证的接口：全程使用 SSO accessToken，中间件转发 SSO 校验
	auth := api.Group("")
	auth.Use(middleware.AuthMiddleware(ssoClient, userRepo))
	{
		// 需要管理员权限的接口
		admin := auth.Group("")
		admin.Use(middleware.AdminMiddleware())
		{
			// 平台管理
			admin.GET("/platforms", platformHandler.List)
			admin.POST("/platforms", platformHandler.Create)
			admin.PUT("/platforms/:id", platformHandler.Update)
			admin.DELETE("/platforms/:id", platformHandler.Delete)

			// 员工管理
			admin.GET("/employees", employeeHandler.List)
			admin.GET("/employees/departments", employeeHandler.GetDepartments)
			admin.POST("/employees", employeeHandler.Create)
			admin.PUT("/employees/:id", employeeHandler.Update)
			admin.DELETE("/employees/:id", employeeHandler.Delete)

			// 重置员工密码(管理员)
			admin.PUT("/employees/:id/reset-password", employeeHandler.ResetPassword)

			// 权限管理
			admin.POST("/employees/permissions/batch", permissionHandler.BatchSet)
			admin.DELETE("/employees/permissions/batch", permissionHandler.BatchDelete)
		}
	}

	return r
}
