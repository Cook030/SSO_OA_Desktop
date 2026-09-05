// Package employee 员工管理：员工基础信息的原子读写与密码管理。
// 员工与角色的跨域编排由 Handler 用例层在事务内组合本包方法与 role.UserRoleService 完成。
package employee

import (
	"errors"
	"fmt"

	"permission-system/internal/api_model/request"
	"permission-system/internal/client"
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/repository"
	"permission-system/internal/service/rbac"
	"permission-system/internal/service/shared"
	"permission-system/internal/utils"

	"gorm.io/gorm"
)

// Service 员工管理服务
type Service struct {
	userRepo     *repository.UserRepository
	userRoleRepo *repository.UserRoleRepository
	enforcer     rbac.Enforcer
	ssoClient    *client.SSOClient
}

// New 创建员工管理服务
func New(
	userRepo *repository.UserRepository,
	userRoleRepo *repository.UserRoleRepository,
	enforcer rbac.Enforcer,
	ssoClient *client.SSOClient,
) *Service {
	return &Service{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		enforcer:     enforcer,
		ssoClient:    ssoClient,
	}
}

// EmployeeListParam 员工列表查询参数(service 层)
type EmployeeListParam struct {
	Keyword    string
	Department string
	UserIDs    []int64
	Page       int
	PageSize   int
}

// FindEmployees 员工列表(搜索/筛选/分页，默认排除管理员账号)
func (s *Service) FindEmployees(param EmployeeListParam) ([]*db_model.SysUser, int64, error) {
	adminIDs, err := s.userRoleRepo.FindUserIDsByRoleCodes([]string{shared.RoleAdmin})
	if err != nil {
		return nil, 0, err
	}

	q := repository.EmployeeListQuery{
		Keyword:        param.Keyword,
		Department:     param.Department,
		UserIDs:        param.UserIDs,
		ExcludeUserIDs: adminIDs,
		Page:           param.Page,
		PageSize:       param.PageSize,
	}

	return s.userRepo.ListEmployees(q)
}

// GetDepartments 获取部门列表(排除管理员账号所属部门数据)
func (s *Service) GetDepartments() ([]string, error) {
	adminIDs, err := s.userRoleRepo.FindUserIDsByRoleCodes([]string{shared.RoleAdmin})
	if err != nil {
		return nil, err
	}
	return s.userRepo.GetDepartments(adminIDs)
}

// CreateUserTx 在调用方事务内新增员工用户（不处理角色）
func (s *Service) CreateUserTx(tx *query.Query, req *request.CreateEmployeeRequest, operatorID int64, requestID string) (*db_model.SysUser, error) {
	// 参数校验
	if err := shared.ValidateAccount(req.Account); err != nil {
		return nil, err
	}
	if err := shared.ValidateEmployeeName(req.Name); err != nil {
		return nil, err
	}
	if err := shared.ValidatePhone(req.Phone); err != nil {
		return nil, err
	}
	if err := shared.ValidateEmailPrefix(req.EmailPrefix); err != nil {
		return nil, err
	}
	if err := shared.ValidateDepartment(req.Department); err != nil {
		return nil, err
	}
	if err := shared.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	account := req.Account
	email := req.EmailPrefix + shared.EmailDomain

	// 唯一性校验（使用事务中的 repo）
	userRepo := repository.NewUserRepository(tx)
	exists, err := userRepo.ExistsByAccount(account, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("该邮箱前缀(账号)已存在")
	}
	exists, err = userRepo.ExistsByPhone(req.Phone, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("手机号已存在")
	}
	exists, err = userRepo.ExistsByEmail(email, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("邮箱已存在")
	}

	// 加密前端传入的密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	dept := req.Department
	phone := req.Phone
	operator := operatorID
	user := &db_model.SysUser{
		Account:         account,
		Password:        hashedPassword,
		Name:            req.Name,
		Phone:           &phone,
		Email:           &email,
		Department:      &dept,
		PasswordVersion: 1,
		CreatedBy:       &operator,
		RequestID:       &requestID,
	}

	if err := userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建员工失败: %w", err)
	}
	return user, nil
}

