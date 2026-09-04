// Package permission 权限域：权限点管理（Service）与
// 用户-平台访问关系的查询（AccessService）。
package permission

import (
	"errors"
	"fmt"

	"permission-system/internal/api_model/request"
	"permission-system/internal/api_model/response"
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/repository"
	"permission-system/internal/service/rbac"
	"permission-system/internal/service/shared"
)

// Service 权限点管理服务
type Service struct {
	permRepo *repository.PermissionRepository
	enforcer rbac.Enforcer
}

// New 创建权限点管理服务
func New(permRepo *repository.PermissionRepository, enforcer rbac.Enforcer) *Service {
	return &Service{
		permRepo: permRepo,
		enforcer: enforcer,
	}
}

// ListTree 查询权限树（按 parent_id 组装，顶级节点 parent_id 为空）
func (s *Service) ListTree() ([]response.PermissionNode, error) {
	perms, err := s.permRepo.ListAll()
	if err != nil {
		return nil, err
	}

	nodes := make([]response.PermissionNode, 0, len(perms))
	index := make(map[int64]*response.PermissionNode, len(perms))
	for _, p := range perms {
		node := response.PermissionNode{
			ID:         p.ID,
			PlatformID: p.PlatformID,
			Code:       p.Code,
			Name:       p.Name,
			Type:       p.Type,
			ParentID:   p.ParentID,
			Sort:       p.Sort,
			Status:     p.Status,
			Children:   []response.PermissionNode{},
		}
		index[p.ID] = &node
		nodes = append(nodes, node)
	}

	roots := make([]response.PermissionNode, 0, len(nodes))
	for i := range nodes {
		node := &nodes[i]
		if node.ParentID == nil {
			roots = append(roots, *node)
			continue
		}
		if parent, ok := index[*node.ParentID]; ok {
			parent.Children = append(parent.Children, *node)
		} else {
			roots = append(roots, *node)
		}
	}
	return roots, nil
}

// CreateTx 在调用方事务内新增权限点（唯一性校验 + 写入 + 策略缓存失效）
func (s *Service) CreateTx(tx *query.Query, req *request.CreatePermissionRequest, operatorID int64, requestID string) (*db_model.SysPermission, error) {
	if err := shared.ValidatePermissionCode(req.Code); err != nil {
		return nil, err
	}
	if err := shared.ValidatePermissionName(req.Name); err != nil {
		return nil, err
	}
	if req.Type != shared.PermissionTypeMenu && req.Type != shared.PermissionTypeAPI {
		return nil, errors.New("权限类型只能是 1菜单 或 2API")
	}

	permRepo := repository.NewPermissionRepository(tx)
	exists, err := permRepo.ExistsByCode(req.Code, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("权限编码已存在")
	}

	operator := operatorID
	perm := &db_model.SysPermission{
		PlatformID: req.PlatformID,
		Code:       req.Code,
		Name:       req.Name,
		Type:       req.Type,
		ParentID:   req.ParentID,
		Sort:       req.Sort,
		Status:     shared.StatusEnabled,
		CreatedBy:  &operator,
		RequestID:  &requestID,
	}
	if err := permRepo.Create(perm); err != nil {
		return nil, fmt.Errorf("创建权限点失败: %w", err)
	}

	_ = s.enforcer.InvalidatePolicy()
	return perm, nil
}
