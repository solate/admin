package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/constants"
	"admin/pkg/idgen"
	"admin/pkg/operationlog"
	"admin/pkg/pagination"
	"admin/pkg/xerr"
	"context"
	"time"

	"gorm.io/gorm"
)

// TenantService 租户服务
type TenantService struct {
	tenantRepo *repository.TenantRepo
}

// NewTenantService 创建租户服务
func NewTenantService(tenantRepo *repository.TenantRepo) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
	}
}

// CreateTenant 创建租户
func (s *TenantService) CreateTenant(ctx context.Context, req *dto.TenantCreateRequest) (*dto.TenantResponse, error) {
	// 检查租户编码是否已存在
	exists, err := s.tenantRepo.CheckExists(ctx, req.Code)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查租户编码失败", err)
	}
	if exists {
		return nil, xerr.New(xerr.ErrConflict.Code, "租户编码已存在")
	}

	// 生成租户ID
	tenantID, err := idgen.GenerateUUID()
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成租户ID失败", err)
	}

	// 构建租户模型
	tenant := &model.Tenant{
		TenantID:   tenantID,
		TenantCode: req.Code,
		Name:       req.Name,
		Status:     1, // 默认启用
	}

	// 设置可选描述
	if req.Description != "" {
		tenant.Description = &req.Description
	}

	// 创建租户
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建租户失败", err)
	}

	// 记录操作日志
	ctx = operationlog.RecordCreate(ctx, constants.ModuleTenant, constants.ResourceTypeTenant, tenant.TenantID, tenant.Name, tenant)

	return s.toTenantResponse(tenant), nil
}

// GetTenantByID 根据ID获取租户
func (s *TenantService) GetTenantByID(ctx context.Context, tenantID string) (*dto.TenantResponse, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户失败", err)
	}

	return s.toTenantResponse(tenant), nil
}

// UpdateTenant 更新租户
func (s *TenantService) UpdateTenant(ctx context.Context, tenantID string, req *dto.TenantUpdateRequest) (*dto.TenantResponse, error) {
	// 检查租户是否存在
	oldTenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户失败", err)
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	} else if req.Description == "" {
		// 空字符串清空描述
		updates["description"] = nil
	}
	if req.Status != 0 {
		updates["status"] = int16(req.Status)
	}
	updates["updated_at"] = time.Now().UnixMilli()

	// 更新租户
	if err := s.tenantRepo.Update(ctx, tenantID, updates); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新租户失败", err)
	}

	// 获取更新后的租户信息
	updatedTenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后租户信息失败", err)
	}

	// 记录操作日志
	ctx = operationlog.RecordUpdate(ctx, constants.ModuleTenant, constants.ResourceTypeTenant, updatedTenant.TenantID, updatedTenant.Name, oldTenant, updatedTenant)

	return s.toTenantResponse(updatedTenant), nil
}

// DeleteTenant 删除租户
func (s *TenantService) DeleteTenant(ctx context.Context, tenantID string) error {
	// 检查租户是否存在
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询租户失败", err)
	}

	// TODO: 检查租户下是否还有用户，如果有则不允许删除
	// userCount, err := s.userRepo.CountByTenantID(ctx, tenantID)
	// if userCount > 0 {
	//     return xerr.New(xerr.ErrBadRequest.Code, "租户下还有用户，无法删除")
	// }

	// 删除租户
	if err := s.tenantRepo.Delete(ctx, tenantID); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "删除租户失败", err)
	}

	// 记录操作日志
	operationlog.RecordDelete(ctx, constants.ModuleTenant, constants.ResourceTypeTenant, tenant.TenantID, tenant.Name, tenant)

	return nil
}

// ListTenants 获取租户列表
func (s *TenantService) ListTenants(ctx context.Context, req *dto.TenantListRequest) (*dto.TenantListResponse, error) {
	// 获取租户列表和总数，支持筛选条件
	tenants, total, err := s.tenantRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.Code, req.Name, req.Status)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户列表失败", err)
	}

	// 转换为响应格式
	tenantResponses := make([]*dto.TenantResponse, len(tenants))
	for i, tenant := range tenants {
		tenantResponses[i] = s.toTenantResponse(tenant)
	}

	// 构建分页响应
	return &dto.TenantListResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     tenantResponses,
	}, nil
}

// UpdateTenantStatus 更新租户状态
func (s *TenantService) UpdateTenantStatus(ctx context.Context, tenantID string, status int) error {
	// 检查租户是否存在
	oldTenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询租户失败", err)
	}

	// 更新租户状态
	if err := s.tenantRepo.UpdateStatus(ctx, tenantID, status); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "更新租户状态失败", err)
	}

	// 获取更新后的租户信息
	updatedTenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后租户信息失败", err)
	}

	// 记录操作日志
	operationlog.RecordUpdate(ctx, constants.ModuleTenant, constants.ResourceTypeTenant, updatedTenant.TenantID, updatedTenant.Name, oldTenant, updatedTenant)

	return nil
}

// toTenantResponse 转换为租户响应格式
func (s *TenantService) toTenantResponse(tenant *model.Tenant) *dto.TenantResponse {
	resp := &dto.TenantResponse{
		TenantID:  tenant.TenantID,
		Code:      tenant.TenantCode,
		Name:      tenant.Name,
		Status:    int(tenant.Status),
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}

	// 处理可选描述字段
	if tenant.Description != nil {
		resp.Description = *tenant.Description
	}

	return resp
}
