package tenant

import (
	"admin/internal/dto"
	"admin/pkg/utils/pagination"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GetTenantByID 根据ID获取租户
func (s *Service) GetTenantByID(ctx context.Context, tenantID string) (*dto.TenantInfo, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("tenant_id", tenantID).Msg("租户不存在")
			return nil, xerr.ErrNotFound
		}
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户失败", err)
	}

	return ModelToTenantInfo(tenant), nil
}

// ListTenants 获取租户列表
func (s *Service) ListTenants(ctx context.Context, req *dto.TenantListRequest) (*dto.TenantListResponse, error) {
	// 获取租户列表和总数，支持筛选条件
	tenants, total, err := s.tenantRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.TenantCode, req.Name, req.Status)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_code", req.TenantCode).
			Str("name", req.Name).
			Int("status", req.Status).
			Msg("查询租户列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户列表失败", err)
	}

	// 转换为响应格式
	tenantInfos := make([]*dto.TenantInfo, len(tenants))
	for i, tenant := range tenants {
		tenantInfos[i] = ModelToTenantInfo(tenant)
	}

	return &dto.TenantListResponse{
		List:     tenantInfos,
		Response: pagination.NewResponse(req.Request, total),
	}, nil
}

// GetAllTenants 获取所有启用的租户列表（不分页）
func (s *Service) GetAllTenants(ctx context.Context) ([]*dto.TenantInfo, error) {
	tenants, err := s.tenantRepo.ListAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("查询所有租户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询所有租户失败", err)
	}

	// 转换为响应格式
	tenantInfos := make([]*dto.TenantInfo, len(tenants))
	for i, tenant := range tenants {
		tenantInfos[i] = ModelToTenantInfo(tenant)
	}

	return tenantInfos, nil
}
