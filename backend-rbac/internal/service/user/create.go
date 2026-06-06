package user

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/idgen"
	"admin/pkg/utils/passwordgen"
	"admin/pkg/utils/rsapwd"
	"admin/pkg/xerr"
	"context"
	"strings"

	"github.com/rs/zerolog/log"
)

// CreateUser 创建用户
func (s *Service) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (resp *dto.CreateUserResponse, err error) {
	var user *model.User
	var plainPassword string

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleUser),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleUser),
				audit.WithResource(constants.ResourceTypeUser, user.UserID, user.UserName),
				audit.WithValue(nil, user),
			)
		}
	}()

	// 租户ID 要传入，创建不同租户用户
	tenantID := req.TenantID
	if tenantID == "" {
		return nil, xerr.ErrInvalidParams
	}

	// 生成用户ID
	userID, err := idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成用户ID失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成用户ID失败", err)
	}

	// 自动生成随机密码（8位字母+数字）
	plainPassword = passwordgen.GenerateRandomPassword(8)

	// 计算密码的 SHA256 哈希值（与前端登录流程保持一致）
	sha256Hash := rsapwd.HashPassword(plainPassword)

	// 生成盐值并加密密码（使用 SHA256 哈希值）
	salt, err := passwordgen.GenerateSalt()
	if err != nil {
		log.Error().Err(err).Msg("生成盐值失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
	}
	hashedPassword, err := passwordgen.Argon2Hash(sha256Hash, salt)
	if err != nil {
		log.Error().Err(err).Msg("密码加密失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "密码加密失败", err)
	}

	// 确定用户名：如果未提供，使用邮箱作为默认值
	userName := req.UserName
	if userName == "" {
		userName = req.Email
	}

	// 创建用户模型
	user = &model.User{
		UserID:             userID,
		TenantID:           tenantID,
		UserName:           userName,
		Password:           hashedPassword,
		Nickname:           req.Nickname,
		Phone:              req.Phone,
		Email:              req.Email,
		Description:        req.Description,
		Remark:             req.Remark,
		Status:             int16(req.Status),
		MustChangePassword: constants.True, // 新用户必须修改密码
	}

	// 如果没有传入昵称，使用邮箱作为默认值
	if user.Nickname == "" {
		user.Nickname = req.Email
	}

	// 设置默认状态
	if user.Status == int16(constants.StatusZero) {
		user.Status = int16(constants.StatusEnabled) // 默认正常状态
	}

	// 创建用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		// 检查是否是唯一约束冲突错误（邮箱或手机号已存在）
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate key") ||
			strings.Contains(errMsg, "uk_users_email") ||
			strings.Contains(errMsg, "uk_users_phone") {
			log.Warn().Err(err).Str("email", req.Email).Str("phone", req.Phone).Msg("邮箱或手机号已存在")
			return nil, xerr.ErrEmailOrPhoneExists
		}

		log.Error().Err(err).Str("user_id", userID).Str("tenant_id", tenantID).Str("username", userName).Msg("创建用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建用户失败", err)
	}

	// 如果传入了角色列表，则为用户分配角色
	var roles []*model.Role
	if len(req.RoleCodes) > 0 {
		// 使用 RoleService 验证并分配角色
		var err error
		roles, err = s.userRoleService.ValidateAndAssignRoles(ctx, user.UserID, req.RoleCodes, tenantID)
		if err != nil {
			return nil, err
		}

		log.Info().
			Str("user_id", userID).
			Str("username", user.UserName).
			Strs("role_codes", req.RoleCodes).
			Msg("创建用户并分配角色成功")
	}

	return &dto.CreateUserResponse{
		User:     modelToUserInfoWithRoles(user, roles),
		Password: plainPassword,
		Message:  "用户创建成功",
	}, nil
}
