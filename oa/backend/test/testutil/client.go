package testutil

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gavv/httpexpect/v2"
)

var (
	// E 管理员身份的全局 HTTP 客户端
	E *httpexpect.Expect

	// EmployeeE 员工身份的全局 HTTP 客户端
	EmployeeE *httpexpect.Expect

	// NoAuthE 未登录身份的全局 HTTP 客户端
	NoAuthE *httpexpect.Expect

	// AdminToken 管理员 SSO Token
	AdminToken string

	// EmployeeToken 员工 SSO Token
	EmployeeToken string

	// GlobalEmployeeID 全局员工 ID，权限不足测试使用
	GlobalEmployeeID int64

	// BaseURL 测试服务器基础 URL
	BaseURL string
)

// NewClient 创建一个 httpexpect 客户端
// token 为空表示不携带认证头
func NewClient(token string) *httpexpect.Expect {
	cfg := httpexpect.Config{
		BaseURL:  BaseURL,
		Reporter: &panicReporter{},
	}
	if token != "" {
		cfg.RequestFactory = &authRequestFactory{token: token}
	}
	return httpexpect.WithConfig(cfg)
}

// authRequestFactory 为每个请求自动附加 mh_sso2_access_token Cookie（认证方式：Cookie 优先，Header 方式已忽略）
type authRequestFactory struct {
	token string
}

func (f *authRequestFactory) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := httpexpect.DefaultRequestFactory{}.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	if f.token != "" {
		req.AddCookie(&http.Cookie{Name: CookieAccessTokenName, Value: f.token})
	}
	return req, nil
}

// panicReporter 在断言失败时直接 panic，保证测试失败可被 Go 测试框架捕获
type panicReporter struct{}

func (r *panicReporter) Errorf(message string, args ...interface{}) {
	panic(fmt.Sprintf(message, args...))
}

func (r *panicReporter) Fatalf(message string, args ...interface{}) {
	panic(fmt.Sprintf(message, args...))
}
