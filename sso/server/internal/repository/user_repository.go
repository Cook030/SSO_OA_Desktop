package repository

import (
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
func (r *UserRepository) UpdatePassword(id uint64, passwordHash string) error {
	_, err := r.q.SysUser.Where(r.q.SysUser.ID.Eq(id)).
		Updates(map[string]interface{}{
			"password":         passwordHash,
			"password_version": gorm.Expr("password_version + 1"),
		})
	return err
}

// ExistsByAccount 登录账号是否已存在
func (r *UserRepository) ExistsByAccount(account string) (bool, error) {
	return r.existsBy(r.q.SysUser.Account.Eq(account))
}

// ExistsByEmail 邮箱是否已存在
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	return r.existsBy(r.q.SysUser.Email.Eq(email))
}

// ExistsByPhone 手机号是否已存在
func (r *UserRepository) ExistsByPhone(phone string) (bool, error) {
	return r.existsBy(r.q.SysUser.Phone.Eq(phone))
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
