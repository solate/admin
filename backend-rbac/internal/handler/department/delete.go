package department

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// DeleteDepartment 删除部门
// @Summary 删除部门
// @Description 删除部门（软删除）
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.DepartmentDeleteRequest true "删除部门请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/departments [delete]
func (h *Handler) DeleteDepartment(c *gin.Context) {
	var req dto.DepartmentDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.DeleteDepartment(c.Request.Context(), req.DepartmentID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// BatchDeleteDepartments 批量删除部门
// @Summary 批量删除部门
// @Description 批量软删除部门（无子部门且无关联用户的部门才能删除）
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.DepartmentBatchDeleteRequest true "批量删除请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/departments/batch-delete [delete]
func (h *Handler) BatchDeleteDepartments(c *gin.Context) {
	var req dto.DepartmentBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.BatchDeleteDepartments(c.Request.Context(), req.DepartmentIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.DepartmentIDs)})
}
