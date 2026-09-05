package handler

import (
	"strconv"

	"permission-system/internal/api_model/request"
	"permission-system/internal/api_model/response"
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/middleware"
	"permission-system/internal/service/employee"
	"permission-system/internal/service/permission"
	"permission-system/internal/service/role"
	"permission-system/internal/service/shared"
	"permission-system/internal/utils"

	"github.com/gin-gonic/gin"
)

// EmployeeHandler 员工管理处理器
//
// 职责：解析请求 -> 编排用例（写用例在事务内组合员工与角色两个领域的原子方法）-> 组装响应。
type EmployeeHandler struct {
	employeeService *employee.Service
	accessService   *permission.AccessService
	userRoleService *role.UserRoleService
	q               *query.Query // 用例层事务控制
}

// NewEmployeeHandler 创建员工管理处理器
func NewEmployeeHandler(
	employeeService *employee.Service,
	accessService *permission.AccessService,
	userRoleService *role.UserRoleService,
	q *query.Query,
) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: employeeService,
		accessService:   accessService,
		userRoleService: userRoleService,
		q:               q,
	}
}

// List	员工列表
//
//	@Summary		获取员工列表
//	@Description	分页获取员工列表，支持搜索和筛选（默认不包含管理员账号）
//	@Tags			员工管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			page		query		int		false	"页码，默认1"
//	@Param			pageSize	query		int		false	"每页数量，默认20"
//	@Param			keyword		query		string	false	"搜索关键词"
//	@Param			department	query		string	false	"部门筛选"
//	@Param			platformId	query		int		false	"平台ID筛选"
//	@Success		200			{object}	utils.Response{data=response.EmployeePageResult}
//	@Failure		401			{object}	utils.Response
//	@Failure		403			{object}	utils.Response
//	@Router			/employees [get]
func (h *EmployeeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	platformID, _ := strconv.ParseInt(c.Query("platformId"), 10, 64)

	// 用例编排：平台筛选 -> 员工列表 -> 角色/平台权限批量组装
	var userIDs []int64
	if platformID > 0 {
		var err error
		userIDs, err = h.accessService.FindUserIDsByPlatformID(platformID)
		if err != nil {
			utils.ServerError(c, "查询平台权限失败")
			return
		}
		if len(userIDs) == 0 {
			utils.OK(c, response.EmployeePageResult{Total: 0, List: []response.EmployeeListItemDTO{}})
			return
		}
	}

	param := employee.EmployeeListParam{
		Keyword:    c.Query("keyword"),
		Department: c.Query("department"),
		UserIDs:    userIDs,
		Page:       page,
		PageSize:   pageSize,
	}

	users, total, err := h.employeeService.FindEmployees(param)
	if err != nil {
		utils.ServerError(c, "查询员工列表失败")
		return
	}

	userIDs = make([]int64, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}
	permMap, err := h.accessService.FindPlatformsByUserIDs(userIDs)
	if err != nil {
		utils.ServerError(c, "查询员工权限失败")
		return
	}
	roleMap, err := h.userRoleService.FindRolesByUserIDs(userIDs)
	if err != nil {
		utils.ServerError(c, "查询员工角色失败")
		return
	}

	// 组装响应
	list := make([]response.EmployeeListItemDTO, len(users))
	for i, u := range users {
		permissions := make([]response.PlatformPermission, 0)
		if perms, ok := permMap[u.ID]; ok {
			permissions = perms
		}
		roles := make([]response.RoleOptionDTO, 0)
		for _, r := range roleMap[u.ID] {
			roles = append(roles, response.RoleOptionDTO{
				ID:        r.ID,
				Code:      r.Code,
				Name:      r.Name,
				IsBuiltin: r.IsBuiltin,
			})
		}

		list[i] = response.EmployeeListItemDTO{
			ID:                  u.ID,
			DisplayID:           shared.FormatEmployeeDisplayID(u.ID),
			Name:                u.Name,
			Account:             u.Account,
			Phone:               shared.DerefString(u.Phone),
			Email:               shared.DerefString(u.Email),
			Department:          shared.DerefString(u.Department),
			Roles:               roles,
			PlatformPermissions: permissions,
		}
	}

	utils.OK(c, response.EmployeePageResult{
		Total: total,
		List:  list,
	})
}

// GetDepartments	获取部门列表
//
//	@Summary		获取部门列表
//	@Description	获取所有部门名称列表
//	@Tags			员工管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Success		200	{object}	utils.Response{data=[]string}
//	@Failure		401	{object}	utils.Response
//	@Failure		403	{object}	utils.Response
//	@Router			/employees/departments [get]
func (h *EmployeeHandler) GetDepartments(c *gin.Context) {
	departments, err := h.employeeService.GetDepartments()
	if err != nil {
		utils.ServerError(c, "查询部门列表失败")
		return
	}

	utils.OK(c, departments)
}

