package dict

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// DeleteSystemDict 删除系统字典（超管专用）
// @Summary 删除系统字典
// @Description 删除系统字典模板（超管专用）
// @Tags 字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.DictTypeDeleteRequest true "删除系统字典请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/system/dicts [delete]
func (h *Handler) DeleteSystemDict(c *gin.Context) {
	var req dto.DictTypeDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.DeleteSystemDict(c.Request.Context(), req.TypeID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// BatchDeleteSystemDicts 批量删除系统字典（超管专用）
// @Summary 批量删除系统字典
// @Description 批量删除系统字典模板（超管专用）
// @Tags 字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.DictTypeBatchDeleteRequest true "批量删除请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/system/dicts/batch-delete [delete]
func (h *Handler) BatchDeleteSystemDicts(c *gin.Context) {
	var req dto.DictTypeBatchDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.BatchDeleteSystemDicts(c.Request.Context(), req.TypeIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.TypeIDs)})
}

// DeleteSystemDictItem 删除系统字典项（超管专用）
// @Summary 删除系统字典项
// @Description 删除系统字典项（超管专用），真正的删除操作
// @Tags 字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.DictItemKeyRequest true "字典项键请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/system/dicts/items [delete]
func (h *Handler) DeleteSystemDictItem(c *gin.Context) {
	var req dto.DictItemKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.DeleteSystemDictItem(c.Request.Context(), req.TypeCode, req.Value); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}
