package handler

import (
	"errors"
	"net/http"

	"mh-sso-svc/internal/consts"
	"mh-sso-svc/internal/middleware"
	"mh-sso-svc/internal/service"
	"mh-sso-svc/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InternalHandler 服务间内部接口处理器（仅供 WebSocket Gateway 等内网服务调用）。
//
// 职责边界：Gateway 只被允许通过这里"问一句这个 access token 现在还能不能用"，
// 不能修改任何登录状态 —— 会话状态的唯一写入方是 SSO 自己。
type InternalHandler struct {
	svc *service.AuthService
}

// NewInternalHandler 创建内部接口处理器
func NewInternalHandler(svc *service.AuthService) *InternalHandler {
	return &InternalHandler{svc: svc}
}

// Introspect POST /internal/v1/sessions/introspect
//
// 请求：Authorization: Bearer {accessToken} + X-MH-Service-Token
// 响应：
//
//	{"active": true,  "userId": 1001, "sessionId": "session_xxx", "deviceId": "...", "expiresAt": "..."}
//	{"active": false, "reason": "SESSION_REPLACED"}
//
// 未激活时返回 HTTP 200 + active=false（而非 401），让 Gateway 能拿到 reason
// 并据此选择 WebSocket 关闭码。
func (h *InternalHandler) Introspect(c *gin.Context) {
	accessToken, errMsg := middleware.ExtractAccessTokenBearerFirst(c)
	if errMsg != "" {
		utils.OK(c, &service.GatewayIntrospectResult{Active: false, Reason: consts.ReasonTokenInvalid})
		return
	}

	result, err := h.svc.IntrospectSession(accessToken)
	if err != nil {
		// Redis / MySQL 故障：Gateway 必须拒绝连接，不能放行
		var biz *utils.BizError
		if errors.As(err, &biz) {
			utils.OK(c, &service.GatewayIntrospectResult{Active: false, Reason: biz.Reason})
			return
		}
		utils.GetLogger().Error("内部 introspect 处理失败",
			zap.String("request_id", middleware.GetRequestID(c)), zap.Error(err))
		utils.Error(c, http.StatusInternalServerError, utils.ErrMsgServerError)
		return
	}
	utils.OK(c, result)
}
