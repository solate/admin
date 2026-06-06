package dict

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

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
func (h *Handler) ListSystemDictTypes(c *gin.Context) {
	var req dto.ListDictTypesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.ListDictTypes(c.Request.Context(), &req)
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
func (h *Handler) GetDict(c *gin.Context) {
	var req dto.DictByCodeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.GetDictByCode(c.Request.Context(), req.TypeCode)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
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
func (h *Handler) ListDictTypes(c *gin.Context) {
	var req dto.ListDictTypesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.ListDictTypes(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
