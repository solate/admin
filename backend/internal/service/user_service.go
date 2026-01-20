package service

import (
	"admin/internal/converter"
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/casbin"
	"admin/pkg/constants"
	"admin/pkg/idgen"
	"admin/pkg/pagination"
	"admin/pkg/passwordgen"
	"admin/pkg/rsapwd"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	userRepo        *repository.UserRepo
	userRoleService *UserRoleService
	roleRepo        *repository.RoleRepo
	tenantRepo      *repository.TenantRepo
	enforcer        *casbin.Enforcer
	recorder        *audit.Recorder
	rsaCipher       *rsapwd.RSACipher // RSA 密码解密器
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepo, userRoleService *UserRoleService, roleRepo *repository.RoleRepo, tenantRepo *repository.TenantRepo, enforcer *casbin.Enforcer, recorder *audit.Recorder, rsaCipher *rsapwd.RSACipher) *UserService {
	return &UserService{
		userRepo:        userRepo,
		userRoleService: userRoleService,
		roleRepo:        roleRepo,
		tenantRepo:      tenantRepo,
		enforcer:        enforcer,
		recorder:        recorder,
		rsaCipher:       rsaCipher,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (resp *dto.CreateUserResponse, err error) {
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
		// 使用 UserRoleService 验证并分配角色
		var err error
		roles, err = s.userRoleService.ValidateAndAssignRoles(ctx, user.UserName, req.RoleCodes, tenantID)
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
		User:     converter.ModelToUserInfoWithRoles(user, roles),
		Password: plainPassword,
		Message:  "用户创建成功",
	}, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*dto.UserInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 获取用户角色信息（使用用户自己的租户ID）
	roles, err := s.userRoleService.GetUserRoleDetails(ctx, user.UserName, user.TenantID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Str("username", user.UserName).Str("tenant_id", user.TenantID).Msg("查询用户角色失败")
		// 即使查询角色失败，也继续处理，但不返回角色信息
		roles = nil
	}

	return converter.ModelToUserInfoWithRoles(user, roles), nil
}

// GetProfile 获取当前用户档案（含角色和租户信息）
func (s *UserService) GetProfile(ctx context.Context) (*dto.ProfileResponse, error) {
	// 从 context 获取当前用户信息
	userID := xcontext.GetUserID(ctx)
	roleCodes := xcontext.GetRoles(ctx)
	tenantID := xcontext.GetTenantID(ctx)

	log.Debug().Str("user_id", userID).Strs("role_codes", roleCodes).Str("tenant_id", tenantID).Msg("[GetProfile] 开始获取用户档案")

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

	log.Debug().Strs("role_codes", roleCodes).Msg("[GetProfile] 准备查询角色详情")

	// 获取角色详情
	// 关键修改：使用 default 租户查询角色详情
	// 因为所有角色都存储在 default 租户中
	var roles []*model.Role
	if len(roleCodes) > 0 {
		// 获取 default 租户ID（使用跨租户查询）
		defaultTenants, err := s.tenantRepo.GetByCodes(ctx, []string{constants.DefaultTenantCode})
		if err != nil || len(defaultTenants) == 0 {
			log.Error().Err(err).Str("default_tenant_code", constants.DefaultTenantCode).Msg("查询 default 租户失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询 default 租户失败", err)
		}
		defaultTenant := defaultTenants[0]

		roles, err = s.roleRepo.ListByCodesWithTenant(ctx, defaultTenant.TenantID, roleCodes)
		if err != nil {
			log.Error().Err(err).Strs("role_codes", roleCodes).Str("tenant_id", defaultTenant.TenantID).Msg("查询角色详情失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色详情失败", err)
		}
	} else {
		log.Debug().Msg("[GetProfile] 角色代码为空，跳过角色查询")
		roles = []*model.Role{}
	}

	log.Debug().Int("roles_count", len(roles)).Msg("[GetProfile] 角色详情查询成功")

	return &dto.ProfileResponse{
		User:   converter.ModelToUserInfoWithRoles(user, roles),
		Tenant: converter.ModelToTenantInfo(tenant),
		Roles:  converter.ModelListToRoleInfoList(roles),
	}, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, userID string, req *dto.UpdateUserRequest) (resp *dto.UserInfo, err error) {
	var oldUser, newUser *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithError(err),
			)
		} else if newUser != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithResource(constants.ResourceTypeUser, newUser.UserID, newUser.UserName),
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

	// 角色列表为必填项，不允许为空（提前验证，避免不必要的数据库操作）
	if len(req.RoleCodes) == 0 {
		log.Warn().Str("user_id", userID).Msg("更新用户时角色列表不能为空")
		return nil, xerr.Wrap(xerr.ErrInvalidParams.Code, "角色列表不能为空", nil)
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if req.Phone != "" {
		updates["phone"] = &req.Phone
	}
	if req.Email != "" {
		updates["email"] = &req.Email
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Description != "" {
		updates["description"] = &req.Description
	}
	if req.Status != constants.StatusZero {
		updates["status"] = req.Status
	}
	if req.Remark != "" {
		updates["remark"] = &req.Remark
	}

	// 处理租户更新
	var newTenantID string
	if req.TenantID != "" && req.TenantID != oldUser.TenantID {
		// 验证目标租户是否存在
		targetTenant, err := s.tenantRepo.GetByIDManual(ctx, req.TenantID)
		if err != nil {
			log.Error().Err(err).Str("target_tenant_id", req.TenantID).Msg("目标租户不存在")
			return nil, xerr.Wrap(xerr.ErrNotFound.Code, "目标租户不存在", err)
		}
		if targetTenant == nil {
			log.Warn().Str("target_tenant_id", req.TenantID).Msg("目标租户不存在")
			return nil, xerr.ErrTenantNotFound
		}

		// 更新租户ID
		updates["tenant_id"] = req.TenantID
		newTenantID = req.TenantID

		log.Info().
			Str("user_id", userID).
			Str("old_tenant_id", oldUser.TenantID).
			Str("new_tenant_id", req.TenantID).
			Msg("用户租户更新")
	} else {
		newTenantID = oldUser.TenantID
	}

	updates["updated_at"] = time.Now().UnixMilli()

	// 更新用户
	if err := s.userRepo.Update(ctx, userID, updates); err != nil {
		// 检查是否是唯一约束冲突错误（邮箱或手机号已存在）
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate key") ||
			strings.Contains(errMsg, "uk_users_email") ||
			strings.Contains(errMsg, "uk_users_phone") {
			log.Warn().Err(err).Str("user_id", userID).Msg("更新用户时邮箱或手机号已存在")
			return nil, xerr.ErrEmailOrPhoneExists
		}

		log.Error().Err(err).Str("user_id", userID).Interface("updates", updates).Msg("更新用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新用户失败", err)
	}

	// 获取更新后的用户信息
	newUser, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("获取更新后用户信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后用户信息失败", err)
	}

	// 更新用户角色
	// 使用新的租户ID（因为可能已经更新）
	var roles []*model.Role
	roles, err = s.userRoleService.ValidateAndAssignRoles(ctx, newUser.UserName, req.RoleCodes, newTenantID)
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("user_id", userID).
		Str("username", newUser.UserName).
		Strs("role_codes", req.RoleCodes).
		Msg("更新用户角色成功")

	return converter.ModelToUserInfoWithRoles(newUser, roles), nil
}

