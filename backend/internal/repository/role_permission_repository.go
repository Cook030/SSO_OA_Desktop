package repository

import (
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/service/shared"

	"gorm.io/gorm/clause"
)

// RolePermissionRepository 角色-权限关系数据访问层
type RolePermissionRepository struct {
	q *query.Query
}

// NewRolePermissionRepository 创建角色-权限关系仓库
func NewRolePermissionRepository(q *query.Query) *RolePermissionRepository {
	return &RolePermissionRepository{q: q}
}

// FindPermissionIDsByRoleID 查询角色持有的权限ID列表
func (r *RolePermissionRepository) FindPermissionIDsByRoleID(roleID int64) ([]int64, error) {
	var ids []int64
	err := r.q.SysRolePermission.
		Where(r.q.SysRolePermission.RoleID.Eq(roleID)).
		Select(r.q.SysRolePermission.PermissionID).
		Scan(&ids)
	return ids, err
}

// BatchInsert 增量批量授权(已存在的不重复插入)，返回新增数量
func (r *RolePermissionRepository) BatchInsert(roleID int64, permissionIDs []int64, operatorID int64, requestID string) (int64, error) {
	if len(permissionIDs) == 0 {
		return 0, nil
	}

	records := make([]db_model.SysRolePermission, 0, len(permissionIDs))
	createdBy := operatorID
	for _, pid := range permissionIDs {
		records = append(records, db_model.SysRolePermission{
			RoleID:       roleID,
			PermissionID: pid,
			CreatedBy:    &createdBy,
			RequestID:    &requestID,
		})
	}

	result := r.q.UnderlyingDB().Clauses(clause.OnConflict{DoNothing: true}).Create(&records)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// DeleteByRoleID 删除角色的全部权限授权
func (r *RolePermissionRepository) DeleteByRoleID(roleID int64) error {
	_, err := r.q.SysRolePermission.Where(r.q.SysRolePermission.RoleID.Eq(roleID)).Delete()
	return err
}

// DeleteByRoleIDAndPermissionIDs 删除角色的指定权限授权
func (r *RolePermissionRepository) DeleteByRoleIDAndPermissionIDs(roleID int64, permissionIDs []int64) error {
	if len(permissionIDs) == 0 {
		return nil
	}
	_, err := r.q.SysRolePermission.Where(
		r.q.SysRolePermission.RoleID.Eq(roleID),
		r.q.SysRolePermission.PermissionID.In(permissionIDs...),
	).Delete()
	return err
}

// FindRoleIDsByPermissionID 查询持有该权限的角色ID列表
func (r *RolePermissionRepository) FindRoleIDsByPermissionID(permissionID int64) ([]int64, error) {
	var ids []int64
	err := r.q.SysRolePermission.
		Where(r.q.SysRolePermission.PermissionID.Eq(permissionID)).
		Select(r.q.SysRolePermission.RoleID).
		Scan(&ids)
	return ids, err
}

// CountByRoleIDs 批量统计角色的授权权限数
func (r *RolePermissionRepository) CountByRoleIDs(roleIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(roleIDs))
	if len(roleIDs) == 0 {
		return result, nil
	}

	type countRow struct {
		RoleID int64 `gorm:"column:role_id"`
		Count  int64 `gorm:"column:cnt"`
	}
	var rows []countRow
	err := r.q.UnderlyingDB().Model(&db_model.SysRolePermission{}).
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

// FindRoleIDsByPermissionIDs 批量查询持有指定权限的角色ID，返回 permissionID -> []roleID
func (r *RolePermissionRepository) FindRoleIDsByPermissionIDs(permissionIDs []int64) (map[int64][]int64, error) {
	result := make(map[int64][]int64, len(permissionIDs))
	if len(permissionIDs) == 0 {
		return result, nil
	}

	type row struct {
		PermissionID int64 `gorm:"column:permission_id"`
		RoleID       int64 `gorm:"column:role_id"`
	}
	var rows []row
	err := r.q.SysRolePermission.
		Where(r.q.SysRolePermission.PermissionID.In(permissionIDs...)).
		Select(r.q.SysRolePermission.PermissionID, r.q.SysRolePermission.RoleID).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.PermissionID] = append(result[r.PermissionID], r.RoleID)
	}
	return result, nil
}

// FindRoleCodesByPermissionIDs 批量查询持有指定权限的角色编码，返回 permissionID -> []roleCode
func (r *RolePermissionRepository) FindRoleCodesByPermissionIDs(permissionIDs []int64) (map[int64][]string, error) {
	result := make(map[int64][]string, len(permissionIDs))
	if len(permissionIDs) == 0 {
		return result, nil
	}

	type row struct {
		PermissionID int64  `gorm:"column:permission_id"`
		RoleCode     string `gorm:"column:role_code"`
	}
	var rows []row
	err := r.q.UnderlyingDB().
		Table("sys_role_permission AS rp").
		Select("rp.permission_id AS permission_id, r.code AS role_code").
		Joins("JOIN sys_role AS r ON r.id = rp.role_id").
		Where("rp.permission_id IN ? AND r.status = ?", permissionIDs, shared.StatusEnabled).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.PermissionID] = append(result[r.PermissionID], r.RoleCode)
	}
	return result, nil
}

// ExistsEnabledPermission 校验权限点是否存在且启用
func (r *RolePermissionRepository) ExistsEnabledPermission(permissionIDs []int64) (int64, error) {
	return r.q.SysPermission.Where(
		r.q.SysPermission.ID.In(permissionIDs...),
		r.q.SysPermission.Status.Eq(shared.StatusEnabled),
	).Count()
}
