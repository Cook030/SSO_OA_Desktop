package consumer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"mh-audit-consumer/internal/config"
)

const deadLetterExt = ".log"

// DeadLetter 解析/映射失败消息的落盘兜底, 便于人工核查与补采。
// 死信按天各写一个文件; maxFiles(文件数) / maxBytes(总大小) 超过上限时,
// 自动删除最旧的死信文件, 防止坏消息洪水把磁盘打满(<=0 表示不限制)。
type DeadLetter struct {
	dir      string
	maxFiles int   // 目录内最多保留的文件数, <=0 不限制
	maxBytes int64 // 目录内文件总字节上限, <=0 不限制
	log      *zap.Logger

	mu sync.Mutex
}

// NewDeadLetter 构造死信写入器。
func NewDeadLetter(cfg config.DeadLetterConfig, log *zap.Logger) *DeadLetter {
	return &DeadLetter{
		dir:      cfg.Dir,
		maxFiles: cfg.MaxFiles,
		maxBytes: int64(cfg.MaxSizeMb) << 20,
		log:      log,
	}
}

// Write 将失败消息追加到当日死信文件, 随后清理超出上限的最旧文件。
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
	day := time.Now()
	name := filepath.Join(d.dir, day.Format("20060102")+deadLetterExt)

	d.mu.Lock()
	defer d.mu.Unlock()

	f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		d.log.Error("打开死信文件失败", zap.String("file", name), zap.Error(err))
		return
	}
	defer f.Close()
	line := fmt.Sprintf("%s\ttopic=%s\tpartition=%d\toffset=%d\terror=%s\tvalue=%s\n",
		day.Format(time.RFC3339), msg.Topic, msg.Partition, msg.Offset,
		reason, string(msg.Value))
	if _, err := f.WriteString(line); err != nil {
		d.log.Error("写入死信文件失败", zap.String("file", name), zap.Error(err))
		return
	}
	d.trim(day)
}

// trim 从最旧的死信文件开始删除, 直到文件数与总大小都不超上限。
// 当日正在写入的文件始终保留(即使它单独就超限, 否则会边写边删丢数据)。
func (d *DeadLetter) trim(day time.Time) {
	if d.maxFiles <= 0 && d.maxBytes <= 0 {
		return
	}
	files := d.list()
	if len(files) == 0 {
		return
	}
	keep := filepath.Join(d.dir, day.Format("20060102")+deadLetterExt)

	total := int64(0)
	for _, f := range files {
		total += f.size
	}
	for i, f := range files { // 按文件名排序, 即按日期从旧到新
		if (d.maxFiles <= 0 || len(files)-i <= d.maxFiles) &&
			(d.maxBytes <= 0 || total <= d.maxBytes) {
			break
		}
		if f.name == keep {
			continue
		}
		if err := os.Remove(f.name); err != nil && !errors.Is(err, os.ErrNotExist) {
			d.log.Error("删除超限死信文件失败", zap.String("file", f.name), zap.Error(err))
			continue
		}
		total -= f.size
	}
}

// list 返回目录内全部死信文件(按名称升序), 删除时从最旧开始。
func (d *DeadLetter) list() []deadFile {
	entries, err := os.ReadDir(d.dir)
	if err != nil {
		d.log.Error("扫描死信目录失败", zap.String("dir", d.dir), zap.Error(err))
		return nil
	}
	files := make([]deadFile, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), deadLetterExt) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, deadFile{name: filepath.Join(d.dir, e.Name()), size: info.Size()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	return files
}

// deadFile 目录内一个死信文件。
type deadFile struct {
	name string
	size int64
}
