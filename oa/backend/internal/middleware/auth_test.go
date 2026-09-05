package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"permission-system/internal/client"

	"github.com/gin-gonic/gin"
)

// newTestContext 构造带请求的 gin.Context（TestMode）
func newTestContext(r *http.Request) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	return c
}

// TestExtractAccessToken_NoCredentials 无任何凭证
func TestExtractAccessToken_NoCredentials(t *testing.T) {
	c := newTestContext(httptest.NewRequest(http.MethodGet, "/", nil))

	_, errMsg := extractAccessToken(c)
	if errMsg != "读取认证信息失败" {
		t.Errorf("errMsg = %q, 期望 读取认证信息失败", errMsg)
	}
}

// TestExtractAccessToken_EmptyCookie Cookie 存在但值为空时按缺少认证信息处理
func TestExtractAccessToken_EmptyCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: client.AccessTokenCookieName, Value: ""})
	c := newTestContext(req)

	_, errMsg := extractAccessToken(c)
	if errMsg != "token为空" {
		t.Errorf("errMsg = %q, 期望 token为空", errMsg)
	}
}

// TestExtractAccessToken_CookieOnly 从 mh_sso2_access_token Cookie 提取
func TestExtractAccessToken_CookieOnly(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: client.AccessTokenCookieName, Value: "cookie-token"})
	c := newTestContext(req)

	token, errMsg := extractAccessToken(c)
	if errMsg != "" {
		t.Fatalf("errMsg = %q, 期望空", errMsg)
	}
	if token != "cookie-token" {
		t.Errorf("token = %q, 期望 cookie-token", token)
	}
}

// TestExtractAccessToken_HeaderIgnored 仅携带 Authorization 头而无 Cookie 时视为缺少认证信息
func TestExtractAccessToken_HeaderIgnored(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer header-token")
	c := newTestContext(req)

	_, errMsg := extractAccessToken(c)
	if errMsg != "读取认证信息失败" {
		t.Errorf("errMsg = %q, 期望 读取认证信息失败", errMsg)
	}
}
