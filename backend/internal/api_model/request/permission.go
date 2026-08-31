package request

// BatchSetRequest 批量设置权限请求
type BatchSetRequest struct {
	UserIDs     []int64 `json:"userIds"`
	PlatformIDs []int64 `json:"platformIds"`
}

// BatchDeleteRequest 批量删除权限请求
type BatchDeleteRequest struct {
	UserIDs     []int64 `json:"userIds"`
	PlatformIDs []int64 `json:"platformIds"`
}