// DeleteUser 删除用户
// 级联删除：删除用户时会自动清理该用户的角色绑定关系
func (s *UserService) DeleteUser(ctx context.Context, userID string) (err error) {
	var user *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleUser),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleUser),
				audit.WithResource(constants.ResourceTypeUser, user.UserID, user.UserName),
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

// BatchDeleteUsers 批量删除用户
func (s *UserService) BatchDeleteUsers(ctx context.Context, userIDs []string) (err error) {
	var users []*model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleUser),
				audit.WithError(err),
			)
		} else if len(users) > 0 {
			// 收集资源信息用于批量审计日志
			ids := make([]string, 0, len(users))
			userNames := make([]string, 0, len(users))
			for _, user := range users {
				ids = append(ids, user.UserID)
				userNames = append(userNames, user.UserName)
			}
			// 记录批量删除审计日志（单条日志记录所有资源）
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleUser),
				audit.WithBatchResource(constants.ResourceTypeUser, ids, userNames),
				audit.WithValue(users, nil),
			)
			log.Info().Strs("user_ids", userIDs).Int("count", len(userIDs)).Msg("批量删除用户成功")
		}
	}()

	// 获取所有用户信息
	users, err = s.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		log.Error().Err(err).Strs("user_ids", userIDs).Msg("查询用户信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户信息失败", err)
	}

	// 验证所有用户都存在
	if len(users) != len(userIDs) {
		log.Warn().Int("requested", len(userIDs)).Int("found", len(users)).Msg("部分用户不存在")
		return xerr.New(xerr.ErrNotFound.Code, "部分用户不存在")
	}

	// 批量删除用户
	if err := s.userRepo.BatchDelete(ctx, userIDs); err != nil {
		log.Error().Err(err).Strs("user_ids", userIDs).Msg("批量删除用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "批量删除用户失败", err)
	}

	// 清理所有用户的所有角色绑定关系
	for _, user := range users {
		// g 策略格式: g, username, role_code, tenant_code
		// 使用 RemoveFilteredGroupingPolicy 按 username 过滤
		s.enforcer.RemoveFilteredGroupingPolicy(0, user.UserName)
	}

	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	// 获取用户列表和总数，支持筛选条件和租户过滤
	users, total, err := s.userRepo.ListWithFiltersAndTenant(ctx, req.GetOffset(), req.GetLimit(), req.Nickname, req.Status, req.TenantID)
	if err != nil {
		log.Error().Err(err).
			Str("nickname", req.Nickname).
			Int("status", req.Status).
			Str("tenant_id", req.TenantID).
			Msg("查询用户列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户列表失败", err)
	}

	// 临时方案：如果不是超级管理员，过滤掉超级管理员用户
	filteredUsers := s.filterSuperAdminUsers(ctx, users)

	// 批量获取租户信息（收集所有租户ID）
	tenantIDs := make([]string, 0, len(filteredUsers))
	tenantMap := make(map[string]*model.Tenant)
	for _, user := range filteredUsers {
		if user.TenantID != "" {
			tenantIDs = append(tenantIDs, user.TenantID)
		}
	}

	// 使用手动模式批量查询租户信息
	if len(tenantIDs) > 0 {
		tenants, err := s.tenantRepo.GetByIDsManual(ctx, tenantIDs)
		if err != nil {
			log.Warn().Err(err).Msg("批量查询租户信息失败，将不返回租户详情")
		} else {
			for _, tenant := range tenants {
				tenantMap[tenant.TenantID] = tenant
			}
		}
	}

	// 转换为响应格式，并为每个用户填充角色和租户信息
	userInfos := make([]*dto.UserInfo, len(filteredUsers))
	for i, user := range filteredUsers {
		// 获取用户角色信息（使用用户自己的租户ID）
		roles, err := s.userRoleService.GetUserRoleDetails(ctx, user.UserName, user.TenantID)
		if err != nil {
			log.Error().Err(err).Str("user_id", user.UserID).Str("username", user.UserName).Str("tenant_id", user.TenantID).Msg("查询用户角色失败")
			// 即使查询角色失败，也继续处理，但不返回角色信息
			roles = nil
		}

		userInfo := converter.ModelToUserInfoWithRoles(user, roles)

		// 填充租户信息
		if tenant, ok := tenantMap[user.TenantID]; ok {
			userInfo.Tenant = converter.ModelToTenantInfo(tenant)
		}

		userInfos[i] = userInfo
	}

	// 如果过滤后的数据量与原始总数不同，重新计算总数（仅当非超管时）
	filteredTotal := int64(len(userInfos))
	if !xcontext.HasRole(ctx, constants.SuperAdmin) && int(total) != len(users) {
		// 如果有分页且需要准确总数，这里简化处理，使用过滤后的数量
		// 注意：这是一个临时方案，可能影响分页准确性
		filteredTotal = int64(len(userInfos))
	} else {
		filteredTotal = total
	}

	return &dto.ListUsersResponse{
		Response: pagination.NewResponse(req.Request, filteredTotal),
		List:     userInfos,
	}, nil
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(ctx context.Context, userID string, status int) (err error) {
	var oldUser, newUser *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithError(err),
			)
		} else if newUser != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithResource(constants.ResourceTypeUser, newUser.UserID, newUser.UserName),
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

