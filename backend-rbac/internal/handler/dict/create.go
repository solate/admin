package dict

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// CreateSystemDict 创建系统字典（超管专用）
// @Summary 创建系统字典
// @Description 创建系统字典模板（超管专用）
// @Tags 字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateSystemDictRequest true "创建系统字典请求参数"
// @Success 200 {object} response.Response "创建成功"
// @Router /api/v1/system/dicts [post]
func (h *Handler) CreateSystemDict(c *gin.Context) {
	var req dto.CreateSystemDictRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.CreateSystemDict(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"created": true})
}
