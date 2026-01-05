package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/casbin"
	"admin/pkg/idgen"
	"admin/pkg/pagination"
	"admin/pkg/passwordgen"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	userRepo     *repository.UserRepo
	userRoleRepo *repository.UserRoleRepo
	roleRepo     *repository.RoleRepo
	tenantRepo   *repository.TenantRepo
	enforcer     *casbin.Enforcer
	recorder     *audit.Recorder
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepo, userRoleRepo *repository.UserRoleRepo, roleRepo *repository.RoleRepo, tenantRepo *repository.TenantRepo, enforcer *casbin.Enforcer, recorder *audit.Recorder) *UserService {
	return &UserService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		tenantRepo:   tenantRepo,
		enforcer:     enforcer,
		recorder:     recorder,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (resp *dto.UserResponse, err error) {
	var user *model.User
	tenantID := xcontext.GetTenantID(ctx)

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(audit.ModuleUser),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(audit.ModuleUser),
				audit.WithResource(audit.ResourceUser, user.UserID, user.UserName),
				audit.WithValue(nil, user),
			)
		}
	}()

	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 检查用户名是否已存在（租户内唯一）
	exists, err := s.userRepo.CheckExists(ctx, tenantID, req.UserName)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("username", req.UserName).Msg("检查用户是否存在失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查用户是否存在失败", err)
	}
	if exists {
		log.Warn().Str("tenant_id", tenantID).Str("username", req.UserName).Msg("用户名已存在")
		return nil, xerr.ErrUserExists
	}

	// 生成用户ID
	userID, err := idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成用户ID失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成用户ID失败", err)
	}

	// 生成盐值并加密密码
	salt, err := passwordgen.GenerateSalt()
	if err != nil {
		log.Error().Err(err).Msg("生成盐值失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
	}
	hashedPassword, err := passwordgen.Argon2Hash(req.Password, salt)
	if err != nil {
		log.Error().Err(err).Msg("密码加密失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "密码加密失败", err)
	}

	// 创建用户模型
	user = &model.User{
		UserID:   userID,
		TenantID: tenantID,
		UserName: req.UserName,
		Password: hashedPassword,
		Nickname: req.Nickname,
		Phone:    req.Phone,
		Email:    req.Email,
		Remark:   req.Remark,
		Status:   int16(req.Status),
	}

	// 如果没有传入昵称，使用用户名作为默认值
	if user.Nickname == "" {
		user.Nickname = req.UserName
	}

	// 设置默认状态
	if user.Status == 0 {
		user.Status = 1 // 默认正常状态
	}

	// 创建用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("tenant_id", tenantID).Str("username", req.UserName).Msg("创建用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建用户失败", err)
	}

	return s.toUserResponse(ctx, user), nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	return s.toUserResponse(ctx, user), nil
}

// GetProfile 获取当前用户档案（含角色和租户信息）
func (s *UserService) GetProfile(ctx context.Context) (*dto.ProfileResponse, error) {
	// 从 context 获取当前用户信息
	userID := xcontext.GetUserID(ctx)
	roleCodes := xcontext.GetRoles(ctx)
	tenantID := xcontext.GetTenantID(ctx)

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 获取租户信息
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户信息失败", err)
	}

	// 获取角色详情
	roles, err := s.roleRepo.ListByCodes(ctx, roleCodes)
	if err != nil {
		log.Error().Err(err).Strs("role_codes", roleCodes).Msg("查询角色详情失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色详情失败", err)
	}

	// 构建角色信息列表
	roleInfos := make([]*dto.RoleInfo, 0, len(roles))
	for _, role := range roles {
		roleInfos = append(roleInfos, &dto.RoleInfo{
			RoleID:      role.RoleID,
			RoleCode:    role.RoleCode,
			Name:        role.Name,
			Description: role.Description,
		})
	}

	return &dto.ProfileResponse{
		User: &dto.UserInfo{
			UserID:        user.UserID,
			UserName:      user.UserName,
			Nickname:      user.Nickname,
			Avatar:        user.Avatar,
			Phone:         user.Phone,
			Email:         user.Email,
			Status:        int(user.Status),
			TenantID:      tenantID,
			LastLoginTime: user.LastLoginTime,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		},
		Tenant: &dto.TenantInfo{
			TenantID:    tenant.TenantID,
			TenantCode:  tenant.TenantCode,
			Name:        tenant.Name,
			Description: tenant.Description,
		},
		Roles: roleInfos,
	}, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, userID string, req *dto.UpdateUserRequest) (resp *dto.UserResponse, err error) {
	var oldUser, newUser *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleUser),
				audit.WithError(err),
			)
		} else if newUser != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleUser),
				audit.WithResource(audit.ResourceUser, newUser.UserID, newUser.UserName),
				audit.WithValue(oldUser, newUser),
			)
		}
	}()

	// 检查用户是否存在
	oldUser, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if req.Password != "" {
		// 生成盐值并加密密码
		salt, err := passwordgen.GenerateSalt()
		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("生成盐值失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
		}
		hashedPassword, err := passwordgen.Argon2Hash(req.Password, salt)
		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("密码加密失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "密码加密失败", err)
		}
		updates["password"] = hashedPassword
	}
	if req.Phone != "" {
		updates["phone"] = &req.Phone
	}
	if req.Email != "" {
		updates["email"] = &req.Email
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Status != 0 {
		updates["status"] = req.Status
	}
	if req.Remark != "" {
		updates["remark"] = &req.Remark
	}
	updates["updated_at"] = time.Now().UnixMilli()

	// 更新用户
	if err := s.userRepo.Update(ctx, userID, updates); err != nil {
		log.Error().Err(err).Str("user_id", userID).Interface("updates", updates).Msg("更新用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新用户失败", err)
	}

	// 获取更新后的用户信息
	newUser, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("获取更新后用户信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后用户信息失败", err)
	}

	return s.toUserResponse(ctx, newUser), nil
}

