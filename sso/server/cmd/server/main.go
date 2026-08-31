package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mh-sso-svc/internal/cache"
	"mh-sso-svc/internal/router"
	"mh-sso-svc/internal/utils"

	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置（运行时工作目录即项目根，支持 ${ENV} 环境变量占位符）
	cfg, err := utils.LoadConfig("config/config.yaml")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	utils.InitLogger(cfg.Server.Mode, &cfg.Log)
	log := utils.GetLogger()

	log.Info("配置加载完成")

	// 3. 校验关键安全配置，缺失时拒绝启动
	validateAuthConfig(&cfg.Auth, log)

	// 4. 初始化数据库连接
	db, closeDB, err := utils.InitDatabase(&cfg.MySQL, cfg.Server.Mode)
	if err != nil {
		log.Fatal("初始化数据库连接失败", zap.Error(err))
	}
	log.Info("数据库连接成功")

	// 5. 初始化 Redis（连接失败仅告警，缓存/限流/锁降级运行，核心认证走 MySQL）
	rdb := cache.NewCache(&cfg.Redis, log)

	// 6. 设置路由（Repository/Service/Handler 在路由层内部组装）
	r := router.SetupRouter(db, rdb, cfg)

	// 7. 启动 HTTP 服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Info("SSO 服务器启动", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 8. 监听关闭信号，优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("收到关闭信号，正在优雅关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("服务器关闭超时，强制退出", zap.Error(err))
	}
	_ = rdb.Close()
	closeDB()
	log.Info("服务器已安全退出")
}

// validateAuthConfig 校验认证相关关键配置，缺失时拒绝启动
func validateAuthConfig(cfg *utils.AuthConfig, log *zap.Logger) {
	if cfg.JWTSecret == "" {
		log.Fatal("auth.jwt_secret 未配置：请设置环境变量 JWT_SECRET 或在配置文件中填写")
	}
	if len(cfg.JWTSecret) < 32 {
		log.Warn("auth.jwt_secret 长度不足 32 位，生产环境建议使用更强的随机密钥")
	}
	if cfg.InternalToken == "" {
		log.Fatal("auth.internal_token 未配置：请设置环境变量 SSO_INTERNAL_TOKEN 或在配置文件中填写")
	}
}
