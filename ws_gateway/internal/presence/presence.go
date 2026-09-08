// Package presence 维护连接在线状态：ws:connections:{sessionId}。
//
// 职责边界：这只是**路由辅助数据**（用于发现"同一会话是否已在另一个 Gateway 实例上建连"），
// 不是安全状态。单设备登录的权威判定在 SSO 的 current_session，
// 即使本 key 全部丢失，也不会出现"两端同时有效"。
package presence

import (
	"context"
	"encoding/json"
	"time"

	"mh-ws-gateway/internal/protocol"

	"github.com/redis/go-redis/v9"
)

// Record 连接在线状态
type Record struct {
	InstanceID    string `json:"gatewayInstanceId"`
	ConnectionID  string `json:"connectionId"`
	SessionID     string `json:"sessionId"`
	LastHeartbeat int64  `json:"lastHeartbeat"` // Unix 毫秒
}

// Store 在线状态存储
type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewStore 创建在线状态存储
func NewStore(rdb *redis.Client, ttl time.Duration) *Store {
	return &Store{rdb: rdb, ttl: ttl}
}

func (s *Store) key(sessionID string) string { return protocol.ConnectionKeyPrefix + sessionID }

// Register 登记连接（SET NX）；返回 false 表示该会话已有存活连接
func (s *Store) Register(ctx context.Context, rec Record) (bool, error) {
	raw, err := json.Marshal(rec)
	if err != nil {
		return false, err
	}
	return s.rdb.SetNX(ctx, s.key(rec.SessionID), raw, s.ttl).Result()
}

// Heartbeat 续期并刷新心跳时间
func (s *Store) Heartbeat(ctx context.Context, rec Record) error {
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.key(rec.SessionID), raw, s.ttl).Err()
}

// Release 释放连接：仅当 key 仍属于本连接时才删除，避免误删新建立的连接
func (s *Store) Release(ctx context.Context, sessionID, connectionID string) error {
	const script = `
local raw = redis.call("GET", KEYS[1])
if not raw then return 0 end
local ok, rec = pcall(cjson.decode, raw)
if ok and type(rec) == "table" and rec["connectionId"] == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`
	return s.rdb.Eval(ctx, script, []string{s.key(sessionID)}, connectionID).Err()
}
