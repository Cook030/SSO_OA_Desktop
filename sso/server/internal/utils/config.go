package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Cookie 与认证 Header 常量（与 SSO 接口文档保持一致）
const (
	AccessTokenCookieName  = "mh_sso_access_token"
	RefreshTokenCookieName = "mh_sso_refresh_token"
	RefreshTokenHeaderName = "X-MH-Refresh-Token"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// MySQLConfig 数据库配置
type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// RedisConfig 缓存配置（仅作缓存/锁/限流，非权威数据源）
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Issuer                string `mapstructure:"issuer"`
	JWTSecret             string `mapstructure:"jwt_secret"`
	AccessTokenTTLSecond  int    `mapstructure:"access_token_ttl_second"`
	RefreshTokenTTLSecond int    `mapstructure:"refresh_token_ttl_second"`
	SessionTTLSecond      int    `mapstructure:"session_ttl_second"`
	CookieDomain          string `mapstructure:"cookie_domain"`
	CookieSecure          bool   `mapstructure:"cookie_secure"`
	CookieSameSite        string `mapstructure:"cookie_same_site"`
}

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// LogConfig 日志配置
type LogConfig struct {
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
	LocalTime  bool   `mapstructure:"local_time"`
}

// Config 全局配置
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Auth   AuthConfig   `mapstructure:"auth"`
	CORS   CORSConfig   `mapstructure:"cors"`
	Log    LogConfig    `mapstructure:"log"`
}

// LoadConfig 加载配置文件，支持 ${ENV} 形式的环境变量占位符展开
// （如 jwt_secret: "${JWT_SECRET}" 读取环境变量 JWT_SECRET）
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

// applyDefaults 填充缺省配置
func (c *Config) applyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "release"
	}
	if c.MySQL.Charset == "" {
		c.MySQL.Charset = "utf8mb4"
	}
	if c.MySQL.MaxIdleConns == 0 {
		c.MySQL.MaxIdleConns = 10
	}
	if c.MySQL.MaxOpenConns == 0 {
		c.MySQL.MaxOpenConns = 100
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 20
	}
	if c.Auth.Issuer == "" {
		c.Auth.Issuer = "mh-sso"
	}
	if c.Auth.AccessTokenTTLSecond == 0 {
		c.Auth.AccessTokenTTLSecond = 900
	}
	if c.Auth.RefreshTokenTTLSecond == 0 {
		c.Auth.RefreshTokenTTLSecond = 2592000
	}
	if c.Auth.SessionTTLSecond == 0 {
		c.Auth.SessionTTLSecond = 2592000
	}
	if c.Auth.CookieSameSite == "" {
		c.Auth.CookieSameSite = "Lax"
	}
	if len(c.CORS.AllowMethods) == 0 {
		c.CORS.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(c.CORS.AllowHeaders) == 0 {
		c.CORS.AllowHeaders = []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			RefreshTokenHeaderName, "X-Request-Id",
		}
	}
	if c.CORS.MaxAge == 0 {
		c.CORS.MaxAge = 43200
	}
}

// DSN 返回 MySQL 连接字符串
func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=10s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
}
