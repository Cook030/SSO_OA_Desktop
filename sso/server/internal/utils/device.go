package utils

import "strings"

// 设备唯一标识相关常量。
//
// 设备标识由客户端生成并持久化（localStorage），仅用于：
//  1. 在会话上标注"本次登录来自哪个设备"，供审计与展示；
//  2. WebSocket 握手时比对，识别"同一会话凭证被搬到另一台设备"的异常。
//
// 它不是安全凭证，也不参与单设备登录的判定 —— 单设备登录的权威依据是
// sso:user:{userId}:current_session。因此设备标识非法时只降级为"未知设备"，不阻断登录。
const (
	// UnknownDeviceID 设备标识缺失或非法时的占位值
	UnknownDeviceID = "unknown"

	// MaxDeviceIDLength 设备标识最大长度，防止超长输入污染 Redis
	MaxDeviceIDLength = 64

	// DeviceIDHeader 登录等 HTTP 接口携带设备标识的 Header
	DeviceIDHeader = "X-MH-Device-Id"

	// DeviceTypeHeader 设备类型（desktop/mobile/tablet）Header
	DeviceTypeHeader = "X-MH-Device-Type"

	// deviceIDAllowedChars 允许的字符集合：UUID、指纹哈希等常见形态
	deviceIDAllowedChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.:"
)

// NormalizeDeviceID 校验并规范化客户端上报的设备标识：
//   - 空、超长或包含非白名单字符 -> 返回 UnknownDeviceID（不报错，降级为未知设备）
//   - 合法 -> 返回去空格后的原值
func NormalizeDeviceID(raw string) string {
	id := strings.TrimSpace(raw)
	if id == "" || len(id) > MaxDeviceIDLength {
		return UnknownDeviceID
	}
	if strings.ContainsFunc(id, func(r rune) bool {
		return !strings.ContainsRune(deviceIDAllowedChars, r)
	}) {
		return UnknownDeviceID
	}
	return id
}

// NormalizeDeviceType 规范化设备类型，仅保留已知取值，其余归为 unknown
func NormalizeDeviceType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "desktop", "mobile", "tablet":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return UnknownDeviceID
	}
}
