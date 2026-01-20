package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// DictHandler 字典处理器
type DictHandler struct {
	dictService *service.DictService
}

// NewDictHandler 创建字典处理器
func NewDictHandler(dictService *service.DictService) *DictHandler {
	return &DictHandler{
		dictService: dictService,
	}
}

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
func (h *DictHandler) CreateSystemDict(c *gin.Context) {
	var req dto.CreateSystemDictRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.dictService.CreateSystemDict(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"created": true})
}

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
func (h *DictHandler) UpdateSystemDict(c *gin.Context) {
	var req dto.UpdateSystemDictRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.dictService.UpdateSystemDict(c.Request.Context(), req.TypeID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}

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
func (h *DictHandler) DeleteSystemDict(c *gin.Context) {
	var req dto.DictTypeDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.dictService.DeleteSystemDict(c.Request.Context(), req.TypeID); err != nil {
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
func (h *DictHandler) BatchDeleteSystemDicts(c *gin.Context) {
	var req dto.DictTypeBatchDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.dictService.BatchDeleteSystemDicts(c.Request.Context(), req.TypeIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.TypeIDs)})
}

// ListSystemDictTypes 获取系统字典类型列表（超管专用）
// @Summary 获取系统字典类型列表
// @Description 分页获取系统字典类型列表（超管专用）
// @Tags 字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param type_name query string false "字典名称(模糊匹配)"
// @Param type_code query string false "字典编码(模糊匹配)"
// @Success 200 {object} response.Response{data=dto.ListDictTypesResponse} "获取成功"
// @Router /api/v1/system/dicts [get]
func (h *DictHandler) ListSystemDictTypes(c *gin.Context) {
	var req dto.ListDictTypesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.dictService.ListDictTypes(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetDict 获取字典（合并系统+覆盖）
// @Summary 获取字典
// @Description 根据字典编码获取字典（自动合并系统默认值和租户覆盖值）
// @Tags 字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param type_code query string true "字典编码"
// @Success 200 {object} response.Response{data=dto.DictInfo} "获取成功"
// @Router /api/v1/dicts [get]
func (h *DictHandler) GetDict(c *gin.Context) {
	var req dto.DictByCodeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.dictService.GetDictByCode(c.Request.Context(), req.TypeCode)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

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
func (h *DictHandler) BatchUpdateDictItems(c *gin.Context) {
	var req dto.BatchUpdateDictItemsRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.dictService.BatchUpdateDictItems(c.Request.Context(), &req); err != nil {
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
func (h *DictHandler) ResetDictItem(c *gin.Context) {
	var req dto.DictItemKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.dictService.ResetDictItem(c.Request.Context(), req.TypeCode, req.Value); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"reset": true})
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
func (h *DictHandler) DeleteSystemDictItem(c *gin.Context) {
	var req dto.DictItemKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.dictService.DeleteSystemDictItem(c.Request.Context(), req.TypeCode, req.Value); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// ListDictTypes 获取字典类型列表（所有用户）
// @Summary 获取字典类型列表
// @Description 分页获取字典类型列表
// @Tags 字典管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param type_name query string false "字典名称(模糊匹配)"
// @Param type_code query string false "字典编码(模糊匹配)"
// @Success 200 {object} response.Response{data=dto.ListDictTypesResponse} "获取成功"
// @Router /api/v1/dicts/types [get]
func (h *DictHandler) ListDictTypes(c *gin.Context) {
	var req dto.ListDictTypesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.dictService.ListDictTypes(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