// ChangePassword 用户修改自己的密码
func (s *UserService) ChangePassword(ctx context.Context, req *dto.ChangePasswordRequest) (err error) {
	var user *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithOperation("修改密码"),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithOperation("修改密码"),
				audit.WithResource(constants.ResourceTypeUser, user.UserID, user.UserName),
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

	// 解密前端传来的旧密码
	// 前端使用 JSEncrypt 库（PKCS#1 v1.5 填充）加密密码
	// 因此后端必须使用 DecryptPKCS1 方法解密
	decryptedOldPassword, err := s.rsaCipher.DecryptPKCS1(req.OldPassword)
	if err != nil {
		log.Error().Err(err).Msg("旧密码解密失败")
		return xerr.Wrap(xerr.ErrInvalidCredentials.Code, "旧密码解密失败", err)
	}

	// 使用 VerifyPassword 验证旧密码（使用解密后的 SHA256 哈希）
	if !passwordgen.VerifyPassword(decryptedOldPassword, user.Password) {
		log.Warn().Str("user_id", userID).Msg("旧密码错误")
		return xerr.New(xerr.ErrUnauthorized.Code, "原密码错误")
	}

	// 解密前端传来的新密码
	// 前端使用 JSEncrypt 库（PKCS#1 v1.5 填充）加密密码
	// 因此后端必须使用 DecryptPKCS1 方法解密
	decryptedNewPassword, err := s.rsaCipher.DecryptPKCS1(req.NewPassword)
	if err != nil {
		log.Error().Err(err).Msg("新密码解密失败")
		return xerr.Wrap(xerr.ErrInvalidCredentials.Code, "新密码解密失败", err)
	}

	// 生成新盐值并加密新密码（使用解密后的 SHA256 哈希）
	newSalt, err := passwordgen.GenerateSalt()
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("生成盐值失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
	}

	newHashedPassword, err := passwordgen.Argon2Hash(decryptedNewPassword, newSalt)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("新密码加密失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "新密码加密失败", err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(ctx, userID, newHashedPassword); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("更新密码失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新密码失败", err)
	}

	// 修改密码成功后，清除"必须修改密码"标记
	if err := s.userRepo.Update(ctx, userID, map[string]interface{}{
		"must_change_password": int16(constants.False),
	}); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("更新must_change_password失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新must_change_password失败", err)
	}

	return nil
}

