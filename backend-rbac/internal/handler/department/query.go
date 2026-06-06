package department

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetDepartment 获取部门详情
// @Summary 获取部门详情
// @Description 根据ID获取部门详情
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param department_id query string true "部门ID"
// @Success 200 {object} response.Response{data=dto.DepartmentInfo} "获取成功"
// @Router /api/v1/departments/detail [get]
func (h *Handler) GetDepartment(c *gin.Context) {
	var req dto.DepartmentDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.GetDepartmentByID(c.Request.Context(), req.DepartmentID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// ListDepartments 获取部门列表
// @Summary 获取部门列表
// @Description 分页获取部门列表，支持按关键词和状态筛选
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param department_name query string false "部门名称(模糊匹配)"
// @Param status query int false "状态筛选(1:启用,2:禁用)" Enums(1, 2)
// @Param parent_id query string false "父部门ID"
// @Success 200 {object} response.Response{data=dto.ListDepartmentsResponse} "获取成功"
// @Router /api/v1/departments [get]
func (h *Handler) ListDepartments(c *gin.Context) {
	var req dto.ListDepartmentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.ListDepartments(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetDepartmentTree 获取部门树
// @Summary 获取部门树
// @Description 获取完整的部门树形结构
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.DepartmentTreeResponse} "获取成功"
// @Router /api/v1/departments/tree [get]
func (h *Handler) GetDepartmentTree(c *gin.Context) {
	resp, err := h.svc.GetDepartmentTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetChildren 获取子部门
// @Summary 获取子部门
// @Description 获取指定部门的直接子部门列表
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param department_id query string true "部门ID"
// @Success 200 {object} response.Response{data=[]dto.DepartmentInfo} "获取成功"
// @Router /api/v1/departments/children [get]
func (h *Handler) GetChildren(c *gin.Context) {
	var req dto.DepartmentDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.GetChildren(c.Request.Context(), req.DepartmentID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
