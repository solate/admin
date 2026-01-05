package audit

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
			e.Status = StatusFailure
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
		e.OperationType = OperationCreate
	}
}

// WithUpdate 更新操作选项
func WithUpdate(module string) LogOption {
	return func(e *LogEntry) {
		e.Module = module
		e.OperationType = OperationUpdate
	}
}

// WithDelete 删除操作选项
func WithDelete(module string) LogOption {
	return func(e *LogEntry) {
		e.Module = module
		e.OperationType = OperationDelete
	}
}

// WithQuery 查询操作选项
func WithQuery(module string) LogOption {
	return func(e *LogEntry) {
		e.Module = module
		e.OperationType = OperationQuery
	}
}

// WithExport 导出操作选项
func WithExport(module string) LogOption {
	return func(e *LogEntry) {
		e.Module = module
		e.OperationType = OperationExport
	}
}

// WithLogin 登录操作选项
func WithLogin() LogOption {
	return func(e *LogEntry) {
		e.Module = LoginTypePassword // 默认密码登录，可后续扩展支持 SSO、OAUTH
		e.OperationType = OperationLogin
	}
}

// WithLogout 登出操作选项
func WithLogout() LogOption {
	return func(e *LogEntry) {
		e.Module = "" // 登出不需要登录类型
		e.OperationType = OperationLogout
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