// DeleteUser 删除用户
// 级联删除：删除用户时会自动清理该用户的角色绑定关系
func (s *UserService) DeleteUser(ctx context.Context, userID string) (err error) {
	var user *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(audit.ModuleUser),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(audit.ModuleUser),
				audit.WithResource(audit.ResourceUser, user.UserID, user.UserName),
				audit.WithValue(user, nil),
			)
			log.Info().Str("user_id", userID).Str("username", user.UserName).Msg("删除用户成功")
		}
	}()

	// 检查用户是否存在
	user, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 删除用户
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("username", user.UserName).Msg("删除用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "删除用户失败", err)
	}

	// 清理该用户的所有角色绑定关系
	// g 策略格式: g, username, role_code, tenant_code
	// 使用 RemoveFilteredGroupingPolicy 按 username 过滤
	s.enforcer.RemoveFilteredGroupingPolicy(0, user.UserName)

	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	// 获取用户列表和总数，支持筛选条件
	users, total, err := s.userRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.UserName, req.Status)
	if err != nil {
		log.Error().Err(err).
			Str("username", req.UserName).
			Int("status", req.Status).
			Msg("查询用户列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户列表失败", err)
	}

	// 转换为响应格式
	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(ctx, user)
	}

	return &dto.ListUsersResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     userResponses,
	}, nil
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(ctx context.Context, userID string, status int) (err error) {
	var oldUser, newUser *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleUser),
				audit.WithError(err),
			)
		} else if newUser != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleUser),
				audit.WithResource(audit.ResourceUser, newUser.UserID, newUser.UserName),
				audit.WithValue(oldUser, newUser),
			)
			log.Info().Str("user_id", userID).Int("status", status).Msg("更新用户状态成功")
		}
	}()

	// 检查用户是否存在
	oldUser, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 更新用户状态
	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		log.Error().Err(err).Str("user_id", userID).Int("status", status).Msg("更新用户状态失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新用户状态失败", err)
	}

	// 获取更新后的用户信息
	newUser, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("获取更新后用户信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后用户信息失败", err)
	}

	return nil
}

// toUserResponse 转换为用户响应格式
func (s *UserService) toUserResponse(ctx context.Context, user *model.User) *dto.UserResponse {
	return &dto.UserResponse{
		User: &dto.UserInfo{
			UserID:        user.UserID,
			UserName:      user.UserName,
			Nickname:      user.Nickname,
			Avatar:        user.Avatar,
			Phone:         user.Phone,
			Email:         user.Email,
			Status:        int(user.Status),
			TenantID:      xcontext.GetTenantID(ctx),
			LastLoginTime: user.LastLoginTime,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		},
	}
}

