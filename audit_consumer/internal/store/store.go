// Package store 负责把审计记录批量写入 sys_audit_log。
package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"mh-audit-consumer/internal/mapper"
)

const (
	insertPrefix = `INSERT INTO sys_audit_log
	(operator_id, action, target_type, target_id, detail, before_data, source, dedup_key, request_id, create_time)
	VALUES `
	rowValues = "(?, ?, ?, ?, ?, ?, 1, ?, ?, ?)"
	// 消费者是at-least-once的即消息至少被处理一次
	// 只处理唯一键冲突，更新dedup_key，其他错误正常抛出
	// 使用 dedup_key 唯一索引保证消息重放时幂等。
	// 当执行INSERT违反了唯一索引dedup_key时，MySQL不会报错，而是转成UPDATE
	onDupNoop = " ON DUPLICATE KEY UPDATE dedup_key = dedup_key"
)

// argsPerRow 单行占位符个数, 必须与 rowValues 保持一致。
const argsPerRow = 9

// maxRowsPerStmt 单条 INSERT 的最大行数。
const maxRowsPerStmt = 500

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

// BatchInsert 批量插入审计记录。任一写入错误都返回上层并阻止提交 offset，
// 以避免审计记录被静默丢弃；重试产生的重复记录由 dedup_key 吸收。
func (s *Store) BatchInsert(ctx context.Context, records []*mapper.Record) error {
	if len(records) == 0 {
		return nil
	}
	for start := 0; start < len(records); start += maxRowsPerStmt {
		end := min(start+maxRowsPerStmt, len(records))
		chunk := records[start:end]

		if err := s.insertBatch(ctx, chunk); err != nil {
			return fmt.Errorf("批量写入审计日志失败: %w", err)
		}
	}
	return nil
}

func (s *Store) insertBatch(ctx context.Context, records []*mapper.Record) error {
	_, err := s.db.ExecContext(ctx, buildInsertSQL(len(records)), batchArgs(records)...)
	return err
}

// buildInsertSQL 按行数动态拼接多组 VALUES, 保证占位符数与 batchArgs 产出的参数数一致。
func buildInsertSQL(rows int) string {
	if rows <= 0 {
		return ""
	}
	var b strings.Builder
	b.Grow(len(insertPrefix) + rows*(len(rowValues)+1) + len(onDupNoop))
	b.WriteString(insertPrefix)
	b.WriteString(rowValues)
	if rows > 1 {
		b.WriteString(strings.Repeat(","+rowValues, rows-1))
	}
	b.WriteString(onDupNoop)
	return b.String()
}

func batchArgs(records []*mapper.Record) []any {
	args := make([]any, 0, len(records)*argsPerRow)
	for _, record := range records {
		args = append(args, insertArgs(record)...)
	}
	return args
}

// insertArgs 组装单条记录参数, 与 rowValues 的 9 个占位符一一对应。
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
