package testutil

import (
	"sync"
)

var (
	cleanupFuncs []func()
	cleanupMu    sync.Mutex
)

// RegisterCleanup 注册一个清理函数，在全部测试结束后执行
func RegisterCleanup(fn func()) {
	cleanupMu.Lock()
	defer cleanupMu.Unlock()
	cleanupFuncs = append(cleanupFuncs, fn)
}

// RunCleanup 执行所有注册的清理函数，按后进先出顺序执行
func RunCleanup() {
	cleanupMu.Lock()
	defer cleanupMu.Unlock()

	for i := len(cleanupFuncs) - 1; i >= 0; i-- {
		cleanupFuncs[i]()
	}
	cleanupFuncs = nil
}
