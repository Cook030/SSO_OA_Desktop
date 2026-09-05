package repository

import (
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/service/shared"

	"gorm.io/gorm/clause"
)

// UserRoleRepository 用户-角色关系数据访问层
type UserRoleRepository struct {
	q *query.Query
}

// NewUserRoleRepository 创建用户-角色关系仓库
func NewUserRoleRepository(q *query.Query) *UserRoleRepository {
	return &UserRoleRepository{q: q}
}

// codeRow 单列查询结果
type codeRow struct {
	Code string `gorm:"column:code"`
}

// FindRoleIDsByUserID 查询用户持有的角色ID列表
func (r *UserRoleRepository) FindRoleIDsByUserID(userID int64) ([]int64, error) {
	var ids []int64
	err := r.q.SysUserRole.
		Where(r.q.SysUserRole.UserID.Eq(userID)).
		Select(r.q.SysUserRole.RoleID).
		Scan(&ids)
	return ids, err
}

// FindRoleCodesByUserID 查询用户持有的角色编码列表(仅启用中的角色)
func (r *UserRoleRepository) FindRoleCodesByUserID(userID int64) ([]string, error) {
	var rows []codeRow
	err := r.q.UnderlyingDB().
		Table("sys_user_role AS ur").
		Select("r.code AS code").
		Joins("JOIN sys_role AS r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.status = ?", userID, shared.StatusEnabled).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	codes := make([]string, 0, len(rows))
	for _, row := range rows {
		codes = append(codes, row.Code)
	}
	return codes, nil
}

// BatchFindRoleCodesByUserIDs 批量查询多个用户的角色编码
func (r *UserRoleRepository) BatchFindRoleCodesByUserIDs(userIDs []int64) (map[int64][]string, error) {
	result := make(map[int64][]string, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	type row struct {
		UserID int64  `gorm:"column:user_id"`
		Code   string `gorm:"column:code"`
	}
	var rows []row
	err := r.q.UnderlyingDB().
		Table("sys_user_role AS ur").
		Select("ur.user_id AS user_id, r.code AS code").
		Joins("JOIN sys_role AS r ON r.id = ur.role_id").
		Where("ur.user_id IN ? AND r.status = ?", userIDs, shared.StatusEnabled).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.UserID] = append(result[r.UserID], r.Code)
	}
	return result, nil
}

// RoleBrief 角色简要信息（用于列表展示，避免 N+1 查询）
type RoleBrief struct {
	ID        int64
	Code      string
	Name      string
	IsBuiltin int32
}

// FindUserIDsByRoleCodes 查询持有任一指定角色编码的用户ID列表(去重)
func (r *UserRoleRepository) FindUserIDsByRoleCodes(roleCodes []string) ([]int64, error) {
	var ids []int64
	if len(roleCodes) == 0 {
		return ids, nil
	}
	err := r.q.UnderlyingDB().
		Table("sys_user_role AS ur").
		Joins("JOIN sys_role AS r ON r.id = ur.role_id").
		Where("r.code IN ?", roleCodes).
		Distinct().
		Pluck("ur.user_id", &ids).Error
	return ids, err
}

// BatchFindRolesByUserIDs 批量查询多个用户持有的角色
func (r *UserRoleRepository) BatchFindRolesByUserIDs(userIDs []int64) (map[int64][]RoleBrief, error) {
	result := make(map[int64][]RoleBrief, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	type row struct {
		UserID    int64  `gorm:"column:user_id"`
		RoleID    int64  `gorm:"column:role_id"`
		Code      string `gorm:"column:code"`
		Name      string `gorm:"column:name"`
		IsBuiltin int32  `gorm:"column:is_builtin"`
	}
	var rows []row
	err := r.q.UnderlyingDB().
		Table("sys_user_role AS ur").
		Select("ur.user_id AS user_id, r.id AS role_id, r.code AS code, r.name AS name, r.is_builtin AS is_builtin").
		Joins("JOIN sys_role AS r ON r.id = ur.role_id").
		Where("ur.user_id IN ?", userIDs).
		Order("r.is_builtin DESC, r.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.UserID] = append(result[r.UserID], RoleBrief{
			ID:        r.RoleID,
			Code:      r.Code,
			Name:      r.Name,
			IsBuiltin: r.IsBuiltin,
		})
	}
	return result, nil
}

// BatchInsert 增量批量授权(已存在的不重复插入)，返回新增数量
func (r *UserRoleRepository) BatchInsert(userIDs, roleIDs []int64, operatorID int64, requestID string) (int64, error) {
	if len(userIDs) == 0 || len(roleIDs) == 0 {
		return 0, nil
	}

	records := make([]db_model.SysUserRole, 0, len(userIDs)*len(roleIDs))
	createdBy := operatorID
	for _, uid := range userIDs {
		for _, rid := range roleIDs {
			records = append(records, db_model.SysUserRole{
				UserID:    uid,
				RoleID:    rid,
				CreatedBy: &createdBy,
				RequestID: &requestID,
			})
		}
	}

	result := r.q.UnderlyingDB().Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// DeleteByUserID 删除用户的全部角色
func (r *UserRoleRepository) DeleteByUserID(userID int64) error {
	_, err := r.q.SysUserRole.Where(r.q.SysUserRole.UserID.Eq(userID)).Delete()
	return err
}

// DeleteByUserIDAndRoleIDs 删除用户的指定角色
func (r *UserRoleRepository) DeleteByUserIDAndRoleIDs(userID int64, roleIDs []int64) error {
	if len(roleIDs) == 0 {
		return nil
	}
	_, err := r.q.SysUserRole.Where(
		r.q.SysUserRole.UserID.Eq(userID),
		r.q.SysUserRole.RoleID.In(roleIDs...),
	).Delete()
	return err
}

// DeleteByRoleID 删除角色的全部用户关联
func (r *UserRoleRepository) DeleteByRoleID(roleID int64) error {
	_, err := r.q.SysUserRole.Where(r.q.SysUserRole.RoleID.Eq(roleID)).Delete()
	return err
}

// FindUserIDsByRoleID 查询持有该角色的用户ID列表
func (r *UserRoleRepository) FindUserIDsByRoleID(roleID int64) ([]int64, error) {
	var ids []int64
	err := r.q.SysUserRole.
		Where(r.q.SysUserRole.RoleID.Eq(roleID)).
		Select(r.q.SysUserRole.UserID).
		Scan(&ids)
	return ids, err
}

// FindUserIDsByRoleIDs 查询持有任一指定角色的用户ID列表(去重)
func (r *UserRoleRepository) FindUserIDsByRoleIDs(roleIDs []int64) ([]int64, error) {
	var ids []int64
	if len(roleIDs) == 0 {
		return ids, nil
	}
	err := r.q.SysUserRole.
		Where(r.q.SysUserRole.RoleID.In(roleIDs...)).
		Distinct(r.q.SysUserRole.UserID).
		Select(r.q.SysUserRole.UserID).
		Scan(&ids)
	return ids, err
}

// CountByRoleIDs 批量统计角色的授权用户数
func (r *UserRoleRepository) CountByRoleIDs(roleIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(roleIDs))
	if len(roleIDs) == 0 {
		return result, nil
	}

	type countRow struct {
		RoleID int64 `gorm:"column:role_id"`
		Count  int64 `gorm:"column:cnt"`
	}
	var rows []countRow
	err := r.q.UnderlyingDB().Model(&db_model.SysUserRole{}).
		Select("role_id, COUNT(*) AS cnt").
		Where("role_id IN ?", roleIDs).
		Group("role_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.RoleID] = row.Count
	}
	return result, nil
}
