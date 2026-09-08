package delivery

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"mh-ws-gateway/internal/rtest"
)

// 本文件验证"至少一次投递"存储语义：
// 落盘(Save) → 补发(Pending) → 确认(Ack)，以及按 eventId 幂等覆盖。
// 只有真实 Redis 上的 Hash 行为才算数。

func testStore(t *testing.T) (*Store, context.Context) {
	rdb := rtest.Start(t)
	rtest.Flush(t, rdb)
	return NewStore(rdb, time.Hour), context.Background()
}

func TestSavePendingAckClear(t *testing.T) {
	s, ctx := testStore(t)
	sid := "session-1"
	p1 := []byte(`{"type":"event","eventId":"e1"}`)
	p2 := []byte(`{"type":"event","eventId":"e2"}`)

	if err := s.Save(ctx, sid, "e1", p1); err != nil {
		t.Fatalf("Save e1 失败: %v", err)
	}
	if err := s.Save(ctx, sid, "e2", p2); err != nil {
		t.Fatalf("Save e2 失败: %v", err)
	}

	pending, err := s.Pending(ctx, sid)
	if err != nil {
		t.Fatalf("Pending 失败: %v", err)
	}
	if len(pending) != 2 || !contains(pending, p1) || !contains(pending, p2) {
		t.Fatalf("Pending 应包含两条消息且内容一致，实际 %d 条: %v", len(pending), pending)
	}

	// 客户端 ACK e1 后仅剩 e2
	if err := s.Ack(ctx, sid, "e1"); err != nil {
		t.Fatalf("Ack e1 失败: %v", err)
	}
	pending, _ = s.Pending(ctx, sid)
	if len(pending) != 1 || !bytes.Equal(pending[0], p2) {
		t.Fatalf("Ack 后应只剩 e2，实际: %v", pending)
	}

	// 全部 ACK 后 key 应消失
	if err := s.Ack(ctx, sid, "e2"); err != nil {
		t.Fatalf("Ack e2 失败: %v", err)
	}
	pending, _ = s.Pending(ctx, sid)
	if len(pending) != 0 {
		t.Fatalf("全部 Ack 后 Pending 应为空，实际 %d", len(pending))
	}
	if n, _ := s.rdb.Exists(ctx, s.key(sid)).Result(); n != 0 {
		t.Fatalf("全部 Ack 后 delivery key 应被删除，仍存在")
	}

	// Clear：对不存在 key 也应幂等成功
	if err := s.Clear(ctx, sid); err != nil {
		t.Fatalf("Clear 失败: %v", err)
	}
	if err := s.Clear(ctx, "no-such-session"); err != nil {
		t.Fatalf("Clear 未知会话应幂等成功: %v", err)
	}
}

// 重复投递同一 eventId 必须覆盖旧值而非累积 —— 这是"至少一次、客户端按 eventId 幂等"的基础
func TestSaveIdempotentByEventID(t *testing.T) {
	s, ctx := testStore(t)
	sid := "session-2"

	if err := s.Save(ctx, sid, "e1", []byte("v1")); err != nil {
		t.Fatalf("Save v1 失败: %v", err)
	}
	if err := s.Save(ctx, sid, "e1", []byte("v2")); err != nil {
		t.Fatalf("Save v2 失败: %v", err)
	}
	pending, err := s.Pending(ctx, sid)
	if err != nil {
		t.Fatalf("Pending 失败: %v", err)
	}
	if len(pending) != 1 || string(pending[0]) != "v2" {
		t.Fatalf("同一 eventId 重复落盘应只剩最新值，实际: %v", pending)
	}
}

func TestPendingMissingSessionEmpty(t *testing.T) {
	s, ctx := testStore(t)
	pending, err := s.Pending(ctx, "never-existed")
	if err != nil {
		t.Fatalf("未知会话 Pending 不应报错: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("未知会话 Pending 应为空，实际 %d", len(pending))
	}
}

func TestSaveSetsTTL(t *testing.T) {
	rdb := rtest.Start(t)
	rtest.Flush(t, rdb)
	s := NewStore(rdb, time.Second)
	ctx := context.Background()

	if err := s.Save(ctx, "session-3", "e1", []byte("v1")); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	ttl, err := rdb.TTL(ctx, s.key("session-3")).Result()
	if err != nil {
		t.Fatalf("读取 TTL 失败: %v", err)
	}
	if ttl <= 0 || ttl > time.Second {
		t.Fatalf("TTL 应在 (0, 1s]，实际 %v", ttl)
	}
	// TTL 到期后待投递集合自然消失，不再补发
	time.Sleep(1500 * time.Millisecond)
	pending, _ := s.Pending(ctx, "session-3")
	if len(pending) != 0 {
		t.Fatalf("TTL 到期后 Pending 应为空，实际 %d", len(pending))
	}
}

// 并发落盘同一会话的多个事件，最终不丢不漏
func TestConcurrentSavePending(t *testing.T) {
	s, ctx := testStore(t)
	sid := "session-concurrent"
	const n = 50

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			payload := []byte(fmt.Sprintf(`{"eventId":"e%d"}`, i))
			if err := s.Save(ctx, sid, fmt.Sprintf("e%d", i), payload); err != nil {
				t.Errorf("并发 Save e%d 失败: %v", i, err)
			}
		}(i)
	}
	wg.Wait()

	pending, err := s.Pending(ctx, sid)
	if err != nil {
		t.Fatalf("Pending 失败: %v", err)
	}
	if len(pending) != n {
		t.Fatalf("并发落盘后应有 %d 条待确认消息，实际 %d", n, len(pending))
	}
	seen := map[string]bool{}
	for _, p := range pending {
		seen[string(p)] = true
	}
	for i := 0; i < n; i++ {
		if !seen[fmt.Sprintf(`{"eventId":"e%d"}`, i)] {
			t.Fatalf("缺少事件 e%d", i)
		}
	}
}

func contains(rows [][]byte, target []byte) bool {
	for _, r := range rows {
		if bytes.Equal(r, target) {
			return true
		}
	}
	return false
}
