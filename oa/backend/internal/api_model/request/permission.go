package request

// CreatePermissionRequest 新增权限点请求
type CreatePermissionRequest struct {
	PlatformID int64  `json:"platformId"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Type       int32  `json:"type"`
	ParentID   *int64 `json:"parentId"`
	Sort       int32  `json:"sort"`
}
