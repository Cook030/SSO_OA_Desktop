package handler

import (
	"permission-system/internal/api_model/request"
	"permission-system/internal/api_model/response"
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/middleware"
	"permission-system/internal/service/permission"
	"permission-system/internal/service/rbac"
	"permission-system/internal/service/shared"
	"permission-system/internal/utils"

	"github.com/gin-gonic/gin"
)

// PermissionHandler 权限管理处理器
//
// 职责：解析请求 -> 编排用例（写用例在事务内调用 service 原子方法）-> 组装响应。
type PermissionHandler struct {
	permissionService *permission.Service
	accessService     *permission.AccessService
	enforcer          rbac.Enforcer // 当前用户权限视图的只读来源
	q                 *query.Query  // 用例层事务控制
}

// NewPermissionHandler 创建权限管理处理器
func NewPermissionHandler(
	permissionService *permission.Service,
	accessService *permission.AccessService,
	enforcer rbac.Enforcer,
	q *query.Query,
) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
		accessService:     accessService,
		enforcer:          enforcer,
		q:                 q,
	}
}

// ListTree	权限树
//
//	@Summary		获取权限树
//	@Description	按父子关系返回全部权限点，供角色授权时勾选
//	@Tags			权限管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Success		200	{object}	utils.Response{data=[]response.PermissionNode}
//	@Failure		401	{object}	utils.Response
//	@Failure		403	{object}	utils.Response
//	@Router			/permissions [get]
func (h *PermissionHandler) ListTree(c *gin.Context) {
	tree, err := h.permissionService.ListTree()
	if err != nil {
		utils.ServerError(c, "查询权限树失败")
		return
	}

	utils.OK(c, tree)
}

// Create	新增权限点
//
//	@Summary		创建权限点
//	@Description	新增一个权限点（菜单或 API）
//	@Tags			权限管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			request	body		request.CreatePermissionRequest			true	"权限点信息"
//	@Success		200		{object}	utils.Response{data=response.PermissionNode}
//	@Failure		400		{object}	utils.Response
//	@Failure		401		{object}	utils.Response
//	@Failure		403		{object}	utils.Response
//	@Router			/permissions [post]
func (h *PermissionHandler) Create(c *gin.Context) {
	var req request.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	operatorID := middleware.CurrentUserID(c)
	requestID := middleware.GetRequestID(c)

	var perm *db_model.SysPermission
	if err := h.q.Transaction(func(tx *query.Query) error {
		var err error
		perm, err = h.permissionService.CreateTx(tx, &req, operatorID, requestID)
		return err
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "权限点创建成功", response.PermissionNode{
		ID:         perm.ID,
		PlatformID: perm.PlatformID,
		Code:       perm.Code,
		Name:       perm.Name,
		Type:       perm.Type,
		ParentID:   perm.ParentID,
		Sort:       perm.Sort,
		Status:     perm.Status,
		Children:   []response.PermissionNode{},
	})
}

// Me	当前用户权限
//
//	@Summary		获取当前用户权限
//	@Description	返回当前登录用户的角色、权限编码与可访问平台，供前端做菜单与按钮控制
//	@Tags			权限管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Success		200	{object}	utils.Response{data=response.MePermissionDTO}
//	@Failure		401	{object}	utils.Response
//	@Router			/me/permissions [get]
func (h *PermissionHandler) Me(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	if userID == 0 {
		utils.Unauthorized(c, "未认证")
		return
	}

	// 用例编排：角色 + 权限码 + 可访问平台，组装用户权限视图
	roles, err := h.enforcer.LoadRoles(userID)
	if err != nil {
		utils.ServerError(c, "查询用户权限失败")
		return
	}
	codes, err := h.enforcer.PermissionCodes(userID)
	if err != nil {
		utils.ServerError(c, "查询用户权限失败")
		return
	}
	platforms, err := h.accessService.FindPlatformsByUserID(userID)
	if err != nil {
		utils.ServerError(c, "查询用户权限失败")
		return
	}

	isAdmin := false
	for _, r := range roles {
		if r == shared.RoleAdmin {
			isAdmin = true
			break
		}
	}

	utils.OK(c, response.MePermissionDTO{
		UserID:      userID,
		Roles:       roles,
		IsAdmin:     isAdmin,
		Permissions: codes,
		Platforms:   platforms,
	})
}
