package user

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新的用户账号，密码自动生成并明文返回，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateUserRequest true "创建用户请求参数"
// @Success 200 {object} response.Response{data=dto.CreateUserResponse} "创建成功"
// @Router /api/v1/users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
