package service

import (
	"admin/internal/converter"
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/rbac"
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/convert"
	"admin/pkg/idgen"
	"admin/pkg/pagination"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// RoleService 角色服务
type RoleService struct {
	roleRepo       *repository.RoleRepo
	permissionRepo *repository.PermissionRepo
	menuRepo       *repository.MenuRepo
	rolePermRepo   *repository.RolePermissionRepo
	userRoleRepo   *repository.UserRoleRepo
	cache          *rbac.PermissionCache
	tenantRepo     *repository.TenantRepo
	recorder       *audit.Recorder
}

// NewRoleService 创建角色服务
func NewRoleService(
	roleRepo *repository.RoleRepo,
	permissionRepo *repository.PermissionRepo,
	menuRepo *repository.MenuRepo,
	rolePermRepo *repository.RolePermissionRepo,
	userRoleRepo *repository.UserRoleRepo,
	cache *rbac.PermissionCache,
	tenantRepo *repository.TenantRepo,
	recorder *audit.Recorder,
) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		menuRepo:       menuRepo,
		rolePermRepo:   rolePermRepo,
		userRoleRepo:   userRoleRepo,
		cache:          cache,
		tenantRepo:     tenantRepo,
		recorder:       recorder,
	}
}

// CreateRole 创建角色（支持继承父角色）
func (s *RoleService) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (resp *dto.RoleInfo, err error) {
	var role *model.Role

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleRole),
				audit.WithError(err),
			)
		} else if role != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleRole),
				audit.WithResource(constants.ResourceTypeRole, role.RoleID, role.Name),
				audit.WithValue(nil, role),
			)
		}
	}()

	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 检查角色编码是否已存在（租户内唯一）
	var exists bool
	exists, err = s.roleRepo.CheckExists(ctx, tenantID, req.RoleCode)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("role_code", req.RoleCode).Msg("检查角色编码是否存在失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查角色编码是否存在失败", err)
	}
	if exists {
		log.Warn().Str("tenant_id", tenantID).Str("role_code", req.RoleCode).Msg("角色编码已存在")
		return nil, xerr.ErrRoleCodeExists
	}

	// 如果有父角色，验证父角色属于 default 租户的角色模板
	var parentRoleCode *string
	if req.ParentRoleCode != nil {
		// 父角色必须在 default 租户中查找
		defaultTenantID, err := s.getDefaultTenantID(ctx)
		if err != nil {
			return nil, err
		}
		parentRole, err := s.roleRepo.GetByCodeWithTenant(ctx, defaultTenantID, *req.ParentRoleCode)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Warn().Str("parent_role_code", *req.ParentRoleCode).Msg("父角色不存在或只能继承 default 租户的角色模板")
				return nil, xerr.New(xerr.ErrInvalidParams.Code, "父角色不存在或只能继承 default 租户的角色模板")
			}
			log.Error().Err(err).Str("parent_role_code", *req.ParentRoleCode).Msg("查询父角色失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询父角色失败", err)
		}
		parentRoleCode = &parentRole.RoleCode
	}

	// 生成角色ID
	var roleID string
	roleID, err = idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成角色ID失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成角色ID失败", err)
	}

	// 构建角色模型，设置 ParentRoleID
	role = &model.Role{
		RoleID:      roleID,
		TenantID:    tenantID,
		RoleCode:    req.RoleCode,
		Name:        req.Name,
		Description: req.Description,
		Status:      int16(req.Status),
	}

	// 设置默认状态
	if role.Status == int16(constants.StatusZero) {
		role.Status = int16(constants.StatusEnabled) // 默认启用状态
	}

	// 如果有父角色，设置 parent_role_id
	if parentRoleCode != nil {
		// 通过 parent_role_code 查找父角色的 RoleID
		defaultTenantID, _ := s.getDefaultTenantID(ctx)
		parentRole, err := s.roleRepo.GetByCodeWithTenant(ctx, defaultTenantID, *parentRoleCode)
		if err == nil && parentRole != nil {
			role.ParentRoleID = parentRole.RoleID
		}
	}

	// 创建角色
	if err := s.roleRepo.Create(ctx, role); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Str("tenant_id", tenantID).Str("role_code", req.RoleCode).Msg("创建角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建角色失败", err)
	}

	return converter.ModelToRoleInfoWithParent(role, parentRoleCode), nil
}

