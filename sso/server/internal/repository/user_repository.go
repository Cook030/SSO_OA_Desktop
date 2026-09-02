package repository

import (
	"context"
	"errors"

	"mh-sso-svc/internal/model"
	"mh-sso-svc/internal/model/query"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// IsDuplicateEntryError 判断是否为唯一索引冲突（MySQL 错误码 1062）
func IsDuplicateEntryError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

// UserRepository 统一用户表（sys_user）数据访问
type UserRepository struct {
	q *query.Query
}

// NewUserRepository 创建用户 Repository
func NewUserRepository(q *query.Query) *UserRepository {
	return &UserRepository{q: q}
}

// Create 新增用户
func (r *UserRepository) Create(user *model.SysUser) error {
	return r.q.SysUser.Create(user)
}

// FindByID 按主键查询用户
func (r *UserRepository) FindByID(id uint64) (*model.SysUser, error) {
	return r.q.SysUser.Where(r.q.SysUser.ID.Eq(id)).First()
}

// FindByAccount 按登录账号查询用户（account / email / phone 任一匹配）
func (r *UserRepository) FindByAccount(account string) (*model.SysUser, error) {
	return r.q.SysUser.
		Where(r.q.SysUser.Account.Eq(account)).
		Or(r.q.SysUser.Email.Eq(account)).
		Or(r.q.SysUser.Phone.Eq(account)).
		First()
}

// UpdatePassword 更新密码哈希并将密码版本 +1
func (r *UserRepository) UpdatePassword(ctx context.Context, id uint64, passwordHash string) error {
	q := r.q.WithContext(ctx)
	_, err := q.SysUser.Where(r.q.SysUser.ID.Eq(id)).
		Updates(map[string]interface{}{
			"password":         passwordHash,
			"password_version": gorm.Expr("password_version + 1"),
		})
	return err
}

// UpdateProfile 更新用户姓名/邮箱/手机号（显式 map 写入，支持清空为 NULL）
func (r *UserRepository) UpdateProfile(ctx context.Context, id uint64, name string, email, phone *string) error {
	q := r.q.WithContext(ctx)
	_, err := q.SysUser.Where(r.q.SysUser.ID.Eq(id)).
		Updates(map[string]interface{}{
			"name":  name,
			"email": email,
			"phone": phone,
		})
	return err
}

// ExistsByEmailExclude 邮箱是否已被其他用户占用（排除自身，用于资料更新）
func (r *UserRepository) ExistsByEmailExclude(email string, excludeID uint64) (bool, error) {
	count, err := r.q.SysUser.
		Where(r.q.SysUser.Email.Eq(email), r.q.SysUser.ID.Neq(excludeID)).
		Count()
	return count > 0, err
}

// ExistsByPhoneExclude 手机号是否已被其他用户占用（排除自身，用于资料更新）
func (r *UserRepository) ExistsByPhoneExclude(phone string, excludeID uint64) (bool, error) {
	count, err := r.q.SysUser.
		Where(r.q.SysUser.Phone.Eq(phone), r.q.SysUser.ID.Neq(excludeID)).
		Count()
	return count > 0, err
}

func (r *UserRepository) existsBy(cond gen.Condition) (bool, error) {
	count, err := r.q.SysUser.Where(cond).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindPasswordVersion 返回用户当前密码版本
func (r *UserRepository) FindPasswordVersion(id uint64) (int32, error) {
	user, err := r.q.SysUser.Select(r.q.SysUser.PasswordVersion).Where(r.q.SysUser.ID.Eq(id)).First()
	if err != nil {
		return 0, err
	}
	return user.PasswordVersion, nil
}
