package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/captcha"
	"admin/pkg/config"
	"admin/pkg/constants"
	"admin/pkg/jwt"
	"admin/pkg/passwordgen"
	"admin/pkg/xerr"
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// AuthService 认证服务
type AuthService struct {
	userRepo           *repository.UserRepo
	userTenantRoleRepo *repository.UserTenantRoleRepo
	roleRepo           *repository.RoleRepo
	tenantRepo         *repository.TenantRepo
	jwt                *jwt.Manager
	captcha            *captcha.Manager
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo *repository.UserRepo, userTenantRoleRepo *repository.UserTenantRoleRepo, roleRepo *repository.RoleRepo, tenantRepo *repository.TenantRepo, jwt *jwt.Manager, rdb redis.UniversalClient, config *config.Config) *AuthService {
	return &AuthService{
		userRepo:           userRepo,
		userTenantRoleRepo: userTenantRoleRepo,
		roleRepo:           roleRepo,
		tenantRepo:         tenantRepo,
		jwt:                jwt,
		captcha:            captcha.NewManager(rdb),
	}
}

// Login 用户登录
// 实现智能租户选择：
// 1. 验证账号密码
// 2. 查询用户有权限的租户列表
// 3. 如果只有一个租户，直接返回 token
// 4. 如果有多个租户，检查 last_tenant_id 是否有效
// 5. 如果 last_tenant_id 有效，返回该租户的 token
// 6. 否则返回租户列表让用户选择
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	log.Info().Str("username", req.UserName).Str("last_tenant_id", req.LastTenantID).Msg("用户登录请求")

	// 验证码校验
	if !s.captcha.Verify(req.CaptchaID, req.Captcha) {
		log.Warn().Str("captcha_id", req.CaptchaID).Msg("验证码验证失败")
		return nil, xerr.ErrCaptchaInvalid
	}
	log.Debug().Msg("验证码验证通过")

	// 查询用户（全局查询，不绑定租户）
	user, err := s.userRepo.GetByUserName(ctx, req.UserName)
	if err != nil {
		log.Error().Err(err).Str("username", req.UserName).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrUserNotFound.Code, "查询用户失败", err)
	}
	log.Info().Str("user_id", user.UserID).Str("username", user.UserName).Msg("找到用户")

	// 验证密码
	if !passwordgen.VerifyPassword(req.Password, user.Password) {
		log.Warn().Str("user_id", user.UserID).Str("username", user.UserName).Msg("密码验证失败")
		return nil, xerr.ErrInvalidCredentials
	}
	log.Debug().Msg("密码验证通过")

	// 检查用户状态
	if user.Status != constants.StatusEnabled {
		log.Warn().Str("user_id", user.UserID).Int32("status", user.Status).Msg("用户状态未启用")
		return nil, xerr.ErrUserDisabled
	}
	log.Debug().Msg("用户状态正常")

	// 查询用户关联的所有租户ID
	tenantIDs, err := s.userTenantRoleRepo.GetUserTenants(ctx, user.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("查询用户租户失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户租户失败", err)
	}
	log.Debug().Int("tenant_count", len(tenantIDs)).Msg("用户关联租户数量")

	// 用户无租户权限
	if len(tenantIDs) == 0 {
		log.Warn().Str("user_id", user.UserID).Msg("用户无关联租户")
		return nil, xerr.ErrUserNoTenants
	}

	// 获取租户详细信息
	tenantMap, err := s.tenantRepo.GetByIDs(ctx, tenantIDs)
	if err != nil {
		log.Error().Err(err).Msg("获取租户详细信息失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "获取租户详细信息失败", err)
	}

	// 构建租户信息列表
	tenantInfos := make([]dto.TenantInfo, 0, len(tenantIDs))
	for _, tenantID := range tenantIDs {
		if tenant, ok := tenantMap[tenantID]; ok && tenant.Status == constants.StatusEnabled {
			// 获取用户在该租户的角色
			roleIDs, err := s.userTenantRoleRepo.GetUserRolesInTenant(ctx, user.UserID, tenantID)
			if err != nil || len(roleIDs) == 0 {
				continue
			}

			tenantInfos = append(tenantInfos, dto.TenantInfo{
				TenantID:   tenant.TenantID,
				TenantName: tenant.Name,
				TenantCode: tenant.TenantCode,
			})
		}
	}

	// 过滤后无可用租户
	if len(tenantInfos) == 0 {
		log.Warn().Str("user_id", user.UserID).Msg("用户无可用租户")
		return nil, xerr.ErrUserNoTenants
	}

	// 智能选择租户逻辑
	var selectedTenant *dto.TenantInfo

	// 场景1：只有一个租户，直接使用
	if len(tenantInfos) == 1 {
		selectedTenant = &tenantInfos[0]
		log.Info().Str("tenant_id", selectedTenant.TenantID).Msg("自动选择唯一租户")
	} else if req.LastTenantID != "" {
		// 场景2：多个租户，检查上次选择的租户是否仍然有效
		for _, t := range tenantInfos {
			if t.TenantID == req.LastTenantID {
				selectedTenant = &t
				log.Info().Str("tenant_id", selectedTenant.TenantID).Msg("使用上次选择的租户")
				break
			}
		}
	}

	// 场景3：需要用户选择租户
	if selectedTenant == nil {
		log.Info().Int("tenant_count", len(tenantInfos)).Msg("用户有多个租户，需要选择")
		return &dto.LoginResponse{
			UserID:  user.UserID,
			Tenants: tenantInfos,
		}, nil
	}

	// 直接登录成功，生成 token
	return s.completeLogin(ctx, user, selectedTenant)
}

