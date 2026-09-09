package cache

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"mh-sso-svc/internal/consts"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// 本文件是单设备登录的正确性核心：
// 登录时的"撤销旧会话 + 写入新会话 + 写下线事件"必须是一个不可分割的操作，
// 否则会出现"状态已变但事件未写"（旧设备不知道自己被踢）或
// "事件已写但状态未变"（旧 access token 仍然可用）的半成功状态。
//
// 采用 Redis Lua 单脚本实现：Redis 单线程执行脚本，脚本内任一步失败整体回滚。
// 因此该会话模型要求所有相关键位于同一 Redis 主节点；跨 slot 脚本会破坏原子性。
//
// 脚本内需要按 sessionId / tokenHash 动态拼键，故通过 fmt.Sprintf 注入
// redis.go 中定义的键前缀，保证 Go 侧与 Lua 侧的键名一致。

// Lua 脚本独立保存在 scripts/ 目录（一段脚本一个文件），以获得语法高亮和更易维护的
// 结构；编译时嵌入二进制，运行时语义与内嵌字符串完全一致。
//
// loginReplaceLua 登录时的会话置换脚本。
//
//go:embed scripts/login_replace.lua
var loginReplaceLua string

// revokeUserLua 撤销用户全部会话并广播下线事件（改密、管理员撤销场景）。
//
//go:embed scripts/revoke_user.lua
var revokeUserLua string

// revokeSessionLua 撤销单个会话并写终止事件。它仅在 current_session
// 恰好是目标 session 时才删除 current_session，因而旧 token 的登出或
// 重放请求不会误伤后来登录产生的新会话。
//
//go:embed scripts/revoke_session.lua
var revokeSessionLua string

var (
	loginReplaceScript  = redis.NewScript(fmt.Sprintf(loginReplaceLua, sessionKeyPrefix, refreshTokenKeyPrefix, sessionTokenKeyPrefix))
	revokeUserScript    = redis.NewScript(fmt.Sprintf(revokeUserLua, sessionKeyPrefix, refreshTokenKeyPrefix, sessionTokenKeyPrefix))
	revokeSessionScript = redis.NewScript(fmt.Sprintf(revokeSessionLua, sessionKeyPrefix, refreshTokenKeyPrefix, sessionTokenKeyPrefix))
)

// SessionEventInput 写入会话事件流的事件描述
type SessionEventInput struct {
	EventType string // 见 consts.EventTypeSession*
	Reason    string // 见 consts.Reason*
}

// LoginReplaceInput 登录时的原子会话置换入参
type LoginReplaceInput struct {
	UserID       uint64
	Mode         consts.LoginMode
	Session      *SessionRecord
	SessionTTL   time.Duration
	RefreshToken *RefreshTokenRecord
	RefreshTTL   time.Duration
	Event        SessionEventInput
}

// LoginReplaceResult 原子会话置换结果
type LoginReplaceResult struct {
	SessionID          string   // 新会话 ID
	ReplacedSessionIDs []string // 被顶下线的旧会话 ID
	Rejected           bool     // reject 模式下因已存在有效会话而拒绝
}

// ReplaceLoginSession 原子执行单设备登录的会话置换。
//
// 一次 EVAL 内完成：撤销旧会话与其 refresh token → 写入新会话与新 token →
// 重置用户会话索引与 current_session → 为每个旧会话写下线事件。
// 返回 error 时 Redis 状态未发生任何变更（脚本整体失败）。
func (c *Cache) ReplaceLoginSession(ctx context.Context, in LoginReplaceInput) (*LoginReplaceResult, error) {
	sessionJSON, err := json.Marshal(in.Session)
	if err != nil {
		return nil, err
	}
	rtJSON, err := json.Marshal(in.RefreshToken)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	keys := []string{
		userSessionKey(in.UserID),
		currentSessionKey(in.UserID),
		consts.SessionEventStreamKey,
		sessionKey(in.Session.SessionID),
		refreshTokenKey(in.RefreshToken.TokenHash),
		sessionTokenKey(in.Session.SessionID),
	}
	args := []interface{}{
		string(in.Mode),
		in.Session.SessionID,
		string(sessionJSON),
		int(in.SessionTTL.Seconds()),
		in.RefreshToken.TokenHash,
		string(rtJSON),
		int(in.RefreshTTL.Seconds()),
		uuid.NewString(),
		in.Event.EventType,
		in.Event.Reason,
		consts.SessionStatusReplaced,
		consts.RefreshTokenStatusRevoked,
		consts.StreamMaxLen,
		now.Format(time.RFC3339Nano),
		consts.SessionStatusActive,
		consts.RefreshTokenStatusActive,
		strconv.FormatUint(in.UserID, 10),
		strconv.FormatInt(now.UnixMilli(), 10),
	}

	res, err := loginReplaceScript.Run(ctx, c.rdb, keys, args...).Result()
	if err != nil {
		c.onError("replace_login_session", err)
		return nil, err
	}
	c.onOK()

	arr, ok := res.([]interface{})
	if !ok || len(arr) < 3 {
		return nil, fmt.Errorf("cache: 会话置换脚本返回结构异常: %v", res)
	}
	status, ok := toInt64(arr[0])
	if !ok {
		return nil, fmt.Errorf("cache: 会话置换脚本返回状态码异常: %v", arr[0])
	}
	if status == 2 {
		return &LoginReplaceResult{Rejected: true}, nil
	}
	replaced, err := toStringSlice(arr, 3)
	if err != nil {
		return nil, err
	}
	return &LoginReplaceResult{
		SessionID:          in.Session.SessionID,
		ReplacedSessionIDs: replaced,
	}, nil
}

