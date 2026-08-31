package service

import (
	"errors"
	"fmt"

	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/repository"
	"permission-system/internal/service/shared"

	"gorm.io/gorm"
)

// PlatformService 平台管理服务(仅负责平台基础信息 CRUD)
type PlatformService struct {
	platformRepo *repository.PlatformRepository
}

// NewPlatformService 创建平台管理服务
func NewPlatformService(platformRepo *repository.PlatformRepository) *PlatformService {
	return &PlatformService{
		platformRepo: platformRepo,
	}
}

// FindPlatforms 平台列表(分页，返回原始平台数据)
func (s *PlatformService) FindPlatforms(page, pageSize int) ([]*db_model.SysPlatform, int64, error) {
	return s.platformRepo.List(page, pageSize)
}

// Create 新增平台
func (s *PlatformService) Create(name, link string) (*db_model.SysPlatform, error) {
	// 参数校验
	if err := shared.ValidatePlatformName(name); err != nil {
		return nil, err
	}
	if err := shared.ValidatePlatformLink(link); err != nil {
		return nil, err
	}

	// 唯一性校验
	exists, err := s.platformRepo.ExistsByName(name, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("平台名称已存在")
	}
	exists, err = s.platformRepo.ExistsByLink(link, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("平台链接已存在")
	}

	platform := &db_model.SysPlatform{
		Name: name,
		Link: link,
	}
	if err := s.platformRepo.Create(platform); err != nil {
		return nil, fmt.Errorf("创建平台失败: %w", err)
	}

	return platform, nil
}

// Update 编辑平台
func (s *PlatformService) Update(id int64, name, link string) (*db_model.SysPlatform, error) {
	platform, err := s.platformRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("平台不存在")
		}
		return nil, err
	}

	// 参数校验
	if err := shared.ValidatePlatformName(name); err != nil {
		return nil, err
	}
	if err := shared.ValidatePlatformLink(link); err != nil {
		return nil, err
	}

	// 唯一性校验
	exists, err := s.platformRepo.ExistsByName(name, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("平台名称已存在")
	}
	exists, err = s.platformRepo.ExistsByLink(link, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("平台链接已存在")
	}

	platform.Name = name
	platform.Link = link
	if err := s.platformRepo.Update(platform); err != nil {
		return nil, fmt.Errorf("更新平台失败: %w", err)
	}

	return platform, nil
}

// Delete 删除平台(物理删除，不处理关联权限)
func (s *PlatformService) Delete(tx *query.Query, id int64) error {
	txPlatformRepo := repository.NewPlatformRepository(tx)

	_, err := txPlatformRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("平台不存在")
		}
		return err
	}

	if err := txPlatformRepo.Delete(id); err != nil {
		return fmt.Errorf("删除平台失败: %w", err)
	}
	return nil
}
