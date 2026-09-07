package consumer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

// Write 将失败消息的可定位元数据追加到当日死信文件, 随后清理超出上限的最旧文件。
// 不落盘原始载荷，避免 Canal 消息中的敏感列绕过脱敏规则。
func (d *DeadLetter) Write(msg kafka.Message, reason error) error {
	if d.dir == "" {
		return errors.New("未配置死信目录")
	}
	if err := os.MkdirAll(d.dir, 0o755); err != nil {
		return fmt.Errorf("创建死信目录(%s)失败: %w", d.dir, err)
	}
	day := time.Now()
	name := filepath.Join(d.dir, day.Format("20060102")+deadLetterExt)

	d.mu.Lock()
	defer d.mu.Unlock()

	f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("打开死信文件(%s)失败: %w", name, err)
	}
	sum := sha256.Sum256(msg.Value)
	line, err := json.Marshal(deadLetterEntry{
		Time:          day.Format(time.RFC3339),
		Topic:         msg.Topic,
		Partition:     msg.Partition,
		Offset:        msg.Offset,
		Reason:        reason.Error(),
		PayloadSHA256: hex.EncodeToString(sum[:]),
		PayloadBytes:  len(msg.Value),
	})
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("编码死信记录失败: %w", err)
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return fmt.Errorf("写入死信文件(%s)失败: %w", name, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("关闭死信文件(%s)失败: %w", name, err)
	}
	d.trim(day)
	return nil
}

type deadLetterEntry struct {
	Time          string `json:"time"`
	Topic         string `json:"topic"`
	Partition     int    `json:"partition"`
	Offset        int64  `json:"offset"`
	Reason        string `json:"reason"`
	PayloadSHA256 string `json:"payload_sha256"`
	PayloadBytes  int    `json:"payload_bytes"`
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
