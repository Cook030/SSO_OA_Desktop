package handler

import (
	"strconv"

	"permission-system/internal/api_model/request"
	"permission-system/internal/api_model/response"
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
	"permission-system/internal/middleware"
	"permission-system/internal/service/permission"
	"permission-system/internal/service/platform"
	"permission-system/internal/utils"

	"github.com/gin-gonic/gin"
)

// PlatformHandler 平台管理处理器
//
// 职责：解析请求 -> 编排用例（写用例在事务内调用 service 原子方法）-> 组装响应。
type PlatformHandler struct {
	platformService *platform.Service
	accessService   *permission.AccessService
	q               *query.Query // 用例层事务控制
}

// NewPlatformHandler 创建平台管理处理器
func NewPlatformHandler(platformService *platform.Service, accessService *permission.AccessService, q *query.Query) *PlatformHandler {
	return &PlatformHandler{
		platformService: platformService,
		accessService:   accessService,
		q:               q,
	}
}

// List	平台列表
//
//	@Summary		获取平台列表
//	@Description	分页获取平台列表
//	@Tags			平台管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			page		query		int	false	"页码，默认1"
//	@Param			pageSize	query		int	false	"每页数量，默认20"
//	@Success		200			{object}	utils.Response{data=response.PlatformPageResult}
//	@Failure		401			{object}	utils.Response
//	@Router			/platforms [get]
func (h *PlatformHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 用例编排：平台分页 + 批量授权人数统计，组装响应
	platforms, total, err := h.platformService.FindPlatforms(page, pageSize)
	if err != nil {
		utils.ServerError(c, "查询平台列表失败")
		return
	}

	platformIDs := make([]int64, len(platforms))
	for i, p := range platforms {
		platformIDs[i] = p.ID
	}
	countMap, err := h.accessService.CountByPlatformIDs(platformIDs)
	if err != nil {
		utils.ServerError(c, "查询平台授权人数失败")
		return
	}

	list := make([]response.PlatformListItemDTO, len(platforms))
	for i, p := range platforms {
		list[i] = response.PlatformListItemDTO{
			ID:              p.ID,
			Name:            p.Name,
			Code:            p.Code,
			PermissionCount: countMap[p.ID],
			CreateTime:      p.CreateTime,
		}
	}

	utils.OK(c, response.PlatformPageResult{
		Total: total,
		List:  list,
	})
}

// Create	新增平台
//
//	@Summary		创建平台
//	@Description	创建新的平台
//	@Tags			平台管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			request	body		request.PlatformRequest	true	"平台信息"
//	@Success		200		{object}	utils.Response{data=response.PlatformDTO}
//	@Failure		400		{object}	utils.Response
//	@Router			/platforms [post]
func (h *PlatformHandler) Create(c *gin.Context) {
	var req request.PlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	operatorID := middleware.CurrentUserID(c)
	requestID := middleware.GetRequestID(c)

	var created *db_model.SysPlatform
	if err := h.q.Transaction(func(tx *query.Query) error {
		var err error
		created, err = h.platformService.CreateTx(tx, req.Name, req.Code, operatorID, requestID)
		return err
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "平台创建成功", response.PlatformDTO{
		ID:   created.ID,
		Name: created.Name,
		Code: created.Code,
	})
}

// Update	编辑平台
//
//	@Summary		更新平台
//	@Description	更新平台信息
//	@Tags			平台管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			id		path		int							true	"平台ID"
//	@Param			request	body		request.PlatformRequest		true	"平台信息"
//	@Success		200		{object}	utils.Response{data=response.PlatformDTO}
//	@Failure		400		{object}	utils.Response
//	@Router			/platforms/{id} [put]
func (h *PlatformHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, "无效的平台ID")
		return
	}

	var req request.PlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数解析失败")
		return
	}

	operatorID := middleware.CurrentUserID(c)
	requestID := middleware.GetRequestID(c)

	var updated *db_model.SysPlatform
	if err := h.q.Transaction(func(tx *query.Query) error {
		var err error
		updated, err = h.platformService.UpdateTx(tx, id, req.Name, req.Code, operatorID, requestID)
		return err
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "平台更新成功", response.PlatformDTO{
		ID:   updated.ID,
		Name: updated.Name,
		Code: updated.Code,
	})
}

// Delete	删除平台
//
//	@Summary		删除平台
//	@Description	删除指定平台
//	@Tags			平台管理
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			id	path		int	true	"平台ID"
//	@Success		200	{object}	utils.Response
//	@Failure		400	{object}	utils.Response
//	@Router			/platforms/{id} [delete]
func (h *PlatformHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, "无效的平台ID")
		return
	}

	if err := h.q.Transaction(func(tx *query.Query) error {
		return h.platformService.DeleteTx(tx, id)
	}); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.OKWithMsg(c, "平台删除成功", nil)
}
