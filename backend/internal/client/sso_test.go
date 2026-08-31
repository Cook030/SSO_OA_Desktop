package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"permission-system/internal/utils"
)

// newMockIntrospect 启动 mock SSO introspect 服务
func newMockIntrospect(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestIntrospectToken_DocFormatNumberUserID 文档格式：userId 为数字、外层 msg、含 sessionId/passwordVersion/valid
func TestIntrospectToken_DocFormatNumberUserID(t *testing.T) {
	srv := newMockIntrospect(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("期望 GET 方法，实际 %s", r.Method)
		}
		if ck, err := r.Cookie(AccessTokenCookieName); err != nil || ck.Value != "doc-token" {
			t.Errorf("accessToken Cookie 错误: %v", ck)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"code": 200,
			"msg":  "success",
			"data": map[string]any{
				"userId":          10001,
				"sessionId":       "session-1",
				"passwordVersion": 2,
				"valid":           true,
			},
		})
	})

	c := NewSSOClient(utils.SSOConfig{BaseURL: srv.URL, IntrospectPath: "/api/v1/auth/introspect", TimeoutSecond: 5})
	resp, err := c.IntrospectToken("doc-token", "")
	if err != nil {
		t.Fatalf("IntrospectToken 返回错误: %v", err)
	}
	if resp.Code != 200 {
		t.Fatalf("code = %d, 期望 200", resp.Code)
	}
	if resp.Data.UserID != "10001" {
		t.Errorf("UserID = %q, 期望 \"10001\"（数字 userId 应转字符串）", resp.Data.UserID)
	}
	if resp.Data.SessionID != "session-1" {
		t.Errorf("SessionID = %q, 期望 session-1", resp.Data.SessionID)
	}
	if resp.Data.PasswordVersion != 2 {
		t.Errorf("PasswordVersion = %d, 期望 2", resp.Data.PasswordVersion)
	}
	if !resp.Data.Valid {
		t.Error("Valid = false, 期望 true")
	}
}

// TestIntrospectToken_DocFormatStringUserID 文档字段 userId 但为字符串
func TestIntrospectToken_DocFormatStringUserID(t *testing.T) {
	srv := newMockIntrospect(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"code": 200,
			"msg":  "success",
			"data": map[string]any{"userId": "zhangsan", "valid": true},
		})
	})

	c := NewSSOClient(utils.SSOConfig{BaseURL: srv.URL, IntrospectPath: "/api/v1/auth/introspect", TimeoutSecond: 5})
	resp, err := c.IntrospectToken("token", "")
	if err != nil {
		t.Fatalf("IntrospectToken 返回错误: %v", err)
	}
	if resp.Data.UserID != "zhangsan" {
		t.Errorf("UserID = %q, 期望 zhangsan", resp.Data.UserID)
	}
}

// TestIntrospectToken_HTTP401 返回 SSOTokenExpiredError，并透传 SSO 的 msg
func TestIntrospectToken_HTTP401(t *testing.T) {
	srv := newMockIntrospect(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"code": 401, "msg": "Token 已过期"})
	})

	c := NewSSOClient(utils.SSOConfig{BaseURL: srv.URL, IntrospectPath: "/api/v1/auth/introspect", TimeoutSecond: 5})
	_, err := c.IntrospectToken("bad-token", "")
	var expiredErr *SSOTokenExpiredError
	if !errors.As(err, &expiredErr) {
		t.Fatalf("期望 SSOTokenExpiredError，实际: %v", err)
	}
	if expiredErr.Message != "Token 已过期" {
		t.Errorf("Message = %q, 期望 'Token 已过期'", expiredErr.Message)
	}
}

// TestIntrospectToken_BizCode401 HTTP 200 但业务码 401，同样视为过期并透传 msg
func TestIntrospectToken_BizCode401(t *testing.T) {
	srv := newMockIntrospect(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"code": 401, "msg": "Token 已过期"})
	})

	c := NewSSOClient(utils.SSOConfig{BaseURL: srv.URL, IntrospectPath: "/api/v1/auth/introspect", TimeoutSecond: 5})
	_, err := c.IntrospectToken("bad-token", "")
	var expiredErr *SSOTokenExpiredError
	if !errors.As(err, &expiredErr) {
		t.Fatalf("期望 SSOTokenExpiredError，实际: %v", err)
	}
	if expiredErr.Message != "Token 已过期" {
		t.Errorf("Message = %q, 期望 'Token 已过期'", expiredErr.Message)
	}
}

