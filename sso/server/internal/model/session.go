package model

import "time"

// SsoSession SSO 会话表（对应 sql/sso_session.sql）
type SsoSession struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SessionID      string    `gorm:"column:session_id;size:128;uniqueIndex:uk_session_id" json:"sessionId"` // 会话唯一标识，同时作为 refresh token family
	UserID         uint64    `gorm:"column:user_id;index:idx_user_status,priority:1" json:"userId"`
	DeviceType     *string   `gorm:"column:device_type;size:32" json:"deviceType"`
	DeviceID       *string   `gorm:"column:device_id;size:128" json:"deviceId"`
	LoginIP        *string   `gorm:"column:login_ip;size:64" json:"loginIp"`
	LoginUserAgent *string   `gorm:"column:login_user_agent;size:512" json:"loginUserAgent"`
	Status         int       `gorm:"column:status;default:1;index:idx_user_status,priority:2" json:"status"`
	LastActiveAt   time.Time `gorm:"column:last_active_at" json:"lastActiveAt"`
	ExpiredAt      time.Time `gorm:"column:expired_at" json:"expiredAt"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 表名
func (SsoSession) TableName() string { return "sso_session" }

// IsActive 会话当前是否有效（状态 active 且未过期）
func (s *SsoSession) IsActive(now time.Time) bool {
	return s.Status == SessionStatusActive && s.ExpiredAt.After(now)
}
