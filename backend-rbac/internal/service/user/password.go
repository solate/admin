package user

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/passwordgen"
	"admin/pkg/utils/rsapwd"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ChangePassword 用户修改自己的密码
func (s *Service) ChangePassword(ctx context.Context, req *dto.ChangePasswordRequest) (err error) {
	var user *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithOperation("修改密码"),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithOperation("修改密码"),
				audit.WithResource(constants.ResourceTypeUser, user.UserID, user.UserName),
				audit.WithValue(nil, user),
			)
			log.Info().Str("user_id", user.UserID).Msg("用户修改密码成功")
		}
	}()

	// 获取当前用户ID
	userID := xcontext.GetUserID(ctx)

	// 查询用户信息
	user, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 解密前端传来的旧密码
	// 前端使用 JSEncrypt 库（PKCS#1 v1.5 填充）加密密码
	// 因此后端必须使用 DecryptPKCS1 方法解密
	decryptedOldPassword, err := s.rsaCipher.DecryptPKCS1(req.OldPassword)
	if err != nil {
		log.Error().Err(err).Msg("旧密码解密失败")
		return xerr.Wrap(xerr.ErrInvalidCredentials.Code, "旧密码解密失败", err)
	}

	// 使用 VerifyPassword 验证旧密码（使用解密后的 SHA256 哈希）
	if !passwordgen.VerifyPassword(decryptedOldPassword, user.Password) {
		log.Warn().Str("user_id", userID).Msg("旧密码错误")
		return xerr.New(xerr.ErrUnauthorized.Code, "原密码错误")
	}

	// 解密前端传来的新密码
	// 前端使用 JSEncrypt 库（PKCS#1 v1.5 填充）加密密码
	// 因此后端必须使用 DecryptPKCS1 方法解密
	decryptedNewPassword, err := s.rsaCipher.DecryptPKCS1(req.NewPassword)
	if err != nil {
		log.Error().Err(err).Msg("新密码解密失败")
		return xerr.Wrap(xerr.ErrInvalidCredentials.Code, "新密码解密失败", err)
	}

	// 生成新盐值并加密新密码（使用解密后的 SHA256 哈希）
	newSalt, err := passwordgen.GenerateSalt()
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("生成盐值失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
	}

	newHashedPassword, err := passwordgen.Argon2Hash(decryptedNewPassword, newSalt)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("新密码加密失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "新密码加密失败", err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(ctx, userID, newHashedPassword); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("更新密码失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新密码失败", err)
	}

	// 修改密码成功后，清除"必须修改密码"标记
	if err := s.userRepo.Update(ctx, userID, map[string]interface{}{
		"must_change_password": int16(constants.False),
	}); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("更新must_change_password失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新must_change_password失败", err)
	}

	return nil
}

// ResetPassword 超管重置用户密码（自动生成，密码只显示一次）
// 权限检查由 RBACMiddleware 处理
func (s *Service) ResetPassword(ctx context.Context, targetUserID string, req *dto.ResetPasswordRequest) (resp *dto.ResetPasswordResponse, err error) {
	var user *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithOperation("重置密码"),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleUser),
				audit.WithOperation("重置密码"),
				audit.WithResource(constants.ResourceTypeUser, user.UserID, user.UserName),
				audit.WithValue(nil, user),
			)
			log.Info().Str("operator_id", xcontext.GetUserID(ctx)).Str("target_user_id", targetUserID).Msg("重置用户密码成功")
		}
	}()

	// 查询目标用户
	user, err = s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("target_user_id", targetUserID).Msg("目标用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 自动生成随机密码（8位字母+数字）
	newPassword := passwordgen.GenerateRandomPassword(8)

	// 计算密码的 SHA256 哈希值（与前端登录流程保持一致）
	sha256Hash := rsapwd.HashPassword(newPassword)

	// 生成盐值并加密密码（使用 SHA256 哈希值）
	salt, err := passwordgen.GenerateSalt()
	if err != nil {
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("生成盐值失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成盐值失败", err)
	}
	hashedPassword, err := passwordgen.Argon2Hash(sha256Hash, salt)
	if err != nil {
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("密码加密失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "密码加密失败", err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(ctx, targetUserID, hashedPassword); err != nil {
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("更新密码失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新密码失败", err)
	}

	// 重置密码后，用户必须修改密码
	if err := s.userRepo.Update(ctx, targetUserID, map[string]interface{}{
		"must_change_password": int16(constants.True),
	}); err != nil {
		log.Error().Err(err).Str("target_user_id", targetUserID).Msg("更新must_change_password失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新must_change_password失败", err)
	}

	// 返回响应（密码只显示这一次）
	return &dto.ResetPasswordResponse{
		Password: newPassword,
		Message:  "密码重置成功，请立即将新密码告知用户，此密码仅显示一次",
	}, nil
}
