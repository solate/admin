package role

import (
	"admin/internal/dto"
	"admin/pkg/constants"
	"admin/pkg/utils/pagination"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GetRoleByID 获取角色详情
// 说明：
//   - 超管可查询任意租户角色
//   - 普通用户通过 RBAC 中间件鉴权 + 数据库自动租户过滤，只能查询本租户角色
func (s *Service) GetRoleByID(ctx context.Context, roleID string) (*dto.RoleInfo, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("role_id", roleID).Msg("角色不存在")
			return nil, xerr.ErrRoleNotFound
		}
		log.Error().Err(err).Str("role_id", roleID).Msg("查询角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	return ModelToRoleInfo(role), nil
}

// ListRoles 获取角色列表
// 说明：
//   - 通过 context 自动获取租户信息，Repository 层自动添加租户过滤
//   - 超管可查询所有租户角色
//   - 租户管理员只能查询本租户角色
//   - 普通用户无权限访问此接口，由 RBAC 中间件拦截
func (s *Service) ListRoles(ctx context.Context, req *dto.ListRolesRequest) (*dto.ListRolesResponse, error) {
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
	roleInfos := ModelListToRoleInfoList(filteredRoles)

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
//   - 返回当前租户的所有角色
func (s *Service) GetAllRoles(ctx context.Context, req *dto.GetAllRolesRequest) (*dto.GetAllRolesResponse, error) {
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
	roleInfos := ModelListToRoleInfoList(roles)

	return &dto.GetAllRolesResponse{
		List: roleInfos,
	}, nil
}
