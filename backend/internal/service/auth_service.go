package service

import (
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/captcha"
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
			log.Error().Err(err).Str("username", req.UserName).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("username", req.UserName).Msg("查询用户失败")
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

	tenantID := xcontext.GetTenantID(ctx)
	tenantCode := xcontext.GetTenantCode(ctx)

	if tenantCode == "" {
		log.Error().Str("username", user.UserName).Msg("租户编码不能为空")
		return nil, xerr.ErrTenantCodeRequired
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

	// 查询角色详情，只获取活跃的角色
	roles, err := s.roleRepo.ListByIDs(ctx, roleCodes)
	if err != nil {
		log.Error().Err(err).Strs("role_codes", roleCodes).Msg("查询角色详情失败")
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
		log.Error().Err(err).Str("user_id", user.UserID).Str("username", user.UserName).Msg("生成JWT令牌失败")
		return nil, err
	}

	// 更新最后登录时间
	now := time.Now()
	if err := s.userRepo.Update(ctx, user.UserID, map[string]interface{}{
		"last_login_time": now,
	}); err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("更新最后登录时间失败")
		// 不影响登录流程，继续返回
	}

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User: &dto.User{
			UserID:        user.UserID,
			UserName:      user.UserName,
			Nickname:      user.Nickname,
			Phone:         user.Phone,
			Email:         user.Email,
			Status:        int(user.Status),
			TenantID:      tenantID,
			LastLoginTime: now.UnixMilli(),
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		},
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
