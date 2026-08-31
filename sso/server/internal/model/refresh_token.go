package model

import "time"

// SsoRefreshToken SSO 刷新令牌表（对应 sql/sso_refresh_token.sql）
// token_hash 只存 refresh token 的 SHA-256 哈希，明文不落库；
// 同一 session_id 下的 token 旋转链即 token family，重放检测时按 family 整体撤销
type SsoRefreshToken struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID   string    `gorm:"column:session_id;size:128;index:idx_session_status,priority:1" json:"sessionId"`
	UserID      uint64    `gorm:"column:user_id;index:idx_user_status,priority:1" json:"userId"`
	TokenHash   string    `gorm:"column:token_hash;size:64;uniqueIndex:uk_token_hash" json:"-"` // SHA-256 哈希
	Status      int       `gorm:"column:status;default:1;index:idx_session_status,priority:2;index:idx_user_status,priority:2" json:"status"`
	ExpiredAt   time.Time `gorm:"column:expired_at" json:"expiredAt"`
	RotatedFrom *uint64   `gorm:"column:rotated_from;index:idx_rotated_from" json:"rotatedFrom"` // 由哪个 token 轮换而来
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"` // 轮换时即 used_at
}

// TableName 表名
func (SsoRefreshToken) TableName() string { return "sso_refresh_token" }
