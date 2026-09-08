package cache

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// 本文件为 cache 包的真实 Redis 集成测试提供支撑。
// Lua 脚本（loginReplaceLua / revokeUserLua / revokeSessionLua）只能由真实 Redis
// 执行，进程内模拟器无法替代，因此测试在随机端口拉起隔离的 redis-server。

// startRedis 启动（或按 TEST_REDIS_ADDR 复用外部）隔离 Redis，返回客户端。
// 外部实例统一使用 db 15 作为专用测试库，每个用例前 FlushDB。
func startRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("TEST_REDIS_ADDR")
	db := 0
	if addr == "" {
		bin := redisServerBin(t)
		port := freePort(t)
		cmd := exec.Command(bin,
			"--port", strconv.Itoa(port),
			"--bind", "127.0.0.1",
			"--save", "",
			"--appendonly", "no",
		)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			t.Fatalf("启动测试 redis-server 失败: %v", err)
		}
		addr = fmt.Sprintf("127.0.0.1:%d", port)
		t.Cleanup(func() {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
		})
	} else {
		db = 15 // 外部共享实例：固定专用测试库
	}

	cli := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if err := cli.Ping(ctx).Err(); err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = cli.Close()
			t.Fatalf("等待测试 Redis（%s db=%d）就绪超时", addr, db)
		}
		time.Sleep(30 * time.Millisecond)
	}
	t.Cleanup(func() { _ = cli.Close() })
	t.Logf("integration redis ready: %s db=%d", addr, db)
	return cli
}

func redisServerBin(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("REDIS_SERVER_BIN"); p != "" {
		return p
	}
	if p, err := exec.LookPath("redis-server"); err == nil {
		return p
	}
	if runtime.GOOS == "windows" {
		for _, cand := range []string{
			`D:\redis\redis-server.exe`,
			`C:\Program Files\Redis\redis-server.exe`,
			`C:\Redis\redis-server.exe`,
		} {
			if _, err := os.Stat(cand); err == nil {
				return cand
			}
		}
	}
	t.Skip("跳过真实 Redis 集成测试：未找到 redis-server，可设置 TEST_REDIS_ADDR 或 REDIS_SERVER_BIN")
	return ""
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("分配空闲端口失败: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

func flushAll(t *testing.T, cli *redis.Client) {
	t.Helper()
	if err := cli.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("清空测试库失败: %v", err)
	}
}

// newTestCache 构造白盒 Cache（真实 Redis + 静默日志）
func newTestCache(cli *redis.Client) *Cache {
	return &Cache{rdb: cli, log: zap.NewNop()}
}

// requireRedisAtLeast 校验测试 Redis 版本。
// 会话置换 Lua 使用 SET ... KEEPTTL（Redis ≥6.0）与 XAUTOCLAIM（Redis ≥6.2），
// 低于目标版本时整个特性不可用，用例跳过（CI 使用新版 Redis 时才会真正执行）。
func requireRedisAtLeast(t *testing.T, cli *redis.Client, major, minor int) {
	t.Helper()
	info := cli.Info(context.Background(), "server").Val()
	idx := strings.Index(info, "redis_version:")
	if idx < 0 {
		t.Skip("无法获取 Redis 版本，跳过")
	}
	rest := info[idx+len("redis_version:"):]
	if nl := strings.Index(rest, "\r\n"); nl >= 0 {
		rest = rest[:nl]
	}
	parts := strings.SplitN(rest, ".", 3)
	gotMajor, _ := strconv.Atoi(parts[0])
	gotMinor := 0
	if len(parts) > 1 {
		gotMinor, _ = strconv.Atoi(parts[1])
	}
	if gotMajor < major || (gotMajor == major && gotMinor < minor) {
		t.Skipf("本用例需要 Redis ≥%d.%d（Lua KEEPTTL 等），当前版本 %s，跳过", major, minor, rest)
	}
}
