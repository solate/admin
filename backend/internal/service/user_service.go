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
	"time"

	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepo
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepo) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// 检查用户名是否已存在（全局唯一）
	exists, err := s.userRepo.CheckExists(ctx, req.UserName)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查用户是否存在失败", err)
	}
	if exists {
		return nil, xerr.ErrUserExists
	}

	// 生成用户ID
	userID, err := idgen.GenerateUUID()
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成用户ID失败", err)
	}

	// 生成盐值并加密密码
	salt, err := passwordgen.GenerateSalt()
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
	}
	hashedPassword, err := passwordgen.Argon2Hash(req.Password, salt)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "密码加密失败", err)
	}

	// 创建用户模型（用户表与租户解耦）
	user := &model.User{
		UserID:   userID,
		UserName: req.UserName,
		Password: hashedPassword,
		Name:     req.Name,
		Status:   int16(req.Status),
	}

	// 如果没有传入姓名，使用用户名作为默认值
	if user.Name == "" {
		user.Name = req.UserName
	}

	// 设置可选字段
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if req.Email != "" {
		user.Email = &req.Email
	}

	// 设置默认状态
	if user.Status == 0 {
		user.Status = 1 // 默认正常状态
	}

	// 创建用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建用户失败", err)
	}

	// 记录操作日志
	ctx = operationlog.RecordCreate(ctx, constants.ModuleUser, constants.ResourceTypeUser, user.UserID, user.Name, user)

	// 注意：创建用户后，还需要通过 user_tenant_role 表关联用户和租户
	// 这里暂时不处理，需要单独的接口来分配角色

	return s.toUserResponse(ctx, user), nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrUserNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	return s.toUserResponse(ctx, user), nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, userID string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// 检查用户是否存在，获取旧值用于日志
	oldUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrUserNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if req.Password != "" {
		// 生成盐值并加密密码
		salt, err := passwordgen.GenerateSalt()
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
		}
		hashedPassword, err := passwordgen.Argon2Hash(req.Password, salt)
		if err != nil {
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
	if req.Name != "" {
		updates["name"] = req.Name
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
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新用户失败", err)
	}

	// 获取更新后的用户信息
	updatedUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后用户信息失败", err)
	}

	// 记录操作日志
	ctx = operationlog.RecordUpdate(ctx, constants.ModuleUser, constants.ResourceTypeUser, updatedUser.UserID, updatedUser.Name, oldUser, updatedUser)

	return s.toUserResponse(ctx, updatedUser), nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	// 检查用户是否存在，获取用户信息用于日志
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrUserNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 删除用户
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "删除用户失败", err)
	}

	// 记录操作日志
	operationlog.RecordDelete(ctx, constants.ModuleUser, constants.ResourceTypeUser, user.UserID, user.Name, user)

	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	// 获取用户列表和总数，支持筛选条件
	users, total, err := s.userRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.UserName, req.Status)
	if err != nil {
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
	// 检查用户是否存在，获取旧值用于日志
	oldUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrUserNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 更新用户状态
	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "更新用户状态失败", err)
	}

	// 获取更新后的用户信息
	updatedUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后用户信息失败", err)
	}

	// 记录操作日志
	operationlog.RecordUpdate(ctx, constants.ModuleUser, constants.ResourceTypeUser, updatedUser.UserID, updatedUser.Name, oldUser, updatedUser)

	return nil
}

// toUserResponse 转换为用户响应格式
func (s *UserService) toUserResponse(ctx context.Context, user *model.User) *dto.UserResponse {
	// 处理可能为nil的指针字段
	var phone, email, avatar string
	if user.Phone != nil {
		phone = *user.Phone
	}
	if user.Email != nil {
		email = *user.Email
	}
	if user.Avatar != nil {
		avatar = *user.Avatar
	}

	var lastLoginTime int64
	if user.LastLoginTime != nil {
		lastLoginTime = *user.LastLoginTime
	}

	return &dto.UserResponse{
		UserID:        user.UserID,
		UserName:      user.UserName,
		Name:          user.Name,
		Avatar:        avatar,
		Phone:         phone,
		Email:         email,
		Status:        int(user.Status),
		TenantID:      xcontext.GetTenantID(ctx),
		LastLoginTime: lastLoginTime,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
