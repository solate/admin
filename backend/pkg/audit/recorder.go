package audit

import (
	"context"
	"time"

	"admin/pkg/constants"
	"admin/pkg/xcontext"
)

// Recorder 审计日志记录器
type Recorder struct {
	db *DB
}

// NewRecorder 创建记录器
func NewRecorder(db *DB) *Recorder {
	return &Recorder{db: db}
}

// Log 记录操作日志（统一的日志记录方法）
func (r *Recorder) Log(ctx context.Context, opts ...LogOption) {
	entry := &LogEntry{
		Status:    constants.OperationStatusSuccess,
		CreatedAt: time.Now().UnixMilli(),
	}

	// 应用选项
	for _, opt := range opts {
		opt(entry)
	}

	// 从 context 获取租户和用户信息
	if entry.TenantID == "" {
		entry.TenantID = xcontext.GetTenantID(ctx)
	}
	if entry.UserID == "" {
		entry.UserID = xcontext.GetUserID(ctx)
	}
	if entry.UserName == "" {
		entry.UserName = xcontext.GetUserName(ctx)
	}

	// 从 context 获取客户端和请求信息
	if clientInfo := GetClientInfo(ctx); clientInfo != nil {
		if entry.IPAddress == "" {
			entry.IPAddress = clientInfo.IP
		}
		if entry.UserAgent == "" {
			entry.UserAgent = clientInfo.UserAgent
		}
	}

	if reqInfo := GetRequestInfo(ctx); reqInfo != nil {
		if entry.RequestMethod == "" {
			entry.RequestMethod = reqInfo.Method
		}
		if entry.RequestPath == "" {
			entry.RequestPath = reqInfo.Path
		}
		if entry.RequestParams == "" {
			entry.RequestParams = reqInfo.Params
		}
	}

	// 异步写入（使用独立的 background context，避免被请求取消影响）
	// 同时保留租户和用户信息，供 GORM 租户插件使用
	go r.db.Write(xcontext.CopyContext(ctx), entry)
}

// Login 记录登录日志
func (r *Recorder) Login(ctx context.Context, tenantID, userID, userName string, err error) {
	opts := []LogOption{
		WithLogin(),
		WithUser(tenantID, userID, userName),
	}
	if err != nil {
		opts = append(opts, WithError(err))
	}
	r.Log(ctx, opts...)
}

// LoginEmail 记录邮箱登录日志
func (r *Recorder) LoginEmail(ctx context.Context, tenantID, userID, userName string, err error) {
	opts := []LogOption{
		WithLoginEmail(),
		WithUser(tenantID, userID, userName),
	}
	if err != nil {
		opts = append(opts, WithError(err))
	}
	r.Log(ctx, opts...)
}

// LoginPhone 记录手机号登录日志
func (r *Recorder) LoginPhone(ctx context.Context, tenantID, userID, userName string, err error) {
	opts := []LogOption{
		WithLoginPhone(),
		WithUser(tenantID, userID, userName),
	}
	if err != nil {
		opts = append(opts, WithError(err))
	}
	r.Log(ctx, opts...)
}

// Logout 记录登出日志
func (r *Recorder) Logout(ctx context.Context) {
	r.Log(ctx, WithLogout())
}

// RecordCreate 记录创建操作
func (r *Recorder) RecordCreate(ctx context.Context, module, resourceType, resourceID, resourceName string, newValue any) {
	r.Log(ctx,
		WithCreate(module),
		WithResource(resourceType, resourceID, resourceName),
		WithValue(nil, newValue),
	)
}

// RecordUpdate 记录更新操作
func (r *Recorder) RecordUpdate(ctx context.Context, module, resourceType, resourceID, resourceName string, oldValue, newValue any) {
	r.Log(ctx,
		WithUpdate(module),
		WithResource(resourceType, resourceID, resourceName),
		WithValue(oldValue, newValue),
	)
}

// RecordDelete 记录删除操作
func (r *Recorder) RecordDelete(ctx context.Context, module, resourceType, resourceID, resourceName string, oldValue any) {
	r.Log(ctx,
		WithDelete(module),
		WithResource(resourceType, resourceID, resourceName),
		WithValue(oldValue, nil),
	)
}

// RecordQuery 记录查询操作
func (r *Recorder) RecordQuery(ctx context.Context, module, resourceType string) {
	r.Log(ctx,
		WithQuery(module),
		WithResource(resourceType, "", ""),
	)
}

// RecordExport 记录导出操作
func (r *Recorder) RecordExport(ctx context.Context, module, resourceType, resourceID, resourceName string) {
	r.Log(ctx,
		WithExport(module),
		WithResource(resourceType, resourceID, resourceName),
	)
}
