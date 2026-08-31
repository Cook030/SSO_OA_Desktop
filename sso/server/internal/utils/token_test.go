package utils

import (
	"strings"
	"testing"
)

func TestRandomHex(t *testing.T) {
	// 长度：n 字节 → 2n 个十六进制字符
	if got := RandomHex(16); len(got) != 32 {
		t.Fatalf("RandomHex(16) 长度 = %d, 期望 32", len(got))
	}
	// 两次生成结果不同（随机性）
	if RandomHex(16) == RandomHex(16) {
		t.Fatal("RandomHex 两次生成结果不应相同")
	}
}

func TestGenerateOpaqueToken(t *testing.T) {
	token := GenerateOpaqueToken("rt_")
	if !strings.HasPrefix(token, "rt_") {
		t.Fatalf("token 应带 rt_ 前缀: %s", token)
	}
	if len(token) != len("rt_")+64 {
		t.Fatalf("token 长度 = %d, 期望 %d", len(token), len("rt_")+64)
	}
}

func TestSHA256Hex(t *testing.T) {
	// 已知值校验：SHA-256("abc")
	got := SHA256Hex("abc")
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Fatalf("SHA256Hex(\"abc\") = %s, 期望 %s", got, want)
	}
}

func TestSafeTokenPrefix(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiJ9.payload.signature"
	got := SafeTokenPrefix(token, 10)
	if got != "eyJhbGciOi..." {
		t.Fatalf("SafeTokenPrefix = %s", got)
	}
	// 短 token 原样返回
	short := "abc"
	if got := SafeTokenPrefix(short, 10); got != short {
		t.Fatalf("短 token 应原样返回: %s", got)
	}
}
