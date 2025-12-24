package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// TenantMemberHandler 租户成员处理器
type TenantMemberHandler struct {
	tenantMemberService *service.TenantMemberService
}

// NewTenantMemberHandler 创建租户成员处理器
func NewTenantMemberHandler(tenantMemberService *service.TenantMemberService) *TenantMemberHandler {
	return &TenantMemberHandler{
		tenantMemberService: tenantMemberService,
	}
}

// AddTenantMember 添加租户成员
// @Summary 添加租户成员
// @Description 租户管理员添加新成员到租户，自动生成初始密码。如果用户已存在则直接添加到租户。
// @Tags 租户成员管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.AddTenantMemberRequest true "添加租户成员请求参数"
// @Success 200 {object} response.Response{data=dto.AddTenantMemberResponse} "添加成功"
// @Failure 200 {object} response.Response "请求参数错误"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "权限不足"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /tenant-members [post]
func (h *TenantMemberHandler) AddTenantMember(c *gin.Context) {
	var req dto.AddTenantMemberRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.tenantMemberService.AddTenantMember(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// RemoveTenantMember 移除租户成员
// @Summary 移除租户成员
// @Description 从租户中移除成员（删除用户在该租户下的所有角色关系）
// @Tags 租户成员管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.RemoveTenantMemberRequest true "移除租户成员请求参数"
// @Success 200 {object} response.Response "移除成功"
// @Failure 200 {object} response.Response "请求参数错误"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "权限不足"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /tenant-members/remove [post]
func (h *TenantMemberHandler) RemoveTenantMember(c *gin.Context) {
	var req dto.RemoveTenantMemberRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.tenantMemberService.RemoveTenantMember(c.Request.Context(), req.UserID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"removed": true})
}

// UpdateMemberRoles 更新成员角色
// @Summary 更新成员角色
// @Description 更新租户成员的角色列表（替换所有角色）
// @Tags 租户成员管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UpdateMemberRolesRequest true "更新成员角色请求参数"
// @Success 200 {object} response.Response{data=dto.UpdateMemberRolesResponse} "更新成功"
// @Failure 200 {object} response.Response "请求参数错误"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "权限不足"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /tenant-members/roles [put]
func (h *TenantMemberHandler) UpdateMemberRoles(c *gin.Context) {
	var req dto.UpdateMemberRolesRequest
	if err := c.BindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.tenantMemberService.UpdateMemberRoles(c.Request.Context(), req.UserID, req.RoleIDs)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// ListTenantMembers 获取租户成员列表
// @Summary 获取租户成员列表
// @Description 分页获取租户成员列表，支持按关键词和状态筛选
// @Tags 租户成员管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "关键词搜索（用户名/姓名）"
// @Param status query int false "状态筛选(0:禁用,1:启用)"
// @Success 200 {object} response.Response{data=dto.ListTenantMembersResponse} "获取成功"
// @Failure 200 {object} response.Response "请求参数错误"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /tenant-members [get]
func (h *TenantMemberHandler) ListTenantMembers(c *gin.Context) {
	var req dto.ListTenantMembersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.tenantMemberService.ListTenantMembers(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// RemoveMemberByPath 移除租户成员（路径参数方式）
// @Summary 移除租户成员
// @Description 从租户中移除成员（通过路径参数指定用户ID）
// @Tags 租户成员管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "用户ID"
// @Success 200 {object} response.Response "移除成功"
// @Failure 200 {object} response.Response "请求参数错误"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "权限不足"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /tenant-members/{user_id} [delete]
func (h *TenantMemberHandler) RemoveMemberByPath(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	if err := h.tenantMemberService.RemoveTenantMember(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"removed": true})
}
