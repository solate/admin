package user

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/constants"
	"admin/pkg/utils/pagination"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	roleconv "admin/internal/service/role"
	tenantconv "admin/internal/service/tenant"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GetUserByID 根据ID获取用户
func (s *Service) GetUserByID(ctx context.Context, userID string) (*dto.UserInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 获取用户角色信息（使用用户自己的租户ID）
	roles, err := s.userRoleService.GetUserRoleDetails(ctx, user.UserID, user.TenantID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Str("username", user.UserName).Str("tenant_id", user.TenantID).Msg("查询用户角色失败")
		// 即使查询角色失败，也继续处理，但不返回角色信息
		roles = nil
	}

	return modelToUserInfoWithRoles(user, roles), nil
}

// GetProfile 获取当前用户档案（含角色和租户信息）
func (s *Service) GetProfile(ctx context.Context) (*dto.ProfileResponse, error) {
	// 从 context 获取当前用户信息
	userID := xcontext.GetUserID(ctx)
	roleCodes := xcontext.GetRoles(ctx)
	tenantID := xcontext.GetTenantID(ctx)

	log.Debug().Str("user_id", userID).Strs("role_codes", roleCodes).Str("tenant_id", tenantID).Msg("[GetProfile] 开始获取用户档案")

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 获取租户信息
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户信息失败", err)
	}

	log.Debug().Strs("role_codes", roleCodes).Msg("[GetProfile] 准备查询角色详情")

	// 获取角色详情
	// 关键修改：使用 default 租户查询角色详情
	// 因为所有角色都存储在 default 租户中
	var roles []*model.Role
	if len(roleCodes) > 0 {
		// 获取 default 租户ID（使用跨租户查询）
		defaultTenants, err := s.tenantRepo.GetByCodes(ctx, []string{constants.DefaultTenantCode})
		if err != nil || len(defaultTenants) == 0 {
			log.Error().Err(err).Str("default_tenant_code", constants.DefaultTenantCode).Msg("查询 default 租户失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询 default 租户失败", err)
		}
		defaultTenant := defaultTenants[0]

		roles, err = s.roleRepo.ListByCodesWithTenant(ctx, defaultTenant.TenantID, roleCodes)
		if err != nil {
			log.Error().Err(err).Strs("role_codes", roleCodes).Str("tenant_id", defaultTenant.TenantID).Msg("查询角色详情失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色详情失败", err)
		}
	} else {
		log.Debug().Msg("[GetProfile] 角色代码为空，跳过角色查询")
		roles = []*model.Role{}
	}

	log.Debug().Int("roles_count", len(roles)).Msg("[GetProfile] 角色详情查询成功")

	return &dto.ProfileResponse{
		User:   modelToUserInfoWithRoles(user, roles),
		Tenant: tenantconv.ModelToTenantInfo(tenant),
		Roles:  roleconv.ModelListToRoleInfoList(roles),
	}, nil
}

// ListUsers 获取用户列表
func (s *Service) ListUsers(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	// 获取用户列表和总数，支持筛选条件和租户过滤
	users, total, err := s.userRepo.ListWithFiltersAndTenant(ctx, req.GetOffset(), req.GetLimit(), req.Nickname, req.Status, req.TenantID)
	if err != nil {
		log.Error().Err(err).
			Str("nickname", req.Nickname).
			Int("status", req.Status).
			Str("tenant_id", req.TenantID).
			Msg("查询用户列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户列表失败", err)
	}

	// 临时方案：如果不是超级管理员，过滤掉超级管理员用户
	filteredUsers := s.filterSuperAdminUsers(ctx, users)

	// 批量获取租户信息（收集所有租户ID）
	tenantIDs := make([]string, 0, len(filteredUsers))
	tenantMap := make(map[string]*model.Tenant)
	for _, user := range filteredUsers {
		if user.TenantID != "" {
			tenantIDs = append(tenantIDs, user.TenantID)
		}
	}

	// 使用手动模式批量查询租户信息
	if len(tenantIDs) > 0 {
		tenants, err := s.tenantRepo.GetByIDsManual(ctx, tenantIDs)
		if err != nil {
			log.Warn().Err(err).Msg("批量查询租户信息失败，将不返回租户详情")
		} else {
			for _, tenant := range tenants {
				tenantMap[tenant.TenantID] = tenant
			}
		}
	}

	// 转换为响应格式，并为每个用户填充角色和租户信息
	userInfos := make([]*dto.UserInfo, len(filteredUsers))
	for i, user := range filteredUsers {
		// 获取用户角色信息（使用用户自己的租户ID）
		roles, err := s.userRoleService.GetUserRoleDetails(ctx, user.UserID, user.TenantID)
		if err != nil {
			log.Error().Err(err).Str("user_id", user.UserID).Str("username", user.UserName).Str("tenant_id", user.TenantID).Msg("查询用户角色失败")
			// 即使查询角色失败，也继续处理，但不返回角色信息
			roles = nil
		}

		userInfo := modelToUserInfoWithRoles(user, roles)

		// 填充租户信息
		if tenant, ok := tenantMap[user.TenantID]; ok {
			userInfo.Tenant = tenantconv.ModelToTenantInfo(tenant)
		}

		userInfos[i] = userInfo
	}

	// 如果过滤后的数据量与原始总数不同，重新计算总数（仅当非超管时）
	filteredTotal := int64(len(userInfos))
	if !xcontext.HasRole(ctx, constants.SuperAdmin) && int(total) != len(users) {
		// 如果有分页且需要准确总数，这里简化处理，使用过滤后的数量
		// 注意：这是一个临时方案，可能影响分页准确性
		filteredTotal = int64(len(userInfos))
	} else {
		filteredTotal = total
	}

	return &dto.ListUsersResponse{
		Response: pagination.NewResponse(req.Request, filteredTotal),
		List:     userInfos,
	}, nil
}
