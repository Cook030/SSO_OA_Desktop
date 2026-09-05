package repository

import (
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	q *query.Query
}

// NewUserRepository 创建用户仓库
func NewUserRepository(q *query.Query) *UserRepository {
	return &UserRepository{q: q}
}

// FindByID 根据ID查找用户
func (r *UserRepository) FindByID(id int64) (*db_model.SysUser, error) {
	return r.q.SysUser.Where(r.q.SysUser.ID.Eq(id)).First()
}

// Create 创建用户
func (r *UserRepository) Create(user *db_model.SysUser) error {
	return r.q.SysUser.Create(user)
}

// Update 更新用户
func (r *UserRepository) Update(user *db_model.SysUser) error {
	_, err := r.q.SysUser.Where(r.q.SysUser.ID.Eq(user.ID)).Updates(user)
	return err
}

// Delete 物理删除用户(关联的角色绑定由 service 层在同一事务中清理)
func (r *UserRepository) Delete(id int64) error {
	_, err := r.q.SysUser.Where(r.q.SysUser.ID.Eq(id)).Delete()
	return err
}

// EmployeeListQuery 员工列表查询参数
type EmployeeListQuery struct {
	Keyword        string
	Department     string
	UserIDs        []int64 // 白名单：仅查这些用户(由 service 层解析后传入)
	ExcludeUserIDs []int64 // 黑名单：排除这些用户(如管理员账号)
	Page           int
	PageSize       int
}

// ListEmployees 查询员工列表(带筛选和分页)
func (r *UserRepository) ListEmployees(q EmployeeListQuery) ([]*db_model.SysUser, int64, error) {
	u := r.q.SysUser.Where()

	// 关键词模糊搜索
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		u = u.Where(
			r.q.SysUser.Or(
				r.q.SysUser.Name.Like(like),
				r.q.SysUser.Phone.Like(like),
				r.q.SysUser.Email.Like(like),
			),
		)
	}

	// 部门精确筛选
	if q.Department != "" {
		u = u.Where(r.q.SysUser.Department.Eq(q.Department))
	}

	// 按用户ID列表筛选(由 service 层从角色/权限表查出后传入)
	if len(q.UserIDs) > 0 {
		u = u.Where(r.q.SysUser.ID.In(q.UserIDs...))
	}

	// 排除指定用户(如管理员账号)
	if len(q.ExcludeUserIDs) > 0 {
		u = u.Where(r.q.SysUser.ID.NotIn(q.ExcludeUserIDs...))
	}

	// 统计总数
	total, err := u.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (q.Page - 1) * q.PageSize
	users, err := u.Order(r.q.SysUser.ID.Desc()).Offset(offset).Limit(q.PageSize).Find()
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetDepartments 获取部门列表(可排除指定用户，如管理员账号)
func (r *UserRepository) GetDepartments(excludeUserIDs []int64) ([]string, error) {
	u := r.q.SysUser.
		Where(r.q.SysUser.Department.Neq(""))
	if len(excludeUserIDs) > 0 {
		u = u.Where(r.q.SysUser.ID.NotIn(excludeUserIDs...))
	}

	var departments []string
	err := u.Distinct(r.q.SysUser.Department).
		Order(r.q.SysUser.Department).
		Scan(&departments)
	return departments, err
}

// FindByAccount 根据账号查找用户
func (r *UserRepository) FindByAccount(account string) (*db_model.SysUser, error) {
	return r.q.SysUser.Where(r.q.SysUser.Account.Eq(account)).First()
}

// ExistsByAccount 检查账号是否存在(排除指定ID)
func (r *UserRepository) ExistsByAccount(account string, excludeID int64) (bool, error) {
	u := r.q.SysUser.Where(r.q.SysUser.Account.Eq(account))
	if excludeID > 0 {
		u = u.Where(r.q.SysUser.ID.Neq(excludeID))
	}
	count, err := u.Count()
	return count > 0, err
}

// ExistsByPhone 检查手机号是否存在
func (r *UserRepository) ExistsByPhone(phone string, excludeID int64) (bool, error) {
	u := r.q.SysUser.Where(r.q.SysUser.Phone.Eq(phone))
	if excludeID > 0 {
		u = u.Where(r.q.SysUser.ID.Neq(excludeID))
	}
	count, err := u.Count()
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(email string, excludeID int64) (bool, error) {
	u := r.q.SysUser.Where(r.q.SysUser.Email.Eq(email))
	if excludeID > 0 {
		u = u.Where(r.q.SysUser.ID.Neq(excludeID))
	}
	count, err := u.Count()
	return count > 0, err
}

// ResetPassword 重置用户密码，并递增密码版本使旧 token 失效
func (r *UserRepository) ResetPassword(id int64, hashedPassword string) error {
	return r.q.UnderlyingDB().Model(&db_model.SysUser{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"password":         hashedPassword,
			"password_version": gorm.Expr("password_version + 1"),
		}).Error
}
