package shared

import "strings"

// SplitPermissionCode 将权限编码拆分为 (object, action)。
// 权限编码格式为 "<object>:<action>"，object 自身可再含冒号（如平台权限 platform:<平台编码>:access），
// 因此以最后一个冒号作为分隔点。
func SplitPermissionCode(code string) (object, action string) {
	idx := strings.LastIndex(code, ":")
	if idx <= 0 || idx == len(code)-1 {
		return code, ""
	}
	return code[:idx], code[idx+1:]
}

// IsPlatformAccessCode 判断权限编码是否为平台访问权限，是则返回其中的平台编码
func IsPlatformAccessCode(code string) (platformCode string, ok bool) {
	if !strings.HasPrefix(code, PlatformAccessPrefix) || !strings.HasSuffix(code, PlatformAccessSuffix) {
		return "", false
	}
	inner := code[len(PlatformAccessPrefix) : len(code)-len(PlatformAccessSuffix)]
	if inner == "" || strings.Contains(inner, ":") {
		return "", false
	}
	return inner, true
}
