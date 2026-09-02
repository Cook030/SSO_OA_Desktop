package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Conversation struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	UpdatedAt string `json:"updatedAt"`
}

type ChatMessage struct {
	ID             int64  `json:"id"`
	ConversationID int64  `json:"conversationId"`
	Role           string `json:"role"`
	Content        string `json:"content"`
	CreatedAt      string `json:"createdAt"`
}

type AgentMessageRequest struct {
	ConversationID int64  `json:"conversationId"`
	Content        string `json:"content"`
}

type AgentReply struct {
	ConversationID int64         `json:"conversationId"`
	Messages       []ChatMessage `json:"messages"`
}

type AgentStatus struct {
	OAMode       string `json:"oaMode"`
	OAConfigured bool   `json:"oaConfigured"`
	ToolCount    int    `json:"toolCount"`
}

type permissionIntent struct {
	EmployeeName string
	PlatformName string
}

type oaResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type oaEmployeePage struct {
	List []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"list"`
}

type oaPlatformPage struct {
	List []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"list"`
}

var assignPermissionPattern = regexp.MustCompile(`(?:给|为)\s*(.+?)\s*(?:分配|开通)\s*(.+?)\s*平台权限`)

func (a *App) AgentStatus() AgentStatus {
	return AgentStatus{OAMode: a.config.OA.Mode, OAConfigured: strings.TrimSpace(a.config.OA.BaseURL) != "", ToolCount: 1}
}

