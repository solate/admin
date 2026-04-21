package tenant

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/idgen"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
)

// CreateTenant 创建租户
func (s *Service) CreateTenant(ctx context.Context, req *dto.TenantCreateRequest) (resp *dto.TenantInfo, err error) {
	var tenant *model.Tenant

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleTenant),
				audit.WithError(err),
			)
		} else if tenant != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleTenant),
				audit.WithResource(constants.ResourceTypeTenant, tenant.TenantID, tenant.Name),
				audit.WithValue(nil, tenant),
			)
		}
	}()

	// 检查租户编码是否已存在
	exists, err := s.tenantRepo.CheckExists(ctx, req.TenantCode)
	if err != nil {
		log.Error().Err(err).Str("tenant_code", req.TenantCode).Msg("检查租户编码失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查租户编码失败", err)
	}
	if exists {
		log.Warn().Str("tenant_code", req.TenantCode).Msg("租户编码已存在")
		return nil, xerr.New(xerr.ErrConflict.Code, "租户编码已存在")
	}

	// 检查租户名称是否已存在
	nameExists, err := s.tenantRepo.CheckNameExists(ctx, req.Name)
	if err != nil {
		log.Error().Err(err).Str("name", req.Name).Msg("检查租户名称失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查租户名称失败", err)
	}
	if nameExists {
		log.Warn().Str("name", req.Name).Msg("租户名称已存在")
		return nil, xerr.New(xerr.ErrConflict.Code, "租户名称已存在")
	}

	// 生成租户ID
	tenantID, err := idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成租户ID失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成租户ID失败", err)
	}

	// 构建租户模型
	tenant = &model.Tenant{
		TenantID:     tenantID,
		TenantCode:   req.TenantCode,
		Name:         req.Name,
		Description:  req.Description,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Status:       int16(constants.StatusEnabled), // 默认启用
	}

	// 创建租户
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("tenant_code", req.TenantCode).Msg("创建租户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建租户失败", err)
	}

	return ModelToTenantInfo(tenant), nil
}
