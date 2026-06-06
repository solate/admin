package department

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// UpdateDepartment 更新部门
// @Summary 更新部门
// @Description 更新部门信息
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UpdateDepartmentRequest true "更新部门请求参数"
// @Success 200 {object} response.Response{data=dto.DepartmentInfo} "更新成功"
// @Router /api/v1/departments [put]
func (h *Handler) UpdateDepartment(c *gin.Context) {
	var req dto.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.UpdateDepartment(c.Request.Context(), req.DepartmentID, &req)
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
// @Param request body dto.DepartmentStatusRequest true "更新状态请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Router /api/v1/departments/status [put]
func (h *Handler) UpdateDepartmentStatus(c *gin.Context) {
	var req dto.DepartmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.UpdateDepartmentStatus(c.Request.Context(), req.DepartmentID, req.Status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}
