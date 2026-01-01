package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/captcha"
	"admin/pkg/casbin"
	"admin/pkg/config"
	"admin/pkg/constants"
	"admin/pkg/jwt"
	"admin/pkg/passwordgen"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	userRepo     *repository.UserRepo
	userRoleRepo *repository.UserRoleRepo
	roleRepo     *repository.RoleRepo
	tenantRepo   *repository.TenantRepo
	jwt          *jwt.Manager
	captcha      *captcha.Manager
	enforcer     *casbin.Enforcer
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo *repository.UserRepo, userRoleRepo *repository.UserRoleRepo, roleRepo *repository.RoleRepo, tenantRepo *repository.TenantRepo, jwt *jwt.Manager, rdb redis.UniversalClient, enforcer *casbin.Enforcer, config *config.Config) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		tenantRepo:   tenantRepo,
		jwt:          jwt,
		captcha:      captcha.NewManager(rdb),
		enforcer:     enforcer,
	}
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {

	// 验证码校验
	if !s.captcha.Verify(req.CaptchaID, req.Captcha) {
		return nil, xerr.ErrCaptchaInvalid
	}

	// 获取租户信息
	tenantID := xcontext.GetTenantID(ctx)
	tenantCode := xcontext.GetTenantCode(ctx)
	if tenantCode == "" {
		log.Error().Str("username", req.UserName).Msg("租户编码不能为空")
		return nil, xerr.ErrTenantCodeRequired
	}

	// 查询用户（租户内用户名唯一）
	user, err := s.userRepo.GetByTenantAndUserName(ctx, tenantID, req.UserName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Err(err).Str("username", req.UserName).Str("tenant_id", tenantID).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("username", req.UserName).Str("tenant_id", tenantID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 验证密码
	if !passwordgen.VerifyPassword(req.Password, user.Password) {
		return nil, xerr.ErrInvalidCredentials
	}

	// 检查用户状态
	if user.Status != constants.StatusEnabled {
		return nil, xerr.ErrUserDisabled
	}

	// 获取用户在租户中的角色（从 Casbin）
	roleCodes, err := s.userRoleRepo.GetUserRoles(ctx, user.UserName, tenantCode)
	if err != nil {
		log.Error().Err(err).Str("username", user.UserName).Str("tenant_code", tenantCode).Msg("查询用户角色失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户角色失败", err)
	}

	// 检查用户是否有角色
	if len(roleCodes) == 0 {
		return nil, xerr.ErrUserNoRoles
	}

	// 生成JWT令牌
	tokenPair, err := s.jwt.GenerateTokenPair(ctx, tenantID, tenantCode, user.UserID, user.UserName, roleCodes)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Str("username", user.UserName).Msg("生成JWT令牌失败")
		return nil, err
	}

	// 更新最后登录时间
	if err := s.userRepo.Update(ctx, user.UserID, map[string]interface{}{
		"last_login_time": time.Now().UnixMilli(),
	}); err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("更新最后登录时间失败")
		// 不影响登录流程，继续返回
	}

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// RefreshToken 刷新用户token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) {

	// 调用JWT manager刷新token
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

// Logout 用户登出
func (s *AuthService) Logout(ctx context.Context) error {
	tokenID := xcontext.GetTokenID(ctx)
	if tokenID == "" {
		return xerr.ErrUnauthorized
	}

	// 撤销当前token（加入黑名单并删除refresh token）
	if err := s.jwt.RevokeToken(ctx, tokenID); err != nil {
		log.Error().Err(err).Str("token_id", tokenID).Msg("撤销token失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "撤销token失败", err)
	}

	return nil
}

// SwitchTenant 切换租户
func (s *AuthService) SwitchTenant(ctx context.Context, req *dto.SwitchTenantRequest) (*dto.LoginResponse, error) {
	// 1. 获取当前用户信息
	userID := xcontext.GetUserID(ctx)
	currentTenantCode := xcontext.GetTenantCode(ctx)
	currentTenantID := xcontext.GetTenantID(ctx)
	userName := xcontext.GetUserName(ctx)

	if userID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 2. 检查用户角色，判断是否有权限切换租户
	isSuperAdmin := xcontext.HasRole(ctx, constants.SuperAdmin)
	isAuditor := xcontext.HasRole(ctx, constants.Auditor)
	roles := xcontext.GetRoles(ctx)

	// 3. 验证目标租户是否存在且状态正常
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

	// 4. 根据角色验证是否有权限切换到目标租户
	switch {
	case isSuperAdmin:
		// 超管：可以切换到任意租户

	case isAuditor:
		// 审计员：只能切换到有权限的租户
		// 验证用户是否在该租户中有 auditor 角色
		hasPermission, err := s.enforcer.HasGroupingPolicy(userName, constants.Auditor, targetTenant.TenantCode)
		if err != nil {
			log.Error().Err(err).Str("username", userName).Str("target_tenant_code", targetTenant.TenantCode).Msg("查询审计员权限失败")
			return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询审计员权限失败", err)
		}

		if !hasPermission {
			log.Error().Str("username", userName).Str("target_tenant_code", targetTenant.TenantCode).Msg("审计员无该租户访问权限")
			return nil, xerr.ErrUserTenantAccessDenied
		}

	default:
		// 其他角色：不能切换租户
		if currentTenantID == req.TenantID {
			// 如果是切换到当前租户，视为无操作
			log.Info().Str("user_id", userID).Str("tenant_id", currentTenantID).Msg("用户切换到当前租户，无需操作")
			return nil, xerr.Wrap(xerr.ErrInvalidParams.Code, "已在当前租户", nil)
		}

		log.Error().Str("user_id", userID).Str("current_tenant", currentTenantCode).Str("target_tenant_id", req.TenantID).Msg("当前角色无权切换租户")
		return nil, xerr.ErrUserTenantAccessDenied
	}

	// 5. 生成新 Token（UserID、userName、roles 都不变，只改变租户信息）
	tokenPair, err := s.jwt.GenerateTokenPair(ctx, targetTenant.TenantID, targetTenant.TenantCode, userID, userName, roles)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("target_tenant_code", targetTenant.TenantCode).Msg("生成JWT令牌失败")
		return nil, err
	}

	log.Info().
		Str("user_id", userID).
		Str("username", userName).
		Str("from_tenant", currentTenantCode).
		Str("to_tenant", targetTenant.TenantCode).
		Msg("用户切换租户成功")

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// GetAvailableTenants 获取用户可切换的租户列表
func (s *AuthService) GetAvailableTenants(ctx context.Context) (*dto.AvailableTenantsResponse, error) {
	var tenants []*dto.TenantInfo

	isSuperAdmin := xcontext.HasRole(ctx, constants.SuperAdmin)
	isAuditor := xcontext.HasRole(ctx, constants.Auditor)

	switch {
	case isSuperAdmin: // 超管
		// 超管：可以切换任意租户，返回所有启用的租户
		tenantList, err := s.tenantRepo.ListAll(ctx)
		if err != nil {
			log.Error().Err(err).Msg("查询所有租户失败")
			return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询所有租户失败", err)
		}

		for _, tenant := range tenantList {
			tenants = append(tenants, s.toTenantInfo(tenant))
		}
		return &dto.AvailableTenantsResponse{
			Tenants: tenants,
		}, nil

	case isAuditor: // 审计员
		// 审计员只能切换到有权限的租户
		username := xcontext.GetUserName(ctx)
		// 根据casbin 获取该用户在auditor角色下的所有租户
		// g 策略格式: g, username, role, tenantCode
		// 使用 GetFilteredGroupingPolicy 获取用户作为 auditor 角色的所有租户
		groupingPolicies, err := s.enforcer.GetFilteredGroupingPolicy(0, username, constants.Auditor)
		if err != nil {
			log.Error().Err(err).Str("username", username).Str("role", constants.Auditor).Msg("查询审计员租户列表失败")
			return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询审计员租户列表失败", err)
		}

		// 从 groupingPolicies 中提取租户编码（第三个字段）
		tenantCodes := make([]string, 0, len(groupingPolicies))
		for _, policy := range groupingPolicies {
			if len(policy) >= 3 {
				tenantCodes = append(tenantCodes, policy[2])
			}
		}

		// 批量查询租户详情
		tenantList, err := s.tenantRepo.GetByCodes(ctx, tenantCodes)
		if err != nil {
			log.Error().Err(err).Strs("tenant_codes", tenantCodes).Msg("批量查询租户信息失败")
			return nil, xerr.Wrap(xerr.ErrQueryError.Code, "批量查询租户信息失败", err)
		}

		// 按原始顺序构建结果
		for _, tenant := range tenantList {
			tenants = append(tenants, s.toTenantInfo(tenant))
		}
		return &dto.AvailableTenantsResponse{
			Tenants: tenants,
		}, nil

	default:
		tenantID := xcontext.GetTenantID(ctx)
		// 其他角色：只能使用当前租户
		tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
		if err != nil {
			log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户信息失败")
			return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询租户信息失败", err)
		}
		tenants = append(tenants, s.toTenantInfo(tenant))
		// 其他角色：返回空列表
		return &dto.AvailableTenantsResponse{
			Tenants: tenants,
		}, nil
	}
}

// toTenantInfo 转换为租户信息格式
func (s *AuthService) toTenantInfo(tenant *model.Tenant) *dto.TenantInfo {
	return &dto.TenantInfo{
		TenantID:    tenant.TenantID,
		TenantCode:  tenant.TenantCode,
		Name:        tenant.Name,
		Description: tenant.Description,
	}
}
