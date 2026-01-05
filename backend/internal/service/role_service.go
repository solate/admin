package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/casbin"
	"admin/pkg/cache"
	"admin/pkg/constants"
	"admin/pkg/idgen"
	"admin/pkg/pagination"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// APIPath API 路径定义（用于解析菜单的 api_paths JSON 字段）
type APIPath struct {
	Path    string   `json:"path"`    // API 路径，如 /api/v1/users
	Methods []string `json:"methods"` // HTTP 方法列表，如 ["GET", "POST"]
}

// RoleService 角色服务
type RoleService struct {
	roleRepo       *repository.RoleRepo
	permissionRepo *repository.PermissionRepo
	menuRepo       *repository.MenuRepo
	enforcer       *casbin.Enforcer
	tenantCache    *cache.TenantCache
}

// NewRoleService 创建角色服务
func NewRoleService(roleRepo *repository.RoleRepo, permissionRepo *repository.PermissionRepo, menuRepo *repository.MenuRepo, enforcer *casbin.Enforcer, tenantCache *cache.TenantCache) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		menuRepo:       menuRepo,
		enforcer:       enforcer,
		tenantCache:    tenantCache,
	}
}

// CreateRole 创建角色（支持继承 default 租户角色模板）
func (s *RoleService) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 检查角色编码是否已存在（租户内唯一）
	exists, err := s.roleRepo.CheckExists(ctx, tenantID, req.RoleCode)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("role_code", req.RoleCode).Msg("检查角色编码是否存在失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查角色编码是否存在失败", err)
	}
	if exists {
		log.Warn().Str("tenant_id", tenantID).Str("role_code", req.RoleCode).Msg("角色编码已存在")
		return nil, xerr.ErrRoleCodeExists
	}

	// 如果有父角色，验证父角色属于 default 租户的角色模板
	var parentRole *model.Role
	if req.ParentRoleCode != nil {
		defaultTenantID := s.tenantCache.GetDefaultTenantID()
		parentRole, err = s.roleRepo.GetByCodeWithTenant(ctx, defaultTenantID, *req.ParentRoleCode)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Warn().Str("parent_role_code", *req.ParentRoleCode).Str("default_tenant_id", defaultTenantID).Msg("父角色不存在或只能继承 default 租户的角色模板")
				return nil, xerr.New(xerr.ErrInvalidParams.Code, "父角色不存在或只能继承 default 租户的角色模板")
			}
			log.Error().Err(err).Str("parent_role_code", *req.ParentRoleCode).Msg("查询父角色失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询父角色失败", err)
		}
	}

	// 生成角色ID
	roleID, err := idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成角色ID失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成角色ID失败", err)
	}

	// 构建角色模型
	role := &model.Role{
		RoleID:      roleID,
		TenantID:    tenantID,
		RoleCode:    req.RoleCode,
		Name:        req.Name,
		Description: req.Description,
		Status:      int16(req.Status),
	}

	// 设置默认状态
	if role.Status == 0 {
		role.Status = 1 // 默认启用状态
	}

	// 创建角色
	if err := s.roleRepo.Create(ctx, role); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Str("tenant_id", tenantID).Str("role_code", req.RoleCode).Msg("创建角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建角色失败", err)
	}

	// 如果有父角色，建立继承关系（不复制权限，实时计算）
	if req.ParentRoleCode != nil && parentRole != nil {
		// 创建 Casbin g2 策略（角色继承，不需要 domain）
		_, err = s.enforcer.AddGroupingPolicy(role.RoleCode, *req.ParentRoleCode)
		if err != nil {
			log.Error().Err(err).Str("role_code", role.RoleCode).Str("parent_role_code", *req.ParentRoleCode).Msg("创建角色继承关系失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建角色继承关系失败", err)
		}
	}

	return s.toRoleResponse(role, req.ParentRoleCode), nil
}

