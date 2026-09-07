package config

import (
	"path/filepath"
	"testing"
)

func TestExampleConfigLoadsWithSourceAndActions(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "config", "config.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MySQL.SourceID == "" {
		t.Fatal("mysql.source_id 不能为空")
	}
	if got := cfg.Mapping["sys_user_role"].Actions["insert"]; got != "user:assign-role" {
		t.Fatalf("关联表 action 配置 = %q", got)
	}
}
