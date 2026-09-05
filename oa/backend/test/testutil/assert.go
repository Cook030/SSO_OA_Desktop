package testutil

import (
	"fmt"

	"github.com/gavv/httpexpect/v2"
)

// MustOK 校验响应 code 为 200
func MustOK(resp *httpexpect.Response) *httpexpect.Object {
	obj := resp.Status(200).JSON().Object()
	obj.Value("code").Number().IsEqual(200)
	return obj
}

// MustCode 校验响应 code 为指定值
func MustCode(resp *httpexpect.Response, code int) *httpexpect.Object {
	obj := resp.Status(200).JSON().Object()
	obj.Value("code").Number().IsEqual(code)
	return obj
}

// MustMessageContains 校验响应 message 包含指定子串
func MustMessageContains(obj *httpexpect.Object, substr string) {
	obj.Value("message").String().Contains(substr)
}

// CookieAccessTokenName SSO Access Token Cookie 名（与后端 middleware 保持一致）
const CookieAccessTokenName = "mh_sso2_access_token"

// TokenInvalidValue 无效的 accessToken 值，用于"Token 无效"场景（通过 Cookie 携带）
const TokenInvalidValue = "invalid-token-12345"

// NoAuthRequest 返回未携带认证头的请求响应，用于"未登录"场景
func NoAuthRequest(method, path string, body ...any) *httpexpect.Response {
	req := requestByMethod(NoAuthE, method, path)
	if len(body) > 0 && body[0] != nil {
		req.WithJSON(body[0])
	}
	return req.Expect()
}

// EmployeeRequest 返回员工身份发起的请求响应，用于"权限不足"场景
func EmployeeRequest(method, path string, body ...any) *httpexpect.Response {
	req := requestByMethod(EmployeeE, method, path)
	if len(body) > 0 && body[0] != nil {
		req.WithJSON(body[0])
	}
	return req.Expect()
}

// requestByMethod 根据 HTTP 方法构造请求
func requestByMethod(client *httpexpect.Expect, method, path string) *httpexpect.Request {
	switch method {
	case "GET":
		return client.GET(path)
	case "POST":
		return client.POST(path)
	case "PUT":
		return client.PUT(path)
	case "DELETE":
		return client.DELETE(path)
	default:
		panic(fmt.Sprintf("不支持的 HTTP 方法: %s", method))
	}
}
