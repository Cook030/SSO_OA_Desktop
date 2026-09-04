package permission

import (
	"permission-system/internal/api_model/response"
	"permission-system/internal/db_model"
	"permission-system/internal/repository"
	"permission-system/internal/service/rbac"
	"permission-system/internal/service/shared"
)

// AccessService 平台访问权解析服务（只读）。
//
// 基于"平台访问权限点 platform:<code>:access + 角色授权 + 用户角色"推导
// 用户与平台之间的访问关系，以及当前用户的完整权限视图。
type AccessService struct {
	platformRepo  *repository.PlatformRepository
	permRepo      *repository.PermissionRepository
	rolePermRepo  *repository.RolePermissionRepository
	userRoleRepo  *repository.UserRoleRepository
	enforcer      rbac.Enforcer
}

// NewAccessService 创建平台访问权解析服务
func NewAccessService(
	platformRepo *repository.PlatformRepository,
	permRepo *repository.PermissionRepository,
	rolePermRepo *repository.RolePermissionRepository,
	userRoleRepo *repository.UserRoleRepository,
	enforcer rbac.Enforcer,
) *AccessService {
	return &AccessService{
		platformRepo: platformRepo,
		permRepo:     permRepo,
		rolePermRepo: rolePermRepo,
		userRoleRepo: userRoleRepo,
		enforcer:     enforcer,
	}
}

// FindPlatformsByUserID 查询单个用户可访问的平台
func (s *AccessService) FindPlatformsByUserID(userID int64) ([]response.PlatformPermission, error) {
	platforms, err := s.platformRepo.FindAll()
	if err != nil {
		return nil, err
	}
	permIDs, err := s.platformPermissionIDs(platforms)
	if err != nil {
		return nil, err
	}
	roleCodesByPerm, err := s.rolePermRepo.FindRoleCodesByPermissionIDs(valuesOf(permIDs))
	if err != nil {
		return nil, err
	}

	userRoles, err := s.enforcer.LoadRoles(userID)
	if err != nil {
		return nil, err
	}

	result := make([]response.PlatformPermission, 0, len(platforms))
	for _, p := range platforms {
		permID, ok := permIDs[p.ID]
		if !ok {
			continue
		}
		if intersects(roleCodesByPerm[permID], userRoles) {
			result = append(result, response.PlatformPermission{ID: p.ID, Name: p.Name})
		}
	}
	return result, nil
}

// FindPlatformsByUserIDs 批量查询多个用户可访问的平台（员工列表用）
func (s *AccessService) FindPlatformsByUserIDs(userIDs []int64) (map[int64][]response.PlatformPermission, error) {
	result := make(map[int64][]response.PlatformPermission, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	platforms, err := s.platformRepo.FindAll()
	if err != nil {
		return nil, err
	}
	permIDs, err := s.platformPermissionIDs(platforms)
	if err != nil {
		return nil, err
	}
	roleCodesByPerm, err := s.rolePermRepo.FindRoleCodesByPermissionIDs(valuesOf(permIDs))
	if err != nil {
		return nil, err
	}
	rolesByUser, err := s.userRoleRepo.BatchFindRoleCodesByUserIDs(userIDs)
	if err != nil {
		return nil, err
	}

	for _, userID := range userIDs {
		userRoles := rolesByUser[userID]
		list := make([]response.PlatformPermission, 0, len(platforms))
		for _, p := range platforms {
			permID, ok := permIDs[p.ID]
			if !ok {
				continue
			}
			if intersects(roleCodesByPerm[permID], userRoles) {
				list = append(list, response.PlatformPermission{ID: p.ID, Name: p.Name})
			}
		}
		result[userID] = list
	}
	return result, nil
}

// FindUserIDsByPlatformID 查询可访问指定平台的用户ID列表
func (s *AccessService) FindUserIDsByPlatformID(platformID int64) ([]int64, error) {
	platform, err := s.platformRepo.FindByID(platformID)
	if err != nil {
		return nil, err
	}

	perm, err := s.permRepo.FindByCode(shared.PlatformAccessCode(platform.Code))
	if err != nil {
		return []int64{}, nil // 平台未生成访问权限点，视为无人可访问
	}

	roleIDs, err := s.rolePermRepo.FindRoleIDsByPermissionIDs([]int64{perm.ID})
	if err != nil {
		return nil, err
	}
	if len(roleIDs[perm.ID]) == 0 {
		return []int64{}, nil
	}
	return s.userRoleRepo.FindUserIDsByRoleIDs(roleIDs[perm.ID])
}

// CountByPlatformIDs 批量统计各平台的可访问用户数
func (s *AccessService) CountByPlatformIDs(platformIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(platformIDs))
	if len(platformIDs) == 0 {
		return result, nil
	}

	platforms, err := s.platformRepo.FindByIDs(platformIDs)
	if err != nil {
		return nil, err
	}
	permIDs, err := s.platformPermissionIDs(platforms)
	if err != nil {
		return nil, err
	}
	roleIDsByPerm, err := s.rolePermRepo.FindRoleIDsByPermissionIDs(valuesOf(permIDs))
	if err != nil {
		return nil, err
	}

	for _, p := range platforms {
		permID, ok := permIDs[p.ID]
		if !ok {
			result[p.ID] = 0
			continue
		}
		userIDs, err := s.userRoleRepo.FindUserIDsByRoleIDs(roleIDsByPerm[permID])
		if err != nil {
			return nil, err
		}
		result[p.ID] = int64(len(dedup(userIDs)))
	}
	return result, nil
}

// platformPermissionIDs 查询各平台对应的访问权限点ID：platformID -> permissionID
func (s *AccessService) platformPermissionIDs(platforms []*db_model.SysPlatform) (map[int64]int64, error) {
	result := make(map[int64]int64, len(platforms))
	if len(platforms) == 0 {
		return result, nil
	}

	codes := make([]string, 0, len(platforms))
	codeToPlatform := make(map[string]int64, len(platforms))
	for _, p := range platforms {
		code := shared.PlatformAccessCode(p.Code)
		codes = append(codes, code)
		codeToPlatform[code] = p.ID
	}

	perms, err := s.permRepo.FindByCodes(codes)
	if err != nil {
		return nil, err
	}
	for _, perm := range perms {
		if platformID, ok := codeToPlatform[perm.Code]; ok {
			result[platformID] = perm.ID
		}
	}
	return result, nil
}
