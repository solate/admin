package service

import (
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
)

// AuthService 认证服务
type AuthService struct {
	userRepo           *repository.UserRepo
	userTenantRoleRepo *repository.UserTenantRoleRepo
	roleRepo           *repository.RoleRepo
	jwt                *jwt.Manager
	captcha            *captcha.Manager
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo *repository.UserRepo, userTenantRoleRepo *repository.UserTenantRoleRepo, roleRepo *repository.RoleRepo, jwt *jwt.Manager, rdb redis.UniversalClient, config *config.Config) *AuthService {
	return &AuthService{
		userRepo:           userRepo,
		userTenantRoleRepo: userTenantRoleRepo,
		roleRepo:           roleRepo,
		jwt:                jwt,
		captcha:            captcha.NewManager(rdb),
	}
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {

	// 验证码校验
	if !s.captcha.Verify(req.CaptchaID, req.Captcha) {
		return nil, xerr.ErrCaptchaInvalid
	}

	// 查询用户
	user, err := s.userRepo.GetUserByName(ctx, req.UserName)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrUserNotFound.Code, "查询用户失败", err)
	}

	// 验证密码
	if !passwordgen.VerifyPassword(user.Password, req.Password) {
		return nil, xerr.ErrInvalidCredentials
	}

	// 检查用户状态
	if user.Status != constants.StatusEnabled {
		return nil, xerr.ErrUserDisabled
	}

	var tenantID string
	if req.TenantID != "" { // 如果给了租户ID，则验证用户是否有该租户的权限
		hasPermission, err := s.userTenantRoleRepo.CheckUserHasTenantRole(ctx, user.UserID, req.TenantID)
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrQueryError.Code, "验证用户租户权限失败", err)
		}
		if !hasPermission {
			return nil, xerr.ErrUserTenantAccessDenied
		}
		tenantID = req.TenantID
	} else { // 如果没有给租户ID，则查询用户关联的租户
		tenants, err := s.userTenantRoleRepo.GetUserTenants(ctx, user.UserID)
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户租户失败", err)
		}
		if len(tenants) == 0 {
			return nil, xerr.ErrUserNoTenants
		} else if len(tenants) == 1 {
			tenantID = tenants[0]
		} else { // 如果用户关联多个租户，则返回错误
			return nil, xerr.ErrUserHasMultipleTenants
		}
	}

	// 查询用户在租户中的角色
	roleIDs, err := s.userTenantRoleRepo.GetUserRolesInTenant(ctx, user.UserID, tenantID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户租户角色失败", err)
	}

	// 检查用户是否有角色
	if len(roleIDs) == 0 {
		return nil, xerr.ErrUserNoRoles
	}

	// 查询角色详情，只获取活跃的角色
	roles, err := s.roleRepo.ListRolesByIDs(ctx, roleIDs)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询角色失败", err)
	}

	// 过滤出活跃的角色
	activeRoleCodes := make([]string, 0, len(roles))
	for _, role := range roles {
		if role.Status == constants.StatusEnabled {
			activeRoleCodes = append(activeRoleCodes, role.RoleCode)
		}
	}

	// 检查是否有活跃的角色
	if len(activeRoleCodes) == 0 {
		return nil, xerr.ErrUserNoRoles
	}

	// 生成JWT令牌 - 注意：这里需要根据实际的JWT方法调整参数
	tokenPair, err := s.jwt.GenerateTokenPair(ctx, tenantID, "", user.UserID, user.UserName, user.RoleType, activeRoleCodes)
	if err != nil {
		return nil, err
	}

	// 处理可选字段
	phone := ""
	if user.Phone != nil {
		phone = *user.Phone
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

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
