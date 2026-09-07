// Package mapper 将 Canal DML 事件映射为 sys_audit_log 记录。
package mapper

import (
	"time"

	"mh-audit-consumer/internal/canal"
	"mh-audit-consumer/internal/sanitize"
)

// TableMapping 单表审计映射(来自配置 mapping 段)。
type TableMapping struct {
	TargetType string            `mapstructure:"target_type"`
	Key        string            `mapstructure:"key"`
	Actions    map[string]string `mapstructure:"actions"`
}

// Record 一条待落库的审计记录。
type Record struct {
	OperatorID *uint64
	Action     string
	TargetType string
	TargetID   string
	Detail     *string // JSON 变更明细
	BeforeData *string // JSON 变更前快照(脱敏后)
	RequestID  *string
	DedupKey   string
	EventTime  *time.Time
}

// Mapper 负责 binlog 行 -> 审计记录 的语义映射。
type Mapper struct {
	tables   map[string]TableMapping
	san      *sanitize.Sanitizer
	sourceID string
}

// New 构造 Mapper。
func New(tables map[string]TableMapping, san *sanitize.Sanitizer) *Mapper {
	return NewWithSource(tables, san, "")
}

// NewWithSource 构造携带稳定数据源标识的 Mapper，用于跨主库切换时保持幂等键隔离。
func NewWithSource(tables map[string]TableMapping, san *sanitize.Sanitizer, sourceID string) *Mapper {
	return &Mapper{tables: tables, san: san, sourceID: sourceID}
}

// Build 将一条 DML 事件映射为 0..n 条记录(多行事件每行一条)。
func (m *Mapper) Build(flat *canal.FlatMessage) ([]*Record, error) {
	if flat == nil || !flat.IsDML() {
		return nil, nil
	}
	tm, ok := m.tables[flat.Table]
	if !ok {
		return nil, nil
	}

	records := make([]*Record, 0, len(flat.Data))
	for idx, row := range flat.Data {
		if record := m.buildRowRecord(flat, tm, idx, row); record != nil {
			records = append(records, record)
		}
	}
	return records, nil
}

// buildRowRecord 负责组装单行变更的公共审计字段与操作快照。
func (m *Mapper) buildRowRecord(flat *canal.FlatMessage, tm TableMapping, idx int, row map[string]*string) *Record {
	if row == nil {
		return nil
	}
	detail, beforeData := m.snapshot(flat, idx, row)
	return &Record{
		OperatorID: pickOperator(row),
		Action:     m.actionName(flat.Table, flat.Type),
		TargetType: tm.TargetType,
		TargetID:   stringValue(row[tm.Key]),
		Detail:     detail,
		BeforeData: beforeData,
		RequestID:  nonEmptyPtr(row["request_id"]),
		DedupKey:   dedupKey(flat, idx, m.sourceID),
		EventTime:  eventTimeOf(flat, row),
	}
}
