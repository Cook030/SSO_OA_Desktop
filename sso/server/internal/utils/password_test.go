package utils

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("Password123")
	if err != nil {
		t.Fatalf("HashPassword 失败: %v", err)
	}
	if hash == "Password123" {
		t.Fatal("哈希值不应等于明文密码")
	}
	if !CheckPassword(hash, "Password123") {
		t.Fatal("正确密码校验应通过")
	}
	if CheckPassword(hash, "WrongPassword") {
		t.Fatal("错误密码校验不应通过")
	}
}

func TestHashPasswordSalts(t *testing.T) {
	// 相同密码两次哈希应产生不同结果（bcrypt 随机盐）
	hash1, _ := HashPassword("Password123")
	hash2, _ := HashPassword("Password123")
	if hash1 == hash2 {
		t.Fatal("两次哈希结果不应相同（应包含随机盐）")
	}
}
