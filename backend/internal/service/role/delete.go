package role

import (
	"admin/internal/dal/model"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/convert"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DeleteRole 删除角色
// 说明：
//   - 超管可删除任意租户角色
//   - 普通用户通过 RBAC 中间件鉴权 + 数据库自动租户过滤，只能删除本租户角色
//   - 级联删除：删除角色时会自动清理该角色的所有权限关联和用户绑定关系
func (s *Service) DeleteRole(ctx context.Context, roleID string) (err error) {
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
func (s *Service) BatchDeleteRoles(ctx context.Context, roleIDs []string) (err error) {
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
