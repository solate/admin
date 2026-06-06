package position

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

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
func (h *Handler) DeletePosition(c *gin.Context) {
	var req dto.PositionDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.DeletePosition(c.Request.Context(), req.PositionID); err != nil {
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
func (h *Handler) BatchDeletePositions(c *gin.Context) {
	var req dto.PositionBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.BatchDeletePositions(c.Request.Context(), req.PositionIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.PositionIDs)})
}
