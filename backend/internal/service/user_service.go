package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
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
	userRepo   *repository.UserRepo
	roleRepo   *repository.RoleRepo
	tenantRepo *repository.TenantRepo
	enforcer   *casbin.Enforcer
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepo, roleRepo *repository.RoleRepo, tenantRepo *repository.TenantRepo, enforcer *casbin.Enforcer) *UserService {
	return &UserService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		tenantRepo: tenantRepo,
		enforcer:   enforcer,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	tenantID := xcontext.GetTenantID(ctx)
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
	user := &model.User{
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
func (s *UserService) UpdateUser(ctx context.Context, userID string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
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
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("获取更新后用户信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后用户信息失败", err)
	}

	return s.toUserResponse(ctx, user), nil
}

// DeleteUser 删除用户
// 级联删除：删除用户时会自动清理该用户的角色绑定关系
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, userID)
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

	log.Info().Str("user_id", userID).Str("username", user.UserName).Msg("删除用户成功")
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
func (s *UserService) UpdateUserStatus(ctx context.Context, userID string, status int) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
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

	log.Info().Str("user_id", userID).Int("status", status).Msg("更新用户状态成功")
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
