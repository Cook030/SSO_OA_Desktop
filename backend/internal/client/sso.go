package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"permission-system/internal/utils"

	"go.uber.org/zap"
)

// SSOClient 封装对 SSO 服务的调用
type SSOClient struct {
	baseURL                string
	introspectPath         string
	revokeUserSessionsPath string
	httpClient             *http.Client
	tokenExpireCode        int
}

// tokenExpire401 SSO 返回的 token 过期状态码（HTTP 状态码或业务码）
const tokenExpire401 = 401

// AccessTokenCookieName SSO accessToken 的 Cookie 名（与 sso 接口设计文档一致）
// 后端仅认 Cookie 通道，不再解析 Authorization: Bearer Header
const AccessTokenCookieName = "mh_sso2_access_token"

// NewSSOClient 创建 SSO 客户端
func NewSSOClient(cfg utils.SSOConfig) *SSOClient {
	timeout := time.Duration(cfg.TimeoutSecond) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &SSOClient{
		baseURL:                cfg.BaseURL,
		introspectPath:         cfg.IntrospectPath,
		revokeUserSessionsPath: cfg.RevokeUserSessionsPath,
		httpClient:             &http.Client{Timeout: timeout},
		tokenExpireCode:        tokenExpire401,
	}
}

// SSOTokenIntrospectData SSO introspect 成功返回的 data 部分
// 字段以 sso 接口设计文档为准；userId 与本地 sys_user.id 对齐，
// 兼容 number / string 两种类型（文档示例 userId 为数字 10001）
type SSOTokenIntrospectData struct {
	UserID          string
	SessionID       string
	PasswordVersion int
	Valid           bool
}

// UnmarshalJSON 按文档字段 userId 解析，兼容 number / string 两种类型
func (d *SSOTokenIntrospectData) UnmarshalJSON(b []byte) error {
	var raw struct {
		UserID      *json.RawMessage `json:"userId"`
		SessionID   string           `json:"sessionId"`
		PassVersion int              `json:"passwordVersion"`
		Valid       bool             `json:"valid"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	d.SessionID = raw.SessionID
	d.PasswordVersion = raw.PassVersion
	d.Valid = raw.Valid

	if raw.UserID == nil {
		return nil
	}
	s, err := unmarshalStringOrNumber(*raw.UserID)
	if err != nil {
		return fmt.Errorf("解析 userId 失败: %s", string(*raw.UserID))
	}
	d.UserID = s
	return nil
}

// unmarshalStringOrNumber 将 JSON 值解析为字符串，兼容 "123" 与 123 两种表示
func unmarshalStringOrNumber(raw json.RawMessage) (string, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err == nil {
		return n.String(), nil
	}
	return "", fmt.Errorf("非法类型")
}

// SSOTokenIntrospectResponse SSO introspect 响应
// 字段以 sso 接口设计文档为准：统一使用 msg 字段
type SSOTokenIntrospectResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"msg"`
	Data    SSOTokenIntrospectData `json:"data"`
}

// SSOTokenExpiredError SSO token 已过期
type SSOTokenExpiredError struct {
	Message string
}

func (e *SSOTokenExpiredError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "SSO token 已过期"
}

// IntrospectToken 转发 access_token 到 SSO introspect 校验
// 返回 SSOTokenIntrospectResponse；当 status code == 401 时返回 *SSOTokenExpiredError
// requestID 为整个 HTTP 请求的 request_id，用于日志链路串联
func (c *SSOClient) IntrospectToken(accessToken string, requestID string) (*SSOTokenIntrospectResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("SSO baseURL 配置缺失")
	}
	if c.introspectPath == "" {
		return nil, fmt.Errorf("SSO introspectPath 配置缺失")
	}

	log := utils.GetLogger().With(zap.String("request_id", requestID))

	url := c.baseURL + c.introspectPath
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("构造 SSO 请求失败: %w", err)
	}
	// 仅以 Cookie 方式携带 accessToken，SSO 按文档优先读 Cookie
	req.AddCookie(&http.Cookie{Name: AccessTokenCookieName, Value: accessToken})

	log.Debug("SSO introspect 请求",
		zap.String("url", url),
		zap.String("tokenPrefix", safeTokenPrefix(accessToken, 20)),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用 SSO 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 SSO 响应失败: %w", err)
	}

	log.Debug("SSO introspect 原始响应",
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(body)),
	)

	// 401 表示 SSO token 已过期，需要把 SSO 返回的 msg 透传出去
	if resp.StatusCode == http.StatusUnauthorized {
		msg := extractSSOMessage(body)
		return nil, &SSOTokenExpiredError{Message: msg}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SSO 返回非 200: status=%d body=%s", resp.StatusCode, string(body))
	}

	var ssoResp SSOTokenIntrospectResponse
	if err := json.Unmarshal(body, &ssoResp); err != nil {
		return nil, fmt.Errorf("解析 SSO 响应失败: %w", err)
	}

	log.Debug("SSO introspect 解析结果",
		zap.Int("code", ssoResp.Code),
		zap.String("msg", ssoResp.Message),
		zap.String("userId", ssoResp.Data.UserID),
		zap.String("sessionId", ssoResp.Data.SessionID),
		zap.Int("passwordVersion", ssoResp.Data.PasswordVersion),
	)

	// 业务码也可能是 401，透传 SSO 的 msg
	if ssoResp.Code == c.tokenExpireCode {
		return nil, &SSOTokenExpiredError{Message: ssoResp.Message}
	}

	return &ssoResp, nil
}

// safeTokenPrefix 安全地返回 token 前缀，避免日志泄露完整凭证
func safeTokenPrefix(token string, n int) string {
	if len(token) <= n {
		return token
	}
	return token[:n]
}

// extractSSOMessage 从 SSO 响应 body 中提取 msg 字段；解析失败时返回默认文案
func extractSSOMessage(body []byte) string {
	var resp SSOTokenIntrospectResponse
	if err := json.Unmarshal(body, &resp); err == nil && resp.Message != "" {
		return resp.Message
	}
	return "Token 已过期"
}

// SSOBaseResponse SSO 通用响应（仅含 code/msg，无 data）
type SSOBaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

// RevokeUserSessions 调用 SSO 撤销指定用户的全部会话和 Refresh Token。
// 请求体为 {"userId": <userID>}，userId 与本地 sys_user.id 对齐。
func (c *SSOClient) RevokeUserSessions(userID int64) error {
	if c.baseURL == "" {
		return fmt.Errorf("SSO baseURL 配置缺失")
	}
	if c.revokeUserSessionsPath == "" {
		return fmt.Errorf("SSO revokeUserSessionsPath 配置缺失")
	}
	if userID <= 0 {
		return fmt.Errorf("userId 必须大于 0")
	}

	body, err := json.Marshal(map[string]any{"userId": userID})
	if err != nil {
		return fmt.Errorf("构造 SSO 请求体失败: %w", err)
	}

	url := c.baseURL + c.revokeUserSessionsPath
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("构造 SSO 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("调用 SSO 失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取 SSO 响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSO 返回非 200: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var ssoResp SSOBaseResponse
	if err := json.Unmarshal(respBody, &ssoResp); err != nil {
		return fmt.Errorf("解析 SSO 响应失败: %w", err)
	}

	if ssoResp.Code != 200 {
		return fmt.Errorf("SSO 撤销会话失败: code=%d msg=%s", ssoResp.Code, ssoResp.Message)
	}

	return nil
}
