package main

import "testing"

func TestParsePermissionIntent(t *testing.T) {
	intent, ok := parsePermissionIntent("给张三分配 A 平台权限")
	if !ok {
		t.Fatal("expected permission intent")
	}
	if intent.EmployeeName != "张三" || intent.PlatformName != "A" {
		t.Fatalf("unexpected intent: %#v", intent)
	}
}

func TestParsePermissionIntentRejectsUnsupportedMessage(t *testing.T) {
	if _, ok := parsePermissionIntent("查询张三有哪些平台权限"); ok {
		t.Fatal("unsupported intent should not match")
	}
}
