package converter

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// ModelToUserInfo 将数据库模型转换为用户信息 DTO
func ModelToUserInfo(user *model.User) *dto.UserInfo {
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

// ModelToUserInfoWithRoles 将数据库模型转换为用户信息 DTO（含角色信息）
func ModelToUserInfoWithRoles(user *model.User, roles []*model.Role) *dto.UserInfo {
	if user == nil {
		return nil
	}

	userInfo := ModelToUserInfo(user)

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
