// Package server 提供 WebSocket 握手入口。
//
// 职责边界（dev.md §4）：Gateway 只负责连接管理与事件投递，
// 会话是否有效的判定完全交给 SSO 的内部 introspect 接口。
// Gateway 不持有 JWT 密钥、不连接用户数据库、不修改任何登录状态。
package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mh-ws-gateway/internal/auth"
	"mh-ws-gateway/internal/config"
	wsconn "mh-ws-gateway/internal/conn"
	"mh-ws-gateway/internal/delivery"
	"mh-ws-gateway/internal/hub"
	"mh-ws-gateway/internal/presence"
	"mh-ws-gateway/internal/protocol"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// errResponse 握手失败时的响应体（与 SSO 的 code/msg/reason 风格保持一致）
type errResponse struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Reason string `json:"reason,omitempty"`
}

// Server WebSocket 服务
type Server struct {
	cfg        *config.Config
	instanceID string
	hub        *hub.Hub
	authClient *auth.Client
	delivery   *delivery.Store
	presence   *presence.Store
	upgrader   websocket.Upgrader
	log        *zap.Logger
}

// New 创建服务实例
func New(cfg *config.Config, instanceID string, h *hub.Hub, ac *auth.Client,
	store *delivery.Store, ps *presence.Store, log *zap.Logger) *Server {
	s := &Server{
		cfg:        cfg,
		instanceID: instanceID,
		hub:        h,
		authClient: ac,
		delivery:   store,
		presence:   ps,
		log:        log,
	}
	s.upgrader = websocket.Upgrader{
		HandshakeTimeout:  10 * time.Second,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		CheckOrigin:       s.checkOrigin,
		EnableCompression: false,
	}
	return s
}

// Handler 注册路由
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(s.cfg.Realtime.WSPath, s.HandleWS)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":     "ok",
			"instanceId": s.instanceID,
			"conns":      s.hub.Len(),
		})
	})
	return mux
}

// HandleWS 处理 WebSocket 握手与连接建立
func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	if s.hub.Len() >= s.cfg.Realtime.MaxConnections {
		writeErr(w, http.StatusServiceUnavailable, "连接数已达上限", "TOO_MANY_CONNECTIONS")
		return
	}

	// 1. 提取 access token：浏览器走 Cookie（dev.md §5 明确禁止放在 URL query 中）
	accessToken := s.extractAccessToken(r)
	if accessToken == "" {
		writeErr(w, http.StatusUnauthorized, "缺少 access token", "TOKEN_MISSING")
		return
	}

	// 2. 向 SSO 询问该 token 当前是否有效
	res, err := s.authClient.Introspect(r.Context(), accessToken)
	if err != nil {
		s.log.Error("调用 SSO introspect 失败", zap.Error(err))
		writeErr(w, http.StatusBadGateway, "会话校验服务不可用", "INTROSPECT_UNAVAILABLE")
		return
	}
	if res == nil || !res.Active {
		reason := ""
		if res != nil {
			reason = res.Reason
		}
		writeErr(w, http.StatusUnauthorized, "会话无效或已失效", reason)
		return
	}

	// 3. 设备标识比对：识别"同一会话凭证被搬到另一台设备"的异常（纵深防御）
	deviceID := r.URL.Query().Get("device_id")
	if res.DeviceID != "" && deviceID != "" && res.DeviceID != deviceID {
		s.log.Warn("设备标识与会话绑定不一致",
			zap.String("session_id", res.SessionID),
			zap.String("session_device", res.DeviceID),
			zap.String("request_device", deviceID))
		writeErr(w, http.StatusForbidden, "设备标识与会话不一致", "DEVICE_MISMATCH")
		return
	}

	// 4. 升级协议
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Warn("WebSocket 升级失败", zap.Error(err))
		return
	}

	c := wsconn.New(wsconn.Params{
		SessionID:  res.SessionID,
		UserID:     res.UserID,
		DeviceID:   deviceID,
		InstanceID: s.instanceID,
		WS:         ws,
		Hub:        s.hub,
		Delivery:   s.delivery,
		Presence:   s.presence,
		Opt: wsconn.Options{
			PingInterval:   s.cfg.PingInterval(),
			PongWait:       s.cfg.PongWait(),
			WriteQueueSize: s.cfg.Realtime.WriteQueueSize,
			ReadLimitBytes: s.cfg.Realtime.ReadLimitBytes,
		},
		Log: s.log,
	})

	// 5. 同一会话的新连接取代旧连接（如浏览器刷新、多标签页）
	if previous := s.hub.Register(c); previous != nil {
		s.log.Info("同一会话建立新连接，关闭旧连接",
			zap.String("session_id", res.SessionID))
		previous.Close(protocol.CloseDuplicateSession, "duplicate session connection")
	}

	// 6. 登记在线状态（路由辅助数据，非安全状态）
	registered, err := s.presence.Register(r.Context(), presence.Record{
		InstanceID:    s.instanceID,
		ConnectionID:  c.ID(),
		SessionID:     res.SessionID,
		LastHeartbeat: time.Now().UnixMilli(),
	})
	if err != nil {
		s.log.Warn("登记在线状态失败", zap.String("session_id", res.SessionID), zap.Error(err))
	} else if !registered {
		// 同一会话在另一个 Gateway 实例上仍有存活连接：
		// 只是提示，不做额外处理 —— 单设备登录的权威判定在 SSO，不在这里
		s.log.Warn("该会话已在其他网关实例上建连", zap.String("session_id", res.SessionID))
	}

	go c.Run()

	// 7. 补发未确认事件：实现"至少一次"，客户端需按 eventId 幂等处理
	s.replayPending(r.Context(), c, res.SessionID)

	s.log.Info("WebSocket 连接已建立",
		zap.Uint64("user_id", res.UserID),
		zap.String("session_id", res.SessionID),
		zap.String("connection_id", c.ID()),
		zap.String("device_id", deviceID))
}

