package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"

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
// @Param log_id query string true "日志ID"
// @Success 200 {object} response.Response{data=dto.LoginLogInfo} "获取成功"
// @Router /api/v1/logs/login/detail [get]
func (h *LoginLogHandler) GetLoginLog(c *gin.Context) {
	var req dto.LoginLogDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.loginLogService.GetLoginLogByID(c.Request.Context(), req.LogID)
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
// @Param login_type query string false "登录类型筛选(PASSWORD:密码登录,EMAIL:邮箱登录,PHONE:手机号登录,SSO:单点登录,OAUTH:第三方登录)" Enums(PASSWORD,EMAIL,PHONE,SSO,OAUTH)
// @Param status query int false "状态筛选(0:失败,1:成功)" Enums(0,1)
// @Param start_date query int false "开始时间(毫秒时间戳)"
// @Param end_date query int false "结束时间(毫秒时间戳)"
// @Param ip_address query string false "IP地址筛选"
// @Success 200 {object} response.Response{data=dto.ListLoginLogsResponse} "获取成功"
// @Router /api/v1/logs/login [get]
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
