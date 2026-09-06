package mapper

import (
	"encoding/json"

	"mh-audit-consumer/internal/canal"
)

var ignoreDiffKeys = map[string]struct{}{
	"created_by": {}, "updated_by": {}, "request_id": {}, "create_time": {}, "update_time": {},
}

// snapshot 先生成脱敏副本，再按操作类型生成 detail 与 before_data。
func (m *Mapper) snapshot(flat *canal.FlatMessage, idx int, row map[string]*string) (detail, beforeData *string) {
	switch flat.Type {
	case "INSERT":
		return jsonPointer(map[string]any{"after": m.san.Row(row)}), nil
	case "UPDATE":
		before := m.san.Row(mergeBefore(row, oldRow(flat, idx)))
		after := m.san.Row(row)
		return jsonPointer(map[string]any{"changed": changedDiff(before, after)}), jsonPointer(before)
	case "DELETE":
		before := m.san.Row(row)
		return jsonPointer(map[string]any{"before": before}), jsonPointer(before)
	default:
		return nil, nil
	}
}

func mergeBefore(after, old map[string]*string) map[string]*string {
	before := cloneRow(after)
	for column, value := range old {
		before[column] = value
	}
	return before
}

func oldRow(flat *canal.FlatMessage, index int) map[string]*string {
	if index >= len(flat.Old) || flat.Old[index] == nil {
		return map[string]*string{}
	}
	return flat.Old[index]
}

func changedDiff(before, after map[string]*string) map[string]map[string]*string {
	changed := make(map[string]map[string]*string)
	for column, newValue := range after {
		if shouldIgnoreDiff(column) {
			continue
		}
		if oldValue := before[column]; !ptrEqual(oldValue, newValue) {
			changed[column] = map[string]*string{"old": oldValue, "new": newValue}
		}
	}
	for column, oldValue := range before {
		if _, existsAfter := after[column]; !existsAfter && !shouldIgnoreDiff(column) {
			changed[column] = map[string]*string{"old": oldValue}
		}
	}
	return changed
}

func shouldIgnoreDiff(column string) bool {
	_, ignored := ignoreDiffKeys[column]
	return ignored
}

func jsonPointer(value any) *string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	json := string(encoded)
	return &json
}

func cloneRow(source map[string]*string) map[string]*string {
	copy := make(map[string]*string, len(source))
	for column, value := range source {
		copy[column] = cloneString(value)
	}
	return copy
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func ptrEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
