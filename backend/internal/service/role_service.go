package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/constants"
	"admin/pkg/idgen"
	"admin/pkg/operationlog"
	"admin/pkg/pagination"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"time"

	"gorm.io/gorm"
)

// RoleService 角色服务
type RoleService struct {
	roleRepo *repository.RoleRepo
}

// NewRoleService 创建角色服务
func NewRoleService(roleRepo *repository.RoleRepo) *RoleService {
	return &RoleService{
		roleRepo: roleRepo,
	}
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 检查角色编码是否已存在（租户内唯一）
	exists, err := s.roleRepo.CheckExists(ctx, tenantID, req.RoleCode)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查角色编码是否存在失败", err)
	}
	if exists {
		return nil, xerr.ErrRoleCodeExists
	}

	// 生成角色ID
	roleID, err := idgen.GenerateUUID()
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成角色ID失败", err)
	}

	// 构建角色模型
	role := &model.Role{
		RoleID:   roleID,
		TenantID: tenantID,
		RoleCode: req.RoleCode,
		Name:     req.Name,
		Status:   int16(req.Status),
	}

	// 设置默认状态
	if role.Status == 0 {
		role.Status = 1 // 默认启用状态
	}

	// 创建角色
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建角色失败", err)
	}

	// 记录操作日志
	ctx = operationlog.RecordCreate(ctx, constants.ModuleRole, constants.ResourceTypeRole, role.RoleID, role.Name, role)

	return s.toRoleResponse(role), nil
}

// GetRoleByID 获取角色详情
// 说明：
// - 超管通过 SkipTenantCheck 可查询任意租户角色
// - 普通用户通过 Casbin 中间件鉴权 + 数据库自动租户过滤，只能查询本租户角色
func (s *RoleService) GetRoleByID(ctx context.Context, roleID string) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrRoleNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	return s.toRoleResponse(role), nil
}

// UpdateRole 更新角色
// 说明：
// - 超管通过 SkipTenantCheck 可更新任意租户角色
// - 普通用户通过 Casbin 中间件鉴权 + 数据库自动租户过滤，只能更新本租户角色
func (s *RoleService) UpdateRole(ctx context.Context, roleID string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	// 检查角色是否存在，获取旧值用于日志
	oldRole, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrRoleNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != 0 {
		updates["status"] = req.Status
	}
	updates["updated_at"] = time.Now().UnixMilli()

	// 更新角色
	if err := s.roleRepo.Update(ctx, roleID, updates); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新角色失败", err)
	}

	// 获取更新后的角色信息
	updatedRole, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后角色信息失败", err)
	}

	// 记录操作日志
	ctx = operationlog.RecordUpdate(ctx, constants.ModuleRole, constants.ResourceTypeRole, updatedRole.RoleID, updatedRole.Name, oldRole, updatedRole)

	return s.toRoleResponse(updatedRole), nil
}

// DeleteRole 删除角色
// 说明：
// - 超管通过 SkipTenantCheck 可删除任意租户角色
// - 普通用户通过 Casbin 中间件鉴权 + 数据库自动租户过滤，只能删除本租户角色
func (s *RoleService) DeleteRole(ctx context.Context, roleID string) error {
	// 检查角色是否存在，获取角色信息用于日志
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrRoleNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	// 删除角色
	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "删除角色失败", err)
	}

	// 记录操作日志
	operationlog.RecordDelete(ctx, constants.ModuleRole, constants.ResourceTypeRole, role.RoleID, role.Name, role)

	return nil
}

// ListRoles 获取角色列表
// 说明：
// - 通过 context 自动获取租户信息，Repository 层自动添加租户过滤
// - 超管通过 SkipTenantCheck 可查询所有租户角色
// - 租户管理员只能查询本租户角色
// - 普通用户无权限访问此接口，由 Casbin 中间件拦截
func (s *RoleService) ListRoles(ctx context.Context, req *dto.ListRolesRequest) (*dto.ListRolesResponse, error) {
	var roles []*model.Role
	var total int64
	var err error

	// 超管和租户管理员使用同一个查询方法
	// - 超管：context 中有 SkipTenantCheck，Repository 自动跳过租户过滤
	// - 租户管理员：Repository 自动添加 tenant_id 过滤
	roles, total, err = s.roleRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.Keyword, req.Status)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色列表失败", err)
	}

	// 转换为响应格式
	roleResponses := make([]*dto.RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = s.toRoleResponse(role)
	}

	return &dto.ListRolesResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     roleResponses,
	}, nil
}

// UpdateRoleStatus 更新角色状态
// 说明：
// - 超管通过 SkipTenantCheck 可更新任意租户角色状态
// - 普通用户通过 Casbin 中间件鉴权 + 数据库自动租户过滤，只能更新本租户角色状态
func (s *RoleService) UpdateRoleStatus(ctx context.Context, roleID string, status int) error {
	// 检查角色是否存在，获取旧值用于日志
	oldRole, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrRoleNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	// 更新角色状态
	if err := s.roleRepo.UpdateStatus(ctx, roleID, status); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "更新角色状态失败", err)
	}

	// 获取更新后的角色信息
	updatedRole, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后角色信息失败", err)
	}

	// 记录操作日志
	operationlog.RecordUpdate(ctx, constants.ModuleRole, constants.ResourceTypeRole, updatedRole.RoleID, updatedRole.Name, oldRole, updatedRole)

	return nil
}

// toRoleResponse 转换为角色响应格式
func (s *RoleService) toRoleResponse(role *model.Role) *dto.RoleResponse {
	return &dto.RoleResponse{
		RoleID:      role.RoleID,
		TenantID:    role.TenantID,
		RoleCode:    role.RoleCode,
		Name:        role.Name,
		Description: role.Description,
		Status:      int(role.Status),
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}
