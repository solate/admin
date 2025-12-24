package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/constants"
	"admin/pkg/idgen"
	"admin/pkg/operationlog"
	"admin/pkg/pagination"
	"admin/pkg/passwordgen"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"gorm.io/gorm"
)

// TenantMemberService 租户成员服务
type TenantMemberService struct {
	userRepo           *repository.UserRepo
	roleRepo           *repository.RoleRepo
	userTenantRoleRepo *repository.UserTenantRoleRepo
}

// NewTenantMemberService 创建租户成员服务
func NewTenantMemberService(
	userRepo *repository.UserRepo,
	roleRepo *repository.RoleRepo,
	userTenantRoleRepo *repository.UserTenantRoleRepo,
) *TenantMemberService {
	return &TenantMemberService{
		userRepo:           userRepo,
		roleRepo:           roleRepo,
		userTenantRoleRepo: userTenantRoleRepo,
	}
}

// AddTenantMember 添加租户成员
// 1. 如果用户不存在，则创建新用户并分配角色
// 2. 如果用户已存在，检查是否已在当前租户中
// 3. 为用户分配指定角色
func (s *TenantMemberService) AddTenantMember(ctx context.Context, req *dto.AddTenantMemberRequest) (*dto.AddTenantMemberResponse, error) {
	// 获取租户ID
	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 验证所有角色ID都属于当前租户
	roles, err := s.roleRepo.ListByIDs(ctx, req.RoleIDs)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}
	if len(roles) != len(req.RoleIDs) {
		return nil, xerr.New(xerr.ErrInvalidParams.Code, "部分角色不存在或不属于当前租户")
	}
	for _, role := range roles {
		if role.TenantID != tenantID {
			return nil, xerr.New(xerr.ErrInvalidParams.Code, "角色不属于当前租户")
		}
	}

	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByUserName(ctx, req.UserName)
	isUserExists := err == nil
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查用户是否存在失败", err)
	}

	var userID string
	var initialPassword string

	if isUserExists {
		// 用户已存在
		userID = existingUser.UserID

		// 检查用户是否已在当前租户中
		hasTenant, err := s.userTenantRoleRepo.CheckUserHasTenant(ctx, userID, tenantID)
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查用户租户关系失败", err)
		}
		if hasTenant {
			return nil, xerr.New(xerr.ErrConflict.Code, "用户已在该租户中")
		}
	} else {
		// 创建新用户
		// 生成用户ID
		userID, err = idgen.GenerateUUID()
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成用户ID失败", err)
		}

		// 生成随机初始密码（16位）
		initialPassword, err = passwordgen.GeneratePassword(16)
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成密码失败", err)
		}

		// 生成盐值并加密密码
		salt, err := passwordgen.GenerateSalt()
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
		}
		hashedPassword, err := passwordgen.Argon2Hash(initialPassword, salt)
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "密码加密失败", err)
		}

		// 创建用户模型
		user := &model.User{
			UserID:   userID,
			UserName: req.UserName,
			Password: hashedPassword,
			Name:     req.Name,
			Status:   1, // 默认启用
		}

		// 设置可选字段
		if req.Phone != "" {
			user.Phone = &req.Phone
		}
		if req.Email != "" {
			user.Email = &req.Email
		}

		// 创建用户
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建用户失败", err)
		}

		// 记录操作日志
		ctx = operationlog.RecordCreate(ctx, constants.ModuleUser, constants.ResourceTypeUser, user.UserID, user.Name, user)
	}

	// 为用户分配角色到当前租户
	for _, roleID := range req.RoleIDs {
		// 检查该关系是否已存在
		exists, err := s.userTenantRoleRepo.CheckExists(ctx, userID, tenantID, roleID)
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查用户角色关系失败", err)
		}
		if exists {
			continue // 跳过已存在的关系
		}

		// 生成关联ID
		userTenantRoleID, err := idgen.GenerateUUID()
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成关联ID失败", err)
		}

		// 创建用户租户角色关系
		userTenantRole := &model.UserTenantRole{
			UserTenantRoleID: userTenantRoleID,
			UserID:           userID,
			TenantID:         tenantID,
			RoleID:           roleID,
		}

		if err := s.userTenantRoleRepo.Create(ctx, userTenantRole); err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建用户租户角色关系失败", err)
		}
	}

	// 获取创建的用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取用户信息失败", err)
	}

	// 如果是新创建的用户，记录添加成员操作日志
	if !isUserExists {
		operationlog.RecordCreate(ctx, constants.ModuleTenantMember, constants.ResourceTypeTenantMember, userID, user.Name,
			map[string]interface{}{
				"user_id":   userID,
				"tenant_id": tenantID,
				"role_ids":  req.RoleIDs,
			})
	}

	return &dto.AddTenantMemberResponse{
		UserID:          user.UserID,
		UserName:        user.UserName,
		Name:            user.Name,
		InitialPassword: initialPassword, // 仅在创建新用户时有值
		TenantID:        tenantID,
		RoleIDs:         req.RoleIDs,
	}, nil
}

// RemoveTenantMember 移除租户成员
// 从租户中移除用户（删除用户在该租户下的所有角色关系）
func (s *TenantMemberService) RemoveTenantMember(ctx context.Context, userID string) error {
	// 获取租户ID
	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return xerr.ErrUnauthorized
	}

	// 检查用户是否在当前租户中
	hasTenant, err := s.userTenantRoleRepo.CheckUserHasTenant(ctx, userID, tenantID)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "检查用户租户关系失败", err)
	}
	if !hasTenant {
		return xerr.New(xerr.ErrNotFound.Code, "用户不在该租户中")
	}

	// 获取用户信息用于日志
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrUserNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "获取用户信息失败", err)
	}

	// 删除用户在当前租户下的所有角色关系
	if err := s.userTenantRoleRepo.DeleteByUserTenant(ctx, userID, tenantID); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "移除租户成员失败", err)
	}

	// 记录操作日志
	operationlog.RecordDelete(ctx, constants.ModuleTenantMember, constants.ResourceTypeTenantMember, userID, user.Name,
		map[string]interface{}{
			"user_id":   userID,
			"tenant_id": tenantID,
		})

	return nil
}

