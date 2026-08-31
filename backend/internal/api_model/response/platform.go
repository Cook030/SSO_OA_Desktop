package response

import "time"

// PlatformPermission 平台权限摘要
type PlatformPermission struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// PlatformListItemDTO 平台列表项
type PlatformListItemDTO struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Link            string    `json:"link"`
	PermissionCount int64     `json:"permissionCount"`
	CreateTime      time.Time `json:"createTime"`
}

// PlatformDTO 平台基本信息
type PlatformDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}

// PlatformPageResult 平台分页结果
type PlatformPageResult struct {
	Total int64                 `json:"total"`
	List  []PlatformListItemDTO `json:"list"`
}
