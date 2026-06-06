package audit

import (
	"encoding/json"
	"fmt"
	"strings"

	"admin/pkg/constants"
)

// LogOption 日志选项函数
type LogOption func(*LogEntry)

// WithModule 设置模块
func WithModule(module string) LogOption {
	return func(e *LogEntry) {
		e.Module = module
	}
}

// WithOperation 设置操作类型
func WithOperation(op string) LogOption {
	return func(e *LogEntry) {
		e.OperationType = op
	}
}

// WithResource 设置资源信息
func WithResource(resourceType, resourceID, resourceName string) LogOption {
	return func(e *LogEntry) {
		e.ResourceType = resourceType
		e.ResourceID = resourceID
		e.ResourceName = resourceName
	}
}

// WithBatchResource 设置批量资源信息
// resourceIDs: 资源ID列表，会被序列化为JSON数组存储
// resourceNames: 资源名称列表（用于汇总描述）
func WithBatchResource(resourceType string, resourceIDs []string, resourceNames []string) LogOption {
	return func(e *LogEntry) {
		e.ResourceType = resourceType
		// 将ID列表序列化为JSON数组字符串
		if len(resourceIDs) > 0 {
			e.ResourceID = toJsonArray(resourceIDs)
		}
		// 生成汇总描述
		if len(resourceNames) > 0 {
			e.ResourceName = formatBatchResourceNames(resourceNames)
		}
	}
}

// WithValue 设置变更值
func WithValue(oldValue, newValue any) LogOption {
	return func(e *LogEntry) {
		e.OldValue = oldValue
		e.NewValue = newValue
	}
}

// WithError 设置错误信息
func WithError(err error) LogOption {
	return func(e *LogEntry) {
		if err != nil {
			e.Status = constants.OperationStatusFailed
			e.ErrorMessage = err.Error()
		}
	}
}

// WithClient 设置客户端信息（可选，中间件已经自动提取）
func WithClient(ip, userAgent string) LogOption {
	return func(e *LogEntry) {
		if ip != "" {
			e.IPAddress = ip
		}
		if userAgent != "" {
			e.UserAgent = userAgent
		}
	}
}

// 便捷操作选项

// WithCreate 创建操作选项
func WithCreate(module string) LogOption {
	return func(e *LogEntry) {
		e.Module = module
		e.OperationType = constants.OperationCreate
	}
}

// WithUpdate 更新操作选项
func WithUpdate(module string) LogOption {
	return func(e *LogEntry) {
		e.Module = module
		e.OperationType = constants.OperationUpdate
	}
}

// WithDelete 删除操作选项
func WithDelete(module string) LogOption {
	return func(e *LogEntry) {
		e.Module = module
		e.OperationType = constants.OperationDelete
	}
}

// WithBatchDelete 批量删除操作选项
func WithBatchDelete(module string) LogOption {
	return func(e *LogEntry) {
		e.Module = module
		e.OperationType = constants.OperationBatchDelete
	}
}

// WithQuery 查询操作选项
func WithQuery(module string) LogOption {
	return func(e *LogEntry) {
		e.Module = module
		e.OperationType = constants.OperationQuery
	}
}

// WithExport 导出操作选项
func WithExport(module string) LogOption {
	return func(e *LogEntry) {
		e.Module = module
		e.OperationType = constants.OperationExport
	}
}

// WithLogin 登录操作选项
func WithLogin() LogOption {
	return func(e *LogEntry) {
		e.Module = constants.LoginTypePassword // 默认密码登录，可后续扩展支持 SSO、OAUTH
		e.OperationType = constants.OperationLogin
	}
}

// WithLoginEmail 邮箱登录操作选项
func WithLoginEmail() LogOption {
	return func(e *LogEntry) {
		e.Module = constants.LoginTypeEmail
		e.OperationType = constants.OperationLogin
	}
}

// WithLoginPhone 手机号登录操作选项
func WithLoginPhone() LogOption {
	return func(e *LogEntry) {
		e.Module = constants.LoginTypePhone
		e.OperationType = constants.OperationLogin
	}
}

// WithLogout 登出操作选项
func WithLogout() LogOption {
	return func(e *LogEntry) {
		e.Module = "" // 登出不需要登录类型
		e.OperationType = constants.OperationLogout
	}
}

// WithUser 设置用户信息（用于登录等场景）
func WithUser(tenantID, userID, userName string) LogOption {
	return func(e *LogEntry) {
		e.TenantID = tenantID
		e.UserID = userID
		e.UserName = userName
	}
}

// WithTenantID 设置租户ID（用于跨租户操作，如创建租户）
func WithTenantID(tenantID string) LogOption {
	return func(e *LogEntry) {
		e.TenantID = tenantID
	}
}

// toJsonArray 将字符串数组序列化为JSON数组字符串
func toJsonArray(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(items)
	return string(data)
}

// formatBatchResourceNames 格式化批量资源名称
// 如果超过3个，显示前3个 + "等N个"
func formatBatchResourceNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) <= 3 {
		return strings.Join(names, ", ")
	}
	return fmt.Sprintf("%s 等 %d 个", strings.Join(names[:3], ", "), len(names))
}