// AssignRoles 为用户分配角色（覆盖式）
func (s *UserService) AssignRoles(ctx context.Context, userID string, req *dto.AssignRolesRequest) (err error) {
	var user *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleRole),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleRole),
				audit.WithResource(audit.ResourceUser, user.UserID, user.UserName),
				audit.WithValue(nil, map[string]interface{}{
					"role_codes": req.RoleCodes,
				}),
			)
		}
	}()

	// 获取用户信息
	user, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	tenantCode := xcontext.GetTenantCode(ctx)

	// 验证所有角色是否存在
	roles, err := s.roleRepo.ListByCodes(ctx, req.RoleCodes)
	if err != nil {
		log.Error().Err(err).Strs("role_codes", req.RoleCodes).Msg("查询角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	if len(roles) != len(req.RoleCodes) {
		log.Warn().Strs("role_codes", req.RoleCodes).Msg("部分角色不存在")
		return xerr.Wrap(xerr.ErrInvalidParams.Code, "部分角色不存在", nil)
	}

	// 分配角色（覆盖式）
	if err := s.userRoleRepo.AssignRoles(ctx, user.UserName, req.RoleCodes, tenantCode); err != nil {
		log.Error().Err(err).Str("user_id", userID).Strs("role_codes", req.RoleCodes).Msg("分配角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "分配角色失败", err)
	}

	log.Info().
		Str("user_id", userID).
		Str("username", user.UserName).
		Strs("role_codes", req.RoleCodes).
		Msg("分配用户角色成功")

	return nil
}

// GetUserRoles 获取用户的角色列表
func (s *UserService) GetUserRoles(ctx context.Context, userID string) (*dto.UserRolesResponse, error) {
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	tenantCode := xcontext.GetTenantCode(ctx)

	// 获取用户角色编码列表
	roleCodes, err := s.userRoleRepo.GetUserRoles(ctx, user.UserName, tenantCode)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户角色失败", err)
	}

	// 获取角色详情
	roles, err := s.roleRepo.ListByCodes(ctx, roleCodes)
	if err != nil {
		log.Error().Err(err).Strs("role_codes", roleCodes).Msg("查询角色详情失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色详情失败", err)
	}

	// 构建角色信息列表
	roleInfos := make([]*dto.RoleInfo, 0, len(roles))
	for _, role := range roles {
		roleInfos = append(roleInfos, &dto.RoleInfo{
			RoleID:      role.RoleID,
			RoleCode:    role.RoleCode,
			Name:        role.Name,
			Description: role.Description,
		})
	}

	return &dto.UserRolesResponse{
		UserID:   user.UserID,
		UserName: user.UserName,
		Roles:    roleInfos,
	}, nil
}

// ChangePassword 用户修改自己的密码
func (s *UserService) ChangePassword(ctx context.Context, req *dto.ChangePasswordRequest) (err error) {
	var user *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleUser),
				audit.WithOperation("修改密码"),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleUser),
				audit.WithOperation("修改密码"),
				audit.WithResource(audit.ResourceUser, user.UserID, user.UserName),
				audit.WithValue(nil, user),
			)
			log.Info().Str("user_id", user.UserID).Msg("用户修改密码成功")
		}
	}()

	// 获取当前用户ID
	userID := xcontext.GetUserID(ctx)

	// 查询用户信息
	user, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 使用 VerifyPassword 验证旧密码
	if !passwordgen.VerifyPassword(req.OldPassword, user.Password) {
		log.Warn().Str("user_id", userID).Msg("旧密码错误")
		return xerr.New(xerr.ErrUnauthorized.Code, "原密码错误")
	}

	// 生成新盐值并加密新密码
	newSalt, err := passwordgen.GenerateSalt()
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("生成盐值失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
	}

	newHashedPassword, err := passwordgen.Argon2Hash(req.NewPassword, newSalt)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("新密码加密失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "新密码加密失败", err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(ctx, userID, newHashedPassword); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("更新密码失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新密码失败", err)
	}

	return nil
}

// ResetPassword 超管重置用户密码（密码只显示一次）
// 权限检查由 CasbinMiddleware 处理
func (s *UserService) ResetPassword(ctx context.Context, targetUserID string, req *dto.ResetPasswordRequest) (resp *dto.ResetPasswordResponse, err error) {
	var user *model.User
	var newPassword string

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleUser),
				audit.WithOperation("重置密码"),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleUser),
				audit.WithOperation("重置密码"),
				audit.WithResource(audit.ResourceUser, user.UserID, user.UserName),
				audit.WithValue(nil, user),
			)
			log.Info().Str("operator_id", xcontext.GetUserID(ctx)).Str("target_user_id", targetUserID).Msg("重置用户密码成功")
		}
	}()

	// 查询目标用户
	user, err = s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("target_user_id", targetUserID).Msg("目标用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 确定新密码：如果未提供则自动生成
	if req.Password != "" {
		newPassword = req.Password
	} else {
		// 自动生成随机密码（8位字母+数字）
		newPassword = passwordgen.GenerateRandomPassword(8)
	}

	// 生成新盐值并加密密码
	salt, err := passwordgen.GenerateSalt()
	if err != nil {
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("生成盐值失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
	}

	hashedPassword, err := passwordgen.Argon2Hash(newPassword, salt)
	if err != nil {
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("密码加密失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "密码加密失败", err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(ctx, targetUserID, hashedPassword); err != nil {
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("更新密码失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新密码失败", err)
	}

	// 返回响应（密码只显示这一次）
	return &dto.ResetPasswordResponse{
		Password:      newPassword,
		AutoGenerated: req.Password == "",
		Message:       "密码重置成功，请立即将新密码告知用户，此密码仅显示一次",
	}, nil
}
