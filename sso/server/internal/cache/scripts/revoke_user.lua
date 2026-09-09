-- 撤销用户全部 active 会话与 refresh token，并广播下线事件（改密、管理员撤销场景）。
--
-- Go 通过 fmt.Sprintf 注入三个 %s 占位符：session / refresh token / token family 前缀。
--
-- KEYS:
--   1 sso:user_sessions:{userId}        用户会话集合
--   2 sso:user:{userId}:current_session 唯一有效会话
--   3 sso:stream:session-events         会话事件流
--
-- ARGV:
--    1 eventIdBase
--    2 eventType
--    3 reason
--    4 sessionStatus
--    5 tokenStatus
--    6 activeSessionStatus
--    7 activeTokenStatus
--    8 userId
--    9 streamMaxLen
--   10 updatedAt (RFC3339Nano)
--   11 nowMs
--
-- 返回：{revokedCount, revokedSessionId...}

-- redis / cjson / KEYS / ARGV 由 Redis 在脚本执行期注入，静态分析无法感知
---@diagnostic disable: undefined-global

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
