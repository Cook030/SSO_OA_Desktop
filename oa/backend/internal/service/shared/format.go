package shared

import "fmt"

// FormatEmployeeDisplayID 将员工数值 ID 格式化为四位补零字符串（如 5 -> "0005"）
func FormatEmployeeDisplayID(id int64) string {
	return fmt.Sprintf("%04d", id)
}

// TimeLayout 统一的时间输出格式
const TimeLayout = "2006-01-02 15:04:05"

// DerefString 解引用字符串指针，nil 时返回空串
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
