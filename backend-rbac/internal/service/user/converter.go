package user

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// modelToUserInfo 将数据库模型转换为用户信息 DTO
func modelToUserInfo(user *model.User) *dto.UserInfo {
	if user == nil {
		return nil
	}

	return &dto.UserInfo{
		UserID:             user.UserID,
		UserName:           user.UserName,
		Nickname:           user.Nickname,
		Avatar:             user.Avatar,
		Phone:              user.Phone,
		Email:              user.Email,
		Description:        user.Description,
		Status:             int(user.Status),
		TenantID:           user.TenantID,
		LastLoginTime:      user.LastLoginTime,
		MustChangePassword: user.MustChangePassword,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
		Remark:             user.Remark,
	}
}

// modelToUserInfoWithRoles 将数据库模型转换为用户信息 DTO（含角色信息）
func modelToUserInfoWithRoles(user *model.User, roles []*model.Role) *dto.UserInfo {
	if user == nil {
		return nil
	}

	userInfo := modelToUserInfo(user)

	// 转换角色信息
	if len(roles) > 0 {
		roleInfos := make([]*dto.RoleInfo, 0, len(roles))
		for _, role := range roles {
			roleInfos = append(roleInfos, &dto.RoleInfo{
				RoleID:      role.RoleID,
				RoleCode:    role.RoleCode,
				Name:        role.Name,
				Description: role.Description,
			})
		}
		userInfo.Roles = roleInfos
	}

	return userInfo
}
