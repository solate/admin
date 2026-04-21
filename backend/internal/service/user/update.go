package user

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/xerr"
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UpdateUser 更新用户
func (s *Service) UpdateUser(ctx context.Context, userID string, req *dto.UpdateUserRequest) (resp *dto.UserInfo, err error) {
	var oldUser, newUser *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithError(err),
			)
		} else if newUser != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithResource(constants.ResourceTypeUser, newUser.UserID, newUser.UserName),
				audit.WithValue(oldUser, newUser),
			)
		}
	}()

	// 检查用户是否存在
	oldUser, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 角色列表为必填项，不允许为空（提前验证，避免不必要的数据库操作）
	if len(req.RoleCodes) == 0 {
		log.Warn().Str("user_id", userID).Msg("更新用户时角色列表不能为空")
		return nil, xerr.Wrap(xerr.ErrInvalidParams.Code, "角色列表不能为空", nil)
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if req.Phone != "" {
		updates["phone"] = &req.Phone
	}
	if req.Email != "" {
		updates["email"] = &req.Email
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Description != "" {
		updates["description"] = &req.Description
	}
	if req.Status != constants.StatusZero {
		updates["status"] = req.Status
	}
	if req.Remark != "" {
		updates["remark"] = &req.Remark
	}

	// 处理租户更新
	var newTenantID string
	if req.TenantID != "" && req.TenantID != oldUser.TenantID {
		// 验证目标租户是否存在
		targetTenant, err := s.tenantRepo.GetByIDManual(ctx, req.TenantID)
		if err != nil {
			log.Error().Err(err).Str("target_tenant_id", req.TenantID).Msg("目标租户不存在")
			return nil, xerr.Wrap(xerr.ErrNotFound.Code, "目标租户不存在", err)
		}
		if targetTenant == nil {
			log.Warn().Str("target_tenant_id", req.TenantID).Msg("目标租户不存在")
			return nil, xerr.ErrTenantNotFound
		}

		// 更新租户ID
		updates["tenant_id"] = req.TenantID
		newTenantID = req.TenantID

		log.Info().
			Str("user_id", userID).
			Str("old_tenant_id", oldUser.TenantID).
			Str("new_tenant_id", req.TenantID).
			Msg("用户租户更新")
	} else {
		newTenantID = oldUser.TenantID
	}

	updates["updated_at"] = time.Now().UnixMilli()

	// 更新用户
	if err := s.userRepo.Update(ctx, userID, updates); err != nil {
		// 检查是否是唯一约束冲突错误（邮箱或手机号已存在）
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate key") ||
			strings.Contains(errMsg, "uk_users_email") ||
			strings.Contains(errMsg, "uk_users_phone") {
			log.Warn().Err(err).Str("user_id", userID).Msg("更新用户时邮箱或手机号已存在")
			return nil, xerr.ErrEmailOrPhoneExists
		}

		log.Error().Err(err).Str("user_id", userID).Interface("updates", updates).Msg("更新用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新用户失败", err)
	}

	// 获取更新后的用户信息
	newUser, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("获取更新后用户信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后用户信息失败", err)
	}

	// 更新用户角色
	// 使用新的租户ID（因为可能已经更新）
	var roles []*model.Role
	roles, err = s.userRoleService.ValidateAndAssignRoles(ctx, newUser.UserID, req.RoleCodes, newTenantID)
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("user_id", userID).
		Str("username", newUser.UserName).
		Strs("role_codes", req.RoleCodes).
		Msg("更新用户角色成功")

	return modelToUserInfoWithRoles(newUser, roles), nil
}

// UpdateUserStatus 更新用户状态
func (s *Service) UpdateUserStatus(ctx context.Context, userID string, status int) (err error) {
	var oldUser, newUser *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithError(err),
			)
		} else if newUser != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithResource(constants.ResourceTypeUser, newUser.UserID, newUser.UserName),
				audit.WithValue(oldUser, newUser),
			)
			log.Info().Str("user_id", userID).Int("status", status).Msg("更新用户状态成功")
		}
	}()

	// 检查用户是否存在
	oldUser, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 更新用户状态
	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		log.Error().Err(err).Str("user_id", userID).Int("status", status).Msg("更新用户状态失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新用户状态失败", err)
	}

	// 获取更新后的用户信息
	newUser, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("获取更新后用户信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后用户信息失败", err)
	}

	return nil
}
