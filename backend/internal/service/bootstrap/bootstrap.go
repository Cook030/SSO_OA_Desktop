// Package bootstrap RBAC 基础数据初始化：
// 根平台、内置角色、内置权限点、角色授权与管理员账号。
package bootstrap

import (
	"errors"

	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/repository"
	"permission-system/internal/service/shared"
	"permission-system/internal/utils"

	"gorm.io/gorm"
)

// 初始化内置数据使用的操作标识
const (
	// RootPlatformCode 后台根平台编码，后台功能权限点统一挂载在该平台下
	RootPlatformCode = "oa_manage"
	// RootPlatformName 后台根平台名称
	RootPlatformName = "OA管理系统"
	// SeedRequestID 初始化数据的 request_id
	SeedRequestID = "system-init"
)

// builtinPermission 内置权限点定义
type builtinPermission struct {
	Code string
	Name string
	Type int32
}

// builtinPermissions 后台内置权限点。
// 编码格式 <object>:<action>，其中 object/action 与路由注册处的声明一一对应。
var builtinPermissions = []builtinPermission{
	// 平台管理
	{"platform:list", "查看平台", shared.PermissionTypeAPI},
	{"platform:create", "创建平台", shared.PermissionTypeAPI},
	{"platform:update", "编辑平台", shared.PermissionTypeAPI},
	{"platform:delete", "删除平台", shared.PermissionTypeAPI},
	// 员工管理
	{"employee:list", "查看员工", shared.PermissionTypeAPI},
	{"employee:create", "创建员工", shared.PermissionTypeAPI},
	{"employee:update", "编辑员工", shared.PermissionTypeAPI},
	{"employee:delete", "删除员工", shared.PermissionTypeAPI},
	{"employee:reset-password", "重置员工密码", shared.PermissionTypeAPI},
	// 角色与权限
	{"role:list", "查看角色", shared.PermissionTypeAPI},
	{"role:create", "创建角色", shared.PermissionTypeAPI},
	{"role:update", "编辑角色", shared.PermissionTypeAPI},
	{"role:delete", "删除角色", shared.PermissionTypeAPI},
	{"role:assign", "配置角色权限", shared.PermissionTypeAPI},
	{"user:role:assign", "分配用户角色", shared.PermissionTypeAPI},
	{"permission:list", "查看权限点", shared.PermissionTypeAPI},
	{"permission:create", "创建权限点", shared.PermissionTypeAPI},
	// 菜单（前端展示用）
	{"menu:platform", "平台管理菜单", shared.PermissionTypeMenu},
	{"menu:employee", "员工管理菜单", shared.PermissionTypeMenu},
	{"menu:role", "角色权限菜单", shared.PermissionTypeMenu},
}

// builtinRolePermissions 各内置角色的默认权限编码。
// admin 使用 "*:*" 全通配，后续新增的权限点自动生效。
var builtinRolePermissions = map[string][]string{
	shared.RoleAdmin: {"*:*"},
	shared.RoleHR: {
		"platform:list", "platform:create", "platform:update", "platform:delete",
		"employee:list", "employee:create", "employee:update", "employee:delete", "employee:reset-password",
		"role:list", "user:role:assign", "permission:list",
		"menu:platform", "menu:employee", "menu:role",
	},
	shared.RoleManager: {
		"platform:list", "employee:list", "role:list", "permission:list",
		"menu:platform", "menu:employee",
	},
	shared.RoleEmployee: {},
}

// InitRBAC 初始化 RBAC 基础数据：根平台、内置角色、内置权限点、角色授权，以及管理员账号。
// 该方法幂等，可安全地在每次启动时执行。
func InitRBAC(q *query.Query, adminCfg *utils.AdminConfig) error {
	return q.Transaction(func(tx *query.Query) error {
		if err := initRootPlatform(tx); err != nil {
			return err
		}
		if err := initBuiltinRoles(tx); err != nil {
			return err
		}
		if err := initBuiltinPermissions(tx); err != nil {
			return err
		}
		if err := initRolePermissions(tx); err != nil {
			return err
		}
		return initAdminUser(tx, adminCfg)
	})
}

// initRootPlatform 创建后台根平台（后台权限点的归属平台）
func initRootPlatform(tx *query.Query) error {
	platformRepo := repository.NewPlatformRepository(tx)
	_, err := platformRepo.FindByCode(RootPlatformCode)
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	platform := &db_model.SysPlatform{
		Name:      RootPlatformName,
		Code:      RootPlatformCode,
		RequestID: strPtr(SeedRequestID),
	}
	return platformRepo.Create(platform)
}

