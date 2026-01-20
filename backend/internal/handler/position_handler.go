package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// PositionHandler 岗位处理器
type PositionHandler struct {
	positionService *service.PositionService
}

// NewPositionHandler 创建岗位处理器
func NewPositionHandler(positionService *service.PositionService) *PositionHandler {
	return &PositionHandler{
		positionService: positionService,
	}
}

// CreatePosition 创建岗位
// @Summary 创建岗位
// @Description 创建新的岗位
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreatePositionRequest true "创建岗位请求参数"
// @Success 200 {object} response.Response{data=dto.PositionInfo} "创建成功"
// @Router /api/v1/positions [post]
func (h *PositionHandler) CreatePosition(c *gin.Context) {
	var req dto.CreatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.positionService.CreatePosition(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

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
func (h *PositionHandler) GetPosition(c *gin.Context) {
	var req dto.PositionDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.positionService.GetPositionByID(c.Request.Context(), req.PositionID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdatePosition 更新岗位
// @Summary 更新岗位
// @Description 更新岗位信息
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UpdatePositionRequest true "更新岗位请求参数"
// @Success 200 {object} response.Response{data=dto.PositionInfo} "更新成功"
// @Router /api/v1/positions [put]
func (h *PositionHandler) UpdatePosition(c *gin.Context) {
	var req dto.UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.positionService.UpdatePosition(c.Request.Context(), req.PositionID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// DeletePosition 删除岗位
// @Summary 删除岗位
// @Description 删除岗位（软删除）
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.PositionDeleteRequest true "删除岗位请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/positions [delete]
func (h *PositionHandler) DeletePosition(c *gin.Context) {
	var req dto.PositionDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.positionService.DeletePosition(c.Request.Context(), req.PositionID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// BatchDeletePositions 批量删除岗位
// @Summary 批量删除岗位
// @Description 批量软删除岗位
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.PositionBatchDeleteRequest true "批量删除请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/positions/batch-delete [delete]
func (h *PositionHandler) BatchDeletePositions(c *gin.Context) {
	var req dto.PositionBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.positionService.BatchDeletePositions(c.Request.Context(), req.PositionIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.PositionIDs)})
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
// @Param status query int false "状态筛选(1:启用,2:禁用)" Enums(1,2)
// @Success 200 {object} response.Response{data=dto.ListPositionsResponse} "获取成功"
// @Router /api/v1/positions [get]
func (h *PositionHandler) ListPositions(c *gin.Context) {
	var req dto.ListPositionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.positionService.ListPositions(c.Request.Context(), &req)
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
func (h *PositionHandler) ListAllPositions(c *gin.Context) {
	resp, err := h.positionService.ListAllPositions(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdatePositionStatus 更新岗位状态
// @Summary 更新岗位状态
// @Description 更新岗位状态
// @Tags 岗位管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.PositionStatusRequest true "更新状态请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Router /api/v1/positions/status [put]
func (h *PositionHandler) UpdatePositionStatus(c *gin.Context) {
	var req dto.PositionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.positionService.UpdatePositionStatus(c.Request.Context(), req.PositionID, req.Status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}
