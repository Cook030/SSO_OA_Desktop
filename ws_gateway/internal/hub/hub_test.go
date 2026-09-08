package hub

import (
	"sync"
	"sync/atomic"
	"testing"
)

// fakeConn 测试用连接，实现 hub.Connection 接口
type fakeConn struct {
	sessionID string
	closed    atomic.Int32
}

func (f *fakeConn) SessionID() string { return f.sessionID }
func (f *fakeConn) Send([]byte) error { return nil }
func (f *fakeConn) Close(_ int, _ string) {
	f.closed.Store(1)
}

func TestHubRegisterReplaceUnregister(t *testing.T) {
	h := NewHub()
	c1 := &fakeConn{sessionID: "s1"}
	c2 := &fakeConn{sessionID: "s1"}
	c3 := &fakeConn{sessionID: "s2"}

	if prev := h.Register(c1); prev != nil {
		t.Fatalf("首次注册应返回 nil，实际 %v", prev)
	}
	if h.Len() != 1 {
		t.Fatalf("Len 期望 1，实际 %d", h.Len())
	}

	// 同会话再次注册：返回被取代的旧连接
	if prev := h.Register(c2); prev != c1 {
		t.Fatalf("重复注册应返回旧连接 c1，实际 %v", prev)
	}
	if cur, ok := h.Get("s1"); !ok || cur != c2 {
		t.Fatalf("注册表应指向新连接 c2")
	}

	// 旧连接注销不能误删当前连接（关键：避免新连接刚建立就被旧连接清理）
	h.Unregister("s1", c1)
	if _, ok := h.Get("s1"); !ok {
		t.Fatalf("旧连接 c1 注销不应移除当前连接 c2")
	}
	if cur, _ := h.Get("s1"); cur != c2 {
		t.Fatalf("当前连接应为 c2")
	}

	h.Register(c3)
	if h.Len() != 2 {
		t.Fatalf("两个会话 Len 应为 2，实际 %d", h.Len())
	}
	// 当前连接注销才真正删除
	h.Unregister("s1", c2)
	if _, ok := h.Get("s1"); ok {
		t.Fatalf("当前连接注销后应移除")
	}
	if h.Len() != 1 {
		t.Fatalf("Len 应为 1，实际 %d", h.Len())
	}
}

func TestHubConcurrentRegisterLastWins(t *testing.T) {
	const n = 64
	h := NewHub()
	var replaced atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if prev := h.Register(&fakeConn{sessionID: "s"}); prev != nil {
				replaced.Add(1)
			}
		}()
	}
	wg.Wait()

	// 同一会话并发建连：最终只剩一条连接，且恰好 n-1 条旧连接被取代
	if h.Len() != 1 {
		t.Fatalf("并发注册后 Len 应为 1，实际 %d", h.Len())
	}
	if got := replaced.Load(); got != n-1 {
		t.Fatalf("被取代连接数应为 %d，实际 %d", n-1, got)
	}
	cur, ok := h.Get("s")
	if !ok || cur == nil {
		t.Fatalf("并发注册后仍应存在一条当前连接")
	}
}

func TestHubConcurrentDifferentSessions(t *testing.T) {
	const n = 200
	h := NewHub()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sid := "session-" + string(rune('a'+i%26))
			h.Register(&fakeConn{sessionID: sid})
			if i%2 == 0 {
				h.Unregister(sid, &fakeConn{sessionID: sid})
			}
			_, _ = h.Get(sid)
			_ = h.Len()
		}(i)
	}
	wg.Wait()
	// 不追求精确数值，只验证并发读写无 panic / 无数据竞争（配合 -race）
}

func TestHubCloseAll(t *testing.T) {
	h := NewHub()
	var conns []*fakeConn
	for i := 0; i < 10; i++ {
		c := &fakeConn{sessionID: "s" + string(rune('0'+i))}
		conns = append(conns, c)
		h.Register(c)
	}
	h.CloseAll(4004, "shutdown")

	if h.Len() != 0 {
		t.Fatalf("CloseAll 后 Len 应为 0，实际 %d", h.Len())
	}
	for i, c := range conns {
		if c.closed.Load() != 1 {
			t.Fatalf("连接 %d 应被关闭", i)
		}
	}
	// CloseAll 后再注册应正常
	h.Register(&fakeConn{sessionID: "new"})
	if h.Len() != 1 {
		t.Fatalf("CloseAll 后新注册失败，Len=%d", h.Len())
	}
}