// getDefaultTenantID 获取 default 租户ID
func (s *RoleService) getDefaultTenantID(ctx context.Context) (string, error) {
	tenant, err := s.tenantRepo.GetByCode(ctx, constants.DefaultTenantCode)
	if err != nil {
		return "", xerr.Wrap(xerr.ErrInternal.Code, "查询默认租户失败", err)
	}
	return tenant.TenantID, nil
}

// GetRoleByID 获取角色详情
// 说明：
// - 超管可查询任意租户角色
// - 普通用户通过 RBAC 中间件鉴权 + 数据库自动租户过滤，只能查询本租户角色
func (s *RoleService) GetRoleByID(ctx context.Context, roleID string) (*dto.RoleInfo, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("role_id", roleID).Msg("角色不存在")
			return nil, xerr.ErrRoleNotFound
		}
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	return converter.ModelToRoleInfo(role), nil
}

// UpdateRole 更新角色
// 说明：
// - 超管可更新任意租户角色
// - 普通用户通过 RBAC 中间件鉴权 + 数据库自动租户过滤，只能更新本租户角色
func (s *RoleService) UpdateRole(ctx context.Context, roleID string, req *dto.UpdateRoleRequest) (resp *dto.RoleInfo, err error) {
	var oldRole, newRole *model.Role

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleRole),
				audit.WithError(err),
			)
		} else if newRole != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleRole),
				audit.WithResource(constants.ResourceTypeRole, newRole.RoleID, newRole.Name),
				audit.WithValue(oldRole, newRole),
			)
		}
	}()

	// 获取旧角色信息
	oldRole, err = s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("role_id", roleID).Msg("角色不存在")
			return nil, xerr.ErrRoleNotFound
		}
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色失败")
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
	if req.Status != constants.StatusZero {
		updates["status"] = req.Status
	}
	updates["updated_at"] = time.Now().UnixMilli()

	// 更新角色
	if err := s.roleRepo.Update(ctx, roleID, updates); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Interface("updates", updates).Msg("更新角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新角色失败", err)
	}

	// 获取更新后的角色信息
	newRole, err = s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleID).Msg("获取更新后角色信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后角色信息失败", err)
	}

	return converter.ModelToRoleInfo(newRole), nil
}

// DeleteRole 删除角色
// 说明：
// - 超管可删除任意租户角色
// - 普通用户通过 RBAC 中间件鉴权 + 数据库自动租户过滤，只能删除本租户角色
// 级联删除：删除角色时会自动清理该角色的所有权限关联和用户绑定关系
func (s *RoleService) DeleteRole(ctx context.Context, roleID string) (err error) {
	var role *model.Role

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleRole),
				audit.WithError(err),
			)
		} else if role != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleRole),
				audit.WithResource(constants.ResourceTypeRole, role.RoleID, role.Name),
				audit.WithValue(role, nil),
			)
			log.Info().Str("role_id", roleID).Str("role_code", role.RoleCode).Msg("删除角色成功")
		}
	}()

	tenantID := xcontext.GetTenantID(ctx)

	// 检查角色是否存在
	role, err = s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("role_id", roleID).Msg("角色不存在")
			return xerr.ErrRoleNotFound
		}
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	// 删除角色
	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Str("role_code", role.RoleCode).Msg("删除角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "删除角色失败", err)
	}

	// 清理该角色的所有权限关联（role_permissions 表）
	if err := s.rolePermRepo.DeleteByRole(ctx, roleID, tenantID); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Msg("清理角色权限关联失败")
		// 不返回错误，角色已删除，权限关联清理失败不影响主流程
	}

	// 清理用户-角色绑定关系（user_roles 表）
	if err := s.userRoleRepo.DeleteRoles(ctx, []string{roleID}, tenantID); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Msg("清理用户角色绑定失败")
		// 不返回错误，角色已删除，绑定关系清理失败不影响主流程
	}

	// 通知权限缓存刷新
	s.cache.NotifyRefresh()

	return nil
}

