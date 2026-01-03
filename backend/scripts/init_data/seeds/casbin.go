package seeds

import (
	"admin/pkg/casbin"
	"bufio"
	_ "embed"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

//go:embed policies.csv
var policiesCSV string

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

// SeedPolicies 初始化 Casbin 权限策略（从内嵌的 CSV 读取）
func SeedPolicies(db *gorm.DB, tenantCode string) error {
	// 从内嵌的 CSV 读取策略
	records, err := parseCSV(policiesCSV)
	if err != nil {
		return fmt.Errorf("解析 CSV 失败: %w", err)
	}

	// 使用本地包创建执行器（会自动 LoadPolicy）
	enforcer, err := casbin.NewEnforcerManager(db, casbin.DefaultModel())
	if err != nil {
		return fmt.Errorf("创建 Casbin 执行器失败: %w", err)
	}

	// 添加策略（如果不存在）
	addedCount := 0
	for _, record := range records {
		// 替换 {tenant} 占位符
		for i := range record {
			record[i] = strings.ReplaceAll(record[i], "{tenant}", tenantCode)
		}

		// 跳过空行和注释
		if len(record) == 0 || record[0] == "#" {
			continue
		}

		// 检查策略类型
		policyType := record[0]

		switch policyType {
		case "p", "P":
			// 权限策略: p, sub, dom, obj, act
			if len(record) < 5 {
				continue
			}
			sub, dom, obj, act := record[1], record[2], record[3], record[4]

			// 检查策略是否已存在
			hasPolicy, err := enforcer.HasPolicy(sub, dom, obj, act)
			if err != nil {
				return fmt.Errorf("检查策略失败: %w", err)
			}

			if hasPolicy {
				continue
			}

			// 添加策略
			if _, err := enforcer.AddPolicy(sub, dom, obj, act); err != nil {
				return fmt.Errorf("添加策略失败 sub=%s obj=%s act=%s: %w", sub, obj, act, err)
			}
			addedCount++

		case "g", "G":
			// 用户角色绑定: g, user, role, tenant
			if len(record) < 4 {
				continue
			}
			user, role, tenant := record[1], record[2], record[3]

			// 检查策略是否已存在
			hasPolicy, err := enforcer.HasGroupingPolicy(user, role, tenant)
			if err != nil {
				return fmt.Errorf("检查用户角色策略失败: %w", err)
			}

			if hasPolicy {
				continue
			}

			// 添加用户角色绑定
			if _, err := enforcer.AddGroupingPolicy(user, role, tenant); err != nil {
				return fmt.Errorf("添加用户角色绑定失败 user=%s role=%s: %w", user, role, err)
			}
			addedCount++

		default:
			// 跳过不识别的类型
			continue
		}
	}

	if addedCount > 0 {
		fmt.Printf("✅ Casbin 策略初始化成功，添加 %d 条策略\n", addedCount)
	} else {
		fmt.Println("ℹ️  Casbin 策略已存在，跳过初始化")
	}

	return nil
}

// parseCSV 解析 CSV 字符串
func parseCSV(csvContent string) ([][]string, error) {
	var records [][]string
	scanner := bufio.NewScanner(strings.NewReader(csvContent))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// 跳过注释行
		if strings.HasPrefix(line, "#") {
			continue
		}
		// 解析 CSV 行
		record := parseCSVLine(line)
		if len(record) > 0 {
			records = append(records, record)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// parseCSVLine 解析单行 CSV
func parseCSVLine(line string) []string {
	// 简单 CSV 解析：按逗号分割，去除空白
	parts := strings.Split(line, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// SeedRoleMenus 为角色分配菜单权限
// roleMenus: map[roleCode][]menuID
func SeedRoleMenus(db *gorm.DB, roleMenus map[string][]string, tenantCode string) error {
	enforcer, err := casbin.NewEnforcerManager(db, casbin.DefaultModel())
	if err != nil {
		return fmt.Errorf("创建 Casbin 执行器失败: %w", err)
	}

	addedCount := 0
	for roleCode, menuIDs := range roleMenus {
		for _, menuID := range menuIDs {
			// 检查策略是否已存在
			resource := "menu:" + menuID
			hasPolicy, err := enforcer.HasPolicy(roleCode, tenantCode, resource, "*")
			if err != nil {
				return fmt.Errorf("检查策略失败: %w", err)
			}

			if hasPolicy {
				continue
			}

			// 添加菜单权限策略
			if _, err := enforcer.AddPolicy(roleCode, tenantCode, resource, "*"); err != nil {
				return fmt.Errorf("添加角色菜单权限失败 role=%s menu=%s: %w", roleCode, menuID, err)
			}
			addedCount++
		}
	}

	if addedCount > 0 {
		fmt.Printf("✅ 角色菜单权限初始化成功，添加 %d 条策略\n", addedCount)
	} else {
		fmt.Println("ℹ️  角色菜单权限已存在，跳过初始化")
	}

	return nil
}
