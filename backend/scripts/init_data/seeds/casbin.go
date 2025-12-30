package seeds

import (
	"admin/pkg/casbin"
	"fmt"

	"gorm.io/gorm"
)

// InitCasbinTable 初始化 Casbin 表（如果不存在）
func InitCasbinTable(db *gorm.DB) error {
	// 使用本地包创建执行器，会自动创建表
	_, err := casbin.NewEnforcerManager(db, casbin.DefaultModel())
	if err != nil {
		return fmt.Errorf("创建 Casbin 适配器失败: %w", err)
	}
	fmt.Println("✅ Casbin 表初始化成功")
	return nil
}

// SeedUserRoles 初始化用户角色关系（通过 Casbin）
func SeedUserRoles(db *gorm.DB, userName, roleCode, tenantCode string) error {
	// 使用本地包创建执行器（会自动 LoadPolicy）
	enforcer, err := casbin.NewEnforcerManager(db, casbin.DefaultModel())
	if err != nil {
		return fmt.Errorf("创建 Casbin 执行器失败: %w", err)
	}

	// 检查策略是否已存在
	hasPolicy, err := enforcer.HasGroupingPolicy(userName, roleCode, tenantCode)
	if err != nil {
		return fmt.Errorf("检查策略失败: %w", err)
	}

	if hasPolicy {
		fmt.Printf("ℹ️  用户角色策略已存在 username=%s role=%s tenant=%s\n", userName, roleCode, tenantCode)
		return nil
	}

	// 添加用户角色策略
	if _, err := enforcer.AddGroupingPolicy(userName, roleCode, tenantCode); err != nil {
		return fmt.Errorf("创建用户角色策略失败: %w", err)
	}

	fmt.Printf("✅ 用户角色策略创建成功 username=%s role=%s tenant=%s\n", userName, roleCode, tenantCode)
	return nil
}

// PolicyDefinition 策略定义
type PolicyDefinition struct {
	Sub   string // 主体：用户名或角色代码
	Domain string // 域：租户代码
	Obj   string // 对象：资源路径
	Act   string // 操作：HTTP 方法
}

// SeedPolicies 初始化 Casbin 权限策略
func SeedPolicies(db *gorm.DB, tenantCode string) error {
	// 使用本地包创建执行器（会自动 LoadPolicy）
	enforcer, err := casbin.NewEnforcerManager(db, casbin.DefaultModel())
	if err != nil {
		return fmt.Errorf("创建 Casbin 执行器失败: %w", err)
	}

	// 定义默认策略
	policies := getDefaultPolicies(tenantCode)

	// 添加策略（如果不存在）
	addedCount := 0
	for _, policy := range policies {
		// 检查策略是否已存在
		hasPolicy, err := enforcer.HasPolicy(policy.Sub, policy.Domain, policy.Obj, policy.Act)
		if err != nil {
			return fmt.Errorf("检查策略失败: %w", err)
		}

		if hasPolicy {
			continue
		}

		// 添加策略
		if _, err := enforcer.AddPolicy(policy.Sub, policy.Domain, policy.Obj, policy.Act); err != nil {
			return fmt.Errorf("添加策略失败 sub=%s obj=%s act=%s: %w", policy.Sub, policy.Obj, policy.Act, err)
		}
		addedCount++
	}

	if addedCount > 0 {
		fmt.Printf("✅ Casbin 策略初始化成功，添加 %d 条策略\n", addedCount)
	} else {
		fmt.Println("ℹ️  Casbin 策略已存在，跳过初始化")
	}

	return nil
}

// getDefaultPolicies 获取默认策略定义
func getDefaultPolicies(tenantCode string) []PolicyDefinition {
	return []PolicyDefinition{
		// ========== 超级管理员策略 ==========
		// 超管拥有所有权限（使用通配符）
		{Sub: "super_admin", Domain: tenantCode, Obj: "/api/v1/*", Act: "*"},
		{Sub: "super_admin", Domain: tenantCode, Obj: "/api/v1/*/*", Act: "*"},
		{Sub: "super_admin", Domain: tenantCode, Obj: "/api/v1/*/*/*", Act: "*"},
		{Sub: "super_admin", Domain: tenantCode, Obj: "/api/v1/*/*/*/*", Act: "*"},

		// ========== 租户管理员策略 ==========
		// 用户管理
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/users", Act: "GET"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/users", Act: "POST"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/users/:user_id", Act: "GET"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/users/:user_id", Act: "PUT"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/users/:user_id", Act: "DELETE"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/users/:user_id/status/:status", Act: "PUT"},

		// 角色管理
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/roles", Act: "GET"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/roles", Act: "POST"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/roles/:role_id", Act: "GET"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/roles/:role_id", Act: "PUT"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/roles/:role_id", Act: "DELETE"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/roles/:role_id/status/:status", Act: "PUT"},

		// 菜单管理
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/menus", Act: "GET"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/menus", Act: "POST"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/menus/all", Act: "GET"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/menus/tree", Act: "GET"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/menus/:menu_id", Act: "GET"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/menus/:menu_id", Act: "PUT"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/menus/:menu_id", Act: "DELETE"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/menus/:menu_id/status/:status", Act: "PUT"},

		// 用户菜单
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/user/menu", Act: "GET"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/user/buttons", Act: "GET"},

		// 操作日志
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/operation-logs", Act: "GET"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/operation-logs/:log_id", Act: "GET"},

		// 认证相关
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/auth/logout", Act: "POST"},
		{Sub: "admin", Domain: tenantCode, Obj: "/api/v1/auth/refresh", Act: "POST"},

		// ========== 普通用户策略 ==========
		// 用户只能查看和修改自己的信息
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/users", Act: "GET"},

		// 角色查看（只能看到自己的角色）
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/roles", Act: "GET"},
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/roles/:role_id", Act: "GET"},

		// 菜单查看
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/menus", Act: "GET"},
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/menus/all", Act: "GET"},
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/menus/tree", Act: "GET"},
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/menus/:menu_id", Act: "GET"},

		// 用户菜单
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/user/menu", Act: "GET"},
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/user/buttons", Act: "GET"},

		// 操作日志
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/operation-logs", Act: "GET"},
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/operation-logs/:log_id", Act: "GET"},

		// 认证相关
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/auth/logout", Act: "POST"},
		{Sub: "user", Domain: tenantCode, Obj: "/api/v1/auth/refresh", Act: "POST"},
	}
}
