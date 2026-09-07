package store

import (
	"strings"
	"testing"
)

func TestBuildInsertSQL(t *testing.T) {
	sql := buildInsertSQL(2)
	if got := strings.Count(sql, "?"); got != 2*argsPerRow {
		t.Fatalf("占位符数 = %d, want %d", got, 2*argsPerRow)
	}
	if !strings.Contains(sql, onDupNoop) {
		t.Fatalf("SQL 必须包含幂等处理: %s", sql)
	}
}