// copyParentMenuPermissions 复制父角色的菜单权限到子角色（已弃用）
// Deprecated: 现在使用实时计算权限，不再需要物理复制。此方法保留用于固化权限功能。
func (s *RoleService) copyParentMenuPermissions(parentRoleCode, childRoleCode, tenantCode string) error {
	// 获取父角色在 default 租户的所有权限
	defaultTenantCode := constants.DefaultTenantCode
	parentPolicies, _ := s.enforcer.GetFilteredPolicy(0, parentRoleCode, defaultTenantCode)

	// 过滤出菜单权限并复制到当前租户
	for _, policy := range parentPolicies {
		if len(policy) >= 4 {
			resource := policy[2]
			// 只复制菜单权限（menu:xxx）
			if strings.HasPrefix(resource, "menu:") {
				// 为子角色添加策略：p, child_role, current_tenant, menu:xxx, *
				_, err := s.enforcer.AddPolicy(childRoleCode, tenantCode, resource, "*")
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// GetRoleByID 获取角色详情
// 说明：
// - 超管通过 SkipTenantCheck 可查询任意租户角色
// - 普通用户通过 Casbin 中间件鉴权 + 数据库自动租户过滤，只能查询本租户角色
func (s *RoleService) GetRoleByID(ctx context.Context, roleID string) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("role_id", roleID).Msg("角色不存在")
			return nil, xerr.ErrRoleNotFound
		}
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	return s.toRoleResponse(role, nil), nil
}

// UpdateRole 更新角色
// 说明：
// - 超管通过 SkipTenantCheck 可更新任意租户角色
// - 普通用户通过 Casbin 中间件鉴权 + 数据库自动租户过滤，只能更新本租户角色
func (s *RoleService) UpdateRole(ctx context.Context, roleID string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	// 检查角色是否存在
	_, err := s.roleRepo.GetByID(ctx, roleID)
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
	if req.Status != 0 {
		updates["status"] = req.Status
	}
	updates["updated_at"] = time.Now().UnixMilli()

	// 更新角色
	if err := s.roleRepo.Update(ctx, roleID, updates); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Interface("updates", updates).Msg("更新角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新角色失败", err)
	}

	// 获取更新后的角色信息
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleID).Msg("获取更新后角色信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后角色信息失败", err)
	}

	return s.toRoleResponse(role, nil), nil
}

// DeleteRole 删除角色
// 说明：
// - 超管通过 SkipTenantCheck 可删除任意租户角色
// - 普通用户通过 Casbin 中间件鉴权 + 数据库自动租户过滤，只能删除本租户角色
// 级联删除：删除角色时会自动清理该角色的所有权限策略和继承关系
func (s *RoleService) DeleteRole(ctx context.Context, roleID string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, roleID)
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

	// 清理该角色的所有权限策略
	// 1. 清理权限策略 (p, role_code, tenant_code, resource, action)
	// 由于需要按 subject (role_code) 过滤，使用 GetFilteredPolicy(0, role.RoleCode)
	policies, _ := s.enforcer.GetFilteredPolicy(0, role.RoleCode)
	for _, policy := range policies {
		if len(policy) >= 4 {
			s.enforcer.RemovePolicy(policy[0], policy[1], policy[2], policy[3])
		}
	}

	// 2. 清理角色继承关系 (g2, child_role, parent_role)
	// g2 策略只有两个参数，使用 RemoveFilteredGroupingPolicy
	s.enforcer.RemoveFilteredGroupingPolicy(0, role.RoleCode)

	// 3. 清理用户-角色绑定关系 (g, username, role_code, tenant_code)
	// 使用 RemoveFilteredGroupingPolicy 按 role_code 过滤
	s.enforcer.RemoveFilteredGroupingPolicy(1, role.RoleCode)

	log.Info().Str("role_id", roleID).Str("role_code", role.RoleCode).Msg("删除角色成功")
	return nil
}

// ListRoles 获取角色列表
// 说明：
// - 通过 context 自动获取租户信息，Repository 层自动添加租户过滤
// - 超管通过 SkipTenantCheck 可查询所有租户角色
// - 租户管理员只能查询本租户角色
// - 普通用户无权限访问此接口，由 Casbin 中间件拦截
func (s *RoleService) ListRoles(ctx context.Context, req *dto.ListRolesRequest) (*dto.ListRolesResponse, error) {
	// 超管和租户管理员使用同一个查询方法
	// - 超管：context 中有 SkipTenantCheck，Repository 自动跳过租户过滤
	// - 租户管理员：Repository 自动添加 tenant_id 过滤
	roles, total, err := s.roleRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.Keyword, req.Status)
	if err != nil {
		log.Error().Err(err).
			Str("keyword", req.Keyword).
			Int("status", req.Status).
			Msg("查询角色列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色列表失败", err)
	}

	// 转换为响应格式
	roleResponses := make([]*dto.RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = s.toRoleResponse(role, nil)
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
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, roleID)
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

	log.Info().Str("role_id", roleID).Str("role_code", role.RoleCode).Int("status", status).Msg("更新角色状态成功")
	return nil
}

// toRoleResponse 转换为角色响应格式
func (s *RoleService) toRoleResponse(role *model.Role, parentRoleCode *string) *dto.RoleResponse {
	return &dto.RoleResponse{
		RoleID:        role.RoleID,
		TenantID:      role.TenantID,
		RoleCode:      role.RoleCode,
		Name:          role.Name,
		Description:   role.Description,
		Status:        int(role.Status),
		ParentRoleCode: parentRoleCode,
		CreatedAt:     role.CreatedAt,
		UpdatedAt:     role.UpdatedAt,
	}
}

