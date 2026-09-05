// Package platform 平台管理：平台 CRUD，
// 以及平台访问权限点（platform:<code>:access）的同步维护。
package platform

import (
	"errors"
	"fmt"

	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/repository"
	"permission-system/internal/service/rbac"
	"permission-system/internal/service/shared"

	"gorm.io/gorm"
)

// Service 平台管理服务
type Service struct {
	platformRepo *repository.PlatformRepository
	enforcer     rbac.Enforcer
}

// New 创建平台管理服务
func New(platformRepo *repository.PlatformRepository, enforcer rbac.Enforcer) *Service {
	return &Service{
		platformRepo: platformRepo,
		enforcer:     enforcer,
	}
}

// FindPlatforms 平台列表(分页，返回原始平台数据)
func (s *Service) FindPlatforms(page, pageSize int) ([]*db_model.SysPlatform, int64, error) {
	return s.platformRepo.List(page, pageSize)
}

// CreateTx 在调用方事务内新增平台，并同步生成该平台的访问权限点
func (s *Service) CreateTx(tx *query.Query, name, code string, operatorID int64, requestID string) (*db_model.SysPlatform, error) {
	if err := shared.ValidatePlatformName(name); err != nil {
		return nil, err
	}
	if err := shared.ValidatePlatformCode(code); err != nil {
		return nil, err
	}

	platformRepo := repository.NewPlatformRepository(tx)
	exists, err := platformRepo.ExistsByName(name, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("平台名称已存在")
	}
	exists, err = platformRepo.ExistsByCode(code, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("平台编码已存在")
	}

	operator := operatorID
	platform := &db_model.SysPlatform{
		Name:      name,
		Code:      code,
		CreatedBy: &operator,
		RequestID: &requestID,
	}
	if err := platformRepo.Create(platform); err != nil {
		return nil, fmt.Errorf("创建平台失败: %w", err)
	}

	// 平台访问权限点：platform:<平台编码>:access，由角色授予后生效
	perm := &db_model.SysPermission{
		PlatformID: platform.ID,
		Code:       shared.PlatformAccessCode(code),
		Name:       "访问" + name,
		Type:       shared.PermissionTypeAPI,
		Sort:       0,
		Status:     shared.StatusEnabled,
		CreatedBy:  &operator,
		RequestID:  &requestID,
	}
	if err := repository.NewPermissionRepository(tx).Create(perm); err != nil {
		return nil, fmt.Errorf("创建平台访问权限失败: %w", err)
	}

	_ = s.enforcer.InvalidatePolicy()
	return platform, nil
}

// UpdateTx 在调用方事务内编辑平台；平台编码变更时同步刷新访问权限点的编码
func (s *Service) UpdateTx(tx *query.Query, id int64, name, code string, operatorID int64, requestID string) (*db_model.SysPlatform, error) {
	if err := shared.ValidatePlatformName(name); err != nil {
		return nil, err
	}
	if err := shared.ValidatePlatformCode(code); err != nil {
		return nil, err
	}

	platformRepo := repository.NewPlatformRepository(tx)
	platform, err := platformRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("平台不存在")
		}
		return nil, err
	}

	exists, err := platformRepo.ExistsByName(name, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("平台名称已存在")
	}
	exists, err = platformRepo.ExistsByCode(code, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("平台编码已存在")
	}

	oldCode := platform.Code
	operator := operatorID
	platform.Name = name
	platform.Code = code
	platform.UpdatedBy = &operator
	platform.RequestID = &requestID

	if err := platformRepo.Update(platform); err != nil {
		return nil, fmt.Errorf("更新平台失败: %w", err)
	}

	// 平台编码是权限码的组成部分，编码变化需同步权限点
	if oldCode != code {
		permRepo := repository.NewPermissionRepository(tx)
		perm, err := permRepo.FindByCode(shared.PlatformAccessCode(oldCode))
		if err == nil {
			perm.Code = shared.PlatformAccessCode(code)
			perm.Name = "访问" + name
			perm.UpdatedBy = &operator
			perm.RequestID = &requestID
			if err := permRepo.UpdateCode(perm); err != nil {
				return nil, fmt.Errorf("同步平台权限编码失败: %w", err)
			}
		}
		_ = s.enforcer.InvalidatePolicy()
	}

	return platform, nil
}

// DeleteTx 在调用方事务内删除平台及其关联权限点（规避外键 RESTRICT）
func (s *Service) DeleteTx(tx *query.Query, id int64) error {
	platformRepo := repository.NewPlatformRepository(tx)

	if _, err := platformRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("平台不存在")
		}
		return err
	}

	if err := repository.NewPermissionRepository(tx).DeleteByPlatformID(id); err != nil {
		return err
	}
	if err := platformRepo.Delete(id); err != nil {
		return fmt.Errorf("删除平台失败: %w", err)
	}

	_ = s.enforcer.InvalidatePolicy()
	return nil
}