// RevokeUserSessionsWithEvents 撤销用户全部 active 会话与 refresh token，
// 并广播下线事件；返回被撤销的会话 ID 列表。
func (c *Cache) RevokeUserSessionsWithEvents(ctx context.Context, userID uint64, sessionStatus int, ev SessionEventInput) ([]string, error) {
	now := time.Now()
	keys := []string{
		userSessionKey(userID),
		currentSessionKey(userID),
		consts.SessionEventStreamKey,
	}
	args := []interface{}{
		uuid.NewString(),
		ev.EventType,
		ev.Reason,
		sessionStatus,
		consts.RefreshTokenStatusRevoked,
		consts.SessionStatusActive,
		consts.RefreshTokenStatusActive,
		strconv.FormatUint(userID, 10),
		consts.StreamMaxLen,
		now.Format(time.RFC3339Nano),
		strconv.FormatInt(now.UnixMilli(), 10),
	}

	res, err := revokeUserScript.Run(ctx, c.rdb, keys, args...).Result()
	if err != nil {
		c.onError("revoke_user_sessions_with_events", err)
		return nil, err
	}
	c.onOK()

	arr, ok := res.([]interface{})
	if !ok || len(arr) < 1 {
		return nil, fmt.Errorf("cache: 撤销会话脚本返回结构异常: %v", res)
	}
	return toStringSlice(arr, 1)
}

// RevokeSessionWithEvent 原子撤销一个指定会话、清理其 token，并写入终止事件。
// 若该 session 已不再 active，返回 false；这保证陈旧凭证不会影响当前会话。
func (c *Cache) RevokeSessionWithEvent(ctx context.Context, userID uint64, sessionID string, sessionStatus int, ev SessionEventInput) (bool, error) {
	now := time.Now()
	keys := []string{
		sessionKey(sessionID),
		sessionTokenKey(sessionID),
		currentSessionKey(userID),
		userSessionKey(userID),
		consts.SessionEventStreamKey,
	}
	args := []interface{}{
		sessionID,
		sessionStatus,
		consts.RefreshTokenStatusRevoked,
		consts.SessionStatusActive,
		consts.RefreshTokenStatusActive,
		uuid.NewString(),
		ev.EventType,
		ev.Reason,
		strconv.FormatUint(userID, 10),
		consts.StreamMaxLen,
		now.Format(time.RFC3339Nano),
		strconv.FormatInt(now.UnixMilli(), 10),
	}
	res, err := revokeSessionScript.Run(ctx, c.rdb, keys, args...).Int64()
	if err != nil {
		c.onError("revoke_session_with_event", err)
		return false, err
	}
	c.onOK()
	return res == 1, nil
}

// GetCurrentSession 读取用户当前唯一有效会话 ID；不存在返回空字符串
func (c *Cache) GetCurrentSession(ctx context.Context, userID uint64) (string, error) {
	val, err := c.rdb.Get(ctx, currentSessionKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		c.onError("get_current_session", err)
		return "", err
	}
	c.onOK()
	return val, nil
}

// toInt64 将 Lua 返回的数值转换为 int64
func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case int:
		return int64(n), true
	default:
		return 0, false
	}
}

// toStringSlice 从 Lua 返回数组的 offset 位置起解析字符串列表
func toStringSlice(arr []interface{}, offset int) ([]string, error) {
	out := make([]string, 0, len(arr)-offset)
	for i := offset; i < len(arr); i++ {
		s, ok := arr[i].(string)
		if !ok {
			return nil, fmt.Errorf("cache: 脚本返回元素类型异常 index=%d: %v", i, arr[i])
		}
		out = append(out, s)
	}
	return out, nil
}
