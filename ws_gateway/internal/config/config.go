// Package config 加载 WebSocket Gateway 的运行配置。
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ServerConfig 服务配置
type ServerConfig struct {
	Port                  int    `mapstructure:"port"`
	Mode                  string `mapstructure:"mode"`
	InstanceID            string `mapstructure:"instance_id"`
	ShutdownTimeoutSecond int    `mapstructure:"shutdown_timeout_second"`
}

// RedisConfig Redis 配置（与 SSO 共用同一实例）
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// SSOConfig 调用 SSO 内部鉴权接口的配置
type SSOConfig struct {
	IntrospectURL string `mapstructure:"introspect_url"`
	ServiceToken  string `mapstructure:"service_token"`
	TimeoutSecond int    `mapstructure:"timeout_second"`
}

// RealtimeConfig 连接与消费相关参数
type RealtimeConfig struct {
	WSPath                string   `mapstructure:"ws_path"`
	AllowedOrigins        []string `mapstructure:"allowed_origins"`
	AccessTokenCookieName string   `mapstructure:"access_token_cookie_name"`

	PingIntervalSecond int   `mapstructure:"ping_interval_second"`
	PongWaitSecond     int   `mapstructure:"pong_wait_second"`
	WriteQueueSize     int   `mapstructure:"write_queue_size"`
	ReadLimitBytes     int64 `mapstructure:"read_limit_bytes"`
	MaxConnections     int   `mapstructure:"max_connections"`

	ConsumerGroup string `mapstructure:"consumer_group"`
	// ConsumerBlockMs XREADGROUP 阻塞时长；超时一轮后继续循环
	ConsumerBlockMs int `mapstructure:"consumer_block_ms"`
	// ConsumerBatchSize 单次 XREADGROUP / XAUTOCLAIM 拉取条数
	ConsumerBatchSize int64 `mapstructure:"consumer_batch_size"`
	// PendingClaimIdleSecond 悬挂超过该时长的消息可被其他实例接管（XAUTOCLAIM）
	PendingClaimIdleSecond int `mapstructure:"pending_claim_idle_second"`
	// PendingClaimIntervalSecond 扫描悬挂消息的间隔
	PendingClaimIntervalSecond int `mapstructure:"pending_claim_interval_second"`

	// DeliveryTTLSecond 未确认事件的保留时长，超时后不再重投
	DeliveryTTLSecond int `mapstructure:"delivery_ttl_second"`
	// ConnectionTTLSecond 连接在线状态 key 的 TTL（短 TTL，靠心跳续期）
	ConnectionTTLSecond int `mapstructure:"connection_ttl_second"`
	// TerminateGraceSecond 会话终结事件投递后强制关闭连接的宽限期
	TerminateGraceSecond int `mapstructure:"terminate_grace_second"`
}

// LogConfig 日志配置
type LogConfig struct {
	Path       string `mapstructure:"path"`
	Level      string `mapstructure:"level"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// Config 全局配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Redis    RedisConfig    `mapstructure:"redis"`
	SSO      SSOConfig      `mapstructure:"sso"`
	Realtime RealtimeConfig `mapstructure:"realtime"`
	Log      LogConfig      `mapstructure:"log"`
}

// LoadConfig 加载配置文件，支持 ${ENV} 占位符
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(os.ExpandEnv(string(data)))); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	cfg.applyDefaults()
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8090
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "release"
	}
	if c.Server.ShutdownTimeoutSecond == 0 {
		c.Server.ShutdownTimeoutSecond = 10
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 50
	}
	if c.SSO.TimeoutSecond == 0 {
		c.SSO.TimeoutSecond = 3
	}
	if c.Realtime.WSPath == "" {
		c.Realtime.WSPath = "/api/v1/realtime/ws"
	}
	if c.Realtime.AccessTokenCookieName == "" {
		c.Realtime.AccessTokenCookieName = "mh_sso_access_token"
	}
	// dev.md §2：服务端每 25 秒 Ping，45 秒未收到 Pong 关闭连接
	if c.Realtime.PingIntervalSecond == 0 {
		c.Realtime.PingIntervalSecond = 25
	}
	if c.Realtime.PongWaitSecond == 0 {
		c.Realtime.PongWaitSecond = 45
	}
	// dev.md §2：写队列上限 128
	if c.Realtime.WriteQueueSize == 0 {
		c.Realtime.WriteQueueSize = 128
	}
	if c.Realtime.ReadLimitBytes == 0 {
		c.Realtime.ReadLimitBytes = 4096
	}
	if c.Realtime.MaxConnections == 0 {
		c.Realtime.MaxConnections = 10000
	}
	if c.Realtime.ConsumerGroup == "" {
		c.Realtime.ConsumerGroup = "ws-gateway"
	}
	if c.Realtime.ConsumerBlockMs == 0 {
		c.Realtime.ConsumerBlockMs = 2000
	}
	if c.Realtime.ConsumerBatchSize == 0 {
		c.Realtime.ConsumerBatchSize = 64
	}
	if c.Realtime.PendingClaimIdleSecond == 0 {
		c.Realtime.PendingClaimIdleSecond = 30
	}
	if c.Realtime.PendingClaimIntervalSecond == 0 {
		c.Realtime.PendingClaimIntervalSecond = 10
	}
	if c.Realtime.DeliveryTTLSecond == 0 {
		c.Realtime.DeliveryTTLSecond = 86400
	}
	if c.Realtime.ConnectionTTLSecond == 0 {
		c.Realtime.ConnectionTTLSecond = 60
	}
	if c.Realtime.TerminateGraceSecond == 0 {
		c.Realtime.TerminateGraceSecond = 3
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
}

// ---------- 派生参数 ----------

func (c *Config) PingInterval() time.Duration {
	return time.Duration(c.Realtime.PingIntervalSecond) * time.Second
}

func (c *Config) PongWait() time.Duration {
	return time.Duration(c.Realtime.PongWaitSecond) * time.Second
}

func (c *Config) ConsumerBlock() time.Duration {
	return time.Duration(c.Realtime.ConsumerBlockMs) * time.Millisecond
}

func (c *Config) PendingClaimIdle() time.Duration {
	return time.Duration(c.Realtime.PendingClaimIdleSecond) * time.Second
}

func (c *Config) PendingClaimInterval() time.Duration {
	return time.Duration(c.Realtime.PendingClaimIntervalSecond) * time.Second
}

func (c *Config) DeliveryTTL() time.Duration {
	return time.Duration(c.Realtime.DeliveryTTLSecond) * time.Second
}

func (c *Config) ConnectionTTL() time.Duration {
	return time.Duration(c.Realtime.ConnectionTTLSecond) * time.Second
}

func (c *Config) TerminateGrace() time.Duration {
	return time.Duration(c.Realtime.TerminateGraceSecond) * time.Second
}

func (c *Config) ShutdownTimeout() time.Duration {
	return time.Duration(c.Server.ShutdownTimeoutSecond) * time.Second
}
