package handler

import (
	"permission-system/internal/api_model/request"
	"permission-system/internal/db_model/query"
	"permission-system/internal/service"
	"permission-system/internal/utils"

	"github.com/gin-gonic/gin"
)

// PermissionHandler 权限管理处理器
type PermissionHandler struct {
	permissionService *service.PermissionService
	q                 *query.Query
}

// NewPermissionHandler 创建权限管理处理器
func NewPermissionHandler(permissionService *service.PermissionService, q *query.Query) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
		q:                 q,
	}
}

// BatchSet	批量设置权限(增量)
//
//	@Summary		批量设置平台权限
//	@Description	为多个用户批量分配平台权限（增量模式）
//	@Tags			权限管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			request	body		request.BatchSetRequest						true	"权限设置请求"
//	@Success		200		{object}	utils.Response{data=response.BatchPermissionResultDTO}
//	@Failure		400		{object}	utils.Response
//	@Failure		401		{object}	utils.Response
//	@Router			/employees/permissions/batch [post]
func (h *PermissionHandler) BatchSet(c *gin.Context) {
	var req request.BatchSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	result, err := h.permissionService.BatchSet(h.q, req.UserIDs, req.PlatformIDs)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "权限设置成功", result)
}

// BatchDelete	批量删除权限
//
//	@Summary		批量删除平台权限
//	@Description	为多个用户批量删除指定的平台权限
//	@Tags			权限管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			request	body		request.BatchDeleteRequest						true	"权限删除请求"
//	@Success		200		{object}	utils.Response{data=response.BatchPermissionResultDTO}
//	@Failure		400		{object}	utils.Response
//	@Failure		401		{object}	utils.Response
//	@Router			/employees/permissions/batch [delete]
func (h *PermissionHandler) BatchDelete(c *gin.Context) {
	var req request.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	result, err := h.permissionService.BatchDelete(h.q, req.UserIDs, req.PlatformIDs)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "权限已删除", result)
}
