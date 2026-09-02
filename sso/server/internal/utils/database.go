package utils

import (
	"fmt"
	"time"

	"mh-sso-svc/internal/audit"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// InitDatabase 初始化数据库连接并 ping 检查，返回清理函数
func InitDatabase(cfg *MySQLConfig, mode string) (*gorm.DB, func(), error) {
	db, err := InitDB(cfg, mode)
	if err != nil {
		return nil, nil, err
	}

	// ping 检查连接是否正常
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("获取sql.DB失败: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("数据库连接ping失败: %w", err)
	}

	// 返回清理函数，用于释放连接池
	closer := func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	return db, closer, nil
}

// InitDB 初始化数据库连接
func InitDB(cfg *MySQLConfig, mode string) (*gorm.DB, error) {
	logLevel := gormlogger.Silent
	if mode == "debug" {
		logLevel = gormlogger.Info
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	if err := audit.RegisterGORMCallbacks(db); err != nil {
		return nil, fmt.Errorf("注册审计数据库回调失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取sql.DB失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
