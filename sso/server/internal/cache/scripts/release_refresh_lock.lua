-- 释放 refresh 并发锁：仅当锁仍由本请求持有时才删除，避免误删其他请求的锁。
--
-- KEYS: 1 sso:refresh_lock:{tokenHash}
-- ARGV: 1 requestId
--
-- 返回：删除的 key 数量（1 释放成功；0 锁已被其他请求持有或已过期）

-- redis / KEYS / ARGV 由 Redis 在脚本执行期注入，静态分析无法感知
---@diagnostic disable: undefined-global

if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
