// Package store 负责把审计记录批量写入 sys_audit_log。
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"

	"mh-audit-consumer/internal/mapper"
)

// 幂等依赖 uk_dedup 唯一键: 重放/重复投递的记录被 INSERT IGNORE 静默跳过。
// source=1 固定表示 binlog 通道来源。
const insertStmt = `INSERT IGNORE INTO sys_audit_log
	(operator_id, action, target_type, target_id, detail, before_data, source, dedup_key, request_id, create_time)
	VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, ?)`

// errNoFKReferenced MySQL 外键不存在错误码(operator_id 引用的用户已被删除)。
const errNoFKReferenced = 1452

// Store 审计库写入器。
type Store struct {
	db *sql.DB
}

// New 打开审计库连接并 ping 校验。
func New(dsn string, maxOpen, maxIdle int) (*Store, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开审计库连接失败: %w", err)
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("审计库连接 ping 失败: %w", err)
	}
	return &Store{db: db}, nil
}

// Close 释放连接池。
func (s *Store) Close() error {
	return s.db.Close()
}

// BatchInsert 批量插入审计记录。
// 返回 skipped: 因外键(操作人已被物理删除)被跳过的记录数; 批量语句失败会回退逐条插入。
func (s *Store) BatchInsert(ctx context.Context, records []*mapper.Record) (int, error) {
	args := make([]any, 0, len(records)*9)
	for _, r := range records {
		args = append(args, insertArgs(r)...)
	}
	if _, err := s.db.ExecContext(ctx, insertStmt, args...); err != nil {
		if isFKError(err) {
			return s.insertOneByOne(ctx, records)
		}
		return 0, fmt.Errorf("批量写入审计日志失败: %w", err)
	}
	return 0, nil
}

// insertOneByOne 逐条插入并隔离外键失败行(单条语句原子, 失败即该行跳过)。
func (s *Store) insertOneByOne(ctx context.Context, records []*mapper.Record) (int, error) {
	skipped := 0
	for _, r := range records {
		if _, err := s.db.ExecContext(ctx, insertStmt, insertArgs(r)...); err != nil {
			if isFKError(err) {
				skipped++
				continue
			}
			return skipped, fmt.Errorf("单条写入审计日志失败: %w", err)
		}
	}
	return skipped, nil
}

// insertArgs 组装单条记录参数, 与 insertStmt 的 9 个占位符一一对应。
func insertArgs(r *mapper.Record) []any {
	event := time.Now()
	if r.EventTime != nil {
		event = *r.EventTime
	}
	return []any{
		r.OperatorID,
		r.Action,
		r.TargetType,
		r.TargetID,
		r.Detail,
		r.BeforeData,
		r.DedupKey,
		r.RequestID,
		event,
	}
}

func isFKError(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == errNoFKReferenced
}
