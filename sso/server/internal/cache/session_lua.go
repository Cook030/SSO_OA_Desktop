package cache

import (
	"context"
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
// 这也是 dev.md 明确不使用 Redis Cluster 的原因 —— 跨 slot 脚本会破坏原子性。
//
// 脚本内需要按 sessionId / tokenHash 动态拼键，故通过 fmt.Sprintf 注入
// redis.go 中定义的键前缀，保证 Go 侧与 Lua 侧的键名一致。

// loginReplaceLua 登录时的会话置换脚本。
//
// KEYS:
//
//	1 sso:user_sessions:{userId}      用户会话集合
//	2 sso:user:{userId}:current_session 唯一有效会话
//	3 sso:stream:session-events       会话事件流
//	4 sso:session:{newSessionId}      新会话
//	5 sso:rt:{newRtHash}              新 refresh token
//	6 sso:rt_family:{newSessionId}    新会话的 token family
//
// ARGV:
//
//	 1 mode              replace | reject
//	 2 newSessionId
//	 3 newSessionJson
//	 4 sessionTtlSeconds
//	 5 newRtHash
//	 6 newRtJson
//	 7 refreshTtlSeconds
//	 8 eventIdBase
//	 9 eventType
//	10 reason
//	11 oldSessionStatus
//	12 oldTokenStatus
//	13 streamMaxLen
//	14 updatedAt (RFC3339Nano)
//	15 activeSessionStatus
//	16 activeTokenStatus
//	17 userId
//	18 nowMs
//
// 返回（Lua 数组）：
//
//	{1, newSessionId, replacedCount, replacedSessionId...} 成功
//	{2, "", 0}                                            reject 模式下已存在有效会话
const loginReplaceLua = `
local SESSION_PREF = "%s"
local RT_PREF = "%s"
local FAMILY_PREF = "%s"

local mode = ARGV[1]
local newSid = ARGV[2]
local newSessionJson = ARGV[3]
local sessionTTL = tonumber(ARGV[4])
local newRtHash = ARGV[5]
local newRtJson = ARGV[6]
local refreshTTL = tonumber(ARGV[7])
local eventIdBase = ARGV[8]
local eventType = ARGV[9]
local reason = ARGV[10]
local oldSessionStatus = tonumber(ARGV[11])
local oldTokenStatus = tonumber(ARGV[12])
local streamMaxLen = tonumber(ARGV[13])
local updatedAt = ARGV[14]
local activeSessionStatus = tonumber(ARGV[15])
local activeTokenStatus = tonumber(ARGV[16])
local userId = ARGV[17]
local nowMs = ARGV[18]

-- reject 模式：存在有效当前会话则整体拒绝，不修改任何状态
if mode == "reject" then
  local cur = redis.call("GET", KEYS[2])
  if cur then
    local raw = redis.call("GET", SESSION_PREF .. cur)
    if raw then
      local ok, rec = pcall(cjson.decode, raw)
      if ok and type(rec) == "table" and rec["status"] == activeSessionStatus then
        return {2, "", 0}
      end
    end
  end
end

-- 1. 撤销全部旧会话及其 active refresh token
local oldSids = redis.call("SMEMBERS", KEYS[1])
local replaced = {}

for i = 1, #oldSids do
  local sid = oldSids[i]
  local raw = redis.call("GET", SESSION_PREF .. sid)
  if raw then
    local ok, rec = pcall(cjson.decode, raw)
    if ok and type(rec) == "table" then
      local hashes = redis.call("SMEMBERS", FAMILY_PREF .. sid)
      for j = 1, #hashes do
        local rtRaw = redis.call("GET", RT_PREF .. hashes[j])
        if rtRaw then
          local okRt, rt = pcall(cjson.decode, rtRaw)
          if okRt and type(rt) == "table" and rt["status"] == activeTokenStatus then
            rt["status"] = oldTokenStatus
            rt["updatedAt"] = updatedAt
            redis.call("SET", RT_PREF .. hashes[j], cjson.encode(rt), "KEEPTTL")
          end
        end
      end
      if rec["status"] == activeSessionStatus then
        rec["status"] = oldSessionStatus
        rec["updatedAt"] = updatedAt
        redis.call("SET", SESSION_PREF .. sid, cjson.encode(rec), "KEEPTTL")
        replaced[#replaced + 1] = sid
      end
    end
  end
end

-- 2. 写入新会话与新 refresh token
redis.call("SET", KEYS[4], newSessionJson, "EX", sessionTTL)
redis.call("SET", KEYS[5], newRtJson, "EX", refreshTTL)
redis.call("SADD", KEYS[6], newRtHash)
redis.call("EXPIRE", KEYS[6], refreshTTL)

-- 3. 重置用户会话索引与唯一有效会话
redis.call("DEL", KEYS[1])
redis.call("SADD", KEYS[1], newSid)
redis.call("EXPIRE", KEYS[1], sessionTTL)
redis.call("SET", KEYS[2], newSid, "EX", sessionTTL)

-- 4. 为每个被替换的会话写下线事件：与状态变更同一次 EVAL，保证"状态变则事件必在"
for i = 1, #replaced do
  redis.call("XADD", KEYS[3], "MAXLEN", "~", streamMaxLen, "*",
    "eventId", eventIdBase .. ":" .. i,
    "type", eventType,
    "userId", userId,
    "targetSessionId", replaced[i],
    "reason", reason,
    "createdAt", nowMs)
end

local out = {1, newSid, #replaced}
for i = 1, #replaced do
  out[3 + i] = replaced[i]
end
return out
`

// revokeUserLua 撤销用户全部会话并广播下线事件（改密、管理员撤销场景）。
//
// KEYS: 同 loginReplaceLua 的 1/2/3。
// ARGV:
//
//	 1 eventIdBase
//	 2 eventType
//	 3 reason
//	 4 sessionStatus
//	 5 tokenStatus
//	 6 activeSessionStatus
//	 7 activeTokenStatus
//	 8 userId
//	 9 streamMaxLen
//	10 updatedAt (RFC3339Nano)
//	11 nowMs
//
// 返回：{revokedCount, revokedSessionId...}
const revokeUserLua = `
local SESSION_PREF = "%s"
local RT_PREF = "%s"
local FAMILY_PREF = "%s"

local eventIdBase = ARGV[1]
local eventType = ARGV[2]
local reason = ARGV[3]
local sessionStatus = tonumber(ARGV[4])
local tokenStatus = tonumber(ARGV[5])
local activeSessionStatus = tonumber(ARGV[6])
local activeTokenStatus = tonumber(ARGV[7])
local userId = ARGV[8]
local streamMaxLen = tonumber(ARGV[9])
local updatedAt = ARGV[10]
local nowMs = ARGV[11]

local sids = redis.call("SMEMBERS", KEYS[1])
local revoked = {}

for i = 1, #sids do
  local sid = sids[i]
  local raw = redis.call("GET", SESSION_PREF .. sid)
  if raw then
    local ok, rec = pcall(cjson.decode, raw)
    if ok and type(rec) == "table" then
      local hashes = redis.call("SMEMBERS", FAMILY_PREF .. sid)
      for j = 1, #hashes do
        local rtRaw = redis.call("GET", RT_PREF .. hashes[j])
        if rtRaw then
          local okRt, rt = pcall(cjson.decode, rtRaw)
          if okRt and type(rt) == "table" and rt["status"] == activeTokenStatus then
            rt["status"] = tokenStatus
            rt["updatedAt"] = updatedAt
            redis.call("SET", RT_PREF .. hashes[j], cjson.encode(rt), "KEEPTTL")
          end
        end
      end
      if rec["status"] == activeSessionStatus then
        rec["status"] = sessionStatus
        rec["updatedAt"] = updatedAt
        redis.call("SET", SESSION_PREF .. sid, cjson.encode(rec), "KEEPTTL")
        revoked[#revoked + 1] = sid
      end
    end
  end
end

-- 当前会话置空：任何 access token 都无法再匹配 current_session
redis.call("DEL", KEYS[2])

for i = 1, #revoked do
  redis.call("XADD", KEYS[3], "MAXLEN", "~", streamMaxLen, "*",
    "eventId", eventIdBase .. ":" .. i,
    "type", eventType,
    "userId", userId,
    "targetSessionId", revoked[i],
    "reason", reason,
    "createdAt", nowMs)
end

local out = {#revoked}
for i = 1, #revoked do
  out[i + 1] = revoked[i]
end
return out
`

// revokeSessionLua 撤销单个会话并写终止事件。它仅在 current_session
// 恰好是目标 session 时才删除 current_session，因而旧 token 的登出或
// 重放请求不会误伤后来登录产生的新会话。
//
// KEYS: session, token-family, current-session, user-session-index, event-stream
const revokeSessionLua = `
local SESSION_PREF = "%s"
local RT_PREF = "%s"
local FAMILY_PREF = "%s"

local sid = ARGV[1]
local sessionStatus = tonumber(ARGV[2])
local tokenStatus = tonumber(ARGV[3])
local activeSessionStatus = tonumber(ARGV[4])
local activeTokenStatus = tonumber(ARGV[5])
local eventId = ARGV[6]
local eventType = ARGV[7]
local reason = ARGV[8]
local userId = ARGV[9]
local streamMaxLen = tonumber(ARGV[10])
local updatedAt = ARGV[11]
local nowMs = ARGV[12]

local raw = redis.call("GET", KEYS[1])
if not raw then return 0 end
local ok, rec = pcall(cjson.decode, raw)
if not ok or type(rec) ~= "table" or rec["status"] ~= activeSessionStatus then return 0 end

local hashes = redis.call("SMEMBERS", KEYS[2])
for i = 1, #hashes do
  local rtRaw = redis.call("GET", RT_PREF .. hashes[i])
  if rtRaw then
    local okRt, rt = pcall(cjson.decode, rtRaw)
    if okRt and type(rt) == "table" and rt["status"] == activeTokenStatus then
      rt["status"] = tokenStatus
      rt["updatedAt"] = updatedAt
      redis.call("SET", RT_PREF .. hashes[i], cjson.encode(rt), "KEEPTTL")
    end
  end
end

rec["status"] = sessionStatus
rec["updatedAt"] = updatedAt
redis.call("SET", KEYS[1], cjson.encode(rec), "KEEPTTL")
if redis.call("GET", KEYS[3]) == sid then
  redis.call("DEL", KEYS[3])
end
redis.call("SREM", KEYS[4], sid)
redis.call("XADD", KEYS[5], "MAXLEN", "~", streamMaxLen, "*",
  "eventId", eventId,
  "type", eventType,
  "userId", userId,
  "targetSessionId", sid,
  "reason", reason,
  "createdAt", nowMs)
return 1
`

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
