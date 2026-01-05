package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/idgen"
	"admin/pkg/pagination"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// TenantService 租户服务
type TenantService struct {
	tenantRepo *repository.TenantRepo
	recorder   *audit.Recorder
}

// NewTenantService 创建租户服务
func NewTenantService(tenantRepo *repository.TenantRepo, recorder *audit.Recorder) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		recorder:   recorder,
	}
}

// CreateTenant 创建租户
func (s *TenantService) CreateTenant(ctx context.Context, req *dto.TenantCreateRequest) (resp *dto.TenantResponse, err error) {
	var tenant *model.Tenant

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(audit.ModuleTenant),
				audit.WithError(err),
			)
		} else if tenant != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(audit.ModuleTenant),
				audit.WithResource(audit.ResourceTenant, tenant.TenantID, tenant.Name),
				audit.WithValue(nil, tenant),
			)
		}
	}()

	// 检查租户编码是否已存在
	exists, err := s.tenantRepo.CheckExists(ctx, req.Code)
	if err != nil {
		log.Error().Err(err).Str("code", req.Code).Msg("检查租户编码失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查租户编码失败", err)
	}
	if exists {
		log.Warn().Str("code", req.Code).Msg("租户编码已存在")
		return nil, xerr.New(xerr.ErrConflict.Code, "租户编码已存在")
	}

	// 生成租户ID
	tenantID, err := idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成租户ID失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成租户ID失败", err)
	}

	// 构建租户模型
	tenant = &model.Tenant{
		TenantID:    tenantID,
		TenantCode:  req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      1, // 默认启用
	}

	// 创建租户
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("code", req.Code).Msg("创建租户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建租户失败", err)
	}

	return s.toTenantResponse(tenant), nil
}

// GetTenantByID 根据ID获取租户
func (s *TenantService) GetTenantByID(ctx context.Context, tenantID string) (*dto.TenantResponse, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("tenant_id", tenantID).Msg("租户不存在")
			return nil, xerr.ErrNotFound
		}
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户失败", err)
	}

	return s.toTenantResponse(tenant), nil
}

// UpdateTenant 更新租户
func (s *TenantService) UpdateTenant(ctx context.Context, tenantID string, req *dto.TenantUpdateRequest) (resp *dto.TenantResponse, err error) {
	var oldTenant, newTenant *model.Tenant

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleTenant),
				audit.WithError(err),
			)
		} else if newTenant != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleTenant),
				audit.WithResource(audit.ResourceTenant, newTenant.TenantID, newTenant.Name),
				audit.WithValue(oldTenant, newTenant),
			)
		}
	}()

	// 获取旧租户信息
	oldTenant, err = s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("tenant_id", tenantID).Msg("租户不存在")
			return nil, xerr.ErrNotFound
		}
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户失败")
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
		log.Error().Err(err).Str("tenant_id", tenantID).Interface("updates", updates).Msg("更新租户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新租户失败", err)
	}

	// 获取更新后的租户信息
	newTenant, err = s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("获取更新后租户信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后租户信息失败", err)
	}

	return s.toTenantResponse(newTenant), nil
}

// DeleteTenant 删除租户
func (s *TenantService) DeleteTenant(ctx context.Context, tenantID string) (err error) {
	var tenant *model.Tenant

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(audit.ModuleTenant),
				audit.WithError(err),
			)
		} else if tenant != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(audit.ModuleTenant),
				audit.WithResource(audit.ResourceTenant, tenant.TenantID, tenant.Name),
				audit.WithValue(tenant, nil),
			)
			log.Info().Str("tenant_id", tenantID).Msg("删除租户成功")
		}
	}()

	// 获取租户信息
	tenant, err = s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("tenant_id", tenantID).Msg("租户不存在")
			return xerr.ErrNotFound
		}
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询租户失败", err)
	}

	// TODO: 检查租户下是否还有用户，如果有则不允许删除
	// userCount, err := s.userRepo.CountByTenantID(ctx, tenantID)
	// if userCount > 0 {
	//     return xerr.New(xerr.ErrBadRequest.Code, "租户下还有用户，无法删除")
	// }

	// 删除租户
	if err := s.tenantRepo.Delete(ctx, tenantID); err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("删除租户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "删除租户失败", err)
	}

	return nil
}

// ListTenants 获取租户列表
func (s *TenantService) ListTenants(ctx context.Context, req *dto.TenantListRequest) (*dto.TenantListResponse, error) {
	// 获取租户列表和总数，支持筛选条件
	tenants, total, err := s.tenantRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.Code, req.Name, req.Status)
	if err != nil {
		log.Error().Err(err).
			Str("code", req.Code).
			Str("name", req.Name).
			Int("status", req.Status).
			Msg("查询租户列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户列表失败", err)
	}

	// 转换为响应格式
	tenantResponses := make([]*dto.TenantResponse, len(tenants))
	for i, tenant := range tenants {
		tenantResponses[i] = s.toTenantResponse(tenant)
	}

	return &dto.TenantListResponse{
		List:     tenantResponses,
		Response: pagination.NewResponse(req.Request, total),
	}, nil
}

// UpdateTenantStatus 更新租户状态
func (s *TenantService) UpdateTenantStatus(ctx context.Context, tenantID string, status int) (err error) {
	var oldTenant, newTenant *model.Tenant

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleTenant),
				audit.WithError(err),
			)
		} else if newTenant != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleTenant),
				audit.WithResource(audit.ResourceTenant, newTenant.TenantID, newTenant.Name),
				audit.WithValue(oldTenant, newTenant),
			)
			log.Info().Str("tenant_id", tenantID).Int("status", status).Msg("更新租户状态成功")
		}
	}()

	// 获取旧租户信息
	oldTenant, err = s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("tenant_id", tenantID).Msg("租户不存在")
			return xerr.ErrNotFound
		}
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询租户失败", err)
	}

	// 更新租户状态
	if err := s.tenantRepo.UpdateStatus(ctx, tenantID, status); err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Int("status", status).Msg("更新租户状态失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新租户状态失败", err)
	}

	// 获取更新后的租户信息
	newTenant, err = s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("获取更新后租户信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后租户信息失败", err)
	}

	return nil
}

// toTenantResponse 转换为租户响应格式
func (s *TenantService) toTenantResponse(tenant *model.Tenant) *dto.TenantResponse {
	return &dto.TenantResponse{
		TenantID:    tenant.TenantID,
		Code:        tenant.TenantCode,
		Name:        tenant.Name,
		Description: tenant.Description,
		Status:      int(tenant.Status),
		CreatedAt:   tenant.CreatedAt,
		UpdatedAt:   tenant.UpdatedAt,
	}
}
