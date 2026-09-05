// Package mapper 将 Canal DML 事件映射为 sys_audit_log 记录。
package mapper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"mh-audit-consumer/internal/canal"
	"mh-audit-consumer/internal/sanitize"
)

// TableMapping 单表审计映射(来自配置 mapping 段)。
type TableMapping struct {
	TargetType string `mapstructure:"target_type"`
	Key        string `mapstructure:"key"`
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
	tables map[string]TableMapping
	san    *sanitize.Sanitizer
}

// New 构造 Mapper。
func New(tables map[string]TableMapping, san *sanitize.Sanitizer) *Mapper {
	return &Mapper{tables: tables, san: san}
}

// 关联表使用更贴合业务的动作语义。
var joinActions = map[string]map[string]string{
	"sys_user_role": {
		"INSERT": "user:assign-role",
		"UPDATE": "user:update",
		"DELETE": "user:unassign-role",
	},
	"sys_role_permission": {
		"INSERT": "role:assign-permission",
		"UPDATE": "role:update",
		"DELETE": "role:unassign-permission",
	},
	"sys_user_platform": {
		"INSERT": "user:assign-platform",
		"UPDATE": "user:update",
		"DELETE": "user:unassign-platform",
	},
}

// 不进入 detail 变更 diff 的技术列(每次写都会变化, 无业务含义)。
var ignoreDiffKeys = map[string]struct{}{
	"created_by": {}, "updated_by": {}, "request_id": {},
	"create_time": {}, "update_time": {},
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
		if row == nil {
			continue
		}
		// 先基于原始行构建快照, 再做脱敏(深拷贝), 避免污染复用。
		var detail *string
		var beforeData *string
		switch flat.Type {
		case "INSERT":
			detail = jsonPointer(map[string]any{"after": m.san.Row(row)})
		case "UPDATE":
			before := m.san.Row(mergeBefore(row, oldRow(flat, idx)))
			after := m.san.Row(row)
			detail = jsonPointer(map[string]any{"changed": changedDiff(before, after)})
			beforeData = jsonPointer(before)
		case "DELETE":
			before := m.san.Row(row)
			detail = jsonPointer(map[string]any{"before": before})
			beforeData = jsonPointer(before)
		}

		records = append(records, &Record{
			OperatorID: pickOperator(row),
			Action:     m.actionName(flat.Table, flat.Type),
			TargetType: tm.TargetType,
			TargetID:   stringValue(row[tm.Key]),
			Detail:     detail,
			BeforeData: beforeData,
			RequestID:  nonEmptyPtr(row["request_id"]),
			DedupKey:   dedupKey(flat, idx),
			EventTime:  eventTimeOf(flat, row),
		})
	}
	return records, nil
}

func (m *Mapper) actionName(table, typ string) string {
	if acts, ok := joinActions[table]; ok {
		if a, ok := acts[typ]; ok {
			return a
		}
	}
	prefix := m.tables[table].TargetType
	switch typ {
	case "INSERT":
		return prefix + ":create"
	case "UPDATE":
		return prefix + ":update"
	case "DELETE":
		return prefix + ":delete"
	}
	return prefix + ":unknown"
}

// mergeBefore 依据 binlog_row_image=FULL 的 old(仅变更列)反推出变更前完整行。
func mergeBefore(after, old map[string]*string) map[string]*string {
	out := cloneRow(after)
	for k, v := range old {
		out[k] = v
	}
	return out
}

func oldRow(flat *canal.FlatMessage, idx int) map[string]*string {
	if flat.Old == nil || idx >= len(flat.Old) {
		return map[string]*string{}
	}
	if flat.Old[idx] == nil {
		return map[string]*string{}
	}
	return flat.Old[idx]
}

// changedDiff 计算 before/after 的字段级差异。
func changedDiff(before, after map[string]*string) map[string]map[string]*string {
	changed := make(map[string]map[string]*string)
	seen := make(map[string]struct{}, len(after))
	for col, newVal := range after {
		seen[col] = struct{}{}
		if _, skip := ignoreDiffKeys[col]; skip {
			continue
		}
		oldVal := before[col]
		if !ptrEqual(oldVal, newVal) {
			changed[col] = map[string]*string{"old": oldVal, "new": newVal}
		}
	}
	for col, oldVal := range before {
		if _, ok := seen[col]; ok {
			continue
		}
		if _, skip := ignoreDiffKeys[col]; skip {
			continue
		}
		changed[col] = map[string]*string{"old": oldVal}
	}
	return changed
}

// dedupKey 由 binlog 位点 + 行号生成, 同一条 binlog 记录重放时幂等。
// binlogFileName/binlogPosition 来自 CanalEntry.Entry.header(flatMessage=false 后真实填充);
// 仅在位点缺失(异常报文)时退回 gtid。
func dedupKey(flat *canal.FlatMessage, rowIdx int) string {
	pos := flat.BinlogFileName
	if pos == "" && flat.GTID != "" {
		pos = "gtid:" + flat.GTID
	}
	raw := fmt.Sprintf("%s|%s|%s|%s|%d|%d",
		flat.Database, flat.Table, flat.Type, pos, flat.BinlogPosition, rowIdx)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// pickOperator 取行内打点的操作人: UPDATE/DELETE 取 updated_by, INSERT 可回退 created_by。
func pickOperator(row map[string]*string) *uint64 {
	for _, col := range []string{"updated_by", "created_by"} {
		if v := row[col]; v != nil && *v != "" {
			if n, err := strconv.ParseUint(*v, 10, 64); err == nil && n > 0 {
				return &n
			}
		}
	}
	return nil
}

// eventTimeOf 优先 Canal es(事件时间), 其次业务行 update/create_time, 缺省由落库时补 now。
func eventTimeOf(flat *canal.FlatMessage, row map[string]*string) *time.Time {
	if flat.ES > 0 {
		t := time.UnixMilli(flat.ES)
		return &t
	}
	key := "update_time"
	if flat.Type == "INSERT" {
		key = "create_time"
	}
	if v := stringValue(row[key]); v != "" {
		for _, layout := range []string{
			"2006-01-02 15:04:05",
			"2006-01-02 15:04:05.999999",
			time.RFC3339Nano,
		} {
			if t, err := time.ParseInLocation(layout, v, time.Local); err == nil {
				return &t
			}
		}
	}
	return nil
}

func jsonPointer(v any) *string {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}

func cloneRow(src map[string]*string) map[string]*string {
	out := make(map[string]*string, len(src))
	for k, v := range src {
		out[k] = cloneString(v)
	}
	return out
}

func cloneString(v *string) *string {
	if v == nil {
		return nil
	}
	s := *v
	return &s
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func nonEmptyPtr(v *string) *string {
	if v == nil || *v == "" {
		return nil
	}
	s := *v
	return &s
}

func ptrEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