// TestIntrospectToken_ServerError 非 200 状态码返回普通错误而非过期错误
func TestIntrospectToken_ServerError(t *testing.T) {
	srv := newMockIntrospect(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"code": 500})
	})

	c := NewSSOClient(utils.SSOConfig{BaseURL: srv.URL, IntrospectPath: "/api/v1/auth/introspect", TimeoutSecond: 5})
	_, err := c.IntrospectToken("token", "")
	if err == nil {
		t.Fatal("期望返回错误，实际 nil")
	}
	var expiredErr *SSOTokenExpiredError
	if errors.As(err, &expiredErr) {
		t.Fatalf("500 不应返回 SSOTokenExpiredError，实际: %v", err)
	}
}

// TestIntrospectToken_ConfigMissing 配置缺失时报错
func TestIntrospectToken_ConfigMissing(t *testing.T) {
	c := NewSSOClient(utils.SSOConfig{})
	if _, err := c.IntrospectToken("token", ""); err == nil {
		t.Fatal("配置缺失时应返回错误")
	}
}

// TestRevokeUserSessions_Success 成功撤销会话
func TestRevokeUserSessions_Success(t *testing.T) {
	var gotUserID int64
	srv := newMockIntrospect(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("期望 POST 方法，实际 %s", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type 错误: %s", got)
		}

		var reqBody struct {
			UserID int64 `json:"userId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("解析请求体失败: %v", err)
		}
		gotUserID = reqBody.UserID

		json.NewEncoder(w).Encode(map[string]any{"code": 200, "msg": "success", "data": nil})
	})

	c := NewSSOClient(utils.SSOConfig{BaseURL: srv.URL, RevokeUserSessionsPath: "/api/v1/auth/revoke-user-sessions", TimeoutSecond: 5})
	if err := c.RevokeUserSessions(521); err != nil {
		t.Fatalf("RevokeUserSessions 返回错误: %v", err)
	}
	if gotUserID != 521 {
		t.Errorf("userId = %d, 期望 521", gotUserID)
	}
}

// TestRevokeUserSessions_BizError SSO 业务码非 200
func TestRevokeUserSessions_BizError(t *testing.T) {
	srv := newMockIntrospect(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"code": 400, "msg": "userId is required"})
	})

	c := NewSSOClient(utils.SSOConfig{BaseURL: srv.URL, RevokeUserSessionsPath: "/api/v1/auth/revoke-user-sessions", TimeoutSecond: 5})
	err := c.RevokeUserSessions(521)
	if err == nil {
		t.Fatal("期望返回错误，实际 nil")
	}
	if !strings.Contains(err.Error(), "userId is required") {
		t.Errorf("错误信息未包含 SSO msg: %v", err)
	}
}

// TestRevokeUserSessions_HTTPError SSO 返回非 200 HTTP 状态码
func TestRevokeUserSessions_HTTPError(t *testing.T) {
	srv := newMockIntrospect(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"code": 500})
	})

	c := NewSSOClient(utils.SSOConfig{BaseURL: srv.URL, RevokeUserSessionsPath: "/api/v1/auth/revoke-user-sessions", TimeoutSecond: 5})
	if err := c.RevokeUserSessions(521); err == nil {
		t.Fatal("期望返回错误，实际 nil")
	}
}

// TestRevokeUserSessions_InvalidUserID userID 非法
func TestRevokeUserSessions_InvalidUserID(t *testing.T) {
	c := NewSSOClient(utils.SSOConfig{BaseURL: "http://localhost", RevokeUserSessionsPath: "/api/v1/auth/revoke-user-sessions"})
	if err := c.RevokeUserSessions(0); err == nil {
		t.Fatal("userId <= 0 时应返回错误")
	}
}

// TestRevokeUserSessions_ConfigMissing 配置缺失时报错
func TestRevokeUserSessions_ConfigMissing(t *testing.T) {
	c := NewSSOClient(utils.SSOConfig{})
	if err := c.RevokeUserSessions(521); err == nil {
		t.Fatal("配置缺失时应返回错误")
	}
}
