package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// RandomHex 生成 n 字节随机数的十六进制字符串（长度为 2n）
func RandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("生成随机数失败: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// GenerateOpaqueToken 生成带前缀的 opaque 随机串，如 session_xxx、rt_xxx
func GenerateOpaqueToken(prefix string) string {
	return prefix + RandomHex(32)
}

// SHA256Hex 返回字符串 SHA-256 哈希的十六进制表示（refresh token 只存哈希）
func SHA256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// SafeTokenPrefix 返回 token 前缀用于日志输出，避免泄露完整凭证
func SafeTokenPrefix(token string, n int) string {
	if len(token) <= n {
		return token
	}
	return token[:n] + "..."
}
