-- Redis Lua scripts for session lifecycle operations.
-- Go injects the key prefixes into the three %s placeholders of each script.

-- loginReplaceLua
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

redis.call("SET", KEYS[4], newSessionJson, "EX", sessionTTL)
redis.call("SET", KEYS[5], newRtJson, "EX", refreshTTL)
redis.call("SADD", KEYS[6], newRtHash)
redis.call("EXPIRE", KEYS[6], refreshTTL)
redis.call("DEL", KEYS[1])
redis.call("SADD", KEYS[1], newSid)
redis.call("EXPIRE", KEYS[1], sessionTTL)
redis.call("SET", KEYS[2], newSid, "EX", sessionTTL)

for i = 1, #replaced do
  redis.call("XADD", KEYS[3], "MAXLEN", "~", streamMaxLen, "*",
    "eventId", eventIdBase .. ":" .. i, "type", eventType,
    "userId", userId, "targetSessionId", replaced[i], "reason", reason,
    "createdAt", nowMs)
end
local out = {1, newSid, #replaced}
for i = 1, #replaced do out[3 + i] = replaced[i] end
return out

-- __SCRIPT_SEPARATOR__

-- revokeUserLua
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
redis.call("DEL", KEYS[2])
for i = 1, #revoked do
  redis.call("XADD", KEYS[3], "MAXLEN", "~", streamMaxLen, "*",
    "eventId", eventIdBase .. ":" .. i, "type", eventType,
    "userId", userId, "targetSessionId", revoked[i], "reason", reason,
    "createdAt", nowMs)
end
local out = {#revoked}
for i = 1, #revoked do out[i + 1] = revoked[i] end
return out

-- __SCRIPT_SEPARATOR__

-- revokeSessionLua
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
if redis.call("GET", KEYS[3]) == sid then redis.call("DEL", KEYS[3]) end
redis.call("SREM", KEYS[4], sid)
redis.call("XADD", KEYS[5], "MAXLEN", "~", streamMaxLen, "*",
  "eventId", eventId, "type", eventType, "userId", userId,
  "targetSessionId", sid, "reason", reason, "createdAt", nowMs)
return 1
