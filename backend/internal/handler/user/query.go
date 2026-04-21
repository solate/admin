package user

import (
	"admin/internal/dto"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id query string true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserInfo} "获取成功"
// @Router /api/v1/users/detail [get]
func (h *Handler) GetUser(c *gin.Context) {
	var req dto.UserDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.GetUserByID(c.Request.Context(), req.UserID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表，支持按昵称、状态和租户ID筛选
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param nickname query string false "昵称模糊搜索"
// @Param status query int false "状态筛选(1:正常,2:禁用)" Enums(1,2)
// @Param tenant_id query string false "租户ID筛选"
// @Success 200 {object} response.Response{data=dto.ListUsersResponse} "获取成功"
// @Router /api/v1/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	var req dto.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.svc.ListUsers(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetProfile 获取当前登录用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息、角色和租户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.ProfileResponse} "获取成功"
// @Router /api/v1/user/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	// 获取当前用户信息（从 context 中获取）
	user, err := h.svc.GetProfile(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}
