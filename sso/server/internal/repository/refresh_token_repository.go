package repository

import (
	"time"

	"mh-sso-svc/internal/model"

	"gorm.io/gorm"
)

// RefreshTokenRepository 刷新令牌表数据访问
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository 创建刷新令牌 Repository
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create 新增刷新令牌记录
func (r *RefreshTokenRepository) Create(token *model.SsoRefreshToken) error {
	return r.db.Create(token).Error
}

// FindByTokenHash 按 token 哈希查询
func (r *RefreshTokenRepository) FindByTokenHash(tokenHash string) (*model.SsoRefreshToken, error) {
	var token model.SsoRefreshToken
	if err := r.db.Where("token_hash = ?", tokenHash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// UpdateStatus 更新令牌状态（轮换/撤销/过期；updated_at 同时充当 used_at）
func (r *RefreshTokenRepository) UpdateStatus(id uint64, status int) error {
	return r.db.Model(&model.SsoRefreshToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// RevokeBySessionID 将指定会话（token family）下全部令牌置为指定状态
func (r *RefreshTokenRepository) RevokeBySessionID(sessionID string, status int) error {
	return r.db.Model(&model.SsoRefreshToken{}).
		Where("session_id = ? AND status = ?", sessionID, model.RefreshTokenStatusActive).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// RevokeActiveByUserID 将用户全部 active 令牌置为指定状态
func (r *RefreshTokenRepository) RevokeActiveByUserID(userID uint64, status int) error {
	return r.db.Model(&model.SsoRefreshToken{}).
		Where("user_id = ? AND status = ?", userID, model.RefreshTokenStatusActive).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}
