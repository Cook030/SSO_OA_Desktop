package utils

import (
	"fmt"

	"github.com/spf13/viper"
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

// AdminConfig 管理员初始化配置
type AdminConfig struct {
	Account  string `mapstructure:"account"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	Phone    string `mapstructure:"phone"`
	Email    string `mapstructure:"email"`
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

// SSOConfig 单点登录配置
type SSOConfig struct {
	BaseURL                string `mapstructure:"base_url"`
	IntrospectPath         string `mapstructure:"introspect_path"`
	RevokeUserSessionsPath string `mapstructure:"revoke_user_sessions_path"`
	TimeoutSecond          int    `mapstructure:"timeout_second"`
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
	Admin  AdminConfig  `mapstructure:"admin"`
	CORS   CORSConfig   `mapstructure:"cors"`
	SSO    SSOConfig    `mapstructure:"sso"`
	Log    LogConfig    `mapstructure:"log"`
}

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &cfg, nil
}

// DSN 返回MySQL连接字符串
func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=10s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
}
