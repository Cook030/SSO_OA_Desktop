package consumer

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// DeadLetter 解析/映射失败消息的落盘兜底, 便于人工核查与补采。
type DeadLetter struct {
	dir string
	log *zap.Logger
	mu  sync.Mutex
}

// NewDeadLetter 构造死信写入器。
func NewDeadLetter(dir string, log *zap.Logger) *DeadLetter {
	return &DeadLetter{dir: dir, log: log}
}

// Write 将失败消息追加到当日死信文件。
func (d *DeadLetter) Write(msg kafka.Message, reason error) {
	if d.dir == "" {
		d.log.Error("丢弃无法解析的Kafka消息(未配置死信目录)",
			zap.Int("partition", msg.Partition), zap.Int64("offset", msg.Offset),
			zap.Error(reason))
		return
	}
	if err := os.MkdirAll(d.dir, 0o755); err != nil {
		d.log.Error("创建死信目录失败", zap.String("dir", d.dir), zap.Error(err))
		return
	}
	line := fmt.Sprintf("%s\ttopic=%s\tpartition=%d\toffset=%d\terror=%s\tvalue=%s\n",
		time.Now().Format(time.RFC3339), msg.Topic, msg.Partition, msg.Offset,
		reason, string(msg.Value))

	d.mu.Lock()
	defer d.mu.Unlock()
	name := filepath.Join(d.dir, time.Now().Format("20060102")+".log")
	f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		d.log.Error("打开死信文件失败", zap.Error(err))
		return
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		d.log.Error("写入死信文件失败", zap.Error(err))
	}
}