// replayPending 重连后补发未 ACK 的事件
func (s *Server) replayPending(ctx context.Context, c *wsconn.Connection, sessionID string) {
	pending, err := s.delivery.Pending(ctx, sessionID)
	if err != nil {
		s.log.Error("读取待投递事件失败", zap.String("session_id", sessionID), zap.Error(err))
		return
	}
	for _, payload := range pending {
		if err := c.Send(payload); errors.Is(err, protocol.ErrQueueFull) {
			s.log.Warn("补发事件时队列已满，关闭连接",
				zap.String("session_id", sessionID))
			c.Close(protocol.CloseSlowConsumer, "write queue full on replay")
			return
		}
	}
	if len(pending) > 0 {
		s.log.Info("已补发未确认事件",
			zap.String("session_id", sessionID), zap.Int("count", len(pending)))
	}
}

// extractAccessToken 按 Cookie → Authorization 的顺序提取 access token。
// dev.md §5：不读取 URL query 参数中的 token，避免泄漏到日志与 Referer。
func (s *Server) extractAccessToken(r *http.Request) string {
	if cookie, err := r.Cookie(s.cfg.Realtime.AccessTokenCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	if auth := r.Header.Get("Authorization"); auth != "" {
		const prefix = "Bearer "
		if strings.HasPrefix(auth, prefix) {
			return strings.TrimPrefix(auth, prefix)
		}
	}
	return ""
}

// checkOrigin 校验 Origin：未配置白名单时只放行同源请求
func (s *Server) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	allowed := s.cfg.Realtime.AllowedOrigins
	if len(allowed) == 0 {
		u, err := url.Parse(origin)
		return err == nil && strings.EqualFold(u.Host, r.Host)
	}
	for _, item := range allowed {
		if strings.EqualFold(item, origin) {
			return true
		}
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeErr(w http.ResponseWriter, status int, msg, reason string) {
	writeJSON(w, status, errResponse{Code: status, Msg: msg, Reason: reason})
}
