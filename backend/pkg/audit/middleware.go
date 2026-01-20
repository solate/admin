package audit

import (
	"admin/pkg/bodyreader"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware 审计日志中间件
// 功能：提取请求信息并存入 context，供日志记录使用
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 提取并存储客户端信息
		clientInfo := &ClientInfo{
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}
		ctx = WithClientInfo(ctx, clientInfo)

		// 提取并存储请求信息
		requestInfo := &RequestInfo{
			Method: c.Request.Method,
			Path:   c.Request.URL.Path,
			Params: extractParams(c),
		}
		ctx = WithRequestInfo(ctx, requestInfo)

		// 更新请求的 context
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// extractParams 提取请求参数（带脱敏）
func extractParams(c *gin.Context) string {
	if c.Request.Method == "GET" {
		return c.Request.URL.RawQuery
	}

	contentType := c.Request.Header.Get("Content-Type")

	// 文件上传请求：只记录摘要信息，不记录文件内容
	if contentType != "" && strings.Contains(contentType, "multipart/form-data") {
		return extractMultipartSummary(c)
	}

	bodyStr, restoredBody := bodyreader.ReadBodyString(c.Request.Body)
	if restoredBody != nil {
		c.Request.Body = restoredBody
	}
	return bodyreader.SanitizeParams(bodyStr)
}

// extractMultipartSummary 提取 multipart 请求的摘要信息
func extractMultipartSummary(c *gin.Context) string {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return "{\"error\": \"无法解析表单数据\"}"
	}

	var parts []string

	// 记录普通字段
	if c.Request.MultipartForm != nil {
		for key, values := range c.Request.MultipartForm.Value {
			for i, value := range values {
				// 对敏感字段进行脱敏
				sanitizedValue := bodyreader.SanitizeParams(fmt.Sprintf(`"%s":"%s"`, key, value))
				parts = append(parts, fmt.Sprintf(`"%s[%d]":%s`, key, i, sanitizedValue))
			}
		}

		// 记录上传的文件信息（只记录文件名，不记录内容）
		for key, files := range c.Request.MultipartForm.File {
			for i, file := range files {
				parts = append(parts, fmt.Sprintf(`"%s[%d]":"文件: %s (%d bytes)"`, key, i, file.Filename, file.Size))
			}
		}
	}

	if len(parts) == 0 {
		return "{}"
	}

	return "{" + strings.Join(parts, ", ") + "}"
}
