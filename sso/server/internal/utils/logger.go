package utils

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var zapLogger *zap.Logger

// InitLogger 初始化 Zap 日志，同时输出到控制台和本地文件
func InitLogger(mode string, logCfg *LogConfig) {
	if logCfg != nil && logCfg.Path != "" {
		if err := os.MkdirAll(filepath.Dir(logCfg.Path), 0755); err != nil {
			panic("创建日志目录失败: " + err.Error())
		}
	}

	var encoderCfg zapcore.EncoderConfig
	var level zapcore.Level
	if mode == "debug" {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		level = zapcore.DebugLevel
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
		level = zapcore.InfoLevel
	}
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)

	cores := []zapcore.Core{
		zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(os.Stdout),
			level,
		),
	}

	if logCfg != nil && logCfg.Path != "" {
		fileEncoder := zapcore.NewJSONEncoder(encoderCfg)
		ws := zapcore.AddSync(&lumberjack.Logger{
			Filename:   logCfg.Path,
			MaxSize:    defaultIfZero(logCfg.MaxSize, 100),
			MaxBackups: defaultIfZero(logCfg.MaxBackups, 10),
			MaxAge:     defaultIfZero(logCfg.MaxAge, 30),
			Compress:   logCfg.Compress,
			LocalTime:  logCfg.LocalTime,
		})
		cores = append(cores, zapcore.NewCore(
			fileEncoder,
			ws,
			level,
		))
	}

	zapLogger = zap.New(zapcore.NewTee(cores...), zap.AddCaller())
}

func defaultIfZero(val, def int) int {
	if val <= 0 {
		return def
	}
	return val
}

// GetLogger 获取日志实例；未初始化时返回 Nop logger（静默丢弃），避免测试环境 panic
func GetLogger() *zap.Logger {
	if zapLogger == nil {
		return zap.NewNop()
	}
	return zapLogger
}
