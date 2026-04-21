package auth

import (
	"admin/internal/dto"
	tenantconv "admin/internal/service/tenant"
	"admin/pkg/constants"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// RefreshToken 刷新用户token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) {
	tokenPair, err := s.jwt.VerifyRefreshToken(ctx, refreshToken)
	if err != nil {
		log.Error().Err(err).Msg("刷新token失败")
		return nil, xerr.Wrap(xerr.ErrTokenInvalid.Code, "刷新token失败", err)
	}

	return &dto.RefreshResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// SwitchTenant 切换租户
func (s *Service) SwitchTenant(ctx context.Context, req *dto.SwitchTenantRequest) (*dto.LoginResponse, error) {
	userID := xcontext.GetUserID(ctx)
	currentTenantCode := xcontext.GetTenantCode(ctx)
	currentTenantID := xcontext.GetTenantID(ctx)
	userName := xcontext.GetUserName(ctx)
	roleIDs := xcontext.GetRoleIDs(ctx)
	roleCodes := xcontext.GetRoles(ctx)

	if userID == "" {
		return nil, xerr.ErrUnauthorized
	}

	isSuperAdmin := xcontext.HasRole(ctx, constants.SuperAdmin)

	// 验证目标租户
	targetTenant, err := s.tenantRepo.GetByIDManual(ctx, req.TenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Str("target_tenant_id", req.TenantID).Msg("目标租户不存在")
			return nil, xerr.ErrTenantNotFound
		}
		log.Error().Err(err).Str("target_tenant_id", req.TenantID).Msg("查询目标租户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询目标租户失败", err)
	}

	if targetTenant.Status != constants.StatusEnabled {
		log.Error().Str("target_tenant_id", req.TenantID).Msg("目标租户已禁用")
		return nil, xerr.ErrTenantNotFound
	}

	switch {
	case isSuperAdmin:
		// 超管可以切换到任意租户

	default:
		// 其他角色检查是否有该租户的 user_roles 记录
		if currentTenantID == req.TenantID {
			return nil, xerr.Wrap(xerr.ErrInvalidParams.Code, "已在当前租户", nil)
		}

		targetRoleIDs, err := s.userRoleRepo.GetUserRoleIDs(ctx, userID, req.TenantID)
		if err != nil || len(targetRoleIDs) == 0 {
			log.Error().Str("user_id", userID).Str("target_tenant_id", req.TenantID).Msg("用户无该租户访问权限")
			return nil, xerr.ErrUserTenantAccessDenied
		}

		// 获取目标租户的角色信息
		targetRoles, err := s.roleRepo.GetByIDs(ctx, targetRoleIDs)
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
		}
		roleIDs = targetRoleIDs
		roleCodes = make([]string, len(targetRoles))
		for i, role := range targetRoles {
			roleCodes[i] = role.RoleCode
		}
	}

	// 生成新 Token
	tokenPair, err := s.jwt.GenerateTokenPair(ctx, targetTenant.TenantID, targetTenant.TenantCode, userID, userName, roleCodes, roleIDs)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("生成JWT令牌失败")
		return nil, err
	}

	log.Info().Str("user_id", userID).Str("username", userName).Str("from_tenant", currentTenantCode).Str("to_tenant", targetTenant.TenantCode).Msg("用户切换租户成功")

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// GetAvailableTenants 获取用户可切换的租户列表
func (s *Service) GetAvailableTenants(ctx context.Context) (*dto.AvailableTenantsResponse, error) {
	isSuperAdmin := xcontext.HasRole(ctx, constants.SuperAdmin)

	if isSuperAdmin {
		tenantList, err := s.tenantRepo.ListAll(ctx)
		if err != nil {
			log.Error().Err(err).Msg("查询所有租户失败")
			return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询所有租户失败", err)
		}

		tenants := make([]*dto.TenantInfo, len(tenantList))
		for i, tenant := range tenantList {
			tenants[i] = tenantconv.ModelToTenantInfo(tenant)
		}
		return &dto.AvailableTenantsResponse{Tenants: tenants}, nil
	}

	tenantID := xcontext.GetTenantID(ctx)
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户信息失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询租户信息失败", err)
	}

	return &dto.AvailableTenantsResponse{
		Tenants: []*dto.TenantInfo{tenantconv.ModelToTenantInfo(tenant)},
	}, nil
}
