package utils

// 业务状态码（与接口文档保持一致）
const (
	CodeOK           = 200
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeConflict     = 409
	CodeServerError  = 500
)

// ErrMsgServerError 对外统一的服务器内部错误提示。
// 内部错误细节只写入日志（参照 PKG/模块 + 错误链），不回传给客户端，避免信息泄露。
const ErrMsgServerError = "服务器内部错误"

// BizError 业务错误：handler 据此返回对应业务码与提示。
// Reason 是给客户端程序消费的机器码（如 SESSION_REPLACED），
// 与 Msg 分离，避免把面向用户的文案当成判断依据。
type BizError struct {
	Code   int
	Msg    string
	Reason string
}

func (e *BizError) Error() string { return e.Msg }

// NewBizError 创建业务错误
func NewBizError(code int, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

// NewBizErrorWithReason 创建带机器原因码的业务错误
func NewBizErrorWithReason(code int, msg, reason string) *BizError {
	return &BizError{Code: code, Msg: msg, Reason: reason}
}
