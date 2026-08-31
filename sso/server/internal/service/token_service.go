package service

import (
	"errors"
	"strconv"
	"time"

	"mh-sso-svc/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

// access token 解析错误（区别过期与非法，便于返回准确的 401 提示）
var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

// AccessClaims JWT payload（需求文档第八节）：
// iss=签发方，sub=用户ID，sid=会话ID，jti=token 唯一ID，account=登录名，passwordVersion=密码版本
type AccessClaims struct {
	UserID          uint64 `json:"-"` // 从 sub 解析，不参与序列化
	SessionID       string `json:"sid"`
	Account         string `json:"account"`
	PasswordVersion int    `json:"passwordVersion"`
	jwt.RegisteredClaims
}

// TokenService access token（JWT/HS256）签发与校验
type TokenService struct {
	issuer    string
	secret    []byte
	accessTTL time.Duration
}

// NewTokenService 创建 TokenService
func NewTokenService(cfg *utils.AuthConfig) *TokenService {
	return &TokenService{
		issuer:    cfg.Issuer,
		secret:    []byte(cfg.JWTSecret),
		accessTTL: time.Duration(cfg.AccessTokenTTLSecond) * time.Second,
	}
}

// AccessTokenTTL access token 有效期
func (s *TokenService) AccessTokenTTL() time.Duration {
	return s.accessTTL
}

// GenerateAccessToken 签发 access token
func (s *TokenService) GenerateAccessToken(userID uint64, sessionID, account string, passwordVersion int, now time.Time) (token, jti string, expiresAt time.Time, err error) {
	jti = "access_" + utils.RandomHex(16)
	expiresAt = now.Add(s.accessTTL)

	claims := &AccessClaims{
		SessionID:       sessionID,
		Account:         account,
		PasswordVersion: passwordVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   strconv.FormatUint(userID, 10),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	return token, jti, expiresAt, err
}

// ParseAccessToken 校验并解析 access token（签名算法、签名、issuer、过期时间）
func (s *TokenService) ParseAccessToken(tokenString string) (*AccessClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		// 仅允许 HMAC 家族签名算法，防止算法混淆攻击
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer), jwt.WithExpirationRequired())
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := parsed.Claims.(*AccessClaims)
	if !ok || !parsed.Valid {
		return nil, ErrTokenInvalid
	}

	userID, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || userID == 0 || claims.SessionID == "" {
		return nil, ErrTokenInvalid
	}
	claims.UserID = userID
	return claims, nil
}
