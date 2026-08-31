package service

import (
	"errors"
	"fmt"

	"permission-system/internal/api_model/request"
	"permission-system/internal/client"
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/repository"
	"permission-system/internal/service/shared"
	"permission-system/internal/utils"

	"gorm.io/gorm"
)

// EmployeeService 员工管理服务(仅负责员工基础信息 CRUD)
type EmployeeService struct {
	q         *query.Query // 仅用于事务控制
	userRepo  *repository.UserRepository
	ssoClient *client.SSOClient
}

// NewEmployeeService 创建员工管理服务
func NewEmployeeService(q *query.Query, userRepo *repository.UserRepository, ssoClient *client.SSOClient) *EmployeeService {
	return &EmployeeService{
		q:         q,
		userRepo:  userRepo,
		ssoClient: ssoClient,
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

// FindEmployees 员工列表(搜索/筛选/分页，返回原始用户数据)
func (s *EmployeeService) FindEmployees(param EmployeeListParam) ([]*db_model.SysUser, int64, error) {
	q := repository.EmployeeListQuery{
		Keyword:    param.Keyword,
		Department: param.Department,
		UserIDs:    param.UserIDs,
		Page:       param.Page,
		PageSize:   param.PageSize,
	}

	return s.userRepo.ListEmployees(q)
}

// Create 新增员工（调用方管理事务，只创建用户信息，不处理权限）
func (s *EmployeeService) Create(tx *query.Query, req *request.CreateEmployeeRequest) (*db_model.SysUser, error) {
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
	txUserRepo := repository.NewUserRepository(tx)
	exists, err := txUserRepo.ExistsByAccount(account, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("该邮箱前缀(账号)已存在")
	}
	exists, err = txUserRepo.ExistsByPhone(req.Phone, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("手机号已存在")
	}
	exists, err = txUserRepo.ExistsByEmail(email, 0)
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
	user := &db_model.SysUser{
		Account:    account,
		Password:   hashedPassword,
		Name:       req.Name,
		Phone:      req.Phone,
		Email:      email,
		Role:       "employee",
		Department: &dept,
	}

	if err := txUserRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建员工失败: %w", err)
	}
	return user, nil
}

// Update 编辑员工（调用方管理事务，只更新用户信息，不处理权限）
func (s *EmployeeService) Update(tx *query.Query, id int64, req *request.UpdateEmployeeRequest) (*db_model.SysUser, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("员工不存在")
		}
		return nil, err
	}

	// 仅允许编辑员工
	if user.Role != "employee" {
		return nil, errors.New("仅可编辑员工账号")
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
	txUserRepo := repository.NewUserRepository(tx)
	if newAccount != user.Account {
		exists, err := txUserRepo.ExistsByAccount(newAccount, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("该邮箱前缀(账号)已存在")
		}
	}
	exists, err := txUserRepo.ExistsByPhone(req.Phone, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("手机号已存在")
	}
	if newEmail != user.Email {
		exists, err = txUserRepo.ExistsByEmail(newEmail, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("邮箱已存在")
		}
	}

	dept := req.Department
	user.Name = req.Name
	user.Phone = req.Phone
	user.Account = newAccount
	user.Email = newEmail
	user.Department = &dept

	if err := txUserRepo.Update(user); err != nil {
		return nil, fmt.Errorf("更新员工信息失败: %w", err)
	}
	return user, nil
}

// Delete 删除员工(物理删除，不处理关联权限)
func (s *EmployeeService) Delete(tx *query.Query, id int64) error {
	txUserRepo := repository.NewUserRepository(tx)

	user, err := txUserRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("员工不存在")
		}
		return err
	}
	if user.Role != "employee" {
		return errors.New("仅可删除员工账号")
	}

	if err := txUserRepo.Delete(id); err != nil {
		return fmt.Errorf("删除员工失败: %w", err)
	}
	return nil
}

// GetDepartments 获取部门列表
func (s *EmployeeService) GetDepartments() ([]string, error) {
	return s.userRepo.GetDepartments()
}

// ResetPassword 重置员工密码为默认密码，并撤销该员工在 SSO 上的全部会话。
func (s *EmployeeService) ResetPassword(employeeID int64) error {
	user, err := s.userRepo.FindByID(employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("员工不存在")
		}
		return err
	}
	if user.Role != "employee" {
		return errors.New("仅可重置员工账号的密码")
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