// BatchDeleteRoles 批量删除角色
func (s *RoleService) BatchDeleteRoles(ctx context.Context, roleIDs []string) (err error) {
	var roleMap map[string]*model.Role

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleRole),
				audit.WithError(err),
			)
		} else if len(roleMap) > 0 {
			// 收集资源信息用于批量审计日志
			ids := make([]string, 0, len(roleMap))
			names := make([]string, 0, len(roleMap))
			for _, role := range roleMap {
				ids = append(ids, role.RoleID)
				names = append(names, role.Name)
			}
			// 记录批量删除审计日志（单条日志记录所有资源）
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleRole),
				audit.WithBatchResource(constants.ResourceTypeRole, ids, names),
				audit.WithValue(roleMap, nil),
			)
			log.Info().Strs("role_ids", roleIDs).Int("count", len(roleIDs)).Msg("批量删除角色成功")
		}
	}()

	tenantID := xcontext.GetTenantID(ctx)

	// 获取所有角色信息
	roles, err := s.roleRepo.GetByIDs(ctx, roleIDs)
	if err != nil {
		log.Error().Err(err).Strs("role_ids", roleIDs).Msg("查询角色信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询角色信息失败", err)
	}
	roleMap = convert.ToMap(roles, func(r *model.Role) string { return r.RoleID })

	// 验证所有角色都存在
	if len(roleMap) != len(roleIDs) {
		var missingIDs []string
		for _, id := range roleIDs {
			if _, exists := roleMap[id]; !exists {
				missingIDs = append(missingIDs, id)
			}
		}
		log.Warn().Strs("missing_ids", missingIDs).Msg("部分角色不存在")
		return xerr.New(xerr.ErrNotFound.Code, "部分角色不存在")
	}

	// 批量删除角色
	if err := s.roleRepo.BatchDelete(ctx, roleIDs); err != nil {
		log.Error().Err(err).Strs("role_ids", roleIDs).Msg("批量删除角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "批量删除角色失败", err)
	}

	// 批量清理所有角色的权限关联（role_permissions 表）
	if err := s.rolePermRepo.DeleteByRoles(ctx, roleIDs, tenantID); err != nil {
		log.Error().Err(err).Strs("role_ids", roleIDs).Msg("批量清理角色权限关联失败")
	}

	// 批量清理用户-角色绑定关系（user_roles 表）
	if err := s.userRoleRepo.DeleteRoles(ctx, roleIDs, tenantID); err != nil {
		log.Error().Err(err).Strs("role_ids", roleIDs).Msg("批量清理用户角色绑定失败")
	}

	// 通知权限缓存刷新
	s.cache.NotifyRefresh()

	return nil
}

// ListRoles 获取角色列表
// 说明：
// - 通过 context 自动获取租户信息，Repository 层自动添加租户过滤
// - 超管可查询所有租户角色
// - 租户管理员只能查询本租户角色
// - 普通用户无权限访问此接口，由 RBAC 中间件拦截
func (s *RoleService) ListRoles(ctx context.Context, req *dto.ListRolesRequest) (*dto.ListRolesResponse, error) {
	// 使用当前租户ID查询角色（Repository 层自动添加租户过滤）
	tenantID := xcontext.GetTenantID(ctx)

	roles, total, err := s.roleRepo.ListByTenantWithFilters(ctx, tenantID, req.GetOffset(), req.GetLimit(), req.RoleName, req.RoleCode, req.Status)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID).
			Str("role_name", req.RoleName).
			Str("role_code", req.RoleCode).
			Int("status", req.Status).
			Msg("查询角色列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色列表失败", err)
	}

	// 临时方案：如果不是超级管理员，过滤掉 super_admin 角色
	filteredRoles := s.filterSuperAdminRoles(ctx, roles)

	// 转换为响应格式
	roleInfos := converter.ModelListToRoleInfoList(filteredRoles)

	// 如果过滤后的数据量与原始总数不同，重新计算总数（仅当非超管时）
	filteredTotal := int64(len(roleInfos))
	if !xcontext.HasRole(ctx, constants.SuperAdmin) && int(total) != len(roles) {
		// 如果有分页且需要准确总数，这里简化处理，使用过滤后的数量
		// 注意：这是一个临时方案，可能影响分页准确性
		filteredTotal = int64(len(roleInfos))
	} else {
		filteredTotal = total
	}

	return &dto.ListRolesResponse{
		Response: pagination.NewResponse(req.Request, filteredTotal),
		List:     roleInfos,
	}, nil
}

