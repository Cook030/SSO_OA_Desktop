package response

// RoleListItemDTO 角色列表项
type RoleListItemDTO struct {
	ID              int64  `json:"id"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	IsBuiltin       int32  `json:"isBuiltin"`
	Status          int32  `json:"status"`
	UserCount       int64  `json:"userCount"`
	PermissionCount int64  `json:"permissionCount"`
	CreateTime      string `json:"createTime"`
}

// RoleDTO 角色详情
type RoleDTO struct {
	ID          int64   `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	IsBuiltin   int32   `json:"isBuiltin"`
	Status      int32   `json:"status"`
	Permissions []int64 `json:"permissions"`
}

// RoleOptionDTO 角色下拉选项
type RoleOptionDTO struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsBuiltin int32  `json:"isBuiltin"`
}

// BatchAssignResultDTO 批量授权结果
type BatchAssignResultDTO struct {
	AffectedCount int64 `json:"affectedCount"`
}
