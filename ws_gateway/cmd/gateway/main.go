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

	"mh-ws-gateway/internal/auth"
	"mh-ws-gateway/internal/config"
	"mh-ws-gateway/internal/delivery"
	"mh-ws-gateway/internal/hub"
	"mh-ws-gateway/internal/presence"
	"mh-ws-gateway/internal/protocol"
	"mh-ws-gateway/internal/server"
	"mh-ws-gateway/internal/stream"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
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

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	log := initLogger(cfg, filepath.Dir(cfgPath))
	defer func() { _ = log.Sync() }()

	instanceID := cfg.Server.InstanceID
	if instanceID == "" {
		instanceID, _ = os.Hostname()
	}
	log.Info("配置加载完成",
		zap.String("config", cfgPath),
		zap.String("instance_id", instanceID))

	// Redis：与 SSO 共用同一实例（会话事件流与投递存储都在这里）
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	ctxBg := context.Background()
	if err := rdb.Ping(ctxBg).Err(); err != nil {
		log.Fatal("Redis 连接失败：Gateway 无法消费会话事件", zap.Error(err))
	}
	log.Info("Redis 连接成功", zap.String("addr", cfg.Redis.Addr))

	// 依赖组装：Hub（连接注册表）→ 投递存储 / 在线状态 → 消费者 → HTTP 服务
	h := hub.NewHub()
	store := delivery.NewStore(rdb, cfg.DeliveryTTL())
	ps := presence.NewStore(rdb, cfg.ConnectionTTL())
	authClient := auth.NewClient(
		cfg.SSO.IntrospectURL,
		cfg.SSO.ServiceToken,
		time.Duration(cfg.SSO.TimeoutSecond)*time.Second,
		log,
	)

	consumer := stream.NewConsumer(rdb, stream.Options{
		StreamKey: protocol.SessionEventStreamKey,
		// 每个实例独立消费同一事件流；共享组会把消息分配给错误实例，
		// 使真正持有目标连接的 Gateway 无法即时投递。
		Group:           cfg.Realtime.ConsumerGroup + ":" + instanceID,
		ConsumerID:      instanceID,
		BatchSize:       cfg.Realtime.ConsumerBatchSize,
		Block:           cfg.ConsumerBlock(),
		PendingIdle:     cfg.PendingClaimIdle(),
		PendingInterval: cfg.PendingClaimInterval(),
		TerminateGrace:  cfg.TerminateGrace(),
	}, h, store, log)

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.EnsureGroup(runCtx); err != nil {
		log.Fatal("初始化消费组失败", zap.Error(err))
	}
	go consumer.Run(runCtx)

	srv := server.New(cfg, instanceID, h, authClient, store, ps, log)
	httpSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("WebSocket Gateway 启动",
			zap.String("addr", httpSrv.Addr),
			zap.String("ws_path", cfg.Realtime.WSPath))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("收到关闭信号，正在优雅关闭...")
	// 1. 先停止接收新连接
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout())
	defer shutdownCancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP 服务关闭超时", zap.Error(err))
	}
	// 2. 通知客户端"网关重启，请重连"（客户端可安全重连，与会话失效区分开）
	h.CloseAll(protocol.CloseServerShutdown, "gateway shutting down")
	// 3. 停止消费并释放资源
	cancel()
	_ = rdb.Close()
	log.Info("WebSocket Gateway 已安全退出")
}

// initLogger 初始化日志；相对路径基于配置文件所在目录解析
func initLogger(cfg *config.Config, baseDir string) *zap.Logger {
	level, err := zap.ParseAtomicLevel(cfg.Log.Level)
	if err != nil {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	enc := zapcore.NewJSONEncoder(encCfg)

	core := zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), level)
	if cfg.Log.Path != "" {
		p := cfg.Log.Path
		if !filepath.IsAbs(p) {
			p = filepath.Join(baseDir, p)
		}
		fileSink := zapcore.AddSync(&lumberjack.Logger{
			Filename:   p,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
			MaxAge:     cfg.Log.MaxAge,
			Compress:   cfg.Log.Compress,
		})
		core = zapcore.NewTee(core, zapcore.NewCore(enc, fileSink, level))
	}
	return zap.New(core, zap.AddCaller())
}

// locateConfig 从当前工作目录向上逐级查找配置文件，
// 使服务可在项目内任意子目录（如 cmd/gateway）直接启动。
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
