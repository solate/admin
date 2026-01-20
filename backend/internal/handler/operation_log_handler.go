package handler

import (
	"admin/internal/dto"
	"admin/internal/service"
	"admin/pkg/response"

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
// @Param log_id query string true "日志ID"
// @Success 200 {object} response.Response{data=dto.OperationLogInfo} "获取成功"
// @Router /api/v1/logs/operation/detail [get]
func (h *OperationLogHandler) GetOperationLog(c *gin.Context) {
	var req dto.OperationLogDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err)
		return
	}

	resp, err := h.operationLogService.GetOperationLogByID(c.Request.Context(), req.LogID)
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
// @Param operation_type query string false "操作类型筛选(CREATE:创建,UPDATE:更新,DELETE:删除,BATCH_DELETE:批量删除,QUERY:查询,EXPORT:导出,IMPORT:导入,LOGIN:登录,LOGOUT:登出)" Enums(CREATE,UPDATE,DELETE,BATCH_DELETE,QUERY,EXPORT,IMPORT,LOGIN,LOGOUT)
// @Param resource_type query string false "资源类型筛选"
// @Param user_name query string false "用户名筛选"
// @Param status query int false "状态筛选(1:成功,2:失败)" Enums(1,2)
// @Param start_date query int false "开始时间(毫秒时间戳)"
// @Param end_date query int false "结束时间(毫秒时间戳)"
// @Success 200 {object} response.Response{data=dto.ListOperationLogsResponse} "获取成功"
// @Router /api/v1/logs/operation [get]
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
