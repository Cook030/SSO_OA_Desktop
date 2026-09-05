package repository

import (
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/service/shared"

	"gorm.io/gen/field"
)

// PermissionRepository 权限点数据访问层
type PermissionRepository struct {
	q *query.Query
}

// NewPermissionRepository 创建权限点仓库
func NewPermissionRepository(q *query.Query) *PermissionRepository {
	return &PermissionRepository{q: q}
}

// PolicyRow 策略行：角色编码 -> 权限编码
type PolicyRow struct {
	RoleCode string `gorm:"column:role_code"`
	PermCode string `gorm:"column:perm_code"`
}

// Create 创建权限点
func (r *PermissionRepository) Create(perm *db_model.SysPermission) error {
	return r.q.SysPermission.Create(perm)
}

// FindByID 根据ID查找权限点
func (r *PermissionRepository) FindByID(id int64) (*db_model.SysPermission, error) {
	return r.q.SysPermission.Where(r.q.SysPermission.ID.Eq(id)).First()
}

// FindByCode 根据权限编码查找权限点
func (r *PermissionRepository) FindByCode(code string) (*db_model.SysPermission, error) {
	return r.q.SysPermission.Where(r.q.SysPermission.Code.Eq(code)).First()
}

// FindByIDs 根据ID列表批量查询权限点
func (r *PermissionRepository) FindByIDs(ids []int64) ([]*db_model.SysPermission, error) {
	if len(ids) == 0 {
		return []*db_model.SysPermission{}, nil
	}
	return r.q.SysPermission.Where(r.q.SysPermission.ID.In(ids...)).Find()
}

// ListAll 查询全部权限点（按平台、排序号、ID）
func (r *PermissionRepository) ListAll() ([]*db_model.SysPermission, error) {
	return r.q.SysPermission.
		Order(r.q.SysPermission.PlatformID, r.q.SysPermission.Sort, r.q.SysPermission.ID).
		Find()
}

// ListByPlatformID 查询指定平台下的全部权限点
func (r *PermissionRepository) ListByPlatformID(platformID int64) ([]*db_model.SysPermission, error) {
	return r.q.SysPermission.
		Where(r.q.SysPermission.PlatformID.Eq(platformID)).
		Order(r.q.SysPermission.Sort, r.q.SysPermission.ID).
		Find()
}

// UpdateCode 更新权限点的编码与名称（平台编码变更时同步）
func (r *PermissionRepository) UpdateCode(perm *db_model.SysPermission) error {
	assigns := []field.AssignExpr{
		r.q.SysPermission.Code.Value(perm.Code),
		r.q.SysPermission.Name.Value(perm.Name),
	}
	if perm.UpdatedBy != nil {
		assigns = append(assigns, r.q.SysPermission.UpdatedBy.Value(*perm.UpdatedBy))
	}
	if perm.RequestID != nil {
		assigns = append(assigns, r.q.SysPermission.RequestID.Value(*perm.RequestID))
	}

	_, err := r.q.SysPermission.Where(r.q.SysPermission.ID.Eq(perm.ID)).UpdateSimple(assigns...)
	return err
}

// FindByCodes 根据权限编码列表批量查询权限点
func (r *PermissionRepository) FindByCodes(codes []string) ([]*db_model.SysPermission, error) {
	if len(codes) == 0 {
		return []*db_model.SysPermission{}, nil
	}
	return r.q.SysPermission.Where(r.q.SysPermission.Code.In(codes...)).Find()
}

// ExistsByCode 检查权限编码是否存在(排除指定ID)
func (r *PermissionRepository) ExistsByCode(code string, excludeID int64) (bool, error) {
	p := r.q.SysPermission.Where(r.q.SysPermission.Code.Eq(code))
	if excludeID > 0 {
		p = p.Where(r.q.SysPermission.ID.Neq(excludeID))
	}
	count, err := p.Count()
	return count > 0, err
}

// DeleteByPlatformID 删除某平台下的全部权限点（角色-权限关联由外键 CASCADE 清理）
func (r *PermissionRepository) DeleteByPlatformID(platformID int64) error {
	_, err := r.q.SysPermission.Where(r.q.SysPermission.PlatformID.Eq(platformID)).Delete()
	return err
}

// ListPolicy 加载全量策略：仅包含启用中的角色与启用中的权限点
func (r *PermissionRepository) ListPolicy() ([]PolicyRow, error) {
	var rows []PolicyRow
	err := r.q.UnderlyingDB().
		Table("sys_role AS r").
		Select("r.code AS role_code, p.code AS perm_code").
		Joins("JOIN sys_role_permission AS rp ON rp.role_id = r.id").
		Joins("JOIN sys_permission AS p ON p.id = rp.permission_id").
		Where("r.status = ? AND p.status = ?", shared.StatusEnabled, shared.StatusEnabled).
		Scan(&rows).Error
	return rows, err
}

// FindByRoleID 查询某角色持有的权限点
func (r *PermissionRepository) FindByRoleID(roleID int64) ([]*db_model.SysPermission, error) {
	var perms []*db_model.SysPermission
	err := r.q.UnderlyingDB().
		Table("sys_permission AS p").
		Select("p.*").
		Joins("JOIN sys_role_permission AS rp ON rp.permission_id = p.id").
		Where("rp.role_id = ?", roleID).
		Order("p.sort, p.id").
		Scan(&perms).Error
	return perms, err
}
