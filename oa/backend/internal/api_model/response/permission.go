package response

// PermissionNode 权限树节点
type PermissionNode struct {
	ID         int64            `json:"id"`
	PlatformID int64            `json:"platformId"`
	Code       string           `json:"code"`
	Name       string           `json:"name"`
	Type       int32            `json:"type"`
	ParentID   *int64           `json:"parentId"`
	Sort       int32            `json:"sort"`
	Status     int32            `json:"status"`
	Children   []PermissionNode `json:"children"`
}

// MePermissionDTO 当前登录用户的权限视图
type MePermissionDTO struct {
	UserID      int64                `json:"userId"`
	Roles       []string             `json:"roles"`
	IsAdmin     bool                 `json:"isAdmin"`
	Permissions []string             `json:"permissions"`
	Platforms   []PlatformPermission `json:"platforms"`
}
