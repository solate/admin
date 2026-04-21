package dict

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// BatchUpdateDictItems 批量更新字典项（租户覆盖）
// @Summary 批量更新字典项
// @Description 批量更新字典项（租户覆盖系统默认值）
// @Tags 字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.BatchUpdateDictItemsRequest true "批量更新字典项请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Router /api/v1/dicts/items [put]
func (h *Handler) BatchUpdateDictItems(c *gin.Context) {
	var req dto.BatchUpdateDictItemsRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.BatchUpdateDictItems(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}

// ResetDictItem 恢复字典项系统默认值
// @Summary 恢复字典项系统默认值
// @Description 删除租户覆盖记录，恢复使用系统默认值
// @Tags 字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.DictItemKeyRequest true "字典项键请求参数"
// @Success 200 {object} response.Response "恢复成功"
// @Router /api/v1/dicts/items/reset [delete]
func (h *Handler) ResetDictItem(c *gin.Context) {
	var req dto.DictItemKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.ResetDictItem(c.Request.Context(), req.TypeCode, req.Value); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"reset": true})
}
