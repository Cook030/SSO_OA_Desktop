package consumer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"mh-audit-consumer/internal/config"
)

type fetchStep struct {
	message kafka.Message
	err     error
}

type stubReader struct {
	steps []fetchStep
	index int
}

func (r *stubReader) FetchMessage(context.Context) (kafka.Message, error) {
	step := r.steps[r.index]
	r.index++
	return step.message, step.err
}

func (*stubReader) CommitMessages(context.Context, ...kafka.Message) error { return nil }
func (*stubReader) Close() error                                           { return nil }

func TestFetchBatchRetainsMessagesWhenSubsequentFetchFails(t *testing.T) {
	first := kafka.Message{Topic: "audit", Partition: 1, Offset: 42, Value: []byte("first")}
	reader := &stubReader{steps: []fetchStep{
		{message: first},
		{err: errors.New("broker unavailable")},
	}}
	cfg := &config.Config{Kafka: config.KafkaConfig{BatchSize: 2, FlushIntervalMs: 100}}
	c := newConsumer(cfg, reader, nil, nil, zap.NewNop())

	batch, err := c.fetchBatch(context.Background())
	if err == nil {
		t.Fatal("后续拉取失败应返回错误")
	}
	if len(batch) != 1 || batch[0].Offset != first.Offset {
		t.Fatalf("已获取的消息必须保留，got %#v", batch)
	}
}

func TestNextBackoffCapsAtConfiguredMaximum(t *testing.T) {
	if got := nextBackoff(20*time.Second, 30_000); got != 30*time.Second {
		t.Fatalf("backoff = %s, want 30s", got)
	}
}
