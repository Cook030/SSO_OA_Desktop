package service

import (
	"time"

	"mh-sso-svc/internal/model"
	"mh-sso-svc/internal/utils"
)

// RequestMeta 请求元信息（审计、限流与设备绑定用）
type RequestMeta struct {
	IP        string
	UserAgent string
	RequestID string
	// DeviceID 客户端设备唯一标识（已规范化）；为空表示未知设备
	DeviceID string
	// DeviceType 设备类型：desktop / mobile / tablet / unknown
	DeviceType string
}

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	// DeviceID 可选：客户端设备唯一标识，缺省时从 X-MH-Device-Id 头读取
	DeviceID string `json:"deviceId"`
	// DeviceType 可选：desktop / mobile / tablet，缺省时从 X-MH-Device-Type 头读取
	DeviceType string `json:"deviceType"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

// UpdateProfileRequest 更新个人资料请求（姓名/邮箱/手机号）
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"required,max=64"`
	Email    string `json:"email" binding:"omitempty,email,max=128"`
	Mobile   string `json:"mobile" binding:"omitempty,min=5,max=32"`
}

// RevokeUserSessionsRequest 撤销指定用户全部会话请求
type RevokeUserSessionsRequest struct {
	UserID uint64 `json:"userId"`
}

// UserInfo 用户展示信息（login / me 响应）
type UserInfo struct {
	ID              uint64    `json:"id"`
	Account         string    `json:"account"`
	Name            string    `json:"name"`
	Phone           string    `json:"phone"`
	Email           string    `json:"email"`
	Role            string    `json:"role"`
	Department      string    `json:"department"`
	PasswordChanged int       `json:"passwordChanged"` // 1 表示修改过密码（password_version > 1）
	CreateTime      time.Time `json:"createTime"`
	UpdateTime      time.Time `json:"updateTime"`
}

// PermItem 用户组/角色展示项
type PermItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// LoginResult 登录响应
type LoginResult struct {
	AccessToken      string   `json:"accessToken"`
	RefreshToken     string   `json:"refreshToken"`
	TokenType        string   `json:"tokenType"`
	ExpiresIn        int      `json:"expiresIn"`
	RefreshExpiresIn int      `json:"refreshExpiresIn"`
	User             UserInfo `json:"user"`
}

// RefreshResult 刷新 token 响应
type RefreshResult struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
}

// MeResult 当前用户信息响应（MVP：基础角色 + 空权限数组）
type MeResult struct {
	User    UserInfo   `json:"user"`
	Groups  []PermItem `json:"groups"`
	Roles   []PermItem `json:"roles"`
	Apps    []string   `json:"apps"`
	Pages   []string   `json:"pages"`
	Apis    []string   `json:"apis"`
	Menus   []string   `json:"menus"`
	Buttons []string   `json:"buttons"`
}

// IntrospectResult token 校验响应
type IntrospectResult struct {
	UserID          uint64 `json:"userId"`
	SessionID       string `json:"sessionId"`
	PasswordVersion int    `json:"passwordVersion"`
	Valid           bool   `json:"valid"`
}

// GatewayIntrospectResult 供 WebSocket Gateway 使用的会话校验结果。
// Active=false 时携带 Reason，Gateway 据此转换为对应的 WebSocket 关闭码。
type GatewayIntrospectResult struct {
	Active    bool      `json:"active"`
	UserID    uint64    `json:"userId,omitempty"`
	SessionID string    `json:"sessionId,omitempty"`
	DeviceID  string    `json:"deviceId,omitempty"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
	Reason    string    `json:"reason,omitempty"`
}

// buildUserInfo 组装用户展示信息（字段与接口文档对齐）
func buildUserInfo(user *model.SysUser) UserInfo {
	passwordChanged := 0
	if user.PasswordVersion > 1 {
		passwordChanged = 1
	}
	return UserInfo{
		ID:              user.ID,
		Account:         user.Account,
		Name:            user.Name,
		Phone:           utils.NilToEmpty(user.Phone),
		Email:           utils.NilToEmpty(user.Email),
		Role:            "user", // MVP 固定基础角色，接入角色表后扩展
		Department:      utils.NilToEmpty(user.Department),
		PasswordChanged: passwordChanged,
		CreateTime:      user.CreateTime,
		UpdateTime:      user.UpdateTime,
	}
}
