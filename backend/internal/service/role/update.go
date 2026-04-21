package role

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UpdateRole 更新角色
// 说明：
//   - 超管可更新任意租户角色
//   - 普通用户通过 RBAC 中间件鉴权 + 数据库自动租户过滤，只能更新本租户角色
func (s *Service) UpdateRole(ctx context.Context, roleID string, req *dto.UpdateRoleRequest) (resp *dto.RoleInfo, err error) {
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

	return ModelToRoleInfo(newRole), nil
}

// UpdateRoleStatus 更新角色状态
// 说明：
//   - 超管可更新任意租户角色状态
//   - 普通用户通过 RBAC 中间件鉴权 + 数据库自动租户过滤，只能更新本租户角色状态
func (s *Service) UpdateRoleStatus(ctx context.Context, roleID string, status int) (err error) {
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
