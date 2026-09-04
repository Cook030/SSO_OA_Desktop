package role

import (
	"errors"
	"fmt"

	"permission-system/internal/db_model/query"
	"permission-system/internal/repository"
	"permission-system/internal/service/rbac"
	"permission-system/internal/service/shared"

	"gorm.io/gorm"
)

// UserRoleService 用户-角色归属服务
type UserRoleService struct {
	userRoleRepo *repository.UserRoleRepository
	enforcer     rbac.Enforcer
}

// NewUserRoleService 创建用户-角色归属服务
func NewUserRoleService(
	userRoleRepo *repository.UserRoleRepository,
	enforcer rbac.Enforcer,
) *UserRoleService {
	return &UserRoleService{
		userRoleRepo: userRoleRepo,
		enforcer:     enforcer,
	}
}

// FindRolesByUserIDs 批量查询用户的角色(员工列表展示用)
func (s *UserRoleService) FindRolesByUserIDs(userIDs []int64) (map[int64][]repository.RoleBrief, error) {
	return s.userRoleRepo.BatchFindRolesByUserIDs(userIDs)
}

// AssignUsersTx 在调用方事务内批量给用户分配角色（增量，已存在的不重复插入）
func (s *UserRoleService) AssignUsersTx(tx *query.Query, userIDs, roleIDs []int64, operatorID int64, requestID string) (int64, error) {
	if len(userIDs) == 0 {
		return 0, errors.New("userIds不能为空")
	}
	if len(roleIDs) == 0 {
		return 0, errors.New("roleIds不能为空")
	}

	roles, err := repository.NewRoleRepository(tx).FindByIDs(uniqueInt64(roleIDs))
	if err != nil {
		return 0, err
	}
	if len(roles) != len(uniqueInt64(roleIDs)) {
		return 0, errors.New("存在无效的角色")
	}

	affected, err := repository.NewUserRoleRepository(tx).BatchInsert(userIDs, roleIDs, operatorID, requestID)
	if err != nil {
		return 0, fmt.Errorf("分配角色失败: %w", err)
	}

	for _, uid := range userIDs {
		s.enforcer.InvalidateUser(uid)
	}
	return affected, nil
}

// SetUserRolesTx 在调用方事务内全量覆盖单个用户的角色
func (s *UserRoleService) SetUserRolesTx(tx *query.Query, userID int64, roleIDs []int64, operatorID int64, requestID string) error {
	roleRepo := repository.NewRoleRepository(tx)
	userRoleRepo := repository.NewUserRoleRepository(tx)

	uniqueIDs := uniqueInt64(roleIDs)
	if len(uniqueIDs) > 0 {
		roles, err := roleRepo.FindByIDs(uniqueIDs)
		if err != nil {
			return err
		}
		if len(roles) != len(uniqueIDs) {
			return errors.New("存在无效的角色")
		}
	}

	if err := userRoleRepo.DeleteByUserID(userID); err != nil {
		return err
	}
	if _, err := userRoleRepo.BatchInsert([]int64{userID}, uniqueIDs, operatorID, requestID); err != nil {
		return fmt.Errorf("设置用户角色失败: %w", err)
	}

	s.enforcer.InvalidateUser(userID)
	return nil
}

// EnsureEmployeeRoleTx 在调用方事务内为无任何角色的用户绑定内置 employee 角色
func (s *UserRoleService) EnsureEmployeeRoleTx(tx *query.Query, userID int64, operatorID int64, requestID string) error {
	userRoleRepo := repository.NewUserRoleRepository(tx)
	roleIDs, err := userRoleRepo.FindRoleIDsByUserID(userID)
	if err != nil {
		return err
	}
	if len(roleIDs) > 0 {
		return nil
	}

	employeeRole, err := repository.NewRoleRepository(tx).FindByCode(shared.RoleEmployee)
	if err != nil {
		// 内置角色缺失时不阻断员工创建，交由初始化脚本补齐
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if _, err := userRoleRepo.BatchInsert([]int64{userID}, []int64{employeeRole.ID}, operatorID, requestID); err != nil {
		return err
	}
	s.enforcer.InvalidateUser(userID)
	return nil
}

// RevokeUserRolesTx 在调用方事务内删除用户的全部角色（删除员工时调用）
func (s *UserRoleService) RevokeUserRolesTx(tx *query.Query, userID int64) error {
	if err := repository.NewUserRoleRepository(tx).DeleteByUserID(userID); err != nil {
		return err
	}
	s.enforcer.InvalidateUser(userID)
	return nil
}
