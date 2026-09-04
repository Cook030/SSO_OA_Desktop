package request

// CreateRoleRequest 新增角色请求
type CreateRoleRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateRoleRequest 编辑角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      *int32 `json:"status"`
}

// AssignPermissionsRequest 给角色配置权限请求(全量覆盖)
type AssignPermissionsRequest struct {
	PermissionIDs []int64 `json:"permissionIds"`
}

// AssignUsersRequest 批量给用户分配角色请求(增量)
type AssignUsersRequest struct {
	UserIDs []int64 `json:"userIds"`
	RoleIDs []int64 `json:"roleIds"`
}

// SetUserRolesRequest 设置单个用户角色请求(全量覆盖)
type SetUserRolesRequest struct {
	RoleIDs []int64 `json:"roleIds"`
}
