// Package auth 封装对 SSO 内部鉴权接口的调用。
//
// Gateway 自身不持有 JWT 密钥、不连接用户数据库，也不判断"会话该不该失效"：
// 它只把浏览器带来的 access token 转给 SSO，并接受 SSO 的结论（dev.md §4）。
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Result SSO introspect 结果
type Result struct {
	Active    bool
	UserID    uint64
	SessionID string
	DeviceID  string
	ExpiresAt time.Time
	// Reason 未激活时的机器原因码（如 SESSION_REPLACED）
	Reason string
}

// ssoEnvelope SSO 统一响应包装 {code,msg,data}
type ssoEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type introspectData struct {
	Active    bool   `json:"active"`
	UserID    uint64 `json:"userId"`
	SessionID string `json:"sessionId"`
	DeviceID  string `json:"deviceId"`
	ExpiresAt string `json:"expiresAt"`
	Reason    string `json:"reason"`
}

// Client SSO 内部鉴权客户端
type Client struct {
	endpoint   string
	httpClient *http.Client
	log        *zap.Logger
}

// NewClient 创建鉴权客户端
func NewClient(endpoint, serviceToken string, timeout time.Duration, log *zap.Logger) *Client {
	transport := serviceTokenRoundTripper{token: serviceToken}
	return &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		log: log,
	}
}

// Introspect 校验 access token 对应的会话当前是否有效
func (c *Client) Introspect(ctx context.Context, accessToken string) (*Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("构造 introspect 请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用 SSO introspect 失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("调用 SSO introspect 返回异常状态: %d", resp.StatusCode)
	}

	var env ssoEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, fmt.Errorf("解析 SSO introspect 响应失败: %w", err)
	}
	if env.Code != http.StatusOK {
		return nil, fmt.Errorf("SSO introspect 返回业务错误: code=%d msg=%s", env.Code, env.Msg)
	}

	var data introspectData
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, &data); err != nil {
			return nil, fmt.Errorf("解析 SSO introspect data 失败: %w", err)
		}
	}

	result := &Result{
		Active:    data.Active,
		UserID:    data.UserID,
		SessionID: data.SessionID,
		DeviceID:  data.DeviceID,
		Reason:    data.Reason,
	}
	if data.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, data.ExpiresAt); err == nil {
			result.ExpiresAt = t
		}
	}
	return result, nil
}

// serviceTokenRoundTripper 为每个请求注入服务间共享密钥
type serviceTokenRoundTripper struct {
	token     string
	transport http.RoundTripper
}

const serviceTokenHeader = "X-MH-Service-Token"

func (t serviceTokenRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.token != "" {
		r.Header.Set(serviceTokenHeader, t.token)
	}
	transport := t.transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	return transport.RoundTrip(r)
}
