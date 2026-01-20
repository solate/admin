package service

import (
	"admin/internal/converter"
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/convert"
	"admin/pkg/idgen"
	"admin/pkg/pagination"
	"admin/pkg/xerr"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// TenantService 租户服务
type TenantService struct {
	tenantRepo *repository.TenantRepo
	userRepo   *repository.UserRepo
	recorder   *audit.Recorder
}

// NewTenantService 创建租户服务
func NewTenantService(tenantRepo *repository.TenantRepo, userRepo *repository.UserRepo, recorder *audit.Recorder) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		recorder:   recorder,
	}
}

// CreateTenant 创建租户
func (s *TenantService) CreateTenant(ctx context.Context, req *dto.TenantCreateRequest) (resp *dto.TenantInfo, err error) {
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

	return converter.ModelToTenantInfo(tenant), nil
}

// GetTenantByID 根据ID获取租户
func (s *TenantService) GetTenantByID(ctx context.Context, tenantID string) (*dto.TenantInfo, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("tenant_id", tenantID).Msg("租户不存在")
			return nil, xerr.ErrNotFound
		}
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户失败", err)
	}

	return converter.ModelToTenantInfo(tenant), nil
}

// UpdateTenant 更新租户
func (s *TenantService) UpdateTenant(ctx context.Context, tenantID string, req *dto.TenantUpdateRequest) (resp *dto.TenantInfo, err error) {
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

	return converter.ModelToTenantInfo(newTenant), nil
}

// DeleteTenant 删除租户
func (s *TenantService) DeleteTenant(ctx context.Context, tenantID string) (err error) {
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
func (s *TenantService) BatchDeleteTenants(ctx context.Context, tenantIDs []string) (err error) {
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

// ListTenants 获取租户列表
func (s *TenantService) ListTenants(ctx context.Context, req *dto.TenantListRequest) (*dto.TenantListResponse, error) {
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
		tenantInfos[i] = converter.ModelToTenantInfo(tenant)
	}

	return &dto.TenantListResponse{
		List:     tenantInfos,
		Response: pagination.NewResponse(req.Request, total),
	}, nil
}

// GetAllTenants 获取所有启用的租户列表（不分页）
func (s *TenantService) GetAllTenants(ctx context.Context) ([]*dto.TenantInfo, error) {
	tenants, err := s.tenantRepo.ListAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("查询所有租户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询所有租户失败", err)
	}

	// 转换为响应格式
	tenantInfos := make([]*dto.TenantInfo, len(tenants))
	for i, tenant := range tenants {
		tenantInfos[i] = converter.ModelToTenantInfo(tenant)
	}

	return tenantInfos, nil
}

// UpdateTenantStatus 更新租户状态
func (s *TenantService) UpdateTenantStatus(ctx context.Context, tenantID string, status int) (err error) {
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
