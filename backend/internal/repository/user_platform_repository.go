package repository

import (
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"

	"gorm.io/gorm/clause"
)

// UserPlatformRepository 用户平台权限关系数据访问层
type UserPlatformRepository struct {
	q *query.Query
}

// NewUserPlatformRepository 创建用户平台权限仓库
func NewUserPlatformRepository(q *query.Query) *UserPlatformRepository {
	return &UserPlatformRepository{q: q}
}

// DeleteByUserID 删除用户的所有平台权限
func (r *UserPlatformRepository) DeleteByUserID(userID int64) error {
	_, err := r.q.SysUserPlatform.Where(r.q.SysUserPlatform.UserID.Eq(userID)).Delete()
	return err
}

// DeleteByPlatformID 删除平台的所有用户权限关联
func (r *UserPlatformRepository) DeleteByPlatformID(platformID int64) error {
	_, err := r.q.SysUserPlatform.Where(r.q.SysUserPlatform.PlatformID.Eq(platformID)).Delete()
	return err
}

// BatchInsert 增量批量插入权限关系(已存在的不重复插入)，返回新增数量
func (r *UserPlatformRepository) BatchInsert(userIDs, platformIDs []int64) (int64, error) {
	if len(userIDs) == 0 || len(platformIDs) == 0 {
		return 0, nil
	}

	var records []db_model.SysUserPlatform
	for _, uid := range userIDs {
		for _, pid := range platformIDs {
			records = append(records, db_model.SysUserPlatform{
				UserID:     uid,
				PlatformID: pid,
			})
		}
	}

	// INSERT IGNORE 不支持 gen 直接表达，通过 UnderlyingDB 实现
	result := r.q.UnderlyingDB().Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// BatchDeleteByUserIDsAndPlatformIDs 批量删除多个用户的指定平台权限，返回删除数量
func (r *UserPlatformRepository) BatchDeleteByUserIDsAndPlatformIDs(userIDs, platformIDs []int64) (int64, error) {
	if len(userIDs) == 0 || len(platformIDs) == 0 {
		return 0, nil
	}
	result, err := r.q.SysUserPlatform.Where(
		r.q.SysUserPlatform.UserID.In(userIDs...),
		r.q.SysUserPlatform.PlatformID.In(platformIDs...),
	).Delete()
	return result.RowsAffected, err
}
