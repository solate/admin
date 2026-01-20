package converter

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// ModelToRoleInfo 将数据库模型转换为角色信息 DTO
func ModelToRoleInfo(role *model.Role) *dto.RoleInfo {
	if role == nil {
		return nil
	}

	return &dto.RoleInfo{
		RoleID:      role.RoleID,
		TenantID:    role.TenantID,
		TenantCode:  "default", // 所有角色模板都来自 default 租户
		RoleCode:    role.RoleCode,
		Name:        role.Name,
		Description: role.Description,
		Status:      int(role.Status),
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

// ModelToRoleInfoWithParent 将数据库模型转换为角色信息 DTO（包含父角色编码）
func ModelToRoleInfoWithParent(role *model.Role, parentRoleCode *string) *dto.RoleInfo {
	info := ModelToRoleInfo(role)
	if info != nil {
		info.ParentRoleCode = parentRoleCode
	}
	return info
}

// ModelListToRoleInfoList 批量将数据库模型转换为角色信息 DTO
func ModelListToRoleInfoList(roles []*model.Role) []*dto.RoleInfo {
	if len(roles) == 0 {
		return nil
	}

	result := make([]*dto.RoleInfo, len(roles))
	for i, role := range roles {
		result[i] = ModelToRoleInfo(role)
	}
	return result
}
