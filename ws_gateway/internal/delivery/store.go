// Package delivery 管理"至少一次投递"语义下的待确认消息。
//
// 投递流程：
//
//	Gateway 收到 Stream 事件 → 先 HSET 到 sso:delivery:{sessionId} → 再 XACK
//	→ 推给连接 → 客户端回 ack → HDEL
//
// 未确认的消息在连接重建时重新发送，因此投递存储必须先于 XACK 落盘，
// 否则"事件已确认但没存下来"会导致消息永久丢失。
package delivery

import (
	"context"
	"time"

	"mh-ws-gateway/internal/protocol"

	"github.com/redis/go-redis/v9"
)

// Store 待确认消息存储：sso:delivery:{sessionId} -> Hash(eventId -> payload)
type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewStore 创建投递存储
func NewStore(rdb *redis.Client, ttl time.Duration) *Store {
	return &Store{rdb: rdb, ttl: ttl}
}

func (s *Store) key(sessionID string) string { return protocol.DeliveryKeyPrefix + sessionID }

// Save 保存待投递消息（幂等：重复投递覆盖同一 eventId）
func (s *Store) Save(ctx context.Context, sessionID, eventID string, payload []byte) error {
	pipe := s.rdb.Pipeline()
	pipe.HSet(ctx, s.key(sessionID), eventID, payload)
	pipe.Expire(ctx, s.key(sessionID), s.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// Pending 取出该会话全部未确认消息（重连后补发）
func (s *Store) Pending(ctx context.Context, sessionID string) ([][]byte, error) {
	values, err := s.rdb.HVals(ctx, s.key(sessionID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	out := make([][]byte, 0, len(values))
	for _, v := range values {
		out = append(out, []byte(v))
	}
	return out, nil
}

// Ack 确认消息：从待投递集合删除
func (s *Store) Ack(ctx context.Context, sessionID, eventID string) error {
	return s.rdb.HDel(ctx, s.key(sessionID), eventID).Err()
}

// Clear 清空该会话的全部待投递消息（会话终止后调用）
func (s *Store) Clear(ctx context.Context, sessionID string) error {
	return s.rdb.Del(ctx, s.key(sessionID)).Err()
}
