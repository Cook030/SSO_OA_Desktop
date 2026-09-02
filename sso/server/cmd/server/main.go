package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"mh-sso-svc/internal/cache"
	"mh-sso-svc/internal/router"
	"mh-sso-svc/internal/utils"

	"go.uber.org/zap"
)

func main() {
	// 0. 定位配置文件：优先 -config 参数，其次环境变量 CONFIG_FILE，
	//    否则从当前工作目录逐级向上自动查找 config/config.yaml，
	//    因此可以在项目内任意子目录（如 cmd/server）直接 go run main.go 启动。
	configFlag := flag.String("config", "", "配置文件路径（默认自动向上查找 config/config.yaml）")
	flag.Parse()

	cfgPath := *configFlag
	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG_FILE")
	}
	if cfgPath == "" {
		var err error
		cfgPath, err = locateConfig("config/config.yaml")
		if err != nil {
			fmt.Printf("加载配置失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 1. 加载配置（支持 ${ENV} 环境变量占位符）
	cfg, err := utils.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志；相对日志路径基于配置文件所在目录解析，保证从任意目录启动时日志落点一致
	if cfg.Log.Path != "" && !filepath.IsAbs(cfg.Log.Path) {
		cfg.Log.Path = filepath.Join(filepath.Dir(cfgPath), cfg.Log.Path)
	}
	utils.InitLogger(cfg.Server.Mode, &cfg.Log)
	log := utils.GetLogger()

	log.Info("配置加载完成", zap.String("config", cfgPath))

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

// locateConfig 从当前工作目录向上逐级查找目标配置文件，
// 使服务可在项目内任意子目录（如 cmd/server）直接启动。
func locateConfig(name string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取工作目录失败: %w", err)
	}
	start := dir
	for {
		candidate := filepath.Join(dir, name)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("在 %s 及其上级目录中未找到 %s；请用 -config 指定或设置环境变量 CONFIG_FILE", start, name)
}

// validateAuthConfig 校验认证相关关键配置，缺失时拒绝启动
func validateAuthConfig(cfg *utils.AuthConfig, log *zap.Logger) {
	if cfg.JWTSecret == "" {
		log.Fatal("auth.jwt_secret 未配置：请设置环境变量 JWT_SECRET 或在配置文件中填写")
	}
	if len(cfg.JWTSecret) < 32 {
		log.Warn("auth.jwt_secret 长度不足 32 位，生产环境建议使用更强的随机密钥")
	}
}