// initBuiltinRoles 创建四个内置角色
func initBuiltinRoles(tx *query.Query) error {
	roleRepo := repository.NewRoleRepository(tx)
	for _, def := range shared.BuiltinRoles {
		if _, err := roleRepo.FindByCode(def.Code); err == nil {
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		description := def.Description
		role := &db_model.SysRole{
			Code:        def.Code,
			Name:        def.Name,
			Description: &description,
			IsBuiltin:   1,
			Status:      shared.StatusEnabled,
			RequestID:   strPtr(SeedRequestID),
		}
		if err := roleRepo.Create(role); err != nil {
			return err
		}
	}
	return nil
}

// initBuiltinPermissions 创建后台内置权限点
func initBuiltinPermissions(tx *query.Query) error {
	platform, err := repository.NewPlatformRepository(tx).FindByCode(RootPlatformCode)
	if err != nil {
		return err
	}

	permRepo := repository.NewPermissionRepository(tx)
	for _, def := range builtinPermissions {
		if _, err := permRepo.FindByCode(def.Code); err == nil {
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		perm := &db_model.SysPermission{
			PlatformID: platform.ID,
			Code:       def.Code,
			Name:       def.Name,
			Type:       def.Type,
			Sort:       0,
			Status:     shared.StatusEnabled,
			RequestID:  strPtr(SeedRequestID),
		}
		if err := permRepo.Create(perm); err != nil {
			return err
		}
	}
	return nil
}

// initRolePermissions 为内置角色初始化授权（仅在角色尚无授权时执行，避免覆盖手工调整）
func initRolePermissions(tx *query.Query) error {
	roleRepo := repository.NewRoleRepository(tx)
	permRepo := repository.NewPermissionRepository(tx)
	rolePermRepo := repository.NewRolePermissionRepository(tx)

	for roleCode, codes := range builtinRolePermissions {
		role, err := roleRepo.FindByCode(roleCode)
		if err != nil {
			return err
		}

		existing, err := rolePermRepo.FindPermissionIDsByRoleID(role.ID)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			continue
		}
		if len(codes) == 0 {
			continue
		}

		// admin 使用通配符，需确保其对应的权限点存在
		if roleCode == shared.RoleAdmin {
			if _, err := permRepo.FindByCode(shared.Wildcard + ":" + shared.Wildcard); err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				platform, err := repository.NewPlatformRepository(tx).FindByCode(RootPlatformCode)
				if err != nil {
					return err
				}
				perm := &db_model.SysPermission{
					PlatformID: platform.ID,
					Code:       shared.Wildcard + ":" + shared.Wildcard,
					Name:       "全部权限",
					Type:       shared.PermissionTypeAPI,
					Status:     shared.StatusEnabled,
					RequestID:  strPtr(SeedRequestID),
				}
				if err := permRepo.Create(perm); err != nil {
					return err
				}
			}
			perm, err := permRepo.FindByCode(shared.Wildcard + ":" + shared.Wildcard)
			if err != nil {
				return err
			}
			if _, err := rolePermRepo.BatchInsert(role.ID, []int64{perm.ID}, 0, SeedRequestID); err != nil {
				return err
			}
			continue
		}

		permissionIDs := make([]int64, 0, len(codes))
		for _, code := range codes {
			perm, err := permRepo.FindByCode(code)
			if err != nil {
				return err
			}
			permissionIDs = append(permissionIDs, perm.ID)
		}
		if _, err := rolePermRepo.BatchInsert(role.ID, permissionIDs, 0, SeedRequestID); err != nil {
			return err
		}
	}
	return nil
}

// initAdminUser 创建管理员账号并绑定 admin 角色
func initAdminUser(tx *query.Query, adminCfg *utils.AdminConfig) error {
	if adminCfg == nil {
		return nil
	}

	userRepo := repository.NewUserRepository(tx)
	roleRepo := repository.NewRoleRepository(tx)
	userRoleRepo := repository.NewUserRoleRepository(tx)

	role, err := roleRepo.FindByCode(shared.RoleAdmin)
	if err != nil {
		return err
	}

	user, err := userRepo.FindByAccount(adminCfg.Account)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		hashedPassword, err := utils.HashPassword(adminCfg.Password)
		if err != nil {
			return err
		}
		phone := adminCfg.Phone
		email := adminCfg.Email
		user = &db_model.SysUser{
			Account:         adminCfg.Account,
			Password:        hashedPassword,
			Name:            adminCfg.Name,
			Phone:           &phone,
			Email:           &email,
			PasswordVersion: 1,
			RequestID:       strPtr(SeedRequestID),
		}
		if err := userRepo.Create(user); err != nil {
			return err
		}
	}

	roleIDs, err := userRoleRepo.FindRoleIDsByUserID(user.ID)
	if err != nil {
		return err
	}
	for _, id := range roleIDs {
		if id == role.ID {
			return nil
		}
	}

	_, err = userRoleRepo.BatchInsert([]int64{user.ID}, []int64{role.ID}, 0, SeedRequestID)
	return err
}

// strPtr 返回字符串指针
func strPtr(s string) *string {
	return &s
}
