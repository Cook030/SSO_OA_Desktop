package mapper_test

import (
	"encoding/json"
	"strings"
	"testing"

	"mh-audit-consumer/internal/canal"
	"mh-audit-consumer/internal/mapper"
	"mh-audit-consumer/internal/sanitize"
)

func newTestMapper() *mapper.Mapper {
	tables := map[string]mapper.TableMapping{
		"sys_user":      {TargetType: "user", Key: "id"},
		"sys_role":      {TargetType: "role", Key: "id"},
		"sys_user_role": {TargetType: "user", Key: "user_id"},
	}
	return mapper.New(tables, sanitize.New([]string{"password", "password_hash"}, "******"))
}

func sp(s string) *string { return &s }

// updateFlat 构造一条 UPDATE sys_user 事件。
// Data 为完整新行; Old 模拟 canal RowChange.beforeColumns(ROW 模式下为变更前完整行,
// 此处仅给变更列以覆盖 mapper 反推 before 的路径)。位点来自 Entry.header。
func updateFlat() *canal.FlatMessage {
	return &canal.FlatMessage{
		Database:        "sso",
		Table:           "sys_user",
		Type:            "UPDATE",
		ES:              1757059200000,
		BinlogFileName:  "mysql-bin.000001",
		BinlogPosition:  500,
		GTID:            "abcdef:1-10",
		Data: []map[string]*string{{
			"id": sp("1"), "account": sp("admin"), "password": sp("hash-new"),
			"name": sp("管理员"), "updated_by": sp("9"), "request_id": sp("req-123"),
			"update_time": sp("2026-09-05 10:00:00"),
		}},
		Old: []map[string]*string{{
			"password": sp("hash-old"), "name": sp("旧名"), "updated_by": sp("5"),
			"request_id": sp("req-000"), "update_time": sp("2026-09-01 10:00:00"),
		}},
	}
}

func TestBuildUpdate(t *testing.T) {
	m := newTestMapper()
	records, err := m.Build(updateFlat())
	if err != nil {
		t.Fatalf("映射失败: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len = %d, want 1", len(records))
	}
	r := records[0]
	if r.Action != "user:update" || r.TargetType != "user" || r.TargetID != "1" {
		t.Fatalf("动作/目标映射错误: %s %s %s", r.Action, r.TargetType, r.TargetID)
	}
	if r.OperatorID == nil || *r.OperatorID != 9 {
		t.Fatalf("操作人应为 updated_by=9, got %v", r.OperatorID)
	}
	if r.RequestID == nil || *r.RequestID != "req-123" {
		t.Fatalf("request_id 映射错误: %v", r.RequestID)
	}
	if len(r.DedupKey) != 64 {
		t.Fatalf("dedup_key 应为 sha256 hex(64), got %d", len(r.DedupKey))
	}

	detail := r.Detail
	if detail == nil || !strings.Contains(*detail, "管理员") || !strings.Contains(*detail, "旧名") {
		t.Fatalf("detail 应含 name 字段前后值: %v", detail)
	}
	// 敏感列脱敏, 原始 hash 不得落库
	if strings.Contains(*detail, "hash-new") || strings.Contains(*detail, "hash-old") {
		t.Fatalf("detail 不应出现明文密码: %v", *detail)
	}
	// updated_by/request_id/update_time 属技术列, 不进 diff
	var parsed struct {
		Changed map[string]any `json:"changed"`
	}
	_ = json.Unmarshal([]byte(*detail), &parsed)
	if _, ok := parsed.Changed["name"]; !ok {
		t.Fatalf("changed 应包含 name: %v", parsed.Changed)
	}
	if _, ok := parsed.Changed["password"]; ok {
		t.Fatalf("password 打码后不应出现在 changed 中: %v", parsed.Changed)
	}
	if _, ok := parsed.Changed["updated_by"]; ok {
		t.Fatalf("updated_by 不应出现在 changed 中: %v", parsed.Changed)
	}

	before := r.BeforeData
	if before == nil || !strings.Contains(*before, "旧名") || strings.Contains(*before, "hash-old") {
		t.Fatalf("before_data 应含脱敏前快照且无明文密码: %v", before)
	}
	if r.EventTime == nil {
		t.Fatalf("应回填 header.executeTime 事件时间")
	}
}

func TestBuildInsertJoinTable(t *testing.T) {
	m := newTestMapper()
	flat := &canal.FlatMessage{
		Database:        "sso",
		Table:           "sys_user_role",
		Type:            "INSERT",
		ES:              1757059200000,
		BinlogFileName:  "mysql-bin.000002",
		BinlogPosition:  10,
		Data: []map[string]*string{{
			"id": sp("88"), "user_id": sp("3"), "role_id": sp("7"),
			"created_by": sp("2"), "request_id": sp("r1"),
		}},
	}
	records, err := m.Build(flat)
	if err != nil {
		t.Fatalf("映射失败: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len = %d, want 1", len(records))
	}
	r := records[0]
	if r.Action != "user:assign-role" {
		t.Fatalf("关联表 INSERT 应映射为 user:assign-role, got %s", r.Action)
	}
	if r.TargetType != "user" || r.TargetID != "3" {
		t.Fatalf("关联表目标应为 user_id, got %s:%s", r.TargetType, r.TargetID)
	}
	if r.OperatorID == nil || *r.OperatorID != 2 {
		t.Fatalf("操作人应为 created_by=2, got %v", r.OperatorID)
	}
}

func TestBuildUnmappedTable(t *testing.T) {
	m := newTestMapper()
	flat := &canal.FlatMessage{
		Database:        "sso",
		Table:           "sys_audit_log",
		Type:            "INSERT",
		ES:              1757059200000,
		BinlogFileName:  "mysql-bin.000003",
		BinlogPosition:  1,
		Data:            []map[string]*string{{"id": sp("1")}},
	}
	records, err := m.Build(flat)
	if err != nil {
		t.Fatalf("不应报错: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("未配置表不应产出记录, got %d", len(records))
	}
}
