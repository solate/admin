package service

import (
	"admin/internal/converter"
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/cache"
	"admin/pkg/casbin"
	"admin/pkg/constants"
	"admin/pkg/convert"
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
	recorder       *audit.Recorder
}

// NewRoleService 创建角色服务
func NewRoleService(roleRepo *repository.RoleRepo, permissionRepo *repository.PermissionRepo, menuRepo *repository.MenuRepo, enforcer *casbin.Enforcer, tenantCache *cache.TenantCache, recorder *audit.Recorder) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		menuRepo:       menuRepo,
		enforcer:       enforcer,
		tenantCache:    tenantCache,
		recorder:       recorder,
	}
}

// CreateRole 创建角色（支持继承 default 租户角色模板）
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
	var roleID string
	roleID, err = idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成角色ID失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成角色ID失败", err)
	}

	// 构建角色模型
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

	// 创建角色
	if err := s.roleRepo.Create(ctx, role); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Str("tenant_id", tenantID).Str("role_code", req.RoleCode).Msg("创建角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建角色失败", err)
	}

	// 获取租户代码
	tenantCode := xcontext.GetTenantCode(ctx)

	// 确定父角色代码（优先使用显式指定的父角色）
	var parentRoleCode string
	if req.ParentRoleCode != nil && parentRole != nil {
		parentRoleCode = *req.ParentRoleCode
	} else if tenantCode != constants.DefaultTenantCode {
		// 如果不是 default 租户且没有显式指定父角色，自动继承 default 租户的同名角色
		// 检查 default 租户是否存在同名角色
		defaultTenantID := s.tenantCache.GetDefaultTenantID()
		defaultRole, err := s.roleRepo.GetByCodeWithTenant(ctx, defaultTenantID, req.RoleCode)
		if err == nil && defaultRole != nil {
			// default 租户存在同名角色，自动建立继承关系
			parentRoleCode = defaultRole.RoleCode
			log.Info().
				Str("role_code", role.RoleCode).
				Str("tenant_code", tenantCode).
				Str("parent_role_code", parentRoleCode).
				Msg("自动建立角色继承关系")
		}
		// 如果 default 租户不存在同名角色，不建立继承关系
	}

	// 如果有父角色，建立 g2 继承关系（不复制权限，实时计算）
	if parentRoleCode != "" {
		// 创建 Casbin g2 策略（角色继承，不需要 domain）
		// 注意：必须使用 AddNamedGroupingPolicy("g2", ...) 来添加 g2 策略
		_, err = s.enforcer.AddNamedGroupingPolicy("g2", role.RoleCode, parentRoleCode)
		if err != nil {
			log.Error().Err(err).Str("role_code", role.RoleCode).Str("parent_role_code", parentRoleCode).Msg("创建角色继承关系失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建角色继承关系失败", err)
		}
		log.Info().
			Str("child_role", role.RoleCode).
			Str("parent_role", parentRoleCode).
			Str("tenant_code", tenantCode).
			Msg("g2 角色继承关系创建成功")
	}

	return converter.ModelToRoleInfoWithParent(role, req.ParentRoleCode), nil
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
// - 超管通过 SkipTenantCheck 可更新任意租户角色
// - 普通用户通过 Casbin 中间件鉴权 + 数据库自动租户过滤，只能更新本租户角色
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
// - 超管通过 SkipTenantCheck 可删除任意租户角色
// - 普通用户通过 Casbin 中间件鉴权 + 数据库自动租户过滤，只能删除本租户角色
// 级联删除：删除角色时会自动清理该角色的所有权限策略和继承关系
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

	// 清理所有角色的所有权限策略和绑定关系
	for _, role := range roleMap {
		// 1. 清理权限策略 (p, role_code, tenant_code, resource, action)
		policies, _ := s.enforcer.GetFilteredPolicy(0, role.RoleCode)
		for _, policy := range policies {
			if len(policy) >= 4 {
				s.enforcer.RemovePolicy(policy[0], policy[1], policy[2], policy[3])
			}
		}

		// 2. 清理角色继承关系 (g2, child_role, parent_role)
		s.enforcer.RemoveFilteredGroupingPolicy(0, role.RoleCode)

		// 3. 清理用户-角色绑定关系 (g, username, role_code, tenant_code)
		s.enforcer.RemoveFilteredGroupingPolicy(1, role.RoleCode)
	}

	return nil
}

// ListRoles 获取角色列表
// 说明：
// - 通过 context 自动获取租户信息，Repository 层自动添加租户过滤
// - 超管通过 SkipTenantCheck 可查询所有租户角色
// - 租户管理员只能查询本租户角色
// - 普通用户无权限访问此接口，由 Casbin 中间件拦截
func (s *RoleService) ListRoles(ctx context.Context, req *dto.ListRolesRequest) (*dto.ListRolesResponse, error) {
	// 获取 default 租户ID
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 查询 default 租户的角色
	roles, total, err := s.roleRepo.ListByTenantWithFilters(ctx, defaultTenantID, req.GetOffset(), req.GetLimit(), req.RoleName, req.RoleCode, req.Status)
	if err != nil {
		log.Error().Err(err).
			Str("default_tenant_id", defaultTenantID).
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
// - 只返回 default 租户的角色模板，所有用户创建用户时都从这些角色中选择
func (s *RoleService) GetAllRoles(ctx context.Context, req *dto.GetAllRolesRequest) (*dto.GetAllRolesResponse, error) {
	// 获取 default 租户ID
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 查询 default 租户的角色
	roles, err := s.roleRepo.ListByTenant(ctx, defaultTenantID, req.RoleName, req.RoleCode, req.Status)
	if err != nil {
		log.Error().Err(err).
			Str("default_tenant_id", defaultTenantID).
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
// - 超管通过 SkipTenantCheck 可更新任意租户角色状态
// - 普通用户通过 Casbin 中间件鉴权 + 数据库自动租户过滤，只能更新本租户角色状态
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
