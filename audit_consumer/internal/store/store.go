// Package store 负责把审计记录批量写入 sys_audit_log。
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"mh-audit-consumer/internal/mapper"
)

const (
	insertPrefix = `INSERT INTO sys_audit_log
	(operator_id, action, target_type, target_id, detail, before_data, source, dedup_key, request_id, create_time)
	VALUES `
	rowValues = "(?, ?, ?, ?, ?, ?, 1, ?, ?, ?)"
	//消费者是at-least-once的，即消息至少被处理一次
	//使用dedup_key这一组合字段保证唯一
	//当执行INSERT违反了唯一索引dedup_key，MySQL不会报错，而是转成UPDATE
	//只处理唯一键冲突，更新dedup_key，其他错误正常抛出
	onDupNoop = " ON DUPLICATE KEY UPDATE dedup_key = dedup_key"
)

// argsPerRow 单行占位符个数, 必须与 rowValues 保持一致。
const argsPerRow = 9

// maxRowsPerStmt 单条 INSERT 的最大行数。
const maxRowsPerStmt = 500

// singleRowStmt 单条插入语句, 逐条回退时使用; 固定字符串可命中 database/sql 的 stmt 缓存。
var singleRowStmt = insertPrefix + rowValues + onDupNoop

// 记录级错误码: 这类错误说明"该行本身写不进去", 重试无意义, 只能跳过该行;
// 其余错误(连接中断、死锁 1213、锁等待超时 1205 等)必须整批重试, 不能静默丢数据。
const (
	errDupEntry       = 1062 // 唯一键冲突(兜底; 正常已由 ON DUPLICATE KEY UPDATE 消化)
	errBadNull        = 1048 // 非空列收到 NULL
	errNoFKReferenced = 1452 // 外键不存在: operator_id 引用的用户已被物理删除
	errDataTooLong    = 1406 // 字段超长: detail/before_data 超出列容量
)

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
// 返回 skipped: 因记录级错误(操作人已删除/字段超长等)被跳过的记录数; 批量语句命中记录级错误时回退逐条插入。
func (s *Store) BatchInsert(ctx context.Context, records []*mapper.Record) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}
	skipped := 0
	for start := 0; start < len(records); start += maxRowsPerStmt {
		end := min(start+maxRowsPerStmt, len(records))
		chunk := records[start:end]

		err := s.insertBatch(ctx, chunk)
		if err == nil {
			continue
		}
		if !isRecordLevelError(err) {
			return skipped, fmt.Errorf("批量写入审计日志失败: %w", err)
		}
		n, err := s.insertOneByOne(ctx, chunk)
		skipped += n
		if err != nil {
			return skipped, err
		}
	}
	return skipped, nil
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

// insertOneByOne 逐条插入并隔离问题行(单条语句原子, 记录级错误跳过该行后继续)。
func (s *Store) insertOneByOne(ctx context.Context, records []*mapper.Record) (int, error) {
	skipped := 0
	for _, record := range records {
		wasSkipped, err := s.insertOne(ctx, record)
		if err != nil {
			return skipped, err
		}
		if wasSkipped {
			skipped++
		}
	}
	return skipped, nil
}

// insertOne 返回该记录是否因记录级错误被跳过。
func (s *Store) insertOne(ctx context.Context, record *mapper.Record) (bool, error) {
	_, err := s.db.ExecContext(ctx, singleRowStmt, insertArgs(record)...)
	if err == nil {
		return false, nil
	}
	if isRecordLevelError(err) {
		return true, nil
	}
	return false, fmt.Errorf("单条写入审计日志失败: %w", err)
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

// isRecordLevelError 判断错误是否只影响当前行。
func isRecordLevelError(err error) bool {
	var me *mysql.MySQLError
	if !errors.As(err, &me) {
		return false
	}
	switch me.Number {
	case errDupEntry, errBadNull, errNoFKReferenced, errDataTooLong:
		return true
	default:
		return false
	}
}
