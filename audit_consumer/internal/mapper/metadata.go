package mapper

import (
	"strconv"
	"time"

	"mh-audit-consumer/internal/canal"
)

func pickOperator(row map[string]*string) *uint64 {
	for _, column := range []string{"updated_by", "created_by"} {
		if operator := positiveUint(row[column]); operator != nil {
			return operator
		}
	}
	return nil
}

func positiveUint(value *string) *uint64 {
	if value == nil || *value == "" {
		return nil
	}
	number, err := strconv.ParseUint(*value, 10, 64)
	if err != nil || number == 0 {
		return nil
	}
	return &number
}

// eventTimeOf 优先 Canal 事件时间，其次读取业务时间字段。
func eventTimeOf(flat *canal.FlatMessage, row map[string]*string) *time.Time {
	if flat.ES > 0 {
		timestamp := time.UnixMilli(flat.ES)
		return &timestamp
	}
	return parseRowTime(row[eventTimeColumn(flat.Type)])
}

func eventTimeColumn(eventType string) string {
	if eventType == "INSERT" {
		return "create_time"
	}
	return "update_time"
}

func parseRowTime(value *string) *time.Time {
	if value == nil || *value == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02 15:04:05.999999", time.RFC3339Nano} {
		if timestamp, err := time.ParseInLocation(layout, *value, time.Local); err == nil {
			return &timestamp
		}
	}
	return nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func nonEmptyPtr(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}
	copy := *value
	return &copy
}