// Create	新增员工
//
//	@Summary		创建员工
//	@Description	创建新的员工账号并分配角色（未指定角色时默认赋予普通员工角色）
//	@Tags			员工管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			request	body		request.CreateEmployeeRequest	true	"员工信息"
//	@Success		200		{object}	utils.Response{data=response.CreateEmployeeDTO}
//	@Failure		400		{object}	utils.Response
//	@Failure		401		{object}	utils.Response
//	@Failure		403		{object}	utils.Response
//	@Router			/employees [post]
func (h *EmployeeHandler) Create(c *gin.Context) {
	var req request.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	operatorID := middleware.CurrentUserID(c)
	requestID := middleware.GetRequestID(c)

	// 跨域用例编排：创建员工用户 + 设置角色，同一事务保证原子性
	var user *db_model.SysUser
	if err := h.q.Transaction(func(tx *query.Query) error {
		var err error
		user, err = h.employeeService.CreateUserTx(tx, &req, operatorID, requestID)
		if err != nil {
			return err
		}
		if len(req.RoleIDs) > 0 {
			return h.userRoleService.SetUserRolesTx(tx, user.ID, req.RoleIDs, operatorID, requestID)
		}
		return h.userRoleService.EnsureEmployeeRoleTx(tx, user.ID, operatorID, requestID)
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "员工添加成功", response.CreateEmployeeDTO{
		ID:        user.ID,
		DisplayID: shared.FormatEmployeeDisplayID(user.ID),
		Name:      user.Name,
		Account:   user.Account,
		Email:     shared.DerefString(user.Email),
	})
}

// Update	编辑员工
//
//	@Summary		更新员工
//	@Description	更新员工信息与角色归属
//	@Tags			员工管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			id		path		int							true	"员工ID"
//	@Param			request	body		request.UpdateEmployeeRequest	true	"员工信息"
//	@Success		200		{object}	utils.Response{data=response.UpdateEmployeeDTO}
//	@Failure		400		{object}	utils.Response
//	@Failure		401		{object}	utils.Response
//	@Failure		403		{object}	utils.Response
//	@Router			/employees/{id} [put]
func (h *EmployeeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, "无效的员工ID")
		return
	}

	var req request.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	operatorID := middleware.CurrentUserID(c)
	requestID := middleware.GetRequestID(c)

	// 跨域用例编排：更新员工信息 + 全量覆盖角色，同一事务保证原子性
	var user *db_model.SysUser
	if err := h.q.Transaction(func(tx *query.Query) error {
		var err error
		user, err = h.employeeService.UpdateUserTx(tx, id, &req, operatorID, requestID)
		if err != nil {
			return err
		}
		if len(req.RoleIDs) > 0 {
			return h.userRoleService.SetUserRolesTx(tx, id, req.RoleIDs, operatorID, requestID)
		}
		return h.userRoleService.EnsureEmployeeRoleTx(tx, id, operatorID, requestID)
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "员工信息更新成功", response.UpdateEmployeeDTO{
		ID:         user.ID,
		DisplayID:  shared.FormatEmployeeDisplayID(user.ID),
		Name:       user.Name,
		Account:    user.Account,
		Phone:      shared.DerefString(user.Phone),
		Email:      shared.DerefString(user.Email),
		Department: shared.DerefString(user.Department),
	})
}

// Delete	删除员工
//
//	@Summary		删除员工
//	@Description	删除指定员工及其角色绑定
//	@Tags			员工管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			id	path		int	true	"员工ID"
//	@Success		200	{object}	utils.Response
//	@Failure		400	{object}	utils.Response
//	@Failure		401	{object}	utils.Response
//	@Failure		403	{object}	utils.Response
//	@Router			/employees/{id} [delete]
func (h *EmployeeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, "无效的员工ID")
		return
	}

	// 跨域用例编排：解除角色绑定 + 删除员工，同一事务保证原子性
	if err := h.q.Transaction(func(tx *query.Query) error {
		if err := h.userRoleService.RevokeUserRolesTx(tx, id); err != nil {
			return err
		}
		return h.employeeService.DeleteUserTx(tx, id)
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "员工删除成功", nil)
}

// ResetPassword	重置密码
//
//	@Summary		重置员工密码
//	@Description	将指定员工的密码重置为默认密码，密码版本递增使旧 token 失效，员工下次登录需修改密码
//	@Tags			员工管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			id	path		int	true	"员工ID"
//	@Success		200	{object}	utils.Response
//	@Failure		400	{object}	utils.Response
//	@Failure		401	{object}	utils.Response
//	@Failure		403	{object}	utils.Response
//	@Router			/employees/{id}/reset-password [put]
func (h *EmployeeHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, "无效的员工ID")
		return
	}

	if err := h.employeeService.ResetPassword(id); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "密码重置成功，员工下次登录需修改密码", nil)
}
