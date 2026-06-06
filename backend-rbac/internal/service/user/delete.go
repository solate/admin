package user

import (
	"admin/internal/dal/model"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DeleteUser 删除用户
// 级联删除：删除用户时会自动清理该用户的角色绑定关系
func (s *Service) DeleteUser(ctx context.Context, userID string) (err error) {
	var user *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleUser),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleUser),
				audit.WithResource(constants.ResourceTypeUser, user.UserID, user.UserName),
				audit.WithValue(user, nil),
			)
			log.Info().Str("user_id", userID).Str("username", user.UserName).Msg("删除用户成功")
		}
	}()

	// 检查用户是否存在
	user, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 删除用户
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("username", user.UserName).Msg("删除用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "删除用户失败", err)
	}

	// 清理该用户的所有角色绑定关系
	_ = s.userRoleRepo.DeleteUserRoles(ctx, user.UserID, user.TenantID)

	return nil
}

// BatchDeleteUsers 批量删除用户
func (s *Service) BatchDeleteUsers(ctx context.Context, userIDs []string) (err error) {
	var users []*model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleUser),
				audit.WithError(err),
			)
		} else if len(users) > 0 {
			// 收集资源信息用于批量审计日志
			ids := make([]string, 0, len(users))
			userNames := make([]string, 0, len(users))
			for _, user := range users {
				ids = append(ids, user.UserID)
				userNames = append(userNames, user.UserName)
			}
			// 记录批量删除审计日志（单条日志记录所有资源）
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleUser),
				audit.WithBatchResource(constants.ResourceTypeUser, ids, userNames),
				audit.WithValue(users, nil),
			)
			log.Info().Strs("user_ids", userIDs).Int("count", len(userIDs)).Msg("批量删除用户成功")
		}
	}()

	// 获取所有用户信息
	users, err = s.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		log.Error().Err(err).Strs("user_ids", userIDs).Msg("查询用户信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户信息失败", err)
	}

	// 验证所有用户都存在
	if len(users) != len(userIDs) {
		log.Warn().Int("requested", len(userIDs)).Int("found", len(users)).Msg("部分用户不存在")
		return xerr.New(xerr.ErrNotFound.Code, "部分用户不存在")
	}

	// 批量删除用户
	if err := s.userRepo.BatchDelete(ctx, userIDs); err != nil {
		log.Error().Err(err).Strs("user_ids", userIDs).Msg("批量删除用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "批量删除用户失败", err)
	}

	// 清理所有用户的所有角色绑定关系
	for _, user := range users {
		_ = s.userRoleRepo.DeleteUserRoles(ctx, user.UserID, user.TenantID)
	}

	return nil
}

// filterSuperAdminUsers 过滤超级管理员用户（临时方案）
// 当调用者不是超级管理员时，过滤掉默认管理员用户
// 判断依据：用户名=admin 或 邮箱=admin@example.com
func (s *Service) filterSuperAdminUsers(ctx context.Context, users []*model.User) []*model.User {
	// 如果是超级管理员，返回所有用户
	if xcontext.HasRole(ctx, constants.SuperAdmin) {
		return users
	}

	// 非超级管理员，过滤掉默认管理员
	filtered := make([]*model.User, 0, len(users))
	for _, user := range users {
		// 跳过默认管理员：用户名=admin 或 邮箱=admin@example.com
		if user.UserName == "admin" || user.Email == "admin@example.com" {
			continue
		}
		filtered = append(filtered, user)
	}

	return filtered
}
