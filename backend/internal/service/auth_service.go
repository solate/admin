package service

import (
	"admin/pkg/captcha"
	"admin/pkg/config"
	"admin/pkg/jwt"

	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db      *gorm.DB
	jwt     *jwt.Manager
	captcha *captcha.Manager
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, jwt *jwt.Manager, captcha *captcha.Manager, config *config.Config) *AuthService {
	return &AuthService{
		db:      db,
		jwt:     jwt,
		captcha: captcha,
	}
}

// // Login 用户登录
// func (s *AuthService) Login(c *gin.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
// 	q := query.Use(s.db)
// 	ctx := c.Request.Context()

// 	// 验证码校验
// 	if !s.captcha.Verify(req.CaptchaID, req.Captcha) {
// 		return nil, xerr.ErrCaptchaInvalid
// 	}

// 	// 查询用户
// 	user, err := q.User.WithContext(ctx).Where(q.User.UserName.Eq(req.UserName)).First()
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, xerr.ErrUserNotFound
// 		}
// 		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户失败", err)
// 	}

// 	// 验证密码
// 	if !passwordgen.VerifyPassword(user.Password, req.Password) {
// 		return nil, xerr.ErrInvalidCredentials
// 	}

// 	// 检查用户状态
// 	if user.Status != constants.StatusEnabled {
// 		return nil, xerr.ErrUserDisabled
// 	}

// 	// 查询租户和角色
// 	query := q.UserTenantRole.WithContext(ctx)
// 	if req.TenantID != "" {
// 		query = query.Where(q.UserTenantRole.TenantID.Eq(req.TenantID))
// 	}
// 	query = query.Where(q.UserTenantRole.UserID.Eq(user.UserID))

// 	userTenantRole, err := query.Find()
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, xerr.ErrRecordNotFound
// 		}
// 		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询用户租户角色失败", err)
// 	}

// 	tenantMap := make(map[string]struct{})

// 	for _, role := range userTenantRole {
// 		tenantMap[role.TenantID] = struct{}{}
// 	}

// 	// 生成JWT令牌
// 	tokenPair, err := s.jwt.GenerateTokenPair(ctx, user.TenantID, user.TenantCode, user.UserID, user.UserName, user.RoleType, user.Roles)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &dto.LoginResponse{
// 		AccessToken:  tokenPair.AccessToken,
// 		RefreshToken: tokenPair.RefreshToken,
// 		ExpiresIn:    tokenPair.ExpiresIn,
// 		UserID:       user.UserID,
// 		TenantID:     user.TenantID,
// 		Phone:        *user.Phone,
// 		Email:        *user.Email,
// 	}, nil
// }
