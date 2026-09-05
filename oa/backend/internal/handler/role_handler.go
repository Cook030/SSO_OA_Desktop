package handler

import (
	"strconv"

	"permission-system/internal/api_model/request"
	"permission-system/internal/api_model/response"
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/middleware"
	"permission-system/internal/service/role"
	"permission-system/internal/service/shared"
	"permission-system/internal/utils"

	"github.com/gin-gonic/gin"
)

// RoleHandler 角色管理处理器
//
// 职责：解析请求 -> 编排用例（写用例在事务内组合 service 的原子方法）-> 组装响应。
type RoleHandler struct {
	roleService     *role.Service
	userRoleService *role.UserRoleService
	q               *query.Query // 用例层事务控制
}

// NewRoleHandler 创建角色管理处理器
func NewRoleHandler(roleService *role.Service, userRoleService *role.UserRoleService, q *query.Query) *RoleHandler {
	return &RoleHandler{
		roleService:     roleService,
		userRoleService: userRoleService,
		q:               q,
	}
}

// List	角色列表
//
//	@Summary		获取角色列表
//	@Description	获取全部角色，并附带动用户数与权限数
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Success		200	{object}	utils.Response{data=[]response.RoleListItemDTO}
//	@Failure		401	{object}	utils.Response
//	@Failure		403	{object}	utils.Response
//	@Router			/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.roleService.List()
	if err != nil {
		utils.ServerError(c, "查询角色列表失败")
		return
	}

	roleIDs := make([]int64, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}
	userCountMap, err := h.roleService.CountUsersByRoleIDs(roleIDs)
	if err != nil {
		utils.ServerError(c, "统计角色用户数失败")
		return
	}
	permCountMap, err := h.roleService.CountPermissionsByRoleIDs(roleIDs)
	if err != nil {
		utils.ServerError(c, "统计角色权限数失败")
		return
	}

	list := make([]response.RoleListItemDTO, len(roles))
	for i, r := range roles {
		list[i] = response.RoleListItemDTO{
			ID:              r.ID,
			Code:            r.Code,
			Name:            r.Name,
			Description:     shared.DerefString(r.Description),
			IsBuiltin:       r.IsBuiltin,
			Status:          r.Status,
			UserCount:       userCountMap[r.ID],
			PermissionCount: permCountMap[r.ID],
			CreateTime:      r.CreateTime.Format(shared.TimeLayout),
		}
	}

	utils.OK(c, list)
}

// Create	新增角色
//
//	@Summary		创建角色
//	@Description	创建新的自定义角色（内置角色由初始化脚本预置）
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			request	body		request.CreateRoleRequest	true	"角色信息"
//	@Success		200		{object}	utils.Response{data=response.RoleDTO}
//	@Failure		400		{object}	utils.Response
//	@Failure		401		{object}	utils.Response
//	@Failure		403		{object}	utils.Response
//	@Router			/roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req request.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	operatorID := middleware.CurrentUserID(c)
	requestID := middleware.GetRequestID(c)

	var created *db_model.SysRole
	if err := h.q.Transaction(func(tx *query.Query) error {
		var err error
		created, err = h.roleService.CreateTx(tx, &req, operatorID, requestID)
		return err
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "角色创建成功", response.RoleDTO{
		ID:          created.ID,
		Code:        created.Code,
		Name:        created.Name,
		Description: shared.DerefString(created.Description),
		IsBuiltin:   created.IsBuiltin,
		Status:      created.Status,
		Permissions: []int64{},
	})
}

// Update	编辑角色
//
//	@Summary		更新角色
//	@Description	更新角色名称、描述与状态（内置角色不可停用）
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			id		path		int						true	"角色ID"
//	@Param			request	body		request.UpdateRoleRequest	true	"角色信息"
//	@Success		200		{object}	utils.Response{data=response.RoleDTO}
//	@Failure		400		{object}	utils.Response
//	@Failure		401		{object}	utils.Response
//	@Failure		403		{object}	utils.Response
//	@Router			/roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, "无效的角色ID")
		return
	}

	var req request.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	operatorID := middleware.CurrentUserID(c)
	requestID := middleware.GetRequestID(c)

	var updated *db_model.SysRole
	if err := h.q.Transaction(func(tx *query.Query) error {
		var err error
		updated, err = h.roleService.UpdateTx(tx, id, &req, operatorID, requestID)
		return err
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	permissions, err := h.roleService.GetPermissionIDs(id)
	if err != nil {
		utils.ServerError(c, "查询角色权限失败")
		return
	}

	utils.OKWithMsg(c, "角色更新成功", response.RoleDTO{
		ID:          updated.ID,
		Code:        updated.Code,
		Name:        updated.Name,
		Description: shared.DerefString(updated.Description),
		IsBuiltin:   updated.IsBuiltin,
		Status:      updated.Status,
		Permissions: permissions,
	})
}

// Delete	删除角色
//
//	@Summary		删除角色
//	@Description	删除指定角色（内置角色、仍有用户绑定的角色不可删除）
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			id	path		int	true	"角色ID"
//	@Success		200	{object}	utils.Response
//	@Failure		400	{object}	utils.Response
//	@Failure		401	{object}	utils.Response
//	@Failure		403	{object}	utils.Response
//	@Router			/roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, "无效的角色ID")
		return
	}

	if err := h.q.Transaction(func(tx *query.Query) error {
		return h.roleService.DeleteTx(tx, id)
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "角色删除成功", nil)
}

