package utils

// EmptyToNil 空字符串转 nil（可空列存 NULL）
func EmptyToNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// NilToEmpty 指针为 nil 时返回空字符串
func NilToEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Truncate 截断字符串到指定字节长度（防超长输入写库失败）
func Truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
