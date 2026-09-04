package middleware

import (
	"time"

	"permission-system/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDKey gin.Context 中 request_id 的键
const RequestIDKey = "request_id"

// RequestIDMiddleware 为整个 HTTP 请求生成唯一 request_id 并注入上下文，
// 后续认证、业务处理、SSO 调用等全部复用该 ID，便于生产环境按 request_id 串联请求链路。
// 同时记录 HTTP 请求开始 / 完成日志（健康检查等探活路径除外）。
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)
		log := utils.GetLogger().With(zap.String("request_id", requestID))

		// 健康检查等高频探活路径不记录访问日志，避免刷屏
		if skipAccessLog(c.Request.URL.Path) {
			c.Next()
			return
		}

		log.Info("HTTP 请求开始",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		start := time.Now()
		c.Next()

		// user_id 由 AuthMiddleware 写入 context；未认证请求为 0
		log.Info("HTTP 请求完成",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("latency_ms", time.Since(start).Milliseconds()),
			zap.Int64("user_id", c.GetInt64(ContextKeyUserID)),
		)
	}
}

// skipAccessLog 命中路径时不输出 HTTP 请求开始/完成日志
func skipAccessLog(path string) bool {
	switch path {
	case "/test/alive":
		return true
	}
	return false
}

// GetRequestID 从 gin.Context 读取 request_id；未设置时返回空字符串
func GetRequestID(c *gin.Context) string {
	requestID, exists := c.Get(RequestIDKey)
	if !exists {
		return ""
	}
	id, ok := requestID.(string)
	if !ok {
		return ""
	}
	return id
}
