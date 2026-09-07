// audit-consumer 入口: 消费 Canal -> Kafka 的 binlog 事件并写入 sys_audit_log。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"mh-audit-consumer/internal/config"
	"mh-audit-consumer/internal/consumer"
	"mh-audit-consumer/internal/mapper"
	"mh-audit-consumer/internal/sanitize"
	"mh-audit-consumer/internal/store"
)

func main() {
	configPath := flag.String("c", "config/config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		newLogger("info").Fatal("加载配置失败", zap.Error(err))
	}
	logger := newLogger(effectiveLogLevel(cfg.Log.Level))
	defer func() { _ = logger.Sync() }()

	st, err := store.New(cfg.MySQL.DSN(), cfg.MySQL.MaxOpenConns, cfg.MySQL.MaxIdleConns)
	if err != nil {
		logger.Fatal("初始化审计库失败", zap.Error(err))
	}
	defer func() { _ = st.Close() }()

	mp := mapper.NewWithSource(cfg.Mapping, sanitize.New(cfg.Sanitize.Fields, cfg.Sanitize.Replacement), cfg.MySQL.SourceID)
	cons := consumer.New(cfg, st, mp, logger)
	defer func() { _ = cons.Close() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("audit-consumer 就绪")
	if err := cons.Run(ctx); err != nil {
		logger.Fatal("audit-consumer 运行异常退出", zap.Error(err))
	}
	logger.Info("audit-consumer 已退出")
}

// newLogger 根据指定级别构建 zap 日志。
func newLogger(level string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := cfg.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	return logger
}

// 环境变量优先于配置文件，便于容器部署时临时调高日志级别。
func effectiveLogLevel(configLevel string) string {
	if envLevel := os.Getenv("AUDIT_LOG_LEVEL"); envLevel != "" {
		return envLevel
	}
	return configLevel
}
