package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// LoginLogHandler 登录日志处理器
type LoginLogHandler struct {
	loginLogService *service.LoginLogService
}

// NewLoginLogHandler 创建登录日志处理器
func NewLoginLogHandler(loginLogService *service.LoginLogService) *LoginLogHandler {
	return &LoginLogHandler{
		loginLogService: loginLogService,
	}
}

// GetLoginLog 获取登录日志详情
// @Summary 获取登录日志详情
// @Description 根据日志ID获取登录日志详细信息
// @Tags 登录日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param log_id path string true "日志ID"
// @Success 200 {object} response.Response{data=dto.LoginLogResponse} "获取成功"
// @Failure 200 {object} response.Response "请求参数错误"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "资源不存在"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /api/v1/login-logs/:log_id [get]
func (h *LoginLogHandler) GetLoginLog(c *gin.Context) {
	logID := c.Param("log_id")
	if logID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	resp, err := h.loginLogService.GetLoginLogByID(c.Request.Context(), logID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// ListLoginLogs 获取登录日志列表
// @Summary 获取登录日志列表
// @Description 分页获取登录日志列表，支持多种筛选条件
// @Tags 登录日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param user_id query string false "用户ID筛选"
// @Param user_name query string false "用户名筛选"
// @Param login_type query string false "登录类型筛选(PASSWORD/SSO/OAUTH)"
// @Param status query int false "状态筛选(1:成功,0:失败)"
// @Param start_date query int false "开始时间(毫秒时间戳)"
// @Param end_date query int false "结束时间(毫秒时间戳)"
// @Param ip_address query string false "IP地址筛选"
// @Success 200 {object} response.Response{data=dto.ListLoginLogsResponse} "获取成功"
// @Failure 200 {object} response.Response "请求参数错误"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /api/v1/login-logs [get]
func (h *LoginLogHandler) ListLoginLogs(c *gin.Context) {
	var req dto.ListLoginLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.loginLogService.ListLoginLogs(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
