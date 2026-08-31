package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"mh-sso-svc/internal/consts"

	"github.com/redis/go-redis/v9"
)

// ErrRecordNotFound Redis 中不存在对应记录（不存在或已过期）
var ErrRecordNotFound = errors.New("cache: record not found")

// SessionRecord 会话记录（sso:session:{sessionId}）
type SessionRecord struct {
	SessionID       string    `json:"sessionId"`
	UserID          uint64    `json:"userId"`
	DeviceType      *string   `json:"deviceType,omitempty"`
	DeviceID        *string   `json:"deviceId,omitempty"`
	LoginIP         *string   `json:"loginIp,omitempty"`
	LoginUserAgent  *string   `json:"loginUserAgent,omitempty"`
	Status          int       `json:"status"`
	PasswordVersion int       `json:"passwordVersion"` // 登录时刻的密码版本快照
	LastActiveAt    time.Time `json:"lastActiveAt"`
	ExpiredAt       time.Time `json:"expiredAt"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// IsActive 会话当前是否有效（状态 active 且未过期）
func (s *SessionRecord) IsActive(now time.Time) bool {
	return s.Status == consts.SessionStatusActive && s.ExpiredAt.After(now)
}

// RefreshTokenRecord refresh token 记录（sso:rt:{tokenHash}）
// 只存 token 的 SHA-256 哈希，明文不落存储；
// 同一 session_id 下的令牌旋转链即 token family，重放检测时按 family 整体撤销
type RefreshTokenRecord struct {
	TokenHash   string    `json:"tokenHash"`
	SessionID   string    `json:"sessionId"`
	UserID      uint64    `json:"userId"`
	Status      int       `json:"status"`
	ExpiredAt   time.Time `json:"expiredAt"`
	RotatedFrom string    `json:"rotatedFrom,omitempty"` // 由哪个 token 哈希轮换而来
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"` // 轮换/撤销时即 used_at
}

// ---------- 会话 ----------

// SaveSession 写入会话记录并登记用户会话索引，TTL 与会话过期时间对齐
func (c *Cache) SaveSession(rec *SessionRecord, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Until(rec.ExpiredAt)
	}
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	ctx := context.Background()
	idx := userSessionKey(rec.UserID)
	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, sessionKey(rec.SessionID), raw, ttl)
	pipe.SAdd(ctx, idx, rec.SessionID)
	pipe.Expire(ctx, idx, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		c.onError("save_session", err)
		return err
	}
	c.onOK()
	return nil
}

// GetSession 读取会话记录；不存在返回 ErrRecordNotFound
func (c *Cache) GetSession(sessionID string) (*SessionRecord, error) {
	var rec SessionRecord
	if err := c.getJSON(sessionKey(sessionID), &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// UpdateSessionStatus 更新会话状态并保留剩余 TTL（登出/撤销场景）
func (c *Cache) UpdateSessionStatus(sessionID string, status int) error {
	rec, err := c.GetSession(sessionID)
	if err != nil {
		return err
	}
	if rec.Status == status {
		return nil
	}
	rec.Status = status
	rec.UpdatedAt = time.Now()
	return c.setSessionKeepTTL(rec)
}

// TouchSession 更新会话最近活跃时间并滑动续期（TTL 同步延长到新的过期时间）
func (c *Cache) TouchSession(sessionID string, lastActiveAt, expiredAt time.Time) error {
	rec, err := c.GetSession(sessionID)
	if err != nil {
		return err
	}

	ttl := time.Until(expiredAt)
	if ttl <= 0 {
		return ErrRecordNotFound
	}

	rec.LastActiveAt = lastActiveAt
	rec.ExpiredAt = expiredAt
	rec.UpdatedAt = lastActiveAt

	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	ctx := context.Background()
	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, sessionKey(sessionID), raw, ttl)
	pipe.Expire(ctx, userSessionKey(rec.UserID), ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		c.onError("touch_session", err)
		return err
	}
	c.onOK()
	return nil
}

// RevokeUserSessions 将用户全部 active 会话置为指定状态，返回被撤销的会话数
func (c *Cache) RevokeUserSessions(userID uint64, status int) (int, error) {
	ctx := context.Background()
	idx := userSessionKey(userID)
	sessionIDs, err := c.rdb.SMembers(ctx, idx).Result()
	if err != nil {
		c.onError("list_user_sessions", err)
		return 0, err
	}

	now := time.Now()
	revoked := 0
	for _, sessionID := range sessionIDs {
		rec, err := c.GetSession(sessionID)
		if errors.Is(err, ErrRecordNotFound) {
			// 记录已过期，顺带清理索引残留
			_ = c.rdb.SRem(ctx, idx, sessionID)
			continue
		}
		if err != nil || rec.Status != consts.SessionStatusActive {
			continue
		}
		rec.Status = status
		rec.UpdatedAt = now
		if err := c.setSessionKeepTTL(rec); err != nil {
			continue
		}
		revoked++
	}
	c.onOK()
	return revoked, nil
}

// setSessionKeepTTL 写回会话记录并保留原有剩余 TTL
func (c *Cache) setSessionKeepTTL(rec *SessionRecord) error {
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	if err := c.rdb.SetArgs(context.Background(), sessionKey(rec.SessionID), raw, redis.SetArgs{KeepTTL: true}).Err(); err != nil {
		c.onError("set_session", err)
		return err
	}
	c.onOK()
	return nil
}

// ---------- Refresh Token ----------

// SaveRefreshToken 写入令牌记录并登记会话令牌索引（token family），TTL 与令牌过期时间对齐
func (c *Cache) SaveRefreshToken(rec *RefreshTokenRecord, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Until(rec.ExpiredAt)
	}
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	ctx := context.Background()
	idx := sessionTokenKey(rec.SessionID)
	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, refreshTokenKey(rec.TokenHash), raw, ttl)
	pipe.SAdd(ctx, idx, rec.TokenHash)
	pipe.Expire(ctx, idx, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		c.onError("save_refresh_token", err)
		return err
	}
	c.onOK()
	return nil
}

