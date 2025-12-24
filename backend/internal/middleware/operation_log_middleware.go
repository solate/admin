package middleware

import (
	"admin/pkg/bodyreader"
	"admin/pkg/operationlog"
	"admin/pkg/useragent"
	"admin/pkg/xcontext"

	"github.com/gin-gonic/gin"
)

// OperationLogMiddleware 操作日志中间件
// 说明：
// - 只在业务代码设置了 LogContext 时才记录（通过 operationlog.Record* 函数）
// - 从 request.Context 获取用户信息（由 AuthMiddleware 注入）
// - 从 request 读取参数并脱敏处理
// - 从 useragent 获取 IP 和 User-Agent
// - 响应后异步写入日志
func OperationLogMiddleware(logger *operationlog.Logger) gin.HandlerFunc {
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
		lc, exists := operationlog.GetLogContext(c.Request.Context())
		if !exists {
			return // 业务代码没有设置 LogContext，跳过记录
		}

		// 5. 从 context 获取用户信息（AuthMiddleware 已注入）
		userID, _ := xcontext.GetUserID(c.Request.Context())
		userName, _ := xcontext.GetUserName(c.Request.Context())

		// 6. 构建日志条目
		entry := &operationlog.LogEntry{
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

		// 7. 根据响应状态更新日志状态
		updateLogStatusFromResponse(c, lc)

		// 8. 异步写入日志（不阻塞响应）
		_ = logger.Write(c.Request.Context(), entry)
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
func updateLogStatusFromResponse(c *gin.Context, lc *operationlog.LogContext) {
	// 检查是否有错误 (response.Error 会调用 c.Error)
	if len(c.Errors) > 0 {
		lc.SetError(c.Errors.Last())
	}
	// 注意：项目统一使用 HTTP 200，错误通过 response.Code 字段区分
	// 因此不需要检查 HTTP 状态码
}