// ResetPassword 超管重置用户密码（自动生成，密码只显示一次）
// 权限检查由 CasbinMiddleware 处理
func (s *UserService) ResetPassword(ctx context.Context, targetUserID string, req *dto.ResetPasswordRequest) (resp *dto.ResetPasswordResponse, err error) {
	var user *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithOperation("重置密码"),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithOperation("重置密码"),
				audit.WithResource(constants.ResourceTypeUser, user.UserID, user.UserName),
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

	// 自动生成随机密码（8位字母+数字）
	newPassword := passwordgen.GenerateRandomPassword(8)

	// 计算密码的 SHA256 哈希值（与前端登录流程保持一致）
	sha256Hash := rsapwd.HashPassword(newPassword)

	// 生成盐值并加密密码（使用 SHA256 哈希值）
	salt, err := passwordgen.GenerateSalt()
	if err != nil {
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("生成盐值失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
	}
	hashedPassword, err := passwordgen.Argon2Hash(sha256Hash, salt)
	if err != nil {
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("密码加密失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "密码加密失败", err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(ctx, targetUserID, hashedPassword); err != nil {
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("更新密码失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新密码失败", err)
	}

	// 重置密码后，用户必须修改密码
	if err := s.userRepo.Update(ctx, targetUserID, map[string]interface{}{
		"must_change_password": int16(constants.True),
	}); err != nil {
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("更新must_change_password失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新must_change_password失败", err)
	}

	// 返回响应（密码只显示这一次）
	return &dto.ResetPasswordResponse{
		Password: newPassword,
		Message:  "密码重置成功，请立即将新密码告知用户，此密码仅显示一次",
	}, nil
}

// filterSuperAdminUsers 过滤超级管理员用户（临时方案）
// 当调用者不是超级管理员时，过滤掉默认管理员用户
// 判断依据：用户名=admin 或 邮箱=admin@example.com
func (s *UserService) filterSuperAdminUsers(ctx context.Context, users []*model.User) []*model.User {
	// 如果是超级管理员，返回所有用户
	if xcontext.HasRole(ctx, constants.SuperAdmin) {
		return users
	}

	// 非超级管理员，过滤掉默认管理员
	filtered := make([]*model.User, 0, len(users))
	for _, user := range users {
		// 跳过默认管理员：用户名=admin 或 邮箱=admin@example.com
		if user.UserName == "admin" || user.Email == "admin@example.com" {
			continue
		}
		filtered = append(filtered, user)
	}

	return filtered
}