// UpdateMemberRoles 更新成员角色
// 替换用户在当前租户下的所有角色
func (s *TenantMemberService) UpdateMemberRoles(ctx context.Context, userID string, roleIDs []string) (*dto.UpdateMemberRolesResponse, error) {
	// 获取租户ID
	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 验证所有角色ID都属于当前租户
	roles, err := s.roleRepo.ListByIDs(ctx, roleIDs)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}
	if len(roles) != len(roleIDs) {
		return nil, xerr.New(xerr.ErrInvalidParams.Code, "部分角色不存在或不属于当前租户")
	}
	for _, role := range roles {
		if role.TenantID != tenantID {
			return nil, xerr.New(xerr.ErrInvalidParams.Code, "角色不属于当前租户")
		}
	}

	// 检查用户是否在当前租户中
	hasTenant, err := s.userTenantRoleRepo.CheckUserHasTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查用户租户关系失败", err)
	}
	if !hasTenant {
		return nil, xerr.New(xerr.ErrNotFound.Code, "用户不在该租户中")
	}

	// 获取用户信息用于日志
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrUserNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取用户信息失败", err)
	}

	// 获取旧的角色列表用于日志
	oldRoles, err := s.userTenantRoleRepo.GetUserRolesInTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取用户旧角色失败", err)
	}

	// 删除用户在当前租户下的所有角色关系
	if err := s.userTenantRoleRepo.DeleteByUserTenant(ctx, userID, tenantID); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "删除旧角色关系失败", err)
	}

	// 创建新的角色关系
	for _, roleID := range roleIDs {
		// 检查该关系是否已存在（防御性检查）
		exists, err := s.userTenantRoleRepo.CheckExists(ctx, userID, tenantID, roleID)
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查用户角色关系失败", err)
		}
		if exists {
			continue
		}

		// 生成关联ID
		userTenantRoleID, err := idgen.GenerateUUID()
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成关联ID失败", err)
		}

		// 创建用户租户角色关系
		userTenantRole := &model.UserTenantRole{
			UserTenantRoleID: userTenantRoleID,
			UserID:           userID,
			TenantID:         tenantID,
			RoleID:           roleID,
		}

		if err := s.userTenantRoleRepo.Create(ctx, userTenantRole); err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建用户租户角色关系失败", err)
		}
	}

	// 记录操作日志
	operationlog.RecordUpdate(ctx, constants.ModuleTenantMember, constants.ResourceTypeTenantMember, userID, user.Name,
		map[string]interface{}{
			"old_role_ids": oldRoles,
			"new_role_ids": roleIDs,
		},
		map[string]interface{}{
			"user_id":   userID,
			"tenant_id": tenantID,
			"role_ids":  roleIDs,
		})

	return &dto.UpdateMemberRolesResponse{
		UserID:  userID,
		RoleIDs: roleIDs,
	}, nil
}

// ListTenantMembers 获取租户成员列表
func (s *TenantMemberService) ListTenantMembers(ctx context.Context, req *dto.ListTenantMembersRequest) (*dto.ListTenantMembersResponse, error) {
	// 获取租户ID
	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 获取租户下的所有用户ID
	userIDs, err := s.userTenantRoleRepo.GetTenantUsers(ctx, tenantID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取租户用户列表失败", err)
	}

	if len(userIDs) == 0 {
		return &dto.ListTenantMembersResponse{
			Response: pagination.NewResponse(req.Request, 0),
			List:     []*dto.TenantMemberResponse{},
		}, nil
	}

	// 获取用户列表和总数，支持筛选条件
	users, total, err := s.userRepo.ListByIDsAndFilters(ctx, userIDs, req.GetOffset(), req.GetLimit(), req.Keyword, req.Status)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户列表失败", err)
	}

	// 转换为响应格式
	memberResponses := make([]*dto.TenantMemberResponse, len(users))
	for i, user := range users {
		memberResponses[i] = s.toTenantMemberResponse(ctx, user, tenantID)
	}

	return &dto.ListTenantMembersResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     memberResponses,
	}, nil
}

// toTenantMemberResponse 转换为租户成员响应格式
func (s *TenantMemberService) toTenantMemberResponse(ctx context.Context, user *model.User, tenantID string) *dto.TenantMemberResponse {
	// 处理可能为nil的指针字段
	var phone, email string
	if user.Phone != nil {
		phone = *user.Phone
	}
	if user.Email != nil {
		email = *user.Email
	}

	var lastLoginTime int64
	if user.LastLoginTime != nil {
		lastLoginTime = *user.LastLoginTime
	}

	// 获取用户在租户中的角色
	roleIDs, _ := s.userTenantRoleRepo.GetUserRolesInTenant(ctx, user.UserID, tenantID)

	return &dto.TenantMemberResponse{
		UserID:        user.UserID,
		UserName:      user.UserName,
		Name:          user.Name,
		Phone:         phone,
		Email:         email,
		Status:        int(user.Status),
		RoleIDs:       roleIDs,
		FirstLogin:    user.LastLoginTime == nil, // 首次登录判断
		LastLoginTime: lastLoginTime,
		CreatedAt:     user.CreatedAt,
	}
}
