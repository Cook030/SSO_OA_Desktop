package handler

import (
	"errors"
	"fmt"
	"net/http"

	"mh-sso-svc/internal/middleware"
	"mh-sso-svc/internal/service"
	"mh-sso-svc/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler 认证接口处理器
type AuthHandler struct {
	svc *service.AuthService
	cfg *utils.AuthConfig
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(svc *service.AuthService, cfg *utils.AuthConfig) *AuthHandler {
	return &AuthHandler{svc: svc, cfg: cfg}
}

// Login 登录（成功写入 HttpOnly Cookie）
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数解析失败: %v", err))
		return
	}
	result, err := h.svc.Login(req.Account, req.Password, requestMeta(c))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	h.setAuthCookies(c, result.AccessToken, result.RefreshToken)
	utils.OK(c, result)
}

// Refresh 刷新 token（成功后更新 Cookie）
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken := extractRefreshToken(c)
	result, err := h.svc.Refresh(refreshToken, requestMeta(c))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	h.setAuthCookies(c, result.AccessToken, result.RefreshToken)
	utils.OK(c, result)
}

// Logout 退出登录（尽力撤销会话，始终返回成功并清空 Cookie）
func (h *AuthHandler) Logout(c *gin.Context) {
	accessToken, _ := middleware.ExtractAccessTokenCookieFirst(c)
	refreshToken := extractRefreshToken(c)

	h.svc.Logout(accessToken, refreshToken, requestMeta(c))
	h.clearAuthCookies(c)
	utils.OK(c, nil)
}

// ChangePassword 修改密码（成功后撤销全部会话并清空 Cookie）
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数解析失败: %v", err))
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.Password, req.ConfirmPassword, requestMeta(c)); err != nil {
		h.writeErr(c, err)
		return
	}
	h.clearAuthCookies(c)
	utils.OK(c, nil)
}

// Me 获取当前用户信息
func (h *AuthHandler) Me(c *gin.Context) {
	result, err := h.svc.Me(middleware.GetUserID(c))
	h.writeResult(c, result, err)
}

// UpdateProfile 更新当前用户资料（姓名/邮箱/手机号）
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数解析失败: %v", err))
		return
	}
	result, err := h.svc.UpdateProfile(c.Request.Context(), middleware.GetUserID(c), req, requestMeta(c))
	h.writeResult(c, result, err)
}

// Introspect 校验 access token 有效性（内部接口，Bearer 优先）
func (h *AuthHandler) Introspect(c *gin.Context) {
	accessToken, errMsg := middleware.ExtractAccessTokenBearerFirst(c)
	if errMsg != "" {
		utils.Unauthorized(c, errMsg)
		return
	}
	result, err := h.svc.Introspect(accessToken)
	h.writeResult(c, result, err)
}

// RevokeUserSessions 撤销指定用户全部会话（内部接口）
func (h *AuthHandler) RevokeUserSessions(c *gin.Context) {
	var req service.RevokeUserSessionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数解析失败: %v", err))
		return
	}
	if err := h.svc.RevokeUserSessions(req.UserID, requestMeta(c)); err != nil {
		h.writeErr(c, err)
		return
	}
	utils.OK(c, nil)
}

// ---------- 私有辅助 ----------

// requestMeta 提取请求元信息（审计与限流用），超长字段按列宽截断
func requestMeta(c *gin.Context) service.RequestMeta {
	return service.RequestMeta{
		IP:        utils.Truncate(c.ClientIP(), 64),
		UserAgent: utils.Truncate(c.Request.UserAgent(), 512),
		RequestID: middleware.GetRequestID(c),
	}
}

// extractRefreshToken 按优先级提取 refresh token：
// Body refreshToken → X-MH-Refresh-Token → Cookie mh_sso_refresh_token
func extractRefreshToken(c *gin.Context) string {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	// body 可为空或非 JSON（如空对象之外的探测请求），解析失败不视为错误
	if err := c.ShouldBindJSON(&body); err == nil && body.RefreshToken != "" {
		return body.RefreshToken
	}
	if token := c.GetHeader(utils.RefreshTokenHeaderName); token != "" {
		return token
	}
	if token, err := c.Cookie(utils.RefreshTokenCookieName); err == nil && token != "" {
		return token
	}
	return ""
}

// setAuthCookies 写入 HttpOnly 认证 Cookie：
// access token Max-Age=accessTTL，refresh token Max-Age=refreshTTL
func (h *AuthHandler) setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	h.setCookie(c, utils.AccessTokenCookieName, accessToken, h.cfg.AccessTokenTTLSecond)
	h.setCookie(c, utils.RefreshTokenCookieName, refreshToken, h.cfg.RefreshTokenTTLSecond)
}

// clearAuthCookies 清空认证 Cookie（Max-Age=0 立即删除）
func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	h.setCookie(c, utils.AccessTokenCookieName, "", -1)
	h.setCookie(c, utils.RefreshTokenCookieName, "", -1)
}

// setCookie 写入单个 Cookie；maxAge < 0 表示删除；
// cookie_domain 为空时不设置 Domain（使用当前域名），本地 HTTP 环境关闭 Secure
func (h *AuthHandler) setCookie(c *gin.Context, name, value string, maxAge int) {
	c.SetSameSite(parseSameSite(h.cfg.CookieSameSite))
	c.SetCookie(name, value, maxAge, "/", h.cfg.CookieDomain, h.cfg.CookieSecure, true)
}

// parseSameSite 解析 SameSite 配置（Lax/Strict/None）
func parseSameSite(s string) http.SameSite {
	switch s {
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// writeResult 统一输出业务结果
func (h *AuthHandler) writeResult(c *gin.Context, data interface{}, err error) {
	if err != nil {
		h.writeErr(c, err)
		return
	}
	utils.OK(c, data)
}

// writeErr 统一错误输出：BizError 按业务码返回，其余按 500 处理；
// 日志不打印密码、完整 token 等敏感信息
func (h *AuthHandler) writeErr(c *gin.Context, err error) {
	var biz *utils.BizError
	if errors.As(err, &biz) {
		utils.Error(c, biz.Code, biz.Msg)
		return
	}
	utils.GetLogger().Error("接口处理失败",
		zap.String("path", c.Request.URL.Path),
		zap.String("request_id", middleware.GetRequestID(c)),
		zap.Error(err))
	utils.ServerError(c, utils.ErrMsgServerError)
}
