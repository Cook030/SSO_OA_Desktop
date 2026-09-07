package consumer

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"mh-audit-consumer/internal/config"
)

func TestDeadLetterWritesMetadataWithoutPayload(t *testing.T) {
	dir := t.TempDir()
	d := NewDeadLetter(config.DeadLetterConfig{Dir: dir}, zap.NewNop())
	secret := "password=do-not-persist"
	msg := kafka.Message{Topic: "audit", Partition: 2, Offset: 7, Value: []byte(secret)}
	if err := d.Write(msg, errors.New("decode failed")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("死信文件 = %v, %v", entries, err)
	}
	body, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if strings.Contains(text, secret) {
		t.Fatalf("死信不得包含原始载荷: %s", text)
	}
	if !strings.Contains(text, `"offset":7`) || !strings.Contains(text, `"payload_sha256"`) {
		t.Fatalf("死信缺少定位元数据: %s", text)
	}
}
