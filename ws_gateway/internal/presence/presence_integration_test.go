package presence

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mh-ws-gateway/internal/rtest"

	"github.com/redis/go-redis/v9"
)

// 本文件验证连接在线状态（路由辅助数据）在真实 Redis 上的行为，
// 重点是 Release 的 Lua 脚本：旧连接关闭时不得误删新连接持有的 key。

func setup(t *testing.T) (*Store, *redis.Client, context.Context) {
	rdb := rtest.Start(t)
	rtest.Flush(t, rdb)
	return NewStore(rdb, time.Minute), rdb, context.Background()
}

// 直接读取当前 key 中的记录
func getRecord(t *testing.T, rdb *redis.Client, key string) Record {
	t.Helper()
	raw, err := rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		t.Fatalf("读取在线状态记录失败: %v", err)
	}
	var rec Record
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("解析在线状态记录失败: %v", err)
	}
	return rec
}

func TestRegisterSemantics(t *testing.T) {
	s, rdb, ctx := setup(t)
	session := "session-a"
	recA := Record{InstanceID: "gw-1", ConnectionID: "conn-1", SessionID: session, LastHeartbeat: 1}

	ok, err := s.Register(ctx, recA)
	if err != nil || !ok {
		t.Fatalf("首次 Register 应成功占位，ok=%v err=%v", ok, err)
	}
	ttl, _ := rdb.TTL(ctx, s.key(session)).Result()
	if ttl <= 0 || ttl > time.Minute {
		t.Fatalf("Register 应设置在线状态 TTL，实际 %v", ttl)
	}

	// 同会话再次 Register（另一实例新连接）：返回 false，key 仍归首连接所有
	recB := Record{InstanceID: "gw-2", ConnectionID: "conn-2", SessionID: session, LastHeartbeat: 2}
	ok, err = s.Register(ctx, recB)
	if err != nil || ok {
		t.Fatalf("同会话第二次 Register 应返回 false，ok=%v err=%v", ok, err)
	}
	got := getRecord(t, rdb, s.key(session))
	if got.ConnectionID != recA.ConnectionID || got.InstanceID != recA.InstanceID {
		t.Fatalf("key 应仍归首连接所有，实际 %+v", got)
	}
}

func TestHeartbeatReplacesOwner(t *testing.T) {
	s, rdb, ctx := setup(t)
	session := "session-b"
	if _, err := s.Register(ctx, Record{InstanceID: "gw-1", ConnectionID: "c1", SessionID: session}); err != nil {
		t.Fatalf("Register 失败: %v", err)
	}
	// 心跳覆盖：连接可能漂移到新实例，路由应指向最新连接
	hb := Record{InstanceID: "gw-2", ConnectionID: "c2", SessionID: session, LastHeartbeat: 99}
	if err := s.Heartbeat(ctx, hb); err != nil {
		t.Fatalf("Heartbeat 失败: %v", err)
	}
	got := getRecord(t, rdb, s.key(session))
	if got.ConnectionID != "c2" || got.InstanceID != "gw-2" || got.LastHeartbeat != 99 {
		t.Fatalf("心跳后记录应被覆盖为最新连接，实际 %+v", got)
	}
}

// 关键可靠性语义：会话被新连接取代后，旧连接关闭时 Release 不得误删新连接。
// 否则会出现"新 Gateway 已接管，旧 Gateway 关闭瞬间把路由状态抹掉"的竞态。
func TestReleaseOnlyRemovesOwnConnection(t *testing.T) {
	s, rdb, ctx := setup(t)
	session := "session-c"
	oldConn := Record{InstanceID: "gw-1", ConnectionID: "old-conn", SessionID: session}
	newConn := Record{InstanceID: "gw-2", ConnectionID: "new-conn", SessionID: session}

	if _, err := s.Register(ctx, oldConn); err != nil {
		t.Fatalf("旧连接 Register 失败: %v", err)
	}
	// 新连接同会话（另一实例）开始心跳，接管路由
	if err := s.Heartbeat(ctx, newConn); err != nil {
		t.Fatalf("新连接 Heartbeat 失败: %v", err)
	}

	// 旧连接关闭清理：不得删除新连接的 key
	if err := s.Release(ctx, session, oldConn.ConnectionID); err != nil {
		t.Fatalf("旧连接 Release 失败: %v", err)
	}
	got := getRecord(t, rdb, s.key(session))
	if got.ConnectionID != "new-conn" {
		t.Fatalf("旧连接 Release 误删新连接路由，实际 %+v", got)
	}

	// 新连接自己释放：key 消失
	if err := s.Release(ctx, session, newConn.ConnectionID); err != nil {
		t.Fatalf("新连接 Release 失败: %v", err)
	}
	if n, _ := rdb.Exists(ctx, s.key(session)).Result(); n != 0 {
		t.Fatalf("owner Release 后 key 应被删除")
	}
}

func TestReleaseIdempotentAndMissing(t *testing.T) {
	s, rdb, ctx := setup(t)
	session := "session-d"
	// key 不存在时 Release 不应报错
	if err := s.Release(ctx, session, "whatever"); err != nil {
		t.Fatalf("key 不存在时 Release 应幂等成功: %v", err)
	}

	if _, err := s.Register(ctx, Record{InstanceID: "gw-1", ConnectionID: "c1", SessionID: session}); err != nil {
		t.Fatalf("Register 失败: %v", err)
	}
	// 重复 Release 幂等
	if err := s.Release(ctx, session, "c1"); err != nil {
		t.Fatalf("首次 Release 失败: %v", err)
	}
	if err := s.Release(ctx, session, "c1"); err != nil {
		t.Fatalf("重复 Release 应幂等成功: %v", err)
	}
	if n, _ := rdb.Exists(ctx, s.key(session)).Result(); n != 0 {
		t.Fatalf("Release 后 key 应已删除")
	}
}

func TestRegisterTTLExpires(t *testing.T) {
	rdb := rtest.Start(t)
	rtest.Flush(t, rdb)
	s := NewStore(rdb, time.Second)
	ctx := context.Background()
	session := "session-e"

	if _, err := s.Register(ctx, Record{InstanceID: "gw-1", ConnectionID: "c1", SessionID: session}); err != nil {
		t.Fatalf("Register 失败: %v", err)
	}
	// 心跳间隔内 key 存活；连接断开后（不再续期）路由自然消失
	time.Sleep(1500 * time.Millisecond)
	if n, _ := rdb.Exists(ctx, s.key(session)).Result(); n != 0 {
		t.Fatalf("TTL 到期后在线状态应消失")
	}
}
