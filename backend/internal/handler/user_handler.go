package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"
	"admin/pkg/xerr"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新的用户账号，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateUserRequest true "创建用户请求参数"
// @Success 200 {object} response.Response{data=dto.UserResponse} "创建成功"
// @Success 200 {object} response.Response "请求参数错误"
// @Success 200 {object} response.Response "未授权访问"
// @Success 200 {object} response.Response "权限不足"
// @Success 200 {object} response.Response "服务器内部错误"
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	resp, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	var req dto.UpdateUserRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.userService.UpdateUser(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表，支持按用户名和状态筛选
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param user_name query string false "用户名筛选"
// @Param status query int false "状态筛选(0:禁用,1:启用)"
// @Success 200 {object} response.Response{data=dto.ListUsersResponse} "获取成功"
// @Success 200 {object} response.Response "请求参数错误"
// @Success 200 {object} response.Response "未授权访问"
// @Success 200 {object} response.Response "服务器内部错误"
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req dto.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.userService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateUserStatus 更新用户状态
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
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

	if err := h.userService.UpdateUserStatus(c.Request.Context(), userID, status); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}

// GetProfile 获取当前登录用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息、角色和租户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.ProfileResponse} "获取成功"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /api/v1/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 获取当前用户信息（从 context 中获取）
	user, err := h.userService.GetProfile(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// AssignRoles 为用户分配角色
// @Summary 为用户分配角色
// @Description 为指定用户分配角色（覆盖式，会替换用户现有的所有角色）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "用户ID"
// @Param request body dto.AssignRolesRequest true "角色编码列表"
// @Success 200 {object} response.Response "分配成功"
// @Router /api/v1/users/:user_id/roles [put]
func (h *UserHandler) AssignRoles(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	var req dto.AssignRolesRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.userService.AssignRoles(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"assigned": true})
}

// GetUserRoles 获取用户的角色列表
// @Summary 获取用户角色
// @Description 获取指定用户的角色列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserRolesResponse} "获取成功"
// @Router /api/v1/users/:user_id/roles [get]
func (h *UserHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	roles, err := h.userService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, roles)
}

// ChangePassword 用户修改自己的密码
// @Summary 修改密码
// @Description 用户修改自己的登录密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.ChangePasswordRequest true "修改密码请求参数"
// @Success 200 {object} response.Response{data=dto.ChangePasswordResponse} "修改成功"
// @Failure 200 {object} response.Response "请求参数错误"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "原密码错误"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /api/v1/password/change [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, &dto.ChangePasswordResponse{
		Success: true,
		Message: "密码修改成功，请重新登录",
	})
}

// ResetPassword 超管重置用户密码
// @Summary 重置用户密码
// @Description 超级管理员重置指定用户的密码，密码仅在响应中显示一次
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "用户ID"
// @Param request body dto.ResetPasswordRequest false "重置密码请求参数"
// @Success 200 {object} response.Response{data=dto.ResetPasswordResponse} "重置成功"
// @Failure 200 {object} response.Response "请求参数错误"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "权限不足"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /api/v1/users/:user_id/password/reset [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		response.Error(c, err)
		return
	}

	// 如果请求体为空，设置默认值（自动生成密码）
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() == "EOF" {
		req.Password = ""
	}

	resp, err := h.userService.ResetPassword(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
