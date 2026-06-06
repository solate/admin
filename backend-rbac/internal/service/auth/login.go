package auth

import (
	"admin/internal/dto"
	"admin/pkg/utils/captcha"
	"admin/pkg/constants"
	"admin/pkg/utils/passwordgen"
	"admin/pkg/xerr"
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Login 用户登录
func (s *Service) Login(ctx context.Context, r *http.Request, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 验证码校验
	captchaMgr := captcha.NewManager(s.rdb)
	if !captchaMgr.Verify(req.CaptchaID, req.Captcha) {
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
func (s *Service) LoginByPhone(ctx context.Context, req *dto.PhoneLoginRequest) (*dto.LoginResponse, error) {
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
