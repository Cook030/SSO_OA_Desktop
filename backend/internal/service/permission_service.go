package service

import (
	"errors"

	"permission-system/internal/api_model/response"
	"permission-system/internal/db_model/query"
	"permission-system/internal/repository"
	"permission-system/internal/service/shared"
)

// PermissionService 权限管理服务(权限操作的唯一入口)
type PermissionService struct {
	q                *query.Query
	userPlatformRepo *repository.UserPlatformRepository
	platformRepo     *repository.PlatformRepository
}

// NewPermissionService 创建权限管理服务
func NewPermissionService(q *query.Query, userPlatformRepo *repository.UserPlatformRepository, platformRepo *repository.PlatformRepository) *PermissionService {
	return &PermissionService{
		q:                q,
		userPlatformRepo: userPlatformRepo,
		platformRepo:     platformRepo,
	}
}

// ---------- 公开查询方法 ----------

// GetPermissionsByUserIDs 批量查询多用户的平台权限
func (s *PermissionService) GetPermissionsByUserIDs(userIDs []int64) (map[int64][]response.PlatformPermission, error) {
	return shared.BuildPermissionsMap(userIDs, s.userPlatformRepo, s.platformRepo)
}

// GetPermissionsByUserID 查询单个用户的平台权限列表
func (s *PermissionService) GetPermissionsByUserID(userID int64) ([]response.PlatformPermission, error) {
	return shared.BuildPermissionsByUserID(userID, s.userPlatformRepo, s.platformRepo)
}

// FindUserIDsByPlatformID 查询拥有指定平台权限的用户ID列表
func (s *PermissionService) FindUserIDsByPlatformID(platformID int64) ([]int64, error) {
	return s.userPlatformRepo.FindUserIDsByPlatformID(platformID)
}

// ---------- 公开写入方法 ----------

// BatchSet 批量设置权限(增量新增，不重复插入)
func (s *PermissionService) BatchSet(tx *query.Query, userIDs, platformIDs []int64) (*response.BatchPermissionResultDTO, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("userIds不能为空")
	}
	if len(platformIDs) == 0 {
		return nil, errors.New("platformIds不能为空")
	}

	txRepo := repository.NewUserPlatformRepository(tx)
	affected, err := txRepo.BatchInsert(userIDs, platformIDs)
	if err != nil {
		return nil, err
	}

	return &response.BatchPermissionResultDTO{AffectedCount: affected}, nil
}

// BatchDelete 批量删除权限(删除选中员工的指定平台权限)
func (s *PermissionService) BatchDelete(tx *query.Query, userIDs, platformIDs []int64) (*response.BatchPermissionResultDTO, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("userIds不能为空")
	}
	if len(platformIDs) == 0 {
		return nil, errors.New("platformIds不能为空")
	}

	txRepo := repository.NewUserPlatformRepository(tx)
	affected, err := txRepo.BatchDeleteByUserIDsAndPlatformIDs(userIDs, platformIDs)
	if err != nil {
		return nil, err
	}

	return &response.BatchPermissionResultDTO{AffectedCount: affected}, nil
}

// DeleteByUserID 删除某用户的全部平台权限
func (s *PermissionService) DeleteByUserID(tx *query.Query, userID int64) error {
	txRepo := repository.NewUserPlatformRepository(tx)
	return txRepo.DeleteByUserID(userID)
}

// DeleteByPlatformID 删除某平台的全部用户关联权限
func (s *PermissionService) DeleteByPlatformID(tx *query.Query, platformID int64) error {
	txRepo := repository.NewUserPlatformRepository(tx)
	return txRepo.DeleteByPlatformID(platformID)
}

// CountByPlatformIDs 批量统计平台的授权用户数
func (s *PermissionService) CountByPlatformIDs(platformIDs []int64) (map[int64]int64, error) {
	return s.userPlatformRepo.CountByPlatformIDs(platformIDs)
}
