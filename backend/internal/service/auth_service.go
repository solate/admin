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
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	log.Info().Str("username", req.UserName).Str("tenant_id", req.TenantID).Msg("用户登录请求")

	// 验证码校验
	if !s.captcha.Verify(req.CaptchaID, req.Captcha) {
		log.Warn().Str("captcha_id", req.CaptchaID).Msg("验证码验证失败")
		return nil, xerr.ErrCaptchaInvalid
	}
	log.Debug().Msg("验证码验证通过")

	// 查询用户
	// 注意：登录接口使用了 SkipTenantCheck 中间件，所以可以跨租户查询用户
	var user *model.User
	var err error
	if req.TenantID != "" {
		log.Debug().Str("username", req.UserName).Str("tenant_id", req.TenantID).Msg("使用租户ID查询用户")
		user, err = s.userRepo.GetUserByNameAndTenant(ctx, req.UserName, req.TenantID)
	} else {
		log.Debug().Str("username", req.UserName).Msg("全局查询用户")
		user, err = s.userRepo.GetUserByName(ctx, req.UserName)
	}
	if err != nil {
		log.Error().Err(err).Str("username", req.UserName).Str("tenant_id", req.TenantID).Msg("查询用户失败")
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

	var tenantID string
	if req.TenantID != "" { // 如果给了租户ID，则验证用户是否有该租户的权限
		log.Debug().Str("tenant_id", req.TenantID).Msg("验证用户租户权限")
		hasPermission, err := s.userTenantRoleRepo.CheckUserHasTenantRole(ctx, user.UserID, req.TenantID)
		if err != nil {
			log.Error().Err(err).Str("user_id", user.UserID).Str("tenant_id", req.TenantID).Msg("验证用户租户权限失败")
			return nil, xerr.Wrap(xerr.ErrQueryError.Code, "验证用户租户权限失败", err)
		}
		if !hasPermission {
			log.Warn().Str("user_id", user.UserID).Str("tenant_id", req.TenantID).Msg("用户无该租户权限")
			return nil, xerr.ErrUserTenantAccessDenied
		}
		tenantID = req.TenantID
		log.Info().Str("tenant_id", tenantID).Msg("用户租户权限验证通过")
	} else { // 如果没有给租户ID，则查询用户关联的租户
		log.Debug().Msg("查询用户关联的租户")
		tenants, err := s.userTenantRoleRepo.GetUserTenants(ctx, user.UserID)
		if err != nil {
			log.Error().Err(err).Str("user_id", user.UserID).Msg("查询用户租户失败")
			return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户租户失败", err)
		}
		log.Debug().Int("tenant_count", len(tenants)).Msg("用户关联租户数量")
		if len(tenants) == 0 {
			log.Warn().Str("user_id", user.UserID).Msg("用户无关联租户")
			return nil, xerr.ErrUserNoTenants
		} else if len(tenants) == 1 {
			tenantID = tenants[0]
			log.Info().Str("tenant_id", tenantID).Msg("自动选择唯一租户")
		} else { // 如果用户关联多个租户，则返回错误
			log.Warn().Str("user_id", user.UserID).Int("count", len(tenants)).Msg("用户关联多个租户")
			return nil, xerr.ErrUserHasMultipleTenants
		}
	}

	// 查询用户在租户中的角色
	roleIDs, err := s.userTenantRoleRepo.GetUserRolesInTenant(ctx, user.UserID, tenantID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Str("tenant_id", tenantID).Msg("查询用户租户角色失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户租户角色失败", err)
	}
	log.Debug().Int("role_count", len(roleIDs)).Msg("用户角色数量")

	// 检查用户是否有角色
	if len(roleIDs) == 0 {
		log.Warn().Str("user_id", user.UserID).Str("tenant_id", tenantID).Msg("用户无角色")
		return nil, xerr.ErrUserNoRoles
	}

	// 查询角色详情，只获取活跃的角色
	roles, err := s.roleRepo.ListRolesByIDs(ctx, roleIDs)
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
		log.Warn().Str("user_id", user.UserID).Str("tenant_id", tenantID).Msg("用户无活跃角色")
		return nil, xerr.ErrUserNoRoles
	}

	// 获取租户编码
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户编码失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询租户编码失败", err)
	}
	log.Debug().Str("tenant_code", tenant.TenantCode).Msg("获取租户编码")

	// 生成JWT令牌
	tokenPair, err := s.jwt.GenerateTokenPair(ctx, tenantID, tenant.TenantCode, user.UserID, user.UserName, user.RoleType, activeRoleCodes)
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

	log.Info().Str("user_id", user.UserID).Str("tenant_id", tenantID).Msg("用户登录成功")

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		UserID:       user.UserID,
		TenantID:     tenantID,
		Phone:        phone,
		Email:        email,
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
