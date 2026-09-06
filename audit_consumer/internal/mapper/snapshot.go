package mapper

import (
	"encoding/json"

	"mh-audit-consumer/internal/canal"
)

// ignoredDiffColumns 技术字段不参与前后值对比(仅体现"谁在什么时候改的")。
var ignoredDiffColumns = map[string]struct{}{
	"created_by": {}, "updated_by": {}, "request_id": {}, "create_time": {}, "update_time": {},
}

// snapshot 先生成脱敏副本，再按操作类型生成 detail 与 before_data。
func (m *Mapper) snapshot(flat *canal.FlatMessage, idx int, row map[string]*string) (detail, beforeData *string) {
	switch flat.Type {
	case "INSERT":
		return marshalToString(map[string]any{"after": m.san.Row(row)}), nil
	case "UPDATE":
		before := m.san.Row(rebuildBeforeRow(row, oldColumnsAt(flat, idx)))
		after := m.san.Row(row)
		return marshalToString(map[string]any{"changed": diffChangedColumns(before, after)}), marshalToString(before)
	case "DELETE":
		before := m.san.Row(row)
		return marshalToString(map[string]any{"before": before}), marshalToString(before)
	default:
		return nil, nil
	}
}

// rebuildBeforeRow 以当前行为底，用 Canal 上报的变更前列逐列覆盖，反推完整的变更前整行。
func rebuildBeforeRow(currentRow, oldColumns map[string]*string) map[string]*string {
	before := cloneRow(currentRow)
	for column, value := range oldColumns {
		before[column] = value
	}
	return before
}

// oldColumnsAt 取指定行对应的变更前列集合(ROW 模式下 Canal 可能只上报变更列)。
func oldColumnsAt(flat *canal.FlatMessage, index int) map[string]*string {
	if index >= len(flat.Old) || flat.Old[index] == nil {
		return map[string]*string{}
	}
	return flat.Old[index]
}

// diffChangedColumns 逐列比对前后两行，返回发生变化的列及其 old/new 值(仅存在于 before 的列只给 old)。
func diffChangedColumns(before, after map[string]*string) map[string]map[string]*string {
	changed := make(map[string]map[string]*string)
	for column, newValue := range after {
		if isIgnoredDiffColumn(column) {
			continue
		}
		if oldValue := before[column]; !ptrEqual(oldValue, newValue) {
			changed[column] = map[string]*string{"old": oldValue, "new": newValue}
		}
	}
	for column, oldValue := range before {
		if _, existsAfter := after[column]; !existsAfter && !isIgnoredDiffColumn(column) {
			changed[column] = map[string]*string{"old": oldValue}
		}
	}
	return changed
}

// isIgnoredDiffColumn 判断列是否属于不参与 diff 的技术字段。
func isIgnoredDiffColumn(column string) bool {
	_, ignored := ignoredDiffColumns[column]
	return ignored
}

// marshalToString 将值序列化为 JSON 字符串指针，序列化失败时返回 nil。
func marshalToString(value any) *string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	text := string(encoded)
	return &text
}

// cloneRow 深拷贝整行，避免后续写回 Canal 原始数据。
func cloneRow(source map[string]*string) map[string]*string {
	row := make(map[string]*string, len(source))
	for column, value := range source {
		row[column] = cloneString(value)
	}
	return row
}

// cloneString 拷贝指针指向的字符串，返回独立副本。
func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

// ptrEqual 比较两个字符串指针的内容是否相同(两个 nil 视为相同)。
func ptrEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