// AssignMenus 为角色分配菜单（已弃用，保留向后兼容）
// Deprecated: 使用 AssignPermissions 代替
func (s *RoleService) AssignMenus(ctx context.Context, roleID string, menuIDs []string) error {
	req := &dto.AssignPermissionsRequest{
		MenuIDs: menuIDs,
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
	return permissions.MenuIDs, nil
}

// AssignPermissions 为角色分配权限（菜单+按钮）
// 根据设计文档：角色权限存储在 Casbin p 策略中
// 策略格式: p, role_code, domain, resource, action
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

	// 2. 获取租户代码
	tenantCode := xcontext.GetTenantCode(ctx)
	if tenantCode == "" {
		return xerr.ErrUnauthorized
	}

	// 3. 清除角色的所有权限（菜单 + 按钮 + API）
	_, err = s.enforcer.RemoveFilteredPolicy(0, role.RoleCode, tenantCode, "menu:", "*")
	if err != nil {
		log.Error().Err(err).Str("role_code", role.RoleCode).Str("tenant_code", tenantCode).Msg("清除旧菜单权限失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "清除旧菜单权限失败", err)
	}

	_, err = s.enforcer.RemoveFilteredPolicy(0, role.RoleCode, tenantCode, "btn:", "*")
	if err != nil {
		log.Error().Err(err).Str("role_code", role.RoleCode).Str("tenant_code", tenantCode).Msg("清除旧按钮权限失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "清除旧按钮权限失败", err)
	}

	// 清除 API 权限（路径以 /api/v1/ 开头的）
	_, err = s.enforcer.RemoveFilteredPolicy(0, role.RoleCode, tenantCode, "/api/v1/", "*")
	if err != nil {
		log.Error().Err(err).Str("role_code", role.RoleCode).Str("tenant_code", tenantCode).Msg("清除旧 API 权限失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "清除旧 API 权限失败", err)
	}

	// 4. 添加新的菜单权限，并自动关联 API 权限
	for _, menuID := range req.MenuIDs {
		// 4.1 添加菜单权限
		// 策略格式: p, role_code, domain, menu:menu_id, *
		_, err := s.enforcer.AddPolicy(role.RoleCode, tenantCode, "menu:"+menuID, "*")
		if err != nil {
			log.Error().Err(err).Str("role_code", role.RoleCode).Str("menu_id", menuID).Msg("添加菜单权限失败")
			return xerr.Wrap(xerr.ErrInternal.Code, "添加菜单权限失败", err)
		}

		// 4.2 查询菜单详情，获取关联的 API 路径
		menu, err := s.menuRepo.GetByID(ctx, menuID)
		if err != nil {
			// 菜单不存在，跳过 API 关联
			continue
		}

		// 4.3 解析 api_paths JSON 并添加 API 权限
		if menu.APIPaths != "" {
			var apiPaths []APIPath
			if err := json.Unmarshal([]byte(menu.APIPaths), &apiPaths); err == nil {
				for _, apiPath := range apiPaths {
					// 为每个 HTTP 方法添加权限策略
					for _, method := range apiPath.Methods {
						// 策略格式: p, role_code, domain, /api/v1/xxx, GET|POST|PUT|DELETE
						_, err = s.enforcer.AddPolicy(role.RoleCode, tenantCode, apiPath.Path, method)
						if err != nil {
							log.Error().Err(err).Str("role_code", role.RoleCode).Str("api_path", apiPath.Path).Str("method", method).Msg("添加 API 权限失败")
							return xerr.Wrap(xerr.ErrInternal.Code, "添加 API 权限失败", err)
						}
					}
				}
			}
		}
	}

	// 5. 添加新的按钮权限
	for _, buttonID := range req.ButtonIDs {
		// 根据 permission_id 查询 resource
		perm, err := s.permissionRepo.GetByID(ctx, buttonID)
		if err != nil {
			continue // 跳过不存在的权限
		}
		// 策略格式: p, role_code, domain, btn:menuID:action, *
		_, err = s.enforcer.AddPolicy(role.RoleCode, tenantCode, perm.Resource, "*")
		if err != nil {
			log.Error().Err(err).Str("role_code", role.RoleCode).Str("button_id", buttonID).Msg("添加按钮权限失败")
			return xerr.Wrap(xerr.ErrInternal.Code, "添加按钮权限失败", err)
		}
	}

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

	// 2. 获取租户代码
	tenantCode := xcontext.GetTenantCode(ctx)
	if tenantCode == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 3. 从 Casbin 获取角色的所有权限
	policies, _ := s.enforcer.GetFilteredPolicy(0, role.RoleCode, tenantCode)

	var menuIDs []string
	var buttonResources []string

	for _, policy := range policies {
		if len(policy) >= 4 {
			resource := policy[2]
			if strings.HasPrefix(resource, "menu:") {
				menuID := strings.TrimPrefix(resource, "menu:")
				menuIDs = append(menuIDs, menuID)
			} else if strings.HasPrefix(resource, "btn:") {
				buttonResources = append(buttonResources, resource)
			}
		}
	}

	// 4. 根据按钮 resource 查询 permission_id
	var buttonIDs []string
	for _, resource := range buttonResources {
		perm, err := s.permissionRepo.GetByResource(ctx, resource)
		if err == nil {
			buttonIDs = append(buttonIDs, perm.PermissionID)
		}
	}

	return &dto.RolePermissionsResponse{
		MenuIDs:   menuIDs,
		ButtonIDs: buttonIDs,
	}, nil
}

// FreezeInheritedPermissions 固化继承的权限（断开继承关系，将权限复制到当前租户）
// 说明：
// - 用于将继承自 default 租户角色模板的权限"固化"到当前租户
// - 固化后，角色将不再跟随模板更新，而是拥有自己独立的权限副本
// - 此操作会删除 g2 继承关系
func (s *RoleService) FreezeInheritedPermissions(ctx context.Context, roleID string) error {
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

	// 2. 获取租户代码
	tenantCode := xcontext.GetTenantCode(ctx)
	if tenantCode == "" {
		return xerr.ErrUnauthorized
	}

	// 3. 获取角色的继承父角色
	parents, _ := s.enforcer.GetRolesForUser(role.RoleCode)
	if len(parents) == 0 {
		log.Warn().Str("role_id", roleID).Str("role_code", role.RoleCode).Msg("该角色没有继承关系，无需固化")
		return xerr.New(xerr.ErrInvalidParams.Code, "该角色没有继承关系，无需固化")
	}

	// 4. 复制父角色的权限到当前租户
	for _, parentRoleCode := range parents {
		if err := s.copyParentMenuPermissions(parentRoleCode, role.RoleCode, tenantCode); err != nil {
			log.Error().Err(err).Str("role_code", role.RoleCode).Str("parent_role_code", parentRoleCode).Str("tenant_code", tenantCode).Msg("复制权限失败")
			return xerr.Wrap(xerr.ErrInternal.Code, "复制权限失败", err)
		}
	}

	// 5. 删除 g2 继承关系
	s.enforcer.DeleteRole(role.RoleCode)

	return nil
}

// CheckTenantAdminPermission 检查租户管理员是否有权限执行某操作
// 用途：在 handler 层调用此方法验证租户管理员的操作权限
func (s *RoleService) CheckTenantAdminPermission(ctx context.Context, operation string, targetID string) error {
	// 获取当前用户角色
	userName := xcontext.GetUserName(ctx)
	tenantCode := xcontext.GetTenantCode(ctx)

	// 获取用户的所有角色
	roles := s.enforcer.GetRolesForUserInDomain(userName, tenantCode)

	// 检查是否是租户管理员
	isTenantAdmin := false
	for _, role := range roles {
		if role == constants.Admin {
			isTenantAdmin = true
			break
		}
	}

	// 如果不是租户管理员，无需检查（由其他权限机制处理）
	if !isTenantAdmin {
		return nil
	}

	// 根据操作类型检查权限
	switch operation {
	case "delete_role":
		if !constants.TenantAdminCanDeleteRoles {
			return xerr.New(xerr.ErrForbidden.Code, "租户管理员不能删除角色")
		}
	case "modify_inherited_permission":
		if !constants.TenantAdminCanModifyInheritedPermissions {
			return xerr.New(xerr.ErrForbidden.Code, "租户管理员不能修改继承的权限")
		}
		// 还需要检查该权限是否是继承的
		role, err := s.roleRepo.GetByID(ctx, targetID)
		if err != nil {
			return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
		}
		// 检查角色是否有继承关系（g2）
		parents, _ := s.enforcer.GetRolesForUser(role.RoleCode)
		if len(parents) > 0 {
			return xerr.New(xerr.ErrForbidden.Code, "不能修改继承的权限，请先固化权限或使用额外权限")
		}
	case "modify_system_menu":
		// 检查菜单是否是系统菜单
		_, err := s.menuRepo.GetByID(ctx, targetID)
		if err != nil {
			return xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
		}
		// 系统菜单的判断：可以通过 menu 的创建租户或其他标识
		// 这里简化处理：假设 default 租户创建的菜单是系统菜单
		// 实际实现可能需要更复杂的逻辑
	}

	return nil
}
