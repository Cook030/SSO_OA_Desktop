-- 释放连接在线状态：仅当 key 仍属于本连接时才删除，避免误删新建立的连接。
--
-- KEYS: 1 ws:connections:{sessionId}
-- ARGV: 1 connectionId
--
-- 返回：删除的 key 数量（1 释放成功；0 key 不存在或已被新连接接管）

-- redis / cjson / KEYS / ARGV 由 Redis 在脚本执行期注入，静态分析无法感知
---@diagnostic disable: undefined-global

local raw = redis.call("GET", KEYS[1])
if not raw then return 0 end
local ok, rec = pcall(cjson.decode, raw)
if ok and type(rec) == "table" and rec["connectionId"] == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
