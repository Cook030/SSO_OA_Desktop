// Package cache 封装 Redis 缓存、并发锁与登录限流。
//
// Redis 不是权威数据源，仅用于加速与防护；所有操作失败均降级处理
// （缓存未命中回源 MySQL、限流放行、锁降级放行），保证核心认证流程可用。
package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"sync/atomic"
	"time"

	"mh-sso-svc/internal/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// SessionCacheData sso:session:{sessionId} 缓存内容
type SessionCacheData struct {
	UserID          uint64 `json:"userId"`
	Status          int    `json:"status"`
	PasswordVersion int    `json:"passwordVersion"`
}

// IntrospectCacheData sso:introspect:{accessTokenHash} 缓存内容
type IntrospectCacheData struct {
	UserID          uint64 `json:"userId"`
	SessionID       string `json:"sessionId"`
	PasswordVersion int    `json:"passwordVersion"`
	Valid           bool   `json:"valid"`
}

// 登录失败限流与刷新锁参数
const (
	loginFailAccountLimit = 5               // 同账号 5 分钟失败 >= 5 次锁定
	loginFailIPLimit      = 20              // 同 IP 5 分钟失败 >= 20 次限制
	loginFailWindow       = 5 * time.Minute // 失败计数窗口
	refreshLockTTL        = 5 * time.Second // refresh 并发锁 TTL
)

// Cache Redis 客户端封装
type Cache struct {
	rdb      *redis.Client
	log      *zap.Logger
	degraded atomic.Bool // 是否处于降级提示状态（避免异常时日志刷屏）
}

// NewCache 初始化 Redis 客户端；连接失败仅告警不阻断启动（降级模式）
func NewCache(cfg *utils.RedisConfig, log *zap.Logger) *Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	c := &Cache{rdb: rdb, log: log}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warn("Redis 连接失败，缓存/限流/并发锁将降级运行（核心认证走 MySQL）",
			zap.String("addr", cfg.Addr), zap.Error(err))
	} else {
		log.Info("Redis 连接成功", zap.String("addr", cfg.Addr))
	}
	return c
}

// Close 关闭连接
func (c *Cache) Close() error {
	return c.rdb.Close()
}

// onError 记录 Redis 异常；降级期间只提示一次，恢复后再次异常会重新提示
func (c *Cache) onError(op string, err error) {
	if err == nil || err == redis.Nil {
		return
	}
	if c.degraded.Swap(true) {
		return
	}
	c.log.Warn("Redis 操作失败，本次降级处理", zap.String("op", op), zap.Error(err))
}

// onOK 操作成功时清除降级提示状态
func (c *Cache) onOK() {
	if c.degraded.Load() {
		c.degraded.Store(false)
	}
}

// ---------- Session 缓存 ----------

// GetSessionCache 读取 session 缓存；未命中或 Redis 异常时返回 ok=false（调用方回源 MySQL）
func (c *Cache) GetSessionCache(sessionID string) (SessionCacheData, bool) {
	var data SessionCacheData
	raw, err := c.rdb.Get(context.Background(), sessionKey(sessionID)).Bytes()
	if err != nil {
		c.onError("get_session", err)
		return data, false
	}
	c.onOK()
	if err := json.Unmarshal(raw, &data); err != nil {
		return data, false
	}
	return data, true
}

// SetSessionCache 写入 session 缓存，TTL 与会话过期时间对齐
func (c *Cache) SetSessionCache(sessionID string, data SessionCacheData, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	if err := c.rdb.Set(context.Background(), sessionKey(sessionID), raw, ttl).Err(); err != nil {
		c.onError("set_session", err)
		return
	}
	c.onOK()
}

// DeleteSessionCache 删除 session 缓存（登出/撤销/改密时调用）
func (c *Cache) DeleteSessionCache(sessionID string) {
	if err := c.rdb.Del(context.Background(), sessionKey(sessionID)).Err(); err != nil {
		c.onError("del_session", err)
		return
	}
	c.onOK()
}

// ---------- Introspect 缓存 ----------

// GetIntrospectCache 读取 introspect 结果缓存；未命中返回 ok=false
func (c *Cache) GetIntrospectCache(tokenHash string) (IntrospectCacheData, bool) {
	var data IntrospectCacheData
	raw, err := c.rdb.Get(context.Background(), introspectKey(tokenHash)).Bytes()
	if err != nil {
		c.onError("get_introspect", err)
		return data, false
	}
	c.onOK()
	if err := json.Unmarshal(raw, &data); err != nil {
		return data, false
	}
	return data, true
}

// SetIntrospectCache 写入 introspect 结果缓存（TTL 不超过 30 秒）
func (c *Cache) SetIntrospectCache(tokenHash string, data IntrospectCacheData, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	if err := c.rdb.Set(context.Background(), introspectKey(tokenHash), raw, ttl).Err(); err != nil {
		c.onError("set_introspect", err)
		return
	}
	c.onOK()
}