// completeLogin 完成登录，生成 token
func (s *AuthService) completeLogin(ctx context.Context, user *model.User, tenantInfo *dto.TenantInfo) (*dto.LoginResponse, error) {
	// 查询用户在租户中的角色
	roleIDs, err := s.userTenantRoleRepo.GetUserRolesInTenant(ctx, user.UserID, tenantInfo.TenantID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Str("tenant_id", tenantInfo.TenantID).Msg("查询用户租户角色失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户租户角色失败", err)
	}
	log.Debug().Int("role_count", len(roleIDs)).Msg("用户角色数量")

	// 检查用户是否有角色
	if len(roleIDs) == 0 {
		log.Warn().Str("user_id", user.UserID).Str("tenant_id", tenantInfo.TenantID).Msg("用户无角色")
		return nil, xerr.ErrUserNoRoles
	}

	// 查询角色详情，只获取活跃的角色
	roles, err := s.roleRepo.ListByIDs(ctx, roleIDs)
	if err != nil {
		log.Error().Err(err).Msg("查询角色失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询角色失败", err)
	}

	// 过滤出活跃的角色
	activeRoleCodes := make([]string, 0, len(roles))
	for _, role := range roles {
		if role.Status == constants.StatusEnabled {
			activeRoleCodes = append(activeRoleCodes, role.RoleCode)
		}
	}
	log.Debug().Strs("active_roles", activeRoleCodes).Msg("活跃角色列表")

	// 检查是否有活跃的角色
	if len(activeRoleCodes) == 0 {
		log.Warn().Str("user_id", user.UserID).Str("tenant_id", tenantInfo.TenantID).Msg("用户无活跃角色")
		return nil, xerr.ErrUserNoRoles
	}

	// 生成JWT令牌
	// roleType 使用 tenantInfo 中的值，默认为普通用户(1)
	tokenPair, err := s.jwt.GenerateTokenPair(ctx, tenantInfo.TenantID, tenantInfo.TenantCode, user.UserID, user.UserName, activeRoleCodes)
	if err != nil {
		log.Error().Err(err).Msg("生成JWT令牌失败")
		return nil, err
	}
	log.Info().Msg("JWT令牌生成成功")

	// 处理可选字段
	phone := ""
	if user.Phone != nil {
		phone = *user.Phone
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	log.Info().Str("user_id", user.UserID).Str("tenant_id", tenantInfo.TenantID).Msg("用户登录成功")

	return &dto.LoginResponse{
		AccessToken:   tokenPair.AccessToken,
		RefreshToken:  tokenPair.RefreshToken,
		ExpiresIn:     tokenPair.ExpiresIn,
		UserID:        user.UserID,
		CurrentTenant: tenantInfo,
		Phone:         phone,
		Email:         email,
	}, nil
}

// SelectTenant 选择租户并完成登录
func (s *AuthService) SelectTenant(ctx context.Context, userID, tenantID string) (*dto.SelectTenantResponse, error) {
	log.Info().Str("user_id", userID).Str("tenant_id", tenantID).Msg("用户选择租户")

	// 查询用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrUserNotFound.Code, "查询用户失败", err)
	}

	// 检查用户是否有该租户权限
	hasPermission, err := s.userTenantRoleRepo.CheckUserHasTenant(ctx, userID, tenantID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("tenant_id", tenantID).Msg("验证用户租户权限失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "验证用户租户权限失败", err)
	}
	if !hasPermission {
		log.Warn().Str("user_id", userID).Str("tenant_id", tenantID).Msg("用户无该租户权限")
		return nil, xerr.ErrUserTenantAccessDenied
	}

	// 获取租户信息
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("获取租户信息失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "获取租户信息失败", err)
	}

	// 获取用户在该租户的角色
	roleIDs, err := s.userTenantRoleRepo.GetUserRolesInTenant(ctx, userID, tenantID)
	if err != nil || len(roleIDs) == 0 {
		log.Warn().Str("user_id", userID).Str("tenant_id", tenantID).Msg("用户在该租户无角色")
		return nil, xerr.ErrUserNoRoles
	}

	// 获取角色详情
	roles, err := s.roleRepo.ListByIDs(ctx, roleIDs)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询角色失败", err)
	}

	// 获取活跃角色代码
	activeRoleCodes := make([]string, 0, len(roles))
	for _, role := range roles {
		if role.Status == constants.StatusEnabled {
			activeRoleCodes = append(activeRoleCodes, role.RoleCode)
		}
	}

	// 构建租户信息
	tenantInfo := &dto.TenantInfo{
		TenantID:   tenant.TenantID,
		TenantName: tenant.Name,
		TenantCode: tenant.TenantCode,
	}

	// 完成登录
	loginResp, err := s.completeLogin(ctx, user, tenantInfo)
	if err != nil {
		return nil, err
	}

	return &dto.SelectTenantResponse{
		AccessToken:   loginResp.AccessToken,
		RefreshToken:  loginResp.RefreshToken,
		ExpiresIn:     loginResp.ExpiresIn,
		CurrentTenant: tenantInfo,
	}, nil
}

// RefreshToken 刷新用户token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) {

	// 调用JWT manager刷新token
	tokenPair, err := s.jwt.VerifyRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrTokenInvalid.Code, "刷新token失败", err)
	}

	return &dto.RefreshResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}
