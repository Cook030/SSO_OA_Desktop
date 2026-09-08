package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一返回结构（SSO 接口文档约定：code/msg/data）
// Reason 为可选的机器原因码（如 SESSION_REPLACED），omitempty 保证既有接口响应结构不变
type Response struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Reason string      `json:"reason,omitempty"`
}

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

// Error 按业务码返回错误（HTTP 状态码与业务码一致）
func Error(c *gin.Context, code int, msg string) {
	ErrorWithReason(c, code, msg, "")
}

// ErrorWithReason 按业务码返回错误，并附加机器原因码供客户端程序化判断
func ErrorWithReason(c *gin.Context, code int, msg, reason string) {
	c.JSON(code, Response{
		Code:   code,
		Msg:    msg,
		Data:   nil,
		Reason: reason,
	})
}

// BadRequest 请求参数错误
func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, msg)
}

// Unauthorized Token 缺失、无效或已过期
func Unauthorized(c *gin.Context, msg string) {
	Error(c, http.StatusUnauthorized, msg)
}

// Conflict 资源冲突（如用户已存在）
func Conflict(c *gin.Context, msg string) {
	Error(c, http.StatusConflict, msg)
}

// ServerError 服务器内部错误
func ServerError(c *gin.Context, msg string) {
	Error(c, http.StatusInternalServerError, msg)
}
