package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// OperationLogHandler 操作日志处理器
type OperationLogHandler struct {
	operationLogService *service.OperationLogService
}

// NewOperationLogHandler 创建操作日志处理器
func NewOperationLogHandler(operationLogService *service.OperationLogService) *OperationLogHandler {
	return &OperationLogHandler{
		operationLogService: operationLogService,
	}
}

// GetOperationLog 获取操作日志详情
// @Summary 获取操作日志详情
// @Description 根据日志ID获取操作日志详细信息
// @Tags 操作日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param log_id path string true "日志ID"
// @Success 200 {object} response.Response{data=dto.OperationLogResponse} "获取成功"
// @Failure 200 {object} response.Response "请求参数错误"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "资源不存在"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /operation-logs/{log_id} [get]
func (h *OperationLogHandler) GetOperationLog(c *gin.Context) {
	logID := c.Param("log_id")
	if logID == "" {
		response.Error(c, xerr.ErrInvalidParams)
		return
	}

	resp, err := h.operationLogService.GetOperationLogByID(c.Request.Context(), logID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// ListOperationLogs 获取操作日志列表
// @Summary 获取操作日志列表
// @Description 分页获取操作日志列表，支持多种筛选条件
// @Tags 操作日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param module query string false "模块筛选"
// @Param operation_type query string false "操作类型筛选(CREATE/UPDATE/DELETE/QUERY)"
// @Param resource_type query string false "资源类型筛选"
// @Param user_name query string false "用户名筛选"
// @Param status query int false "状态筛选(1:成功,2:失败)"
// @Param start_date query int false "开始时间(毫秒时间戳)"
// @Param end_date query int false "结束时间(毫秒时间戳)"
// @Success 200 {object} response.Response{data=dto.ListOperationLogsResponse} "获取成功"
// @Failure 200 {object} response.Response "请求参数错误"
// @Failure 200 {object} response.Response "未授权访问"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /operation-logs [get]
func (h *OperationLogHandler) ListOperationLogs(c *gin.Context) {
	var req dto.ListOperationLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.operationLogService.ListOperationLogs(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
