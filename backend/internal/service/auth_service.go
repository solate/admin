package service

import (
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/captcha"
	"admin/pkg/cache"
	"admin/pkg/config"
	"admin/pkg/constants"
	"admin/pkg/jwt"
	"admin/pkg/passwordgen"
	"admin/pkg/xerr"
	"context"

	"github.com/redis/go-redis/v9"
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
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo *repository.UserRepo, userRoleRepo *repository.UserRoleRepo, roleRepo *repository.RoleRepo, tenantRepo *repository.TenantRepo, jwt *jwt.Manager, rdb redis.UniversalClient, config *config.Config) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		tenantRepo:   tenantRepo,
		jwt:          jwt,
		captcha:      captcha.NewManager(rdb),
	}
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {

	// 验证码校验
	if !s.captcha.Verify(req.CaptchaID, req.Captcha) {
		return nil, xerr.ErrCaptchaInvalid
	}

	// 查询用户（用户名全局唯一）
	user, err := s.userRepo.GetByUserName(ctx, req.UserName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrUserNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
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
	var tenantCode string

	if req.LastTenantID != "" {
		// 客户端指定了上次登录的租户ID，验证租户是否存在
		tenant, err := s.tenantRepo.GetByID(ctx, req.LastTenantID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, xerr.ErrNotFound
			}
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户失败", err)
		}
		tenantID = tenant.TenantID
		tenantCode = tenant.TenantCode
	} else {
		// 客户端未指定租户，使用默认租户
		tenantID = cache.Get().Tenant.GetDefaultTenantID()
		tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取默认租户失败", err)
		}
		tenantCode = tenant.TenantCode
	}

	// 获取用户在租户中的角色（从 Casbin）
	roleCodes, err := s.userRoleRepo.GetUserRoles(ctx, user.UserName, tenantCode)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户角色失败", err)
	}

	// 检查用户是否有角色
	if len(roleCodes) == 0 {
		return nil, xerr.ErrUserNoRoles
	}

	// 查询角色详情，只获取活跃的角色
	roles, err := s.roleRepo.ListByIDs(ctx, roleCodes)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询角色详情失败", err)
	}

	// 过滤出活跃的角色编码
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

	// 生成JWT令牌
	tokenPair, err := s.jwt.GenerateTokenPair(ctx, tenantID, tenantCode, user.UserID, user.UserName, activeRoleCodes)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		UserID:       user.UserID,
		CurrentTenant: &dto.TenantInfo{
			TenantID:   tenantID,
			TenantCode: tenantCode,
		},
		Phone: user.Phone,
		Email: user.Email,
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
