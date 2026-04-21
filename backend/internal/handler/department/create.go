package department

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// CreateDepartment 创建部门
// @Summary 创建部门
// @Description 创建新的部门
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateDepartmentRequest true "创建部门请求参数"
// @Success 200 {object} response.Response{data=dto.DepartmentInfo} "创建成功"
// @Router /api/v1/departments [post]
func (h *Handler) CreateDepartment(c *gin.Context) {
	var req dto.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.CreateDepartment(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
