package main

import (
	"fmt"

	"admin/internal/dal/model"
	"admin/pkg/constants"

	"gorm.io/gorm"
)

// MenuDefinition 菜单定义
type MenuDefinition struct {
	ID        string
	Name      string
	Type      string
	ParentID  string
	Path      string
	Component string
	Redirect  string
	Icon      string
	Sort      int16
}

var defaultSystemMenus = []MenuDefinition{
	{
		ID:       "menu_dashboard",
		Name:     "仪表盘",
		Type:     constants.TypeMenu,
		Path:     "/dashboard",
		Component: "views/Dashboard.vue",
		Icon:     "Dashboard",
		Sort:     1,
	},
	{
		ID:       "menu_system",
		Name:     "系统管理",
		Type:     constants.TypeMenu,
		Path:     "/system",
		Redirect: "/system/users",
		Icon:     "Setting",
		Sort:     100,
	},
	{
		ID:       "menu_users",
		Name:     "用户管理",
		Type:     constants.TypeMenu,
		ParentID: "menu_system",
		Path:     "/system/users",
		Component: "views/system/Users.vue",
		Icon:     "User",
		Sort:     101,
	},
	{
		ID:       "menu_roles",
		Name:     "角色管理",
		Type:     constants.TypeMenu,
		ParentID: "menu_system",
		Path:     "/system/roles",
		Component: "views/system/Roles.vue",
		Icon:     "UserGroup",
		Sort:     102,
	},
	{
		ID:       "menu_tenants",
		Name:     "租户管理",
		Type:     constants.TypeMenu,
		ParentID: "menu_system",
		Path:     "/system/tenants",
		Component: "views/system/Tenants.vue",
		Icon:     "Building",
		Sort:     103,
	},
	{
		ID:       "menu_menus",
		Name:     "菜单管理",
		Type:     constants.TypeMenu,
		ParentID: "menu_system",
		Path:     "/system/menus",
		Component: "views/system/Menus.vue",
		Icon:     "Menu",
		Sort:     104,
	},
}

// SeedSystemMenus 初始化系统菜单
func SeedSystemMenus(db *gorm.DB) error {
	for _, menuDef := range defaultSystemMenus {
		var existing model.Permission
		err := db.Unscoped().Where("permission_id = ?", menuDef.ID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			status := int16(constants.MenuStatusShow)
			menu := &model.Permission{
				PermissionID: menuDef.ID,
				TenantID:     constants.DefaultTenant, // 使用默认租户
				Name:         menuDef.Name,
				Type:         menuDef.Type,
				Status:       &status,
			}
			if menuDef.ParentID != "" {
				menu.ParentID = &menuDef.ParentID
			}
			if menuDef.Path != "" {
				menu.Path = &menuDef.Path
			}
			if menuDef.Component != "" {
				menu.Component = &menuDef.Component
			}
			if menuDef.Redirect != "" {
				menu.Redirect = &menuDef.Redirect
			}
			if menuDef.Icon != "" {
				menu.Icon = &menuDef.Icon
			}
			menu.Sort = &menuDef.Sort

			if err := db.Create(menu).Error; err != nil {
				return err
			}
			fmt.Printf("✅ 菜单创建成功 permission_id=%s name=%s\n", menu.PermissionID, menu.Name)
		} else if existing.DeletedAt > 0 {
			// 如果是软删除的记录，恢复它
			existing.DeletedAt = 0
			db.Save(&existing)
			fmt.Printf("ℹ️  菜单已恢复 permission_id=%s name=%s\n", existing.PermissionID, existing.Name)
		} else {
			fmt.Printf("ℹ️  菜单已存在 permission_id=%s name=%s\n", existing.PermissionID, existing.Name)
		}
	}
	return nil
}
