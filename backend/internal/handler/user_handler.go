package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService     *service.UserService
	userRoleService *service.UserRoleService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService, userRoleService *service.UserRoleService) *UserHandler {
	return &UserHandler{
		userService:     userService,
		userRoleService: userRoleService,
	}
}

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
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
// @Summary 获取用户详情
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id query string true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserInfo} "获取成功"
// @Router /api/v1/users/detail [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	var req dto.UserDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.userService.GetUserByID(c.Request.Context(), req.UserID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户基本信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UpdateUserRequest true "更新用户请求参数"
// @Success 200 {object} response.Response{data=dto.UserInfo} "更新成功"
// @Router /api/v1/users [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.userService.UpdateUser(c.Request.Context(), req.UserID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除用户（软删除）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UserDeleteRequest true "删除用户请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/users [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	var req dto.UserDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), req.UserID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

// BatchDeleteUsers 批量删除用户
// @Summary 批量删除用户
// @Description 批量软删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UserBatchDeleteRequest true "批量删除请求参数"
// @Success 200 {object} response.Response "删除成功"
// @Router /api/v1/users/batch-delete [delete]
func (h *UserHandler) BatchDeleteUsers(c *gin.Context) {
	var req dto.UserBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.userService.BatchDeleteUsers(c.Request.Context(), req.UserIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true, "count": len(req.UserIDs)})
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
// @Summary 更新用户状态
// @Description 启用或禁用用户账号
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UserStatusRequest true "更新状态请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Router /api/v1/users/status [put]
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	var req dto.UserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.userService.UpdateUserStatus(c.Request.Context(), req.UserID, req.Status); err != nil {
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
// @Router /api/v1/user/profile [get]
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
// @Param request body dto.AssignRolesRequest true "分配角色请求参数"
// @Success 200 {object} response.Response "分配成功"
// @Router /api/v1/users/roles [put]
func (h *UserHandler) AssignRoles(c *gin.Context) {
	var req dto.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.userRoleService.AssignRoles(c.Request.Context(), req.UserID, &req); err != nil {
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
// @Param user_id query string true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserRolesResponse} "获取成功"
// @Router /api/v1/users/roles [get]
func (h *UserHandler) GetUserRoles(c *gin.Context) {
	var req dto.UserDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	roles, err := h.userRoleService.GetUserRoles(c.Request.Context(), req.UserID)
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
// @Router /api/v1/user/password/change [post]
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
// @Description 超级管理员重置指定用户的密码，密码自动生成并在响应中显示一次
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.ResetPasswordRequest true "重置密码请求参数"
// @Success 200 {object} response.Response{data=dto.ResetPasswordResponse} "重置成功"
// @Router /api/v1/users/password/reset [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.userService.ResetPassword(c.Request.Context(), req.UserID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
