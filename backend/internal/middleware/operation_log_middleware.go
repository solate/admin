package middleware

import (
	"admin/pkg/bodyreader"
	"admin/pkg/constants"
	"admin/pkg/auditlog"
	"admin/pkg/useragent"
	"admin/pkg/xcontext"

	"github.com/gin-gonic/gin"
)

// OperationLogMiddleware 操作日志中间件
// 说明：
// - 只处理操作日志（CREATE/UPDATE/DELETE/QUERY）
// - 跳过登录日志（LOGIN/LOGOUT 由 AuthService 直接写入）
// - 业务代码设置 LogContext 后，中间件自动收集 HTTP 请求信息并写入
func OperationLogMiddleware(writer *auditlog.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 读取并保存请求参数（需要提前读取，因为 Body 只能读一次）
		requestParams := extractRequestParams(c)

		// 2. 获取客户端信息（一次性获取，存入 gin.Context）
		clientInfo := useragent.GetClientInfo(c.Request)
		c.Set("client_info", clientInfo)
		c.Set("request_params", requestParams)

		// 3. 处理请求
		c.Next()

		// 4. 请求处理完成后，检查是否需要记录操作日志
		lc, exists := auditlog.GetLogContext(c.Request.Context())
		if !exists {
			return // 业务代码没有设置 LogContext，跳过记录
		}

		// 5. 跳过登录日志（由 AuthService 直接写入）
		if lc.OperationType == constants.OperationLogin || lc.OperationType == constants.OperationLogout {
			return
		}

		// 6. 从 context 获取用户信息（AuthMiddleware 已注入）
		userID := xcontext.GetUserID(c.Request.Context())
		userName := xcontext.GetUserName(c.Request.Context())

		// 7. 构建日志条目
		entry := &auditlog.LogEntry{
			TenantID:      lc.TenantID,
			UserID:        userID,
			UserName:      userName,
			RequestMethod: c.Request.Method,
			RequestPath:   c.Request.URL.Path,
			RequestParams: requestParams,
			IPAddress:     clientInfo.IP,
			UserAgent:     clientInfo.UserAgent,
			LogContext:    lc,
		}

		// 8. 根据响应状态更新日志状态
		updateLogStatusFromResponse(c, lc)

		// 9. 异步写入日志（不阻塞响应）
		_ = writer.Write(c.Request.Context(), entry)
	}
}

// extractRequestParams 提取请求参数（带脱敏）
func extractRequestParams(c *gin.Context) string {
	if c.Request.Method == "GET" {
		return c.Request.URL.RawQuery
	}

	if c.Request.Body == nil {
		return ""
	}

	// 使用 bodyreader 读取并恢复 Body
	bodyStr, restoredBody := bodyreader.ReadBodyString(c.Request.Body)
	if restoredBody != nil {
		c.Request.Body = restoredBody
	}

	// 脱敏处理
	return bodyreader.SanitizeParams(bodyStr)
}

// updateLogStatusFromResponse 根据响应状态更新日志状态
func updateLogStatusFromResponse(c *gin.Context, lc *auditlog.LogContext) {
	// 检查是否有错误 (response.Error 会调用 c.Error)
	if len(c.Errors) > 0 {
		lc.Status = 2
		lc.ErrorMessage = c.Errors.Last().Error()
	}
	// 注意：项目统一使用 HTTP 200，错误通过 response.Code 字段区分
	// 因此不需要检查 HTTP 状态码
}
