package position

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetPosition 获取岗位详情
// @Summary 获取岗位详情
// @Description 根据ID获取岗位详情
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param position_id query string true "岗位ID"
// @Success 200 {object} response.Response{data=dto.PositionInfo} "获取成功"
// @Router /api/v1/positions/detail [get]
func (h *Handler) GetPosition(c *gin.Context) {
	var req dto.PositionDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.GetPositionByID(c.Request.Context(), req.PositionID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// ListPositions 获取岗位列表
// @Summary 获取岗位列表
// @Description 分页获取岗位列表，支持按关键词和状态筛选
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param position_name query string false "岗位名称(模糊匹配)"
// @Param position_code query string false "岗位编码(模糊匹配)"
// @Param status query int false "状态筛选(1:启用,2:禁用)" Enums(1, 2)
// @Success 200 {object} response.Response{data=dto.ListPositionsResponse} "获取成功"
// @Router /api/v1/positions [get]
func (h *Handler) ListPositions(c *gin.Context) {
	var req dto.ListPositionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.ListPositions(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// ListAllPositions 获取所有岗位
// @Summary 获取所有岗位
// @Description 获取所有岗位列表（不分页，用于下拉选择）
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]dto.PositionInfo} "获取成功"
// @Router /api/v1/positions/all [get]
func (h *Handler) ListAllPositions(c *gin.Context) {
	resp, err := h.svc.ListAllPositions(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
