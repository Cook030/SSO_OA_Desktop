// Package rtest 为真实 Redis 集成测试提供支撑。
//
// 为什么需要真实 Redis：本仓库要验证的可靠性语义 —— Redis Lua 原子脚本、
// Consumer Group 的 XAUTOCLAIM/PEL 恢复、XREADGROUP 多实例消费 ——
// 均无法用 miniredis 这类进程内模拟器表达，只有真实 Redis 的结果才算数。
//
// 为避免污染开发者本机 6379，默认策略是在随机端口拉起一个隔离的
// redis-server 进程（不持久化、仅监听 127.0.0.1），测试结束自动杀掉。
//
// 环境变量：
//   - TEST_REDIS_ADDR：显式指定外部 Redis 地址（CI 容器场景）。此时不自启实例，
//     统一使用 TEST_REDIS_DB（默认 15）作为专用测试库，每个用例前执行 FLUSHDB。
//   - REDIS_SERVER_BIN：redis-server 可执行文件路径（默认在 PATH 中查找）。
//
// 当既找不到 redis-server 又没有外部地址时，调用方用 t.Skip 跳过集成用例，
// 普通单测不受影响。
package rtest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// Start 返回一个连接就绪的隔离 Redis 客户端。
// 资源（进程 / 连接）随测试结束自动回收；无 Redis 可用时 t.Skip。
func Start(t *testing.T) *redis.Client {
	t.Helper()

	db := 0
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		bin, err := findRedisServer()
		if err != nil {
			t.Skip("跳过真实 Redis 集成测试：未找到 redis-server，可设置 TEST_REDIS_ADDR 或 REDIS_SERVER_BIN")
		}
		port := freePort(t)
		cmd := exec.Command(bin,
			"--port", strconv.Itoa(port),
			"--bind", "127.0.0.1",
			"--save", "", // 不落盘，避免测试垃圾
			"--appendonly", "no",
			"--databases", "16",
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
		// 外部实例：固定使用专用 DB，避免误伤其它数据。
		if v := os.Getenv("TEST_REDIS_DB"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				db = n
			}
		} else {
			db = 15
		}
	}

	cli := redis.NewClient(&redis.Options{Addr: addr, DB: db, PoolSize: 8})
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

// Flush 清空当前测试库；集成用例应在一个干净的 Redis 上开始。
func Flush(t *testing.T, cli *redis.Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cli.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("清空测试库失败: %v", err)
	}
}

// Eventually 轮询直到 cond 成立或超时，避免脆弱的固定 sleep。
func Eventually(t *testing.T, timeout time.Duration, desc string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("等待超时（%v）：%s", timeout, desc)
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

func findRedisServer() (string, error) {
	if p := os.Getenv("REDIS_SERVER_BIN"); p != "" {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	if p, err := exec.LookPath("redis-server"); err == nil {
		return p, nil
	}
	if runtime.GOOS == "windows" {
		for _, cand := range []string{
			`D:\redis\redis-server.exe`,
			`C:\Program Files\Redis\redis-server.exe`,
			`C:\Redis\redis-server.exe`,
		} {
			if st, err := os.Stat(cand); err == nil && !st.IsDir() {
				return cand, nil
			}
		}
	}
	return "", errors.New("redis-server 不在 PATH（可设置 REDIS_SERVER_BIN）")
}