func (a *App) ListConversations() ([]Conversation, error) {
	if _, err := a.requireRole("admin", "maintainer", "user"); err != nil {
		return nil, err
	}
	rows, err := a.db.Query(`SELECT id,title,updated_at FROM conversations ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Conversation{}
	for rows.Next() {
		var item Conversation
		if err := rows.Scan(&item.ID, &item.Title, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (a *App) NewConversation() (*Conversation, error) {
	if _, err := a.requireRole("admin", "maintainer", "user"); err != nil {
		return nil, err
	}
	created := now()
	result, err := a.db.Exec(`INSERT INTO conversations(title,created_at,updated_at) VALUES(?,?,?)`, "新会话", created, created)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &Conversation{ID: id, Title: "新会话", UpdatedAt: created}, nil
}

func (a *App) ListMessages(conversationID int64) ([]ChatMessage, error) {
	if _, err := a.requireRole("admin", "maintainer", "user"); err != nil {
		return nil, err
	}
	rows, err := a.db.Query(`SELECT id,conversation_id,role,content,created_at FROM messages WHERE conversation_id=? ORDER BY id`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ChatMessage{}
	for rows.Next() {
		var item ChatMessage
		if err := rows.Scan(&item.ID, &item.ConversationID, &item.Role, &item.Content, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (a *App) SendAgentMessage(req AgentMessageRequest) (*AgentReply, error) {
	actor, err := a.requireRole("admin", "maintainer", "user")
	if err != nil {
		return nil, err
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errors.New("请输入任务")
	}
	if req.ConversationID == 0 {
		conversation, err := a.NewConversation()
		if err != nil {
			return nil, err
		}
		req.ConversationID = conversation.ID
	}
	if err := a.saveMessage(req.ConversationID, "user", content); err != nil {
		return nil, err
	}
	_ = a.updateConversationTitle(req.ConversationID, content)
	intent, matched := parsePermissionIntent(content)
	response := "我目前可以处理 OA 平台权限分配。请使用类似“给张三分配 A 平台权限”的指令。"
	if matched {
		response = a.handleAssignPermission(actor.Account, intent)
	}
	if err := a.saveMessage(req.ConversationID, "assistant", response); err != nil {
		return nil, err
	}
	messages, err := a.ListMessages(req.ConversationID)
	if err != nil {
		return nil, err
	}
	return &AgentReply{ConversationID: req.ConversationID, Messages: messages}, nil
}

func parsePermissionIntent(content string) (permissionIntent, bool) {
	matches := assignPermissionPattern.FindStringSubmatch(strings.TrimSpace(content))
	if len(matches) != 3 {
		return permissionIntent{}, false
	}
	employee := strings.Trim(strings.TrimSpace(matches[1]), "，,。.")
	platform := strings.Trim(strings.TrimSpace(matches[2]), "，,。.")
	return permissionIntent{EmployeeName: employee, PlatformName: platform}, employee != "" && platform != ""
}

func (a *App) handleAssignPermission(actor string, intent permissionIntent) string {
	if strings.TrimSpace(a.config.OA.BaseURL) == "" {
		return fmt.Sprintf("已识别操作：为「%s」分配「%s」平台权限。\n\nOA 尚未配置，未执行。请在 workbench.yaml 设置 oa.base_url 和 OA_ACCESS_TOKEN。", intent.EmployeeName, intent.PlatformName)
	}
	employee, err := a.findOAEmployee(intent.EmployeeName)
	if err != nil {
		return fmt.Sprintf("未执行：查询员工「%s」失败：%v", intent.EmployeeName, err)
	}
	platform, err := a.findOAPlatform(intent.PlatformName)
	if err != nil {
		return fmt.Sprintf("未执行：查询平台「%s」失败：%v", intent.PlatformName, err)
	}
	if a.config.OA.Mode != "execute" {
		return fmt.Sprintf("已完成 OA 演练：\n\n1. 找到员工：%s（ID: %d）\n2. 找到平台：%s（ID: %d）\n3. 已生成授权命令，但当前 `oa.mode: %s`，未写入 OA。\n\n将配置改为 `oa.mode: execute` 后，同一类指令会自动执行。", employee.Name, employee.ID, platform.Name, platform.ID, a.config.OA.Mode)
	}
	if err := a.assignOAPermission(employee.ID, platform.ID); err != nil {
		return fmt.Sprintf("授权失败：%v", err)
	}
	a.recordEvent(actor, "oa.permission.assigned", "oa_permission", fmt.Sprintf("%d:%d", employee.ID, platform.ID), fmt.Sprintf("%s -> %s", employee.Name, platform.Name))
	return fmt.Sprintf("授权已完成：已为「%s」分配「%s」平台权限。", employee.Name, platform.Name)
}

type oaEmployee struct {
	ID   int64
	Name string
}
type oaPlatform struct {
	ID   int64
	Name string
}

func (a *App) findOAEmployee(name string) (oaEmployee, error) {
	response, err := a.callOA(http.MethodGet, "/api/employees?keyword="+url.QueryEscape(name)+"&page=1&pageSize=20", nil)
	if err != nil {
		return oaEmployee{}, err
	}
	var page oaEmployeePage
	if err := json.Unmarshal(response.Data, &page); err != nil {
		return oaEmployee{}, errors.New("OA 员工响应格式异常")
	}
	matches := []oaEmployee{}
	for _, item := range page.List {
		if item.Name == name {
			matches = append(matches, oaEmployee{ID: item.ID, Name: item.Name})
		}
	}
	if len(matches) == 0 {
		return oaEmployee{}, errors.New("未找到同名员工")
	}
	if len(matches) > 1 {
		return oaEmployee{}, errors.New("存在多个同名员工，请改用“姓名（账号）”的明确指令")
	}
	return matches[0], nil
}

func (a *App) findOAPlatform(name string) (oaPlatform, error) {
	response, err := a.callOA(http.MethodGet, "/api/platforms?page=1&pageSize=100", nil)
	if err != nil {
		return oaPlatform{}, err
	}
	var page oaPlatformPage
	if err := json.Unmarshal(response.Data, &page); err != nil {
		return oaPlatform{}, errors.New("OA 平台响应格式异常")
	}
	matches := []oaPlatform{}
	for _, item := range page.List {
		if item.Name == name {
			matches = append(matches, oaPlatform{ID: item.ID, Name: item.Name})
		}
	}
	if len(matches) == 0 {
		return oaPlatform{}, errors.New("未找到同名平台")
	}
	if len(matches) > 1 {
		return oaPlatform{}, errors.New("存在多个同名平台")
	}
	return matches[0], nil
}

func (a *App) assignOAPermission(userID, platformID int64) error {
	payload, _ := json.Marshal(map[string]any{"userIds": []int64{userID}, "platformIds": []int64{platformID}})
	_, err := a.callOA(http.MethodPost, "/api/employees/permissions/batch", payload)
	return err
}

func (a *App) callOA(method, path string, body []byte) (*oaResponse, error) {
	token := strings.TrimSpace(os.Getenv(a.config.OA.AccessTokenEnv))
	if token == "" {
		return nil, fmt.Errorf("未设置环境变量 %s", a.config.OA.AccessTokenEnv)
	}
	request, err := http.NewRequest(method, strings.TrimRight(a.config.OA.BaseURL, "/")+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		request.Header.Set("Content-Type", "application/json")
	}
	request.AddCookie(&http.Cookie{Name: a.config.OA.CookieName, Value: token})
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("OA 不可访问：%w", err)
	}
	defer resp.Body.Close()
	var result oaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.New("OA 返回非预期响应")
	}
	if result.Code != 200 {
		return nil, errors.New(result.Message)
	}
	return &result, nil
}

func (a *App) saveMessage(conversationID int64, role, content string) error {
	_, err := a.db.Exec(`INSERT INTO messages(conversation_id,role,content,created_at) VALUES(?,?,?,?)`, conversationID, role, content, now())
	if err != nil {
		return err
	}
	_, _ = a.db.Exec(`UPDATE conversations SET updated_at=? WHERE id=?`, now(), conversationID)
	return nil
}

func (a *App) updateConversationTitle(conversationID int64, content string) error {
	var count int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM messages WHERE conversation_id=? AND role='user'`, conversationID).Scan(&count); err != nil || count != 1 {
		return err
	}
	title := []rune(content)
	if len(title) > 22 {
		title = append(title[:22], '…')
	}
	_, err := a.db.Exec(`UPDATE conversations SET title=?,updated_at=? WHERE id=?`, string(title), now(), conversationID)
	return err
}
