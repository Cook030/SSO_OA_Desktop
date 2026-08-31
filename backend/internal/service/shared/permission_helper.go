package shared

import (
	"permission-system/internal/api_model/response"
	"permission-system/internal/db_model"
	"permission-system/internal/repository"
)

// BuildPermissionsMap 批量查询多用户的平台权限，返回 map[userID][]PlatformPermission
func BuildPermissionsMap(userIDs []int64, upRepo *repository.UserPlatformRepository, pRepo *repository.PlatformRepository) (map[int64][]response.PlatformPermission, error) {
	if len(userIDs) == 0 {
		return map[int64][]response.PlatformPermission{}, nil
	}

	// 查询用户-平台关联记录
	userPlatforms, err := upRepo.FindByUserIDs(userIDs)
	if err != nil {
		return nil, err
	}

	// 收集所有不重复的 platform_id
	platformIDSet := make(map[int64]struct{})
	for _, up := range userPlatforms {
		platformIDSet[up.PlatformID] = struct{}{}
	}
	platformIDs := make([]int64, 0, len(platformIDSet))
	for pid := range platformIDSet {
		platformIDs = append(platformIDs, pid)
	}

	// 查询平台详情
	platforms, err := pRepo.FindByIDs(platformIDs)
	if err != nil {
		return nil, err
	}
	platformMap := make(map[int64]*db_model.SysPlatform, len(platforms))
	for _, p := range platforms {
		platformMap[p.ID] = p
	}

	// 内存聚合: user_id -> []PlatformPermission
	result := make(map[int64][]response.PlatformPermission, len(userIDs))
	for _, up := range userPlatforms {
		if p, ok := platformMap[up.PlatformID]; ok {
			result[up.UserID] = append(result[up.UserID], response.PlatformPermission{
				ID:   p.ID,
				Name: p.Name,
			})
		}
	}
	return result, nil
}

// BuildPermissionsByUserID 查询单个用户的平台权限列表
func BuildPermissionsByUserID(userID int64, upRepo *repository.UserPlatformRepository, pRepo *repository.PlatformRepository) ([]response.PlatformPermission, error) {
	m, err := BuildPermissionsMap([]int64{userID}, upRepo, pRepo)
	if err != nil {
		return nil, err
	}
	if perms, ok := m[userID]; ok {
		return perms, nil
	}
	return []response.PlatformPermission{}, nil
}
