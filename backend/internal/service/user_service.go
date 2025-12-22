package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/idgen"
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
	// 确保有租户ID
	tenantID := req.TenantID
	if tenantID == "" {
		// 从上下文中获取租户ID
		if ctxTenantID, exists := xcontext.GetTenantID(ctx); exists {
			tenantID = ctxTenantID
		} else {
			return nil, xerr.Wrap(xerr.ErrUnauthorized.Code, "缺少租户信息", nil)
		}
	}

	// 检查用户名是否已存在（在指定租户内）
	exists, err := s.userRepo.CheckUserExists(ctx, req.UserName, tenantID)
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

	// 创建用户模型
	user := &model.User{
		UserID:    userID,
		UserName:  req.UserName,
		Password:  hashedPassword,
		Name:      req.UserName, // 临时使用用户名作为姓名，可以后续修改
		TenantID:  tenantID,
		Status:    req.Status,
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
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
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建用户失败", err)
	}

	// 返回用户响应
	return s.toUserResponse(user), nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrUserNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	return s.toUserResponse(user), nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, userID string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// 检查用户是否存在
	_, err := s.userRepo.GetUserByID(ctx, userID)
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
	if err := s.userRepo.UpdateUser(ctx, userID, updates); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新用户失败", err)
	}

	// 获取更新后的用户信息
	updatedUser, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后用户信息失败", err)
	}

	return s.toUserResponse(updatedUser), nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrUserNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 删除用户
	if err := s.userRepo.DeleteUser(ctx, userID); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "删除用户失败", err)
	}

	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	// 设置默认分页参数
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取用户列表和总数，支持筛选条件
	users, total, err := s.userRepo.ListUsersWithFilters(ctx, offset, pageSize, req.UserName, req.Status)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户列表失败", err)
	}

	// 转换为响应格式
	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(user)
	}

	return &dto.ListUsersResponse{
		List:      userResponses,
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		TotalPage: (total + int64(pageSize) - 1) / int64(pageSize), // 计算总页数
	}, nil
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(ctx context.Context, userID string, status int32) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrUserNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 更新用户状态
	if err := s.userRepo.UpdateUserStatus(ctx, userID, status); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "更新用户状态失败", err)
	}

	return nil
}

// toUserResponse 转换为用户响应格式
func (s *UserService) toUserResponse(user *model.User) *dto.UserResponse {
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
		Status:        user.Status,
		TenantID:      user.TenantID,
		RoleType:      user.RoleType,
		LastLoginTime: lastLoginTime,
		CreatedAt:     time.UnixMilli(user.CreatedAt),
		UpdatedAt:     time.UnixMilli(user.UpdatedAt),
	}
}
