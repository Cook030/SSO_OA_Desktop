package service

import (
	"errors"
	"testing"
	"time"

	"mh-sso-svc/internal/utils"
)

func newTestTokenService(issuer, secret string) *TokenService {
	return NewTokenService(&utils.AuthConfig{
		Issuer:               issuer,
		JWTSecret:            secret,
		AccessTokenTTLSecond: 900,
	})
}

func TestTokenRoundTrip(t *testing.T) {
	svc := newTestTokenService("mh-sso", "test-secret-32-bytes-0123456789ab")
	now := time.Now()

	token, jti, expiresAt, err := svc.GenerateAccessToken(10001, "session_xxx", "zhangsan", 3, now)
	if err != nil {
		t.Fatalf("签发 access token 失败: %v", err)
	}
	if jti == "" {
		t.Fatal("jti 不应为空")
	}
	if !expiresAt.After(now) {
		t.Fatal("过期时间应晚于签发时间")
	}
	if svc.AccessTokenTTL() != 900*time.Second {
		t.Fatalf("AccessTokenTTL = %v, 期望 900s", svc.AccessTokenTTL())
	}

	claims, err := svc.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("解析 access token 失败: %v", err)
	}
	if claims.UserID != 10001 {
		t.Fatalf("UserID = %d, 期望 10001", claims.UserID)
	}
	if claims.SessionID != "session_xxx" {
		t.Fatalf("SessionID = %s, 期望 session_xxx", claims.SessionID)
	}
	if claims.Account != "zhangsan" {
		t.Fatalf("Account = %s, 期望 zhangsan", claims.Account)
	}
	if claims.PasswordVersion != 3 {
		t.Fatalf("PasswordVersion = %d, 期望 3", claims.PasswordVersion)
	}
	if claims.Issuer != "mh-sso" {
		t.Fatalf("Issuer = %s, 期望 mh-sso", claims.Issuer)
	}
}

func TestParseExpiredToken(t *testing.T) {
	svc := newTestTokenService("mh-sso", "test-secret-32-bytes-0123456789ab")
	// 签发一张已过期的 token
	token, _, _, err := svc.GenerateAccessToken(1, "s", "a", 1, time.Now().Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}

	if _, err := svc.ParseAccessToken(token); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("期望 ErrTokenExpired, 实际: %v", err)
	}
}

func TestParseTamperedToken(t *testing.T) {
	svc := newTestTokenService("mh-sso", "test-secret-32-bytes-0123456789ab")
	token, _, _, err := svc.GenerateAccessToken(1, "s", "a", 1, time.Now())
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}

	if _, err := svc.ParseAccessToken(token + "x"); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("期望 ErrTokenInvalid, 实际: %v", err)
	}
}

func TestParseTokenWithWrongSecret(t *testing.T) {
	signer := newTestTokenService("mh-sso", "secret-aaaaaaaaaaaaaaaaaaaaaaaaa")
	verifier := newTestTokenService("mh-sso", "secret-bbbbbbbbbbbbbbbbbbbbbbbb")

	token, _, _, err := signer.GenerateAccessToken(1, "s", "a", 1, time.Now())
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}

	if _, err := verifier.ParseAccessToken(token); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("期望 ErrTokenInvalid, 实际: %v", err)
	}
}

func TestParseTokenWithWrongIssuer(t *testing.T) {
	signer := newTestTokenService("mh-sso", "secret-aaaaaaaaaaaaaaaaaaaaaaaaa")
	verifier := newTestTokenService("other-issuer", "secret-aaaaaaaaaaaaaaaaaaaaaaaaa")

	token, _, _, err := signer.GenerateAccessToken(1, "s", "a", 1, time.Now())
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}

	if _, err := verifier.ParseAccessToken(token); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("期望 ErrTokenInvalid, 实际: %v", err)
	}
}

func TestParseGarbageToken(t *testing.T) {
	svc := newTestTokenService("mh-sso", "test-secret-32-bytes-0123456789ab")
	for _, token := range []string{"", "not-a-jwt", "aaa.bbb.ccc"} {
		if _, err := svc.ParseAccessToken(token); !errors.Is(err, ErrTokenInvalid) {
			t.Fatalf("token %q 期望 ErrTokenInvalid, 实际: %v", token, err)
		}
	}
}
