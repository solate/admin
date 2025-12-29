package seeds

import (
	"fmt"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"gorm.io/gorm"
)

// InitCasbinTable 初始化 Casbin 表（如果不存在）
func InitCasbinTable(db *gorm.DB) error {
	// 使用 gorm-adapter 自动创建表
	_, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return fmt.Errorf("创建 Casbin 适配器失败: %w", err)
	}
	fmt.Println("✅ Casbin 表初始化成功")
	return nil
}

// SeedUserRoles 初始化用户角色关系（通过 Casbin）
func SeedUserRoles(db *gorm.DB, userName, roleCode, tenantCode string) error {
	// 创建适配器（会自动创建表）
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return fmt.Errorf("创建 Casbin 适配器失败: %w", err)
	}

	// 创建简单的 RBAC 模型
	modelText := `
[request_definition]
r = sub, domain, obj, act

[policy_definition]
p = sub, domain, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.domain == p.domain && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return fmt.Errorf("创建 Casbin 模型失败: %w", err)
	}

	// 创建执行器
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return fmt.Errorf("创建 Casbin 执行器失败: %w", err)
	}

	// 加载现有策略
	if err := enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("加载策略失败: %w", err)
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
