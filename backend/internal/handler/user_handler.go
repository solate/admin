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
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*xerr.AppError); ok {
			response.Error(c, appErr)
		} else {
			response.Error(c, err)
		}
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
		if appErr, ok := err.(*xerr.AppError); ok {
			response.Error(c, appErr)
		} else {
			response.Error(c, err)
		}
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
		if appErr, ok := err.(*xerr.AppError); ok {
			response.Error(c, appErr)
		} else {
			response.Error(c, err)
		}
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
		if appErr, ok := err.(*xerr.AppError); ok {
			response.Error(c, appErr)
		} else {
			response.Error(c, err)
		}
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
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req dto.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.userService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*xerr.AppError); ok {
			response.Error(c, appErr)
		} else {
			response.Error(c, err)
		}
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

	status, err := strconv.ParseInt(statusStr, 10, 32)
	if err != nil {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	if err := h.userService.UpdateUserStatus(c.Request.Context(), userID, int32(status)); err != nil {
		if appErr, ok := err.(*xerr.AppError); ok {
			response.Error(c, appErr)
		} else {
			response.Error(c, err)
		}
		return
	}

	response.Success(c, gin.H{"updated": true})
}
