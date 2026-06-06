package dict

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// UpdateSystemDict 更新系统字典（超管专用）
// @Summary 更新系统字典
// @Description 更新系统字典模板（超管专用）
// @Tags 字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UpdateSystemDictRequest true "更新系统字典请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Router /api/v1/system/dicts [put]
func (h *Handler) UpdateSystemDict(c *gin.Context) {
	var req dto.UpdateSystemDictRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.UpdateSystemDict(c.Request.Context(), req.TypeID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}
