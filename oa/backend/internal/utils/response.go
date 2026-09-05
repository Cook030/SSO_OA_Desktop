package utils

import "github.com/gin-gonic/gin"

// Response 统一返回结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// OKWithMsg 成功响应(自定义消息)
func OKWithMsg(c *gin.Context, message string, data interface{}) {
	c.JSON(200, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// Fail 参数错误
func Fail(c *gin.Context, message string) {
	c.JSON(200, Response{
		Code:    400,
		Message: message,
		Data:    nil,
	})
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, message string) {
	c.JSON(200, Response{
		Code:    401,
		Message: message,
		Data:    nil,
	})
}

// Forbidden 无权限
func Forbidden(c *gin.Context, message string) {
	c.JSON(200, Response{
		Code:    403,
		Message: message,
		Data:    nil,
	})
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	c.JSON(200, Response{
		Code:    404,
		Message: message,
		Data:    nil,
	})
}

// ServerError 服务器错误
func ServerError(c *gin.Context, message string) {
	c.JSON(200, Response{
		Code:    500,
		Message: message,
		Data:    nil,
	})
}
