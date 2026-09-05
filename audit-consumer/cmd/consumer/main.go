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

	logger := newLogger()
	defer func() { _ = logger.Sync() }()

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("加载配置失败", zap.Error(err))
	}

	st, err := store.New(cfg.MySQL.DSN(), cfg.MySQL.MaxOpenConns, cfg.MySQL.MaxIdleConns)
	if err != nil {
		logger.Fatal("初始化审计库失败", zap.Error(err))
	}
	defer func() { _ = st.Close() }()

	mp := mapper.New(cfg.Mapping, sanitize.New(cfg.Sanitize.Fields, cfg.Sanitize.Replacement))
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

// newLogger 根据 AUDIT_LOG_LEVEL 构建 zap 日志(默认 info)。
func newLogger() *zap.Logger {
	level := os.Getenv("AUDIT_LOG_LEVEL")
	if level == "" {
		level = "info"
	}
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
