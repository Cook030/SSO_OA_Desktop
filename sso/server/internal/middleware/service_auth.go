package middleware

import (
	"crypto/subtle"
	"net"
	"net/http"

	"mh-sso-svc/internal/utils"

	"github.com/gin-gonic/gin"
)

// ServiceTokenHeader 服务间共享密钥的请求头
const ServiceTokenHeader = "X-MH-Service-Token"

// ServiceAuthMiddleware 内部接口的服务间鉴权（当前用于 WebSocket Gateway → SSO）。
//
// 两道检查：
//  1. 共享密钥：X-MH-Service-Token 与 internal.service_token 常量时间比较，防时序探测；
//  2. 来源网段：internal.allow_cidrs，为空表示不限制（生产环境应显式配置）。
//
// 未配置 service_token 时一律拒绝 —— 内部接口必须显式开启，避免"部署后默认裸奔"。
func ServiceAuthMiddleware(cfg *utils.InternalConfig) gin.HandlerFunc {
	secret := cfg.ServiceToken
	cidrs := parseCIDRs(cfg.AllowCIDRs)

	return func(c *gin.Context) {
		if secret == "" {
			utils.Error(c, http.StatusServiceUnavailable, "内部接口未配置服务间密钥")
			c.Abort()
			return
		}
		if subtle.ConstantTimeCompare([]byte(c.GetHeader(ServiceTokenHeader)), []byte(secret)) != 1 {
			utils.Error(c, http.StatusUnauthorized, "服务间鉴权失败")
			c.Abort()
			return
		}
		if len(cidrs) > 0 {
			ip := net.ParseIP(c.ClientIP())
			if ip == nil || !ipInCIDRs(ip, cidrs) {
				utils.Error(c, http.StatusForbidden, "来源地址不允许访问内部接口")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// parseCIDRs 解析网段白名单，跳过非法项
func parseCIDRs(list []string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(list))
	for _, item := range list {
		_, n, err := net.ParseCIDR(item)
		if err != nil {
			continue
		}
		out = append(out, n)
	}
	return out
}

func ipInCIDRs(ip net.IP, cidrs []*net.IPNet) bool {
	for _, n := range cidrs {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
