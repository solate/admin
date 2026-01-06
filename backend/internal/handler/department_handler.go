package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"
	"admin/pkg/xerr"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DepartmentHandler 部门处理器
type DepartmentHandler struct {
	deptService *service.DepartmentService
}

// NewDepartmentHandler 创建部门处理器
func NewDepartmentHandler(deptService *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{
		deptService: deptService,
	}
}

// CreateDepartment 创建部门
// @Summary 创建部门
// @Description 创建新的部门
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateDepartmentRequest true "创建部门请求参数"
// @Success 200 {object} response.Response{data=dto.DepartmentResponse} "创建成功"
// @Router /api/v1/departments [post]
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req dto.CreateDepartmentRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.deptService.CreateDepartment(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetDepartment 获取部门详情
// @Summary 获取部门详情
// @Description 根据ID获取部门详情
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param department_id path string true "部门ID"
// @Success 200 {object} response.Response{data=dto.DepartmentResponse} "获取成功"
// @Router /api/v1/departments/{department_id} [get]
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	departmentID := c.Param("department_id")
	if departmentID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	resp, err := h.deptService.GetDepartmentByID(c.Request.Context(), departmentID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateDepartment 更新部门
// @Summary 更新部门
// @Description 更新部门信息
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param department_id path string true "部门ID"
// @Param request body dto.UpdateDepartmentRequest true "更新部门请求参数"
// @Success 200 {object} response.Response{data=dto.DepartmentResponse} "更新成功"
// @Router /api/v1/departments/{department_id} [put]
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	departmentID := c.Param("department_id")
	if departmentID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.deptService.UpdateDepartment(c.Request.Context(), departmentID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteDepartment 删除部门
// @Summary 删除部门
// @Description 删除部门（软删除）
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param department_id path string true "部门ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/departments/{department_id} [delete]
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	departmentID := c.Param("department_id")
	if departmentID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	if err := h.deptService.DeleteDepartment(c.Request.Context(), departmentID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
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
// @Param keyword query string false "关键词搜索（部门名称）"
// @Param status query int false "状态筛选(1:启用,2:禁用)"
// @Param parent_id query string false "父部门ID"
// @Success 200 {object} response.Response{data=dto.ListDepartmentsResponse} "获取成功"
// @Router /api/v1/departments [get]
func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
	var req dto.ListDepartmentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.deptService.ListDepartments(c.Request.Context(), &req)
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
func (h *DepartmentHandler) GetDepartmentTree(c *gin.Context) {
	resp, err := h.deptService.GetDepartmentTree(c.Request.Context())
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
// @Param department_id path string true "部门ID"
// @Success 200 {object} response.Response{data=[]dto.DepartmentResponse} "获取成功"
// @Router /api/v1/departments/{department_id}/children [get]
func (h *DepartmentHandler) GetChildren(c *gin.Context) {
	departmentID := c.Param("department_id")
	if departmentID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	resp, err := h.deptService.GetChildren(c.Request.Context(), departmentID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateDepartmentStatus 更新部门状态
// @Summary 更新部门状态
// @Description 更新部门状态
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param department_id path string true "部门ID"
// @Param status path int true "状态(1:启用,2:禁用)"
// @Success 200 {object} response.Response "更新成功"
// @Router /api/v1/departments/{department_id}/status/{status} [put]
func (h *DepartmentHandler) UpdateDepartmentStatus(c *gin.Context) {
	departmentID := c.Param("department_id")
	if departmentID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	statusStr := c.Param("status")
	if statusStr == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	if err := h.deptService.UpdateDepartmentStatus(c.Request.Context(), departmentID, status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}
