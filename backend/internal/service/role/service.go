// Package role 角色域：角色管理与角色授权（Service），
// 以及用户-角色归属关系（UserRoleService）。
package role

import (
	"errors"
	"fmt"

	"permission-system/internal/api_model/request"
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/repository"
	"permission-system/internal/service/rbac"
	"permission-system/internal/service/shared"

	"gorm.io/gorm"
)

// Service 角色管理服务。
// 读方法直接转发 repository；写方法（Tx 后缀）为原子操作，
// 只做领域校验与数据变更，不自行开启事务，由 Handler 用例层统一编排事务。
type Service struct {
	roleRepo     *repository.RoleRepository
	rolePermRepo *repository.RolePermissionRepository
	userRoleRepo *repository.UserRoleRepository
	enforcer     rbac.Enforcer
}

// New 创建角色管理服务
func New(
	roleRepo *repository.RoleRepository,
	rolePermRepo *repository.RolePermissionRepository,
	userRoleRepo *repository.UserRoleRepository,
	enforcer rbac.Enforcer,
) *Service {
	return &Service{
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
		userRoleRepo: userRoleRepo,
		enforcer:     enforcer,
	}
}

// ---------- 读 ----------

// List 查询全部角色
func (s *Service) List() ([]*db_model.SysRole, error) {
	return s.roleRepo.List()
}

// GetPermissionIDs 查询角色当前持有的权限ID列表
func (s *Service) GetPermissionIDs(roleID int64) ([]int64, error) {
	return s.rolePermRepo.FindPermissionIDsByRoleID(roleID)
}

// GetUserRoles 查询用户当前持有的角色
func (s *Service) GetUserRoles(userID int64) ([]*db_model.SysRole, error) {
	roleIDs, err := s.userRoleRepo.FindRoleIDsByUserID(userID)
	if err != nil {
		return nil, err
	}
	return s.roleRepo.FindByIDs(roleIDs)
}

// CountUsersByRoleIDs 批量统计角色的授权用户数
func (s *Service) CountUsersByRoleIDs(roleIDs []int64) (map[int64]int64, error) {
	return s.userRoleRepo.CountByRoleIDs(roleIDs)
}

// CountPermissionsByRoleIDs 批量统计角色持有的权限数
func (s *Service) CountPermissionsByRoleIDs(roleIDs []int64) (map[int64]int64, error) {
	return s.rolePermRepo.CountByRoleIDs(roleIDs)
}

// ---------- 写（原子操作，事务由调用方开启） ----------

// CreateTx 在调用方事务内新增角色（仅可创建非内置角色）
func (s *Service) CreateTx(tx *query.Query, req *request.CreateRoleRequest, operatorID int64, requestID string) (*db_model.SysRole, error) {
	if err := shared.ValidateRoleCode(req.Code); err != nil {
		return nil, err
	}
	if err := shared.ValidateRoleName(req.Name); err != nil {
		return nil, err
	}

	roleRepo := repository.NewRoleRepository(tx)
	exists, err := roleRepo.ExistsByCode(req.Code, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("角色编码已存在")
	}

	var description *string
	if req.Description != "" {
		d := req.Description
		description = &d
	}
	operator := operatorID
	role := &db_model.SysRole{
		Code:        req.Code,
		Name:        req.Name,
		Description: description,
		IsBuiltin:   0,
		Status:      shared.StatusEnabled,
		CreatedBy:   &operator,
		RequestID:   &requestID,
	}
	if err := roleRepo.Create(role); err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}
	return role, nil
}

// UpdateTx 在调用方事务内编辑角色（内置角色不允许停用）
func (s *Service) UpdateTx(tx *query.Query, id int64, req *request.UpdateRoleRequest, operatorID int64, requestID string) (*db_model.SysRole, error) {
	if err := shared.ValidateRoleName(req.Name); err != nil {
		return nil, err
	}

	roleRepo := repository.NewRoleRepository(tx)
	role, err := roleRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, err
	}

	if req.Status != nil && *req.Status == shared.StatusDisabled && role.IsBuiltin == 1 {
		return nil, errors.New("内置角色不可停用")
	}

	var description *string
	if req.Description != "" {
		d := req.Description
		description = &d
	}
	operator := operatorID
	role.Name = req.Name
	role.Description = description
	role.UpdatedBy = &operator
	role.RequestID = &requestID
	if req.Status != nil {
		role.Status = *req.Status
	}

	if err := roleRepo.Update(role); err != nil {
		return nil, fmt.Errorf("更新角色失败: %w", err)
	}

	// 角色状态变更会同时影响角色归属与策略，统一失效
	_ = s.enforcer.InvalidatePolicy()
	return role, nil
}

// DeleteTx 在调用方事务内删除角色（内置角色、仍有用户绑定的角色不可删除）
func (s *Service) DeleteTx(tx *query.Query, id int64) error {
	roleRepo := repository.NewRoleRepository(tx)
	userRoleRepo := repository.NewUserRoleRepository(tx)

	role, err := roleRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}
	if role.IsBuiltin == 1 {
		return errors.New("内置角色不可删除")
	}

	userIDs, err := userRoleRepo.FindUserIDsByRoleID(id)
	if err != nil {
		return err
	}
	if len(userIDs) > 0 {
		return fmt.Errorf("该角色下仍有 %d 个用户，请先解除绑定", len(userIDs))
	}

	if err := roleRepo.Delete(id); err != nil {
		return fmt.Errorf("删除角色失败: %w", err)
	}

	// 级联清理用户-角色（外键已 CASCADE，此处兜底）
	if err := userRoleRepo.DeleteByRoleID(id); err != nil {
		return err
	}

	_ = s.enforcer.InvalidatePolicy()
	return nil
}

// AssignPermissionsTx 在调用方事务内给角色配置权限（全量覆盖：先删旧授权，再插新授权）
func (s *Service) AssignPermissionsTx(tx *query.Query, roleID int64, permissionIDs []int64, operatorID int64, requestID string) error {
	roleRepo := repository.NewRoleRepository(tx)
	rolePermRepo := repository.NewRolePermissionRepository(tx)

	if _, err := roleRepo.FindByID(roleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	uniqueIDs := uniqueInt64(permissionIDs)
	if len(uniqueIDs) > 0 {
		count, err := rolePermRepo.ExistsEnabledPermission(uniqueIDs)
		if err != nil {
			return err
		}
		if count != int64(len(uniqueIDs)) {
			return errors.New("存在无效或已禁用的权限点")
		}
	}

	if err := rolePermRepo.DeleteByRoleID(roleID); err != nil {
		return err
	}
	if _, err := rolePermRepo.BatchInsert(roleID, uniqueIDs, operatorID, requestID); err != nil {
		return fmt.Errorf("角色授权失败: %w", err)
	}

	_ = s.enforcer.InvalidatePolicy()
	return nil
}
