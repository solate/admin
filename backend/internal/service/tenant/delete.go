package tenant

import (
	"admin/internal/dal/model"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/convert"
	"admin/pkg/xerr"
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DeleteTenant 删除租户
func (s *Service) DeleteTenant(ctx context.Context, tenantID string) (err error) {
	var tenant *model.Tenant

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleTenant),
				audit.WithError(err),
			)
		} else if tenant != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleTenant),
				audit.WithResource(constants.ResourceTypeTenant, tenant.TenantID, tenant.Name),
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

// BatchDeleteTenants 批量删除租户
func (s *Service) BatchDeleteTenants(ctx context.Context, tenantIDs []string) (err error) {
	var tenants []*model.Tenant

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleTenant),
				audit.WithError(err),
			)
		} else if len(tenants) > 0 {
			// 收集资源信息用于批量审计日志
			ids := make([]string, 0, len(tenants))
			names := make([]string, 0, len(tenants))
			for _, tenant := range tenants {
				ids = append(ids, tenant.TenantID)
				names = append(names, tenant.Name)
			}
			// 记录批量删除审计日志（单条日志记录所有资源）
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleTenant),
				audit.WithBatchResource(constants.ResourceTypeTenant, ids, names),
				audit.WithValue(tenants, nil),
			)
			log.Info().Strs("tenant_ids", tenantIDs).Int("count", len(tenantIDs)).Msg("批量删除租户成功")
		}
	}()

	// 获取所有租户信息
	tenants, err = s.tenantRepo.GetByIDs(ctx, tenantIDs)
	if err != nil {
		log.Error().Err(err).Strs("tenant_ids", tenantIDs).Msg("查询租户信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询租户信息失败", err)
	}

	// 验证所有租户都存在
	if len(tenants) != len(tenantIDs) {
		log.Warn().Int("requested", len(tenantIDs)).Int("found", len(tenants)).Msg("部分租户不存在")
		return xerr.New(xerr.ErrNotFound.Code, "部分租户不存在")
	}

	// 检查租户下是否还有用户
	userCounts, err := s.userRepo.CountByTenantIDs(ctx, tenantIDs)
	if err != nil {
		log.Error().Err(err).Strs("tenant_ids", tenantIDs).Msg("检查租户用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "检查租户用户失败", err)
	}

	// 构造租户 map 用于快速查找
	tenantMap := convert.ToMap(tenants, func(t *model.Tenant) string { return t.TenantID })

	// 收集有用户的租户
	var tenantsWithUsers []string
	for _, tenantID := range tenantIDs {
		if count, exists := userCounts[tenantID]; exists && count > 0 {
			if tenant, ok := tenantMap[tenantID]; ok {
				tenantsWithUsers = append(tenantsWithUsers, fmt.Sprintf("%s(%d个用户)", tenant.Name, count))
			}
		}
	}

	if len(tenantsWithUsers) > 0 {
		log.Warn().Strs("tenants_with_users", tenantsWithUsers).Msg("以下租户下还有用户，无法删除")
		return xerr.New(xerr.ErrInvalidParams.Code, fmt.Sprintf("以下租户下还有用户，无法删除：%s", strings.Join(tenantsWithUsers, "、")))
	}

	// 批量删除租户
	if err := s.tenantRepo.BatchDelete(ctx, tenantIDs); err != nil {
		log.Error().Err(err).Strs("tenant_ids", tenantIDs).Msg("批量删除租户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "批量删除租户失败", err)
	}

	return nil
}
