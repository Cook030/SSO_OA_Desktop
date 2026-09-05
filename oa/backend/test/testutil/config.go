package testutil

import (
	"os"
	"strconv"

	"permission-system/internal/utils"
)

const (
	// DefaultConfigPath 默认配置文件路径
	DefaultConfigPath = "../config/config.example.yaml"

	// EmailDomain 员工邮箱域名
	EmailDomain = "@maplehaze.cn"
)

// Cfg 加载后的配置
var Cfg *utils.Config

// LoadTestConfig 加载测试配置文件
// 优先使用环境变量 TEST_CONFIG_PATH 指定的路径，否则使用默认路径
func LoadTestConfig() *utils.Config {
	path := DefaultConfigPath
	if envPath := os.Getenv("TEST_CONFIG_PATH"); envPath != "" {
		path = envPath
	}
	cfg, err := utils.LoadConfig(path)
	if err != nil {
		panic("加载测试配置失败: " + err.Error())
	}
	return cfg
}

// ServerPort 获取配置中的服务端口号，用于日志或调试
func ServerPort() string {
	return strconv.Itoa(Cfg.Server.Port)
}