// ---------- 用户密码版本缓存 ----------

// GetPasswordVersion 读取用户当前密码版本；未命中返回 ok=false
func (c *Cache) GetPasswordVersion(userID uint64) (int, bool) {
	val, err := c.rdb.Get(context.Background(), passwordVersionKey(userID)).Int()
	if err != nil {
		c.onError("get_password_version", err)
		return 0, false
	}
	c.onOK()
	return val, true
}

// SetPasswordVersion 写入用户当前密码版本
func (c *Cache) SetPasswordVersion(userID uint64, version int, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	if err := c.rdb.Set(context.Background(), passwordVersionKey(userID), version, ttl).Err(); err != nil {
		c.onError("set_password_version", err)
		return
	}
	c.onOK()
}

// DeletePasswordVersion 删除用户密码版本缓存（改密后调用，强制下次回源）
func (c *Cache) DeletePasswordVersion(userID uint64) {
	if err := c.rdb.Del(context.Background(), passwordVersionKey(userID)).Err(); err != nil {
		c.onError("del_password_version", err)
		return
	}
	c.onOK()
}

// ---------- 登录失败限流 ----------

// IsLoginRateLimited 判断账号或 IP 是否触发登录失败限流；
// Redis 异常时放行（降级），保证核心登录可用
func (c *Cache) IsLoginRateLimited(account, ip string) bool {
	ctx := context.Background()

	if n, err := c.rdb.Get(ctx, loginFailAccountKey(account)).Int(); err == nil {
		if n >= loginFailAccountLimit {
			return true
		}
	} else if err != redis.Nil {
		c.onError("get_login_fail_account", err)
	}

	if ip != "" {
		if n, err := c.rdb.Get(ctx, loginFailIPKey(ip)).Int(); err == nil {
			if n >= loginFailIPLimit {
				return true
			}
		} else if err != redis.Nil {
			c.onError("get_login_fail_ip", err)
		}
	}
	c.onOK()
	return false
}

// RecordLoginFailure 记录一次登录失败（5 分钟窗口内累计）
func (c *Cache) RecordLoginFailure(account, ip string) {
	ctx := context.Background()
	c.incrWithTTL(ctx, loginFailAccountKey(account))
	if ip != "" {
		c.incrWithTTL(ctx, loginFailIPKey(ip))
	}
}

// ClearLoginFailures 登录成功后清理账号失败计数
func (c *Cache) ClearLoginFailures(account string) {
	if err := c.rdb.Del(context.Background(), loginFailAccountKey(account)).Err(); err != nil {
		c.onError("clear_login_fail", err)
		return
	}
	c.onOK()
}

// incrWithTTL 自增计数，首次自增时设置窗口 TTL
func (c *Cache) incrWithTTL(ctx context.Context, key string) {
	n, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		c.onError("incr_login_fail", err)
		return
	}
	if n == 1 {
		if err := c.rdb.Expire(ctx, key, loginFailWindow).Err(); err != nil {
			c.onError("expire_login_fail", err)
			return
		}
	}
	c.onOK()
}

// ---------- Refresh 并发锁 ----------

// AcquireRefreshLock 以 SET key requestId NX EX 5 方式获取 refresh 并发锁；
// 拿不到锁返回 false；Redis 异常时降级放行（返回 true），保证核心刷新流程可用
func (c *Cache) AcquireRefreshLock(tokenHash, requestID string) bool {
	ok, err := c.rdb.SetNX(context.Background(), refreshLockKey(tokenHash), requestID, refreshLockTTL).Result()
	if err != nil {
		c.onError("acquire_refresh_lock", err)
		return true
	}
	c.onOK()
	return ok
}

// ReleaseRefreshLock 释放并发锁（仅删除自己持有的锁，避免误删其他请求的锁）
func (c *Cache) ReleaseRefreshLock(tokenHash, requestID string) {
	const delIfOwner = `if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) else return 0 end`
	if err := c.rdb.Eval(context.Background(), delIfOwner, []string{refreshLockKey(tokenHash)}, requestID).Err(); err != nil {
		c.onError("release_refresh_lock", err)
		return
	}
	c.onOK()
}

// ---------- Key 构造 ----------

func sessionKey(sessionID string) string        { return "sso:session:" + sessionID }
func introspectKey(tokenHash string) string     { return "sso:introspect:" + tokenHash }
func refreshLockKey(tokenHash string) string    { return "sso:refresh_lock:" + tokenHash }
func loginFailAccountKey(account string) string { return "sso:login_fail:account:" + account }
func loginFailIPKey(ip string) string           { return "sso:login_fail:ip:" + ip }
func passwordVersionKey(userID uint64) string {
	return "sso:user:password_version:" + strconv.FormatUint(userID, 10)
}
