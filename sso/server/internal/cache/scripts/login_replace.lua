-- 登录时的会话置换：撤销全部旧会话及其 active refresh token → 写入新会话与新 token
-- → 重置用户会话索引与 current_session → 为每个被替换的旧会话写下线事件。
--
-- Go 通过 fmt.Sprintf 注入三个 %s 占位符：session / refresh token / token family 前缀。
--
-- KEYS:
--   1 sso:user_sessions:{userId}        用户会话集合
--   2 sso:user:{userId}:current_session 唯一有效会话
--   3 sso:stream:session-events         会话事件流
--   4 sso:session:{newSessionId}        新会话
--   5 sso:rt:{newRtHash}                新 refresh token
--   6 sso:rt_family:{newSessionId}      新会话的 token family
--
-- ARGV:
--    1 mode              replace | reject
--    2 newSessionId
--    3 newSessionJson
--    4 sessionTtlSeconds
--    5 newRtHash
--    6 newRtJson
--    7 refreshTtlSeconds
--    8 eventIdBase
--    9 eventType
--   10 reason
--   11 oldSessionStatus
--   12 oldTokenStatus
--   13 streamMaxLen
--   14 updatedAt (RFC3339Nano)
--   15 activeSessionStatus
--   16 activeTokenStatus
--   17 userId
--   18 nowMs
--
-- 返回（Lua 数组）：
--   {1, newSessionId, replacedCount, replacedSessionId...} 成功
--   {2, "", 0}                                            reject 模式下已存在有效会话

-- redis / cjson / KEYS / ARGV 由 Redis 在脚本执行期注入，静态分析无法感知
---@diagnostic disable: undefined-global

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
