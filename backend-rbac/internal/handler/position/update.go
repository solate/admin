package position

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

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
func (h *Handler) UpdatePosition(c *gin.Context) {
	var req dto.UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.UpdatePosition(c.Request.Context(), req.PositionID, &req)
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
func (h *Handler) UpdatePositionStatus(c *gin.Context) {
	var req dto.PositionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.UpdatePositionStatus(c.Request.Context(), req.PositionID, req.Status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}
