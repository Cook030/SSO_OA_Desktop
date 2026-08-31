package repository

import (
	"permission-system/internal/db_model"
)

// FindByUserIDs 根据用户ID列表批量查询用户-平台关联记录
func (r *UserPlatformRepository) FindByUserIDs(userIDs []int64) ([]*db_model.SysUserPlatform, error) {
	if len(userIDs) == 0 {
		return []*db_model.SysUserPlatform{}, nil
	}
	return r.q.SysUserPlatform.
		Where(r.q.SysUserPlatform.UserID.In(userIDs...)).
		Order(r.q.SysUserPlatform.UserID, r.q.SysUserPlatform.PlatformID).
		Find()
}

// CountByPlatformIDs 按平台ID批量统计每个平台的授权用户数
func (r *UserPlatformRepository) CountByPlatformIDs(platformIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64)
	if len(platformIDs) == 0 {
		return result, nil
	}

	// 复杂聚合查询通过 UnderlyingDB 实现
	type countRow struct {
		PlatformID int64 `gorm:"column:platform_id"`
		Count      int64 `gorm:"column:cnt"`
	}
	var rows []countRow
	err := r.q.UnderlyingDB().Model(&db_model.SysUserPlatform{}).
		Select("platform_id, COUNT(*) AS cnt").
		Where("platform_id IN ?", platformIDs).
		Group("platform_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.PlatformID] = row.Count
	}
	return result, nil
}

// FindUserIDsByPlatformID 查询拥有指定平台权限的用户ID列表
func (r *UserPlatformRepository) FindUserIDsByPlatformID(platformID int64) ([]int64, error) {
	var userIDs []int64
	err := r.q.SysUserPlatform.
		Where(r.q.SysUserPlatform.PlatformID.Eq(platformID)).
		Select(r.q.SysUserPlatform.UserID).
		Scan(&userIDs)
	return userIDs, err
}
