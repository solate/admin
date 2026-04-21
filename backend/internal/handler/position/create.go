package position

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

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
func (h *Handler) CreatePosition(c *gin.Context) {
	var req dto.CreatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.CreatePosition(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
