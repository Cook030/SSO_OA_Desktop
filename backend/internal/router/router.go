package router

import (
	"net/http"
	"time"

	"permission-system/internal/client"
	"permission-system/internal/db_model/query"
	"permission-system/internal/handler"
	"permission-system/internal/middleware"
	"permission-system/internal/repository"
	"permission-system/internal/service/employee"
	"permission-system/internal/service/permission"
	"permission-system/internal/service/platform"
	"permission-system/internal/service/rbac"
	"permission-system/internal/service/role"
	"permission-system/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
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

	// 组装 Repository 层
	userRepo := repository.NewUserRepository(q)
	platformRepo := repository.NewPlatformRepository(q)
	permRepo := repository.NewPermissionRepository(q)
	roleRepo := repository.NewRoleRepository(q)
	rolePermRepo := repository.NewRolePermissionRepository(q)
	userRoleRepo := repository.NewUserRoleRepository(q)

	// Enforcer：策略在内存中缓存，权限变更时由 Service 主动失效
	enforcer := rbac.NewCachedEnforcer(permRepo, userRoleRepo)
	if err := enforcer.ReloadPolicy(); err != nil {
		utils.GetLogger().Error("加载权限策略失败", zap.Error(err))
	}

	// 组装 Service 层（读方法直接转发 repository；写方法为 Tx 原子操作，事务由 Handler 用例层编排）
	ssoClient := client.NewSSOClient(cfg.SSO)
	permissionService := permission.New(permRepo, enforcer)
	accessService := permission.NewAccessService(platformRepo, permRepo, rolePermRepo, userRoleRepo, enforcer)
	platformService := platform.New(platformRepo, enforcer)
	roleService := role.New(roleRepo, rolePermRepo, userRoleRepo, enforcer)
	userRoleService := role.NewUserRoleService(userRoleRepo, enforcer)
	employeeService := employee.New(userRepo, userRoleRepo, enforcer, ssoClient)

	// 组装 Handler 层（Handler 持有 q，负责用例事务编排与响应组装）
	platformHandler := handler.NewPlatformHandler(platformService, accessService, q)
	employeeHandler := handler.NewEmployeeHandler(employeeService, accessService, userRoleService, q)
	roleHandler := handler.NewRoleHandler(roleService, userRoleService, q)
	permissionHandler := handler.NewPermissionHandler(permissionService, accessService, enforcer, q)

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
		// 当前用户权限视图（所有登录用户均可访问）
		auth.GET("/me/permissions", permissionHandler.Me)

		// 平台管理
		auth.GET("/platforms", middleware.RequirePermission(enforcer, "platform", "list"), platformHandler.List)
		auth.POST("/platforms", middleware.RequirePermission(enforcer, "platform", "create"), platformHandler.Create)
		auth.PUT("/platforms/:id", middleware.RequirePermission(enforcer, "platform", "update"), platformHandler.Update)
		auth.DELETE("/platforms/:id", middleware.RequirePermission(enforcer, "platform", "delete"), platformHandler.Delete)

		// 员工管理
		auth.GET("/employees", middleware.RequirePermission(enforcer, "employee", "list"), employeeHandler.List)
		auth.GET("/employees/departments", middleware.RequirePermission(enforcer, "employee", "list"), employeeHandler.GetDepartments)
		auth.POST("/employees", middleware.RequirePermission(enforcer, "employee", "create"), employeeHandler.Create)
		auth.PUT("/employees/:id", middleware.RequirePermission(enforcer, "employee", "update"), employeeHandler.Update)
		auth.DELETE("/employees/:id", middleware.RequirePermission(enforcer, "employee", "delete"), employeeHandler.Delete)
		auth.PUT("/employees/:id/reset-password",
			middleware.RequirePermission(enforcer, "employee", "reset-password"), employeeHandler.ResetPassword)

		// 角色管理
		auth.GET("/roles", middleware.RequirePermission(enforcer, "role", "list"), roleHandler.List)
		auth.POST("/roles", middleware.RequirePermission(enforcer, "role", "create"), roleHandler.Create)
		auth.PUT("/roles/:id", middleware.RequirePermission(enforcer, "role", "update"), roleHandler.Update)
		auth.DELETE("/roles/:id", middleware.RequirePermission(enforcer, "role", "delete"), roleHandler.Delete)
		auth.GET("/roles/:id/permissions", middleware.RequirePermission(enforcer, "role", "list"), roleHandler.GetPermissions)
		auth.PUT("/roles/:id/permissions", middleware.RequirePermission(enforcer, "role", "assign"), roleHandler.AssignPermissions)
		auth.POST("/roles/users", middleware.RequirePermission(enforcer, "user", "role:assign"), roleHandler.AssignUsers)

		// 用户-角色
		auth.GET("/users/:id/roles", middleware.RequirePermission(enforcer, "role", "list"), roleHandler.GetUserRoles)
		auth.PUT("/users/:id/roles", middleware.RequirePermission(enforcer, "user", "role:assign"), roleHandler.SetUserRoles)

		// 权限点
		auth.GET("/permissions", middleware.RequirePermission(enforcer, "permission", "list"), permissionHandler.ListTree)
		auth.POST("/permissions", middleware.RequirePermission(enforcer, "permission", "create"), permissionHandler.Create)
	}

	return r
}