// GetAllRoles 获取所有角色（不分页）
// 说明：
// - 返回当前租户的所有角色
func (s *RoleService) GetAllRoles(ctx context.Context, req *dto.GetAllRolesRequest) (*dto.GetAllRolesResponse, error) {
	// 使用当前租户ID查询角色
	tenantID := xcontext.GetTenantID(ctx)

	roles, err := s.roleRepo.ListByTenant(ctx, tenantID, req.RoleName, req.RoleCode, req.Status)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID).
			Str("role_name", req.RoleName).
			Str("role_code", req.RoleCode).
			Int("status", req.Status).
			Msg("查询所有角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询所有角色失败", err)
	}

	// 转换为响应格式
	roleInfos := converter.ModelListToRoleInfoList(roles)

	return &dto.GetAllRolesResponse{
		List: roleInfos,
	}, nil
}

// UpdateRoleStatus 更新角色状态
// 说明：
// - 超管可更新任意租户角色状态
// - 普通用户通过 RBAC 中间件鉴权 + 数据库自动租户过滤，只能更新本租户角色状态
func (s *RoleService) UpdateRoleStatus(ctx context.Context, roleID string, status int) (err error) {
	var oldRole, newRole *model.Role

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleRole),
				audit.WithError(err),
			)
		} else if newRole != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleRole),
				audit.WithResource(constants.ResourceTypeRole, newRole.RoleID, newRole.Name),
				audit.WithValue(oldRole, newRole),
			)
			log.Info().Str("role_id", roleID).Str("role_code", newRole.RoleCode).Int("status", status).Msg("更新角色状态成功")
		}
	}()

	// 获取旧角色信息
	oldRole, err = s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("role_id", roleID).Msg("角色不存在")
			return xerr.ErrRoleNotFound
		}
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	// 更新角色状态
	if err := s.roleRepo.UpdateStatus(ctx, roleID, status); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Int("status", status).Msg("更新角色状态失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新角色状态失败", err)
	}

	// 获取更新后的角色信息
	newRole, err = s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleID).Msg("获取更新后角色信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后角色信息失败", err)
	}

	return nil
}

// AssignMenus 为角色分配菜单（已弃用，保留向后兼容）
// Deprecated: 使用 AssignPermissions 代替
func (s *RoleService) AssignMenus(ctx context.Context, roleID string, menuIDs []string) error {
	req := &dto.AssignPermissionsRequest{
		MenuPermIDs: menuIDs,
	}
	return s.AssignPermissions(ctx, roleID, req)
}

// GetRoleMenus 获取角色的菜单ID列表（已弃用，保留向后兼容）
// Deprecated: 使用 GetRolePermissions 代替
func (s *RoleService) GetRoleMenus(ctx context.Context, roleID string) ([]string, error) {
	permissions, err := s.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return permissions.MenuPermIDs, nil
}

