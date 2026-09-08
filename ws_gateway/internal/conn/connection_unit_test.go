package conn

import (
	"strings"
	"testing"
	"time"

	"mh-ws-gateway/internal/protocol"

	"go.uber.org/zap"
)

func TestConnectionSendQueueFull(t *testing.T) {
	c := New(Params{
		SessionID: "s1",
		Opt: Options{
			WriteQueueSize: 2,
			PingInterval:   time.Minute,
			PongWait:       time.Minute,
		},
		Log: zap.NewNop(),
	})
	for i := 0; i < 2; i++ {
		if err := c.Send([]byte("x")); err != nil {
			t.Fatalf("第 %d 次 Send 应成功: %v", i+1, err)
		}
	}
	if err := c.Send([]byte("overflow")); err != protocol.ErrQueueFull {
		t.Fatalf("写队列已满时应返回 protocol.ErrQueueFull，实际: %v", err)
	}
}

func TestConnectionSendRejectedAfterClose(t *testing.T) {
	c := New(Params{
		SessionID: "s1",
		Opt: Options{
			WriteQueueSize: 8,
			PingInterval:   time.Minute,
			PongWait:       time.Minute,
		},
		Log: zap.NewNop(),
	})
	// white-box：直接置为关闭态（避免依赖真实 socket）
	close(c.closed)
	if err := c.Send([]byte("x")); err == nil {
		t.Fatalf("连接关闭后 Send 应返回错误")
	} else if !strings.Contains(err.Error(), "已关闭") {
		t.Fatalf("Send 错误信息不符合预期: %v", err)
	}
}

func TestConnectionDefaultQueueSize(t *testing.T) {
	c := New(Params{SessionID: "s1", Log: zap.NewNop()}) // Opt 零值
	if cap(c.send) != 128 {
		t.Fatalf("零值 WriteQueueSize 应回退为 128，实际 %d", cap(c.send))
	}
}
