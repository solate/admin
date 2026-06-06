package tenant

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UpdateTenant 更新租户
func (s *Service) UpdateTenant(ctx context.Context, tenantID string, req *dto.TenantUpdateRequest) (resp *dto.TenantInfo, err error) {
	var oldTenant, newTenant *model.Tenant

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleTenant),
				audit.WithError(err),
			)
		} else if newTenant != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleTenant),
				audit.WithResource(constants.ResourceTypeTenant, newTenant.TenantID, newTenant.Name),
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
		// 检查租户名称是否已被其他租户使用
		nameExists, err := s.tenantRepo.CheckNameExists(ctx, req.Name, tenantID)
		if err != nil {
			log.Error().Err(err).Str("name", req.Name).Msg("检查租户名称失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查租户名称失败", err)
		}
		if nameExists {
			log.Warn().Str("name", req.Name).Msg("租户名称已存在")
			return nil, xerr.New(xerr.ErrConflict.Code, "租户名称已存在")
		}
		updates["name"] = req.Name
	}
	// description 可以为空字符串
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ContactName != "" {
		updates["contact_name"] = req.ContactName
	}
	if req.ContactPhone != "" {
		updates["contact_phone"] = req.ContactPhone
	}
	if req.Status != constants.StatusZero {
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

	return ModelToTenantInfo(newTenant), nil
}

// UpdateTenantStatus 更新租户状态
func (s *Service) UpdateTenantStatus(ctx context.Context, tenantID string, status int) (err error) {
	var oldTenant, newTenant *model.Tenant

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleTenant),
				audit.WithError(err),
			)
		} else if newTenant != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleTenant),
				audit.WithResource(constants.ResourceTypeTenant, newTenant.TenantID, newTenant.Name),
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