// AssignPermissions 为角色分配权限（菜单+按钮）
// 权限存储在 role_permissions 表中，通过 PermissionCache 实时计算
//
// 重要变更：分配菜单权限时，会自动关联该菜单的 API 权限
// 这解决了"菜单权限粒度"问题：前端隐藏菜单时，后端 API 也会被拦截
func (s *RoleService) AssignPermissions(ctx context.Context, roleID string, req *dto.AssignPermissionsRequest) error {
	// 1. 查询角色信息
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("role_id", roleID).Msg("角色不存在")
			return xerr.ErrRoleNotFound
		}
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	tenantID := xcontext.GetTenantID(ctx)

	// 2. 清除角色的所有现有权限（role_permissions 表）
	if err := s.rolePermRepo.DeleteByRole(ctx, roleID, tenantID); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Msg("清除旧权限失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "清除旧权限失败", err)
	}

	// 3. 收集所有需要添加的权限ID
	var allPermIDs []string

	// 3.1 添加菜单权限
	for _, menuID := range req.MenuPermIDs {
		// 菜单权限的 permission_id 可以是 "menu:<menuID>" 形式
		// 或者直接使用 menuID 查找对应的权限记录
		allPermIDs = append(allPermIDs, menuID)
	}

	// 3.2 添加按钮权限
	for _, buttonID := range req.ButtonPermIDs {
		allPermIDs = append(allPermIDs, buttonID)
	}

	// 4. 批量添加权限到 role_permissions 表
	if len(allPermIDs) > 0 {
		items := make([]*model.RolePermission, 0, len(allPermIDs))
		for _, permID := range allPermIDs {
			items = append(items, &model.RolePermission{
				RoleID:       role.RoleID,
				PermissionID: permID,
				TenantID:     tenantID,
			})
		}
		if err := s.rolePermRepo.AddPermissions(ctx, items); err != nil {
			log.Error().Err(err).Str("role_id", roleID).Int("count", len(items)).Msg("添加权限失败")
			return xerr.Wrap(xerr.ErrInternal.Code, "添加权限失败", err)
		}
	}

	// 5. 通知权限缓存刷新
	s.cache.NotifyRefresh()

	return nil
}

// GetRolePermissions 获取角色的所有权限（菜单+按钮）
func (s *RoleService) GetRolePermissions(ctx context.Context, roleID string) (*dto.RolePermissionsResponse, error) {
	// 1. 查询角色信息
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("role_id", roleID).Msg("角色不存在")
			return nil, xerr.ErrRoleNotFound
		}
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	tenantID := xcontext.GetTenantID(ctx)

	// 2. 从 role_permissions 表获取角色的所有权限ID
	permIDs, err := s.rolePermRepo.GetPermissionIDsByRole(ctx, role.RoleID, tenantID)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色权限失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色权限失败", err)
	}

	// 3. 根据权限ID查询权限详情，区分菜单和按钮
	var menuPermIDs []string
	var buttonPermIDs []string

	if len(permIDs) > 0 {
		perms, err := s.permissionRepo.GetByIDs(ctx, permIDs)
		if err != nil {
			log.Error().Err(err).Str("role_id", roleID).Msg("查询权限详情失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询权限详情失败", err)
		}

		for _, perm := range perms {
			switch perm.Type {
			case constants.TypeMenu:
				menuPermIDs = append(menuPermIDs, perm.PermissionID)
			case constants.TypeButton:
				buttonPermIDs = append(buttonPermIDs, perm.PermissionID)
			}
		}
	}

	return &dto.RolePermissionsResponse{
		MenuPermIDs:   menuPermIDs,
		ButtonPermIDs: buttonPermIDs,
	}, nil
}

// filterSuperAdminRoles 过滤超级管理员角色（临时方案）
// 当调用者不是超级管理员时，过滤掉 super_admin 角色
// 判断依据：role_code = super_admin
func (s *RoleService) filterSuperAdminRoles(ctx context.Context, roles []*model.Role) []*model.Role {
	// 如果是超级管理员，返回所有角色
	if xcontext.HasRole(ctx, constants.SuperAdmin) {
		return roles
	}

	// 非超级管理员，过滤掉 super_admin 角色
	filtered := make([]*model.Role, 0, len(roles))
	for _, role := range roles {
		// 跳过 super_admin 角色
		if role.RoleCode == constants.SuperAdmin {
			continue
		}
		filtered = append(filtered, role)
	}

	return filtered
}