// UpdateUserTx 在调用方事务内更新员工信息（不处理角色）
func (s *Service) UpdateUserTx(tx *query.Query, id int64, req *request.UpdateEmployeeRequest, operatorID int64, requestID string) (*db_model.SysUser, error) {
	userRepo := repository.NewUserRepository(tx)
	user, err := userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("员工不存在")
		}
		return nil, err
	}

	// 管理员账号不在员工管理范围内
	if isAdmin, err := s.isAdmin(id); err != nil {
		return nil, err
	} else if isAdmin {
		return nil, errors.New("管理员账号不可在员工管理中编辑")
	}

	// 参数校验
	if err := shared.ValidateAccount(req.Account); err != nil {
		return nil, err
	}
	if err := shared.ValidateEmployeeName(req.Name); err != nil {
		return nil, err
	}
	if err := shared.ValidatePhone(req.Phone); err != nil {
		return nil, err
	}
	if err := shared.ValidateEmailPrefix(req.EmailPrefix); err != nil {
		return nil, err
	}
	if err := shared.ValidateDepartment(req.Department); err != nil {
		return nil, err
	}

	newAccount := req.Account
	newEmail := req.EmailPrefix + shared.EmailDomain

	// 唯一性校验（使用事务中的 repo）
	if newAccount != user.Account {
		exists, err := userRepo.ExistsByAccount(newAccount, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("该邮箱前缀(账号)已存在")
		}
	}
	exists, err := userRepo.ExistsByPhone(req.Phone, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("手机号已存在")
	}
	if newEmail != deref(user.Email) {
		exists, err = userRepo.ExistsByEmail(newEmail, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("邮箱已存在")
		}
	}

	dept := req.Department
	phone := req.Phone
	operator := operatorID
	user.Name = req.Name
	user.Phone = &phone
	user.Account = newAccount
	user.Email = &newEmail
	user.Department = &dept
	user.UpdatedBy = &operator
	user.RequestID = &requestID

	if err := userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("更新员工信息失败: %w", err)
	}
	return user, nil
}

// DeleteUserTx 在调用方事务内删除员工用户（不处理角色）
func (s *Service) DeleteUserTx(tx *query.Query, id int64) error {
	userRepo := repository.NewUserRepository(tx)

	if _, err := userRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("员工不存在")
		}
		return err
	}
	if isAdmin, err := s.isAdmin(id); err != nil {
		return err
	} else if isAdmin {
		return errors.New("管理员账号不可删除")
	}

	if err := userRepo.Delete(id); err != nil {
		return fmt.Errorf("删除员工失败: %w", err)
	}
	return nil
}

// ResetPassword 重置员工密码为默认密码，递增密码版本并撤销该员工在 SSO 上的全部会话
func (s *Service) ResetPassword(employeeID int64) error {
	user, err := s.userRepo.FindByID(employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("员工不存在")
		}
		return err
	}
	if isAdmin, err := s.isAdmin(employeeID); err != nil {
		return err
	} else if isAdmin {
		return errors.New("管理员账号的密码不可在此重置")
	}

	hashedPassword, err := utils.HashPassword(shared.DefaultPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	if err := s.userRepo.ResetPassword(employeeID, hashedPassword); err != nil {
		return err
	}

	// 密码重置成功后，通知 SSO 撤销该用户的全部会话，强制其重新登录。
	if err := s.ssoClient.RevokeUserSessions(user.ID); err != nil {
		return fmt.Errorf("密码已重置，但撤销用户会话失败: %w", err)
	}

	return nil
}

// isAdmin 判断用户是否为超级管理员（内置 admin 角色）
func (s *Service) isAdmin(userID int64) (bool, error) {
	return s.enforcer.HasRole(userID, shared.RoleAdmin)
}

// deref 解引用字符串指针
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
