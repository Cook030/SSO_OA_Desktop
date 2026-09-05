package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"permission-system/internal/db_model/query"
	"permission-system/internal/router"
	"permission-system/internal/service/bootstrap"
	"permission-system/internal/utils"

	_ "permission-system/docs"

	"go.uber.org/zap"
)

//	@title			权限管理系统 API
//	@version		1.0
//	@description	这是一个基于 Go + Gin 的权限管理系统 API，支持用户认证、平台管理、员工管理和权限管理等功能。
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/api
//	@schemes	http

//	@securityDefinitions.apikey	CookieAuth
//	@in							cookie
//	@name						mh_sso2_access_token
//	@description				SSO Access Token Cookie 认证（后端仅认 Cookie，不再解析 Authorization Header）

func main() {
	// 1. 加载配置（运行时工作目录即项目根）
	configPath := "config/config.yaml"
	cfg, err := utils.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	utils.InitLogger(cfg.Server.Mode, &cfg.Log)
	log := utils.GetLogger()

	log.Info("配置加载完成")

	// 3. 初始化数据库连接
	db, closeDB, err := utils.InitDatabase(&cfg.MySQL, cfg.Server.Mode)
	if err != nil {
		log.Fatal("初始化数据库连接失败", zap.Error(err))
	}
	log.Info("数据库连接成功")

	// 4. 初始化 RBAC 基础数据（内置角色、权限点、管理员账号）
	if err := bootstrap.InitRBAC(query.Use(db), &cfg.Admin); err != nil {
		log.Fatal("初始化 RBAC 数据失败", zap.Error(err))
	}

	// 5. 设置路由（Repository/Service/Handler 在路由层内部组装）
	r := router.SetupRouter(db, cfg)

	// 6. 启动 HTTP 服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Info("服务器启动", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 7. 监听关闭信号，优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("收到关闭信号，正在优雅关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("服务器关闭超时，强制退出", zap.Error(err))
	}
	closeDB()
	log.Info("服务器已安全退出")
}