// GetRefreshToken 按 token 哈希读取记录；不存在返回 ErrRecordNotFound
func (c *Cache) GetRefreshToken(tokenHash string) (*RefreshTokenRecord, error) {
	var rec RefreshTokenRecord
	if err := c.getJSON(refreshTokenKey(tokenHash), &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// UpdateRefreshTokenStatus 更新令牌状态并保留剩余 TTL；updated_at 同时充当 used_at
func (c *Cache) UpdateRefreshTokenStatus(tokenHash string, status int) error {
	rec, err := c.GetRefreshToken(tokenHash)
	if err != nil {
		return err
	}
	if rec.Status == status {
		return nil
	}
	rec.Status = status
	rec.UpdatedAt = time.Now()
	return c.setRefreshTokenKeepTTL(rec)
}

// RevokeSessionTokens 撤销指定会话（token family）下全部 active 令牌
func (c *Cache) RevokeSessionTokens(sessionID string, status int) error {
	ctx := context.Background()
	idx := sessionTokenKey(sessionID)
	tokenHashes, err := c.rdb.SMembers(ctx, idx).Result()
	if err != nil {
		c.onError("list_session_tokens", err)
		return err
	}

	now := time.Now()
	for _, tokenHash := range tokenHashes {
		rec, err := c.GetRefreshToken(tokenHash)
		if errors.Is(err, ErrRecordNotFound) {
			// 令牌已过期，顺带清理索引残留
			_ = c.rdb.SRem(ctx, idx, tokenHash)
			continue
		}
		if err != nil || rec.Status != consts.RefreshTokenStatusActive {
			continue
		}
		rec.Status = status
		rec.UpdatedAt = now
		_ = c.setRefreshTokenKeepTTL(rec)
	}
	c.onOK()
	return nil
}

// RevokeUserTokens 撤销用户全部 active 令牌（遍历其全部会话的令牌索引）
func (c *Cache) RevokeUserTokens(userID uint64, status int) error {
	ctx := context.Background()
	sessionIDs, err := c.rdb.SMembers(ctx, userSessionKey(userID)).Result()
	if err != nil {
		c.onError("list_user_sessions", err)
		return err
	}
	for _, sessionID := range sessionIDs {
		if err := c.RevokeSessionTokens(sessionID, status); err != nil {
			return err
		}
	}
	return nil
}

// setRefreshTokenKeepTTL 写回令牌记录并保留原有剩余 TTL
func (c *Cache) setRefreshTokenKeepTTL(rec *RefreshTokenRecord) error {
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	if err := c.rdb.SetArgs(context.Background(), refreshTokenKey(rec.TokenHash), raw, redis.SetArgs{KeepTTL: true}).Err(); err != nil {
		c.onError("set_refresh_token", err)
		return err
	}
	c.onOK()
	return nil
}

// ---------- 通用 ----------

// getJSON 读取并反序列化 JSON 记录；键不存在返回 ErrRecordNotFound
func (c *Cache) getJSON(key string, dst any) error {
	raw, err := c.rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrRecordNotFound
		}
		c.onError("get_json", err)
		return err
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return err
	}
	c.onOK()
	return nil
}
