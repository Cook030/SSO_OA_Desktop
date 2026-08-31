package repository

import (
	"time"

	"mh-sso-svc/internal/model"

	"gorm.io/gorm"
)

// SessionRepository 会话表数据访问
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository 创建会话 Repository
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create 新增会话
func (r *SessionRepository) Create(session *model.SsoSession) error {
	return r.db.Create(session).Error
}

// FindBySessionID 按会话 ID 查询
func (r *SessionRepository) FindBySessionID(sessionID string) (*model.SsoSession, error) {
	var session model.SsoSession
	if err := r.db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateStatusBySessionID 更新指定会话状态
func (r *SessionRepository) UpdateStatusBySessionID(sessionID string, status int) error {
	return r.db.Model(&model.SsoSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// RevokeActiveByUserID 将用户全部 active 会话置为指定状态
func (r *SessionRepository) RevokeActiveByUserID(userID uint64, status int) error {
	return r.db.Model(&model.SsoSession{}).
		Where("user_id = ? AND status = ?", userID, model.SessionStatusActive).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// ListActiveSessionIDs 查询用户全部 active 会话的 session_id（用于清理缓存）
func (r *SessionRepository) ListActiveSessionIDs(userID uint64) ([]string, error) {
	var sessionIDs []string
	err := r.db.Model(&model.SsoSession{}).
		Where("user_id = ? AND status = ?", userID, model.SessionStatusActive).
		Pluck("session_id", &sessionIDs).Error
	if err != nil {
		return nil, err
	}
	return sessionIDs, nil
}

// Touch 更新会话最近活跃时间并滑动续期
func (r *SessionRepository) Touch(sessionID string, lastActiveAt, expiredAt time.Time) error {
	return r.db.Model(&model.SsoSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"last_active_at": lastActiveAt,
			"expired_at":     expiredAt,
			"updated_at":     lastActiveAt,
		}).Error
}