// GetPermissions	查询角色权限
//
//	@Summary		查询角色持有的权限
//	@Description	返回该角色已授权的权限点ID列表
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			id	path		int	true	"角色ID"
//	@Success		200	{object}	utils.Response{data=[]int64}
//	@Failure		401	{object}	utils.Response
//	@Failure		403	{object}	utils.Response
//	@Router			/roles/{id}/permissions [get]
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, "无效的角色ID")
		return
	}

	permissionIDs, err := h.roleService.GetPermissionIDs(id)
	if err != nil {
		utils.ServerError(c, "查询角色权限失败")
		return
	}
	if permissionIDs == nil {
		permissionIDs = []int64{}
	}

	utils.OK(c, permissionIDs)
}

// AssignPermissions	配置角色权限
//
//	@Summary		给角色配置权限
//	@Description	全量覆盖角色的权限点授权
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			id		path		int							true	"角色ID"
//	@Param			request	body		request.AssignPermissionsRequest	true	"权限ID列表"
//	@Success		200		{object}	utils.Response
//	@Failure		400		{object}	utils.Response
//	@Failure		401		{object}	utils.Response
//	@Failure		403		{object}	utils.Response
//	@Router			/roles/{id}/permissions [put]
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, "无效的角色ID")
		return
	}

	var req request.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	operatorID := middleware.CurrentUserID(c)
	requestID := middleware.GetRequestID(c)

	if err := h.q.Transaction(func(tx *query.Query) error {
		return h.roleService.AssignPermissionsTx(tx, id, req.PermissionIDs, operatorID, requestID)
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "角色权限配置成功", nil)
}

// AssignUsers	批量分配角色
//
//	@Summary		批量给用户分配角色
//	@Description	为多个用户增量分配指定角色（已存在的不重复插入）
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			request	body		request.AssignUsersRequest				true	"用户与角色ID"
//	@Success		200		{object}	utils.Response{data=response.BatchAssignResultDTO}
//	@Failure		400		{object}	utils.Response
//	@Failure		401		{object}	utils.Response
//	@Failure		403		{object}	utils.Response
//	@Router			/roles/users [post]
func (h *RoleHandler) AssignUsers(c *gin.Context) {
	var req request.AssignUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	operatorID := middleware.CurrentUserID(c)
	requestID := middleware.GetRequestID(c)

	var affected int64
	if err := h.q.Transaction(func(tx *query.Query) error {
		var err error
		affected, err = h.userRoleService.AssignUsersTx(tx, req.UserIDs, req.RoleIDs, operatorID, requestID)
		return err
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "角色分配成功", response.BatchAssignResultDTO{AffectedCount: affected})
}

// GetUserRoles	查询用户角色
//
//	@Summary		查询用户持有的角色
//	@Description	返回指定用户当前的角色列表
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			id	path		int	true	"用户ID"
//	@Success		200	{object}	utils.Response{data=[]response.RoleOptionDTO}
//	@Failure		401	{object}	utils.Response
//	@Failure		403	{object}	utils.Response
//	@Router			/users/{id}/roles [get]
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, "无效的用户ID")
		return
	}

	roles, err := h.roleService.GetUserRoles(id)
	if err != nil {
		utils.ServerError(c, "查询用户角色失败")
		return
	}

	list := make([]response.RoleOptionDTO, len(roles))
	for i, r := range roles {
		list[i] = response.RoleOptionDTO{
			ID:        r.ID,
			Code:      r.Code,
			Name:      r.Name,
			IsBuiltin: r.IsBuiltin,
		}
	}

	utils.OK(c, list)
}

// SetUserRoles	设置用户角色
//
//	@Summary		设置用户的角色
//	@Description	全量覆盖指定用户的角色
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			id		path		int						true	"用户ID"
//	@Param			request	body		request.SetUserRolesRequest	true	"角色ID列表"
//	@Success		200		{object}	utils.Response
//	@Failure		400		{object}	utils.Response
//	@Failure		401		{object}	utils.Response
//	@Failure		403		{object}	utils.Response
//	@Router			/users/{id}/roles [put]
func (h *RoleHandler) SetUserRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, "无效的用户ID")
		return
	}

	var req request.SetUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	operatorID := middleware.CurrentUserID(c)
	requestID := middleware.GetRequestID(c)

	if err := h.q.Transaction(func(tx *query.Query) error {
		return h.userRoleService.SetUserRolesTx(tx, id, req.RoleIDs, operatorID, requestID)
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "用户角色设置成功", nil)
}
