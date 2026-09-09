-- 撤销单个会话、清理其 refresh token，并写入终止事件。
--
-- 仅在 current_session 恰好是目标 session 时才删除 current_session，
-- 因而旧 token 的登出或重放请求不会误伤后来登录产生的新会话。
--
-- Go 通过 fmt.Sprintf 注入三个 %s 占位符：session / refresh token / token family 前缀。
--
-- KEYS:
--   1 sso:session:{sessionId}           目标会话
--   2 sso:rt_family:{sessionId}         该会话的 token family
--   3 sso:user:{userId}:current_session 唯一有效会话
--   4 sso:user_sessions:{userId}        用户会话集合
--   5 sso:stream:session-events         会话事件流
--
-- ARGV:
--    1 sessionId
--    2 sessionStatus
--    3 tokenStatus
--    4 activeSessionStatus
--    5 activeTokenStatus
--    6 eventId
--    7 eventType
--    8 reason
--    9 userId
--   10 streamMaxLen
--   11 updatedAt (RFC3339Nano)
--   12 nowMs
--
-- 返回：1 已撤销；0 会话不存在或已不再 active

-- redis / cjson / KEYS / ARGV 由 Redis 在脚本执行期注入，静态分析无法感知
---@diagnostic disable: undefined-global

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
