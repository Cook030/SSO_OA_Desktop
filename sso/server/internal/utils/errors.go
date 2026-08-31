package utils

// 业务状态码（与接口文档保持一致）
const (
	CodeOK           = 200
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeConflict     = 409
	CodeServerError  = 500
)

// BizError 业务错误：handler 据此返回对应业务码与提示
type BizError struct {
	Code int
	Msg  string
}

func (e *BizError) Error() string { return e.Msg }

// NewBizError 创建业务错误
func NewBizError(code int, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}
