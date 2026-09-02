package service

import (
	"time"

	"mh-sso-svc/internal/model"
	"mh-sso-svc/internal/utils"
)

// RequestMeta 请求元信息（审计与限流用）
type RequestMeta struct {
	IP        string
	UserAgent string
	RequestID string
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
	Email           string `json:"email" binding:"omitempty,email"`
	Mobile          string `json:"mobile" binding:"omitempty,min=5,max=32"`
	Nickname        string `json:"nickname"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
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

// RegisterResult 注册响应
type RegisterResult struct {
	UserID   uint64 `json:"userId"`
	Username string `json:"username"`
	Status   string `json:"status"`
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
