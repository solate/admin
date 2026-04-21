package service

import (
	"admin/internal/converter"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/captcha"
	"admin/pkg/config"
	"admin/pkg/constants"
	"admin/pkg/jwt"
	"admin/pkg/passwordgen"
	"admin/pkg/rsapwd"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"net/http"
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
	recorder     *audit.Recorder
	rsaCipher    *rsapwd.RSACipher
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo *repository.UserRepo, userRoleRepo *repository.UserRoleRepo, roleRepo *repository.RoleRepo, tenantRepo *repository.TenantRepo, jwtManager *jwt.Manager, rdb redis.UniversalClient, recorder *audit.Recorder, config *config.Config, rsaCipher *rsapwd.RSACipher) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		tenantRepo:   tenantRepo,
		jwt:          jwtManager,
		captcha:      captcha.NewManager(rdb),
		recorder:     recorder,
		rsaCipher:    rsaCipher,
	}
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, r *http.Request, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 验证码校验
	if !s.captcha.Verify(req.CaptchaID, req.Captcha) {
		return nil, xerr.ErrCaptchaInvalid
	}

	// 解密前端传来的加密密码
	decryptedPassword, err := s.rsaCipher.DecryptPKCS1(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("密码解密失败")
		return nil, xerr.Wrap(xerr.ErrInvalidCredentials.Code, "密码解密失败", err)
	}

	// 查询用户（通过邮箱全局唯一）
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Err(err).Str("email", req.Email).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("email", req.Email).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 验证密码
	if !passwordgen.VerifyPassword(decryptedPassword, user.Password) {
		return nil, xerr.ErrInvalidCredentials
	}

	// 检查用户状态
	if user.Status != constants.StatusEnabled {
		return nil, xerr.ErrUserDisabled
	}

	// 查询用户所属租户信息
	tenant, err := s.tenantRepo.GetByIDManual(ctx, user.TenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Err(err).Str("user_id", user.UserID).Str("tenant_id", user.TenantID).Msg("用户所属租户不存在")
			return nil, xerr.ErrTenantNotFound
		}
		log.Error().Err(err).Str("tenant_id", user.TenantID).Msg("查询租户信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户信息失败", err)
	}

	// 检查租户状态
	if tenant.Status != constants.StatusEnabled {
		log.Error().Str("tenant_id", tenant.TenantID).Msg("用户所属租户已禁用")
		return nil, xerr.ErrTenantDisabled
	}

	// 获取用户角色ID列表（从 user_roles 表）
	roleIDs, err := s.userRoleRepo.GetUserRoleIDs(ctx, user.UserID, user.TenantID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("查询用户角色失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户角色失败", err)
	}

	if len(roleIDs) == 0 {
		return nil, xerr.ErrUserNoRoles
	}

	// 获取角色详情（用于提取角色编码）
	roles, err := s.roleRepo.GetByIDs(ctx, roleIDs)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("查询角色详情失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询角色详情失败", err)
	}

	roleCodes := make([]string, len(roles))
	for i, role := range roles {
		roleCodes[i] = role.RoleCode
	}

	// 生成JWT令牌（包含角色编码和角色ID）
	tokenPair, err := s.jwt.GenerateTokenPair(ctx, tenant.TenantID, tenant.TenantCode, user.UserID, user.UserName, roleCodes, roleIDs)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("生成JWT令牌失败")
		return nil, err
	}

	// 更新最后登录时间
	if err := s.userRepo.UpdateManual(ctx, user.UserID, map[string]interface{}{
		"last_login_time": time.Now().UnixMilli(),
	}); err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("更新最后登录时间失败")
	}

	// 记录登录日志
	s.recorder.LoginEmail(ctx, tenant.TenantID, user.UserID, user.UserName, nil)

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// LoginByPhone 手机号登录
func (s *AuthService) LoginByPhone(ctx context.Context, req *dto.PhoneLoginRequest) (*dto.LoginResponse, error) {
	// 解密前端传来的加密密码
	decryptedPassword, err := s.rsaCipher.DecryptPKCS1(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("密码解密失败")
		return nil, xerr.Wrap(xerr.ErrInvalidCredentials.Code, "密码解密失败", err)
	}

	// 查询用户（通过手机号全局唯一）
	user, err := s.userRepo.GetByPhone(ctx, req.Phone)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Err(err).Str("phone", req.Phone).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("phone", req.Phone).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 验证密码
	if !passwordgen.VerifyPassword(decryptedPassword, user.Password) {
		return nil, xerr.ErrInvalidCredentials
	}

	// 检查用户状态
	if user.Status != constants.StatusEnabled {
		return nil, xerr.ErrUserDisabled
	}

	// 查询用户所属租户信息
	tenant, err := s.tenantRepo.GetByIDManual(ctx, user.TenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Err(err).Str("user_id", user.UserID).Str("tenant_id", user.TenantID).Msg("用户所属租户不存在")
			return nil, xerr.ErrTenantNotFound
		}
		log.Error().Err(err).Str("tenant_id", user.TenantID).Msg("查询租户信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户信息失败", err)
	}

	// 检查租户状态
	if tenant.Status != constants.StatusEnabled {
		log.Error().Str("tenant_id", tenant.TenantID).Msg("用户所属租户已禁用")
		return nil, xerr.ErrTenantDisabled
	}

	// 获取用户角色
	roleIDs, err := s.userRoleRepo.GetUserRoleIDs(ctx, user.UserID, user.TenantID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("查询用户角色失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户角色失败", err)
	}

	if len(roleIDs) == 0 {
		return nil, xerr.ErrUserNoRoles
	}

	roles, err := s.roleRepo.GetByIDs(ctx, roleIDs)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("查询角色详情失败")
		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询角色详情失败", err)
	}

	roleCodes := make([]string, len(roles))
	for i, role := range roles {
		roleCodes[i] = role.RoleCode
	}

	tokenPair, err := s.jwt.GenerateTokenPair(ctx, tenant.TenantID, tenant.TenantCode, user.UserID, user.UserName, roleCodes, roleIDs)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("生成JWT令牌失败")
		return nil, err
	}

	if err := s.userRepo.UpdateManual(ctx, user.UserID, map[string]interface{}{
		"last_login_time": time.Now().UnixMilli(),
	}); err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("更新最后登录时间失败")
	}

	s.recorder.LoginPhone(ctx, tenant.TenantID, user.UserID, user.UserName, nil)

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// RefreshToken 刷新用户token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) {
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
func (s *AuthService) Logout(ctx context.Context, r *http.Request) error {
	tokenID := xcontext.GetTokenID(ctx)
	if tokenID == "" {
		return xerr.ErrUnauthorized
	}

	if err := s.jwt.RevokeToken(ctx, tokenID); err != nil {
		log.Error().Err(err).Str("token_id", tokenID).Msg("撤销token失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "撤销token失败", err)
	}

	s.recorder.Logout(ctx)

	return nil
}

// SwitchTenant 切换租户
func (s *AuthService) SwitchTenant(ctx context.Context, req *dto.SwitchTenantRequest) (*dto.LoginResponse, error) {
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
	isSuperAdmin := xcontext.HasRole(ctx, constants.SuperAdmin)

	if isSuperAdmin {
		tenantList, err := s.tenantRepo.ListAll(ctx)
		if err != nil {
			log.Error().Err(err).Msg("查询所有租户失败")
			return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询所有租户失败", err)
		}

		tenants := make([]*dto.TenantInfo, len(tenantList))
		for i, tenant := range tenantList {
			tenants[i] = converter.ModelToTenantInfo(tenant)
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
		Tenants: []*dto.TenantInfo{converter.ModelToTenantInfo(tenant)},
	}, nil
}
