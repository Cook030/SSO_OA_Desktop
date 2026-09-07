// Package config 负责加载 audit-consumer 配置(YAML, 支持 ${ENV} 占位)。
package config

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"

	"mh-audit-consumer/internal/mapper"
)

// Config 顶层配置。
type Config struct {
	Kafka      KafkaConfig
	MySQL      MySQLConfig
	Sanitize   SanitizeConfig
	Mapping    map[string]mapper.TableMapping
	Log        LogConfig
	DeadLetter DeadLetterConfig `mapstructure:"dead_letter"`
}

// KafkaConfig Kafka 消费组配置。
type KafkaConfig struct {
	Brokers           []string
	Topic             string
	GroupID           string `mapstructure:"group_id"`
	BatchSize         int    `mapstructure:"batch_size"`
	FlushIntervalMs   int    `mapstructure:"flush_interval_ms"`
	FetchMaxWaitMs    int    `mapstructure:"fetch_max_wait_ms"`
	SessionTimeoutMs  int    `mapstructure:"session_timeout_ms"`
	RetryMaxBackoffMs int    `mapstructure:"retry_max_backoff_ms"`
}

// FlushInterval 攒批时间上限。
func (k KafkaConfig) FlushInterval() time.Duration {
	return time.Duration(k.FlushIntervalMs) * time.Millisecond
}

// MySQLConfig 审计库(sys_audit_log)连接配置。
type MySQLConfig struct {
	SourceID     string `mapstructure:"source_id"`
	Host         string
	Port         int
	Username     string
	Password     string
	Database     string
	Charset      string
	MaxOpenConns int `mapstructure:"max_open_conns"`
	MaxIdleConns int `mapstructure:"max_idle_conns"`
}

// DSN 拼接 go-sql-driver DSN, 本地时区与库内 DATETIME 保持一致。
func (m MySQLConfig) DSN() string {
	charset := m.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		m.Username, m.Password, m.Host, m.Port, m.Database, charset)
}

// SanitizeConfig 脱敏规则。
type SanitizeConfig struct {
	Fields      []string
	Replacement string
}

// LogConfig 日志级别。
type LogConfig struct {
	Level string
}

// DeadLetterConfig 死信落盘配置。
// 两个上限均为: 0 或未配置 = 使用默认值, 负数 = 不限制。
type DeadLetterConfig struct {
	Dir       string `mapstructure:"dir"`
	MaxFiles  int    `mapstructure:"max_files"`
	MaxSizeMb int    `mapstructure:"max_size_mb"`
}

// Load 读取并解析配置文件。
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败(%s): %w", path, err)
	}

	v := viper.New()
	v.SetConfigType("yaml")
	content := os.ExpandEnv(string(raw))
	if err := v.ReadConfig(bytes.NewReader([]byte(content))); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	cfg := new(Config)
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Kafka.BatchSize <= 0 {
		c.Kafka.BatchSize = 500
	}
	if c.Kafka.FlushIntervalMs <= 0 {
		c.Kafka.FlushIntervalMs = 1000
	}
	if c.Kafka.FetchMaxWaitMs <= 0 {
		c.Kafka.FetchMaxWaitMs = 500
	}
	if c.Kafka.SessionTimeoutMs <= 0 {
		c.Kafka.SessionTimeoutMs = 10000
	}
	if c.Kafka.RetryMaxBackoffMs <= 0 {
		c.Kafka.RetryMaxBackoffMs = 30000
	}
	if c.MySQL.Port <= 0 {
		c.MySQL.Port = 3306
	}
	if c.MySQL.Charset == "" {
		c.MySQL.Charset = "utf8mb4"
	}
	if c.MySQL.MaxOpenConns <= 0 {
		c.MySQL.MaxOpenConns = 20
	}
	if c.MySQL.MaxIdleConns <= 0 {
		c.MySQL.MaxIdleConns = 5
	}
	if c.Sanitize.Replacement == "" {
		c.Sanitize.Replacement = "******"
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.DeadLetter.Dir == "" {
		c.DeadLetter.Dir = "dead_letter"
	}
	if c.DeadLetter.MaxFiles == 0 {
		c.DeadLetter.MaxFiles = 30
	}
	if c.DeadLetter.MaxSizeMb == 0 {
		c.DeadLetter.MaxSizeMb = 512
	}
}

func (c *Config) validate() error {
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("配置缺失: kafka.brokers 不能为空")
	}
	if c.Kafka.Topic == "" {
		return fmt.Errorf("配置缺失: kafka.topic 不能为空")
	}
	if c.Kafka.GroupID == "" {
		return fmt.Errorf("配置缺失: kafka.group_id 不能为空")
	}
	if c.MySQL.Host == "" {
		return fmt.Errorf("配置缺失: mysql.host 不能为空")
	}
	if c.MySQL.SourceID == "" {
		return fmt.Errorf("配置缺失: mysql.source_id 不能为空")
	}
	if c.MySQL.Database == "" {
		return fmt.Errorf("配置缺失: mysql.database 不能为空")
	}
	if len(c.Mapping) == 0 {
		return fmt.Errorf("配置缺失: mapping 未配置任何审计表")
	}
	for table, mapping := range c.Mapping {
		if mapping.TargetType == "" || mapping.Key == "" {
			return fmt.Errorf("配置错误: mapping.%s 的 target_type 与 key 不能为空", table)
		}
	}
	return nil
}
