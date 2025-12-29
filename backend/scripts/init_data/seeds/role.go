package seeds

import (
	"admin/internal/dal/model"
	"fmt"

	"gorm.io/gorm"
)

// RoleDefinition 角色定义
type RoleDefinition struct {
	RoleID   string
	RoleCode string
	Name     string
}

// SeedRoles 初始化默认角色
func SeedRoles(db *gorm.DB, roleDefs []RoleDefinition, tenantID string) ([]model.Role, error) {
	roles := make([]model.Role, 0, len(roleDefs))
	for _, def := range roleDefs {
		var role model.Role
		if err := db.Where("role_code = ? AND tenant_id = ?", def.RoleCode, tenantID).First(&role).Error; err != nil {
			// 角色不存在，创建新角色
			role = model.Role{
				RoleID:   def.RoleID,
				TenantID: tenantID,
				RoleCode: def.RoleCode,
				Name:     def.Name,
				Status:   1,
			}
			if err := db.Create(&role).Error; err != nil {
				return nil, fmt.Errorf("创建角色 %s 失败: %w", def.Name, err)
			}
			fmt.Printf("✅ 角色创建成功 role_id=%s role_code=%s name=%s\n", role.RoleID, role.RoleCode, role.Name)
		} else {
			fmt.Printf("ℹ️  角色已存在 role_id=%s role_code=%s name=%s\n", role.RoleID, role.RoleCode, role.Name)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// DefaultRoleDefinitions 返回默认角色定义
func DefaultRoleDefinitions(roleIDs []string) []RoleDefinition {
	return []RoleDefinition{
		{roleIDs[0], "super_admin", "超级管理员"},
		{roleIDs[1], "admin", "租户管理员"},
		{roleIDs[2], "user", "普通用户"},
	}
}
