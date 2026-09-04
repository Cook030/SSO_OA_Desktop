package repository

import (
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"

	"gorm.io/gen/field"
)

// RoleRepository 角色数据访问层
type RoleRepository struct {
	q *query.Query
}

// NewRoleRepository 创建角色仓库
func NewRoleRepository(q *query.Query) *RoleRepository {
	return &RoleRepository{q: q}
}

// Create 创建角色
func (r *RoleRepository) Create(role *db_model.SysRole) error {
	return r.q.SysRole.Create(role)
}

// FindByID 根据ID查找角色
func (r *RoleRepository) FindByID(id int64) (*db_model.SysRole, error) {
	return r.q.SysRole.Where(r.q.SysRole.ID.Eq(id)).First()
}

// FindByCode 根据编码查找角色
func (r *RoleRepository) FindByCode(code string) (*db_model.SysRole, error) {
	return r.q.SysRole.Where(r.q.SysRole.Code.Eq(code)).First()
}

// Update 更新角色（status 可能为 0，故显式指定字段，避免零值被忽略）
func (r *RoleRepository) Update(role *db_model.SysRole) error {
	description := ""
	if role.Description != nil {
		description = *role.Description
	}
	assigns := []field.AssignExpr{
		r.q.SysRole.Name.Value(role.Name),
		r.q.SysRole.Description.Value(description),
		r.q.SysRole.Status.Value(role.Status),
	}
	if role.UpdatedBy != nil {
		assigns = append(assigns, r.q.SysRole.UpdatedBy.Value(*role.UpdatedBy))
	}
	if role.RequestID != nil {
		assigns = append(assigns, r.q.SysRole.RequestID.Value(*role.RequestID))
	}

	_, err := r.q.SysRole.Where(r.q.SysRole.ID.Eq(role.ID)).UpdateSimple(assigns...)
	return err
}

// Delete 物理删除角色（关联的角色-权限、用户-角色由外键 ON DELETE CASCADE 清理）
func (r *RoleRepository) Delete(id int64) error {
	_, err := r.q.SysRole.Where(r.q.SysRole.ID.Eq(id)).Delete()
	return err
}

// List 查询全部角色（内置角色在前，其余按ID升序）
func (r *RoleRepository) List() ([]*db_model.SysRole, error) {
	return r.q.SysRole.
		Order(r.q.SysRole.IsBuiltin.Desc(), r.q.SysRole.ID).
		Find()
}

// ExistsByCode 检查角色编码是否存在(排除指定ID)
func (r *RoleRepository) ExistsByCode(code string, excludeID int64) (bool, error) {
	p := r.q.SysRole.Where(r.q.SysRole.Code.Eq(code))
	if excludeID > 0 {
		p = p.Where(r.q.SysRole.ID.Neq(excludeID))
	}
	count, err := p.Count()
	return count > 0, err
}

// FindByIDs 根据ID列表批量查询角色
func (r *RoleRepository) FindByIDs(ids []int64) ([]*db_model.SysRole, error) {
	if len(ids) == 0 {
		return []*db_model.SysRole{}, nil
	}
	return r.q.SysRole.Where(r.q.SysRole.ID.In(ids...)).Find()
}

// FindByCodes 根据编码列表批量查询角色
func (r *RoleRepository) FindByCodes(codes []string) ([]*db_model.SysRole, error) {
	if len(codes) == 0 {
		return []*db_model.SysRole{}, nil
	}
	return r.q.SysRole.Where(r.q.SysRole.Code.In(codes...)).Find()
}
