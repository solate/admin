package seeds

import (
	"admin/internal/dal/model"
	"admin/pkg/idgen"
	"fmt"

	"gorm.io/gorm"
)

// MenuDefinition 菜单定义
type MenuDefinition struct {
	MenuID    string
	Name      string
	Path      string
	Component string
	Icon      string
	Sort      int
	Status    int
}

// SeedSystemMenus 初始化系统菜单
func SeedSystemMenus(db *gorm.DB) error {
	// 检查是否已有菜单
	var count int64
	if err := db.Model(&model.Menu{}).Where("deleted_at = 0").Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		fmt.Println("ℹ️  系统菜单已存在，跳过初始化")
		return nil
	}

	// 生成菜单ID
	ids, err := idgen.GenerateUUIDs(8)
	if err != nil {
		return fmt.Errorf("生成菜单ID失败: %w", err)
	}

	// 根菜单的 ParentID 为空字符串
	parentID0 := ids[0]

	menus := []model.Menu{
		// 系统管理
		{
			MenuID: ids[0], ParentID: "", Name: "系统管理", Path: "/system",
			Component: "", Icon: "Setting", Sort: 100, Status: 1,
		},
		// 用户管理
		{
			MenuID: ids[1], ParentID: parentID0, Name: "用户管理", Path: "/system/users",
			Component: "system/users/index", Icon: "User", Sort: 1, Status: 1,
		},
		// 角色管理
		{
			MenuID: ids[2], ParentID: parentID0, Name: "角色管理", Path: "/system/roles",
			Component: "system/roles/index", Icon: "UserFilled", Sort: 2, Status: 1,
		},
		// 租户管理
		{
			MenuID: ids[3], ParentID: parentID0, Name: "租户管理", Path: "/system/tenants",
			Component: "system/tenants/index", Icon: "OfficeBuilding", Sort: 3, Status: 1,
		},
		// 菜单管理
		{
			MenuID: ids[4], ParentID: parentID0, Name: "菜单管理", Path: "/system/menus",
			Component: "system/menus/index", Icon: "Menu", Sort: 4, Status: 1,
		},
		// 操作日志
		{
			MenuID: ids[5], ParentID: parentID0, Name: "操作日志", Path: "/system/operation-logs",
			Component: "system/operation-logs/index", Icon: "Document", Sort: 5, Status: 1,
		},
	}

	for _, menu := range menus {
		if err := db.Create(&menu).Error; err != nil {
			return fmt.Errorf("创建菜单 %s 失败: %w", menu.Name, err)
		}
		fmt.Printf("✅ 菜单创建成功 %s\n", menu.Name)
	}

	return nil
}
