package shared

import "fmt"

// FormatEmployeeDisplayID 将员工数值 ID 格式化为四位补零字符串（如 5 -> "0005"）
func FormatEmployeeDisplayID(id int64) string {
	return fmt.Sprintf("%04d", id)
}
