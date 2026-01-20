package main

import (
	"admin/internal/dal/model"
	"admin/pkg/config"
	"admin/pkg/database"
	"admin/pkg/idgen"
	"admin/pkg/xcontext"
	"admin/scripts/init_data/seeds"
	"context"
	"fmt"
	"os"
	"strings"

	"gorm.io/gorm"
)

// SeedResult 初始化结果
type SeedResult struct {
	Tenant      model.Tenant
	User        model.User
	Roles       []model.Role
	Departments []model.Department
	Positions   []model.Position
	DictTypes   []model.DictType
	Password    string // 仅用于输出，不存储
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔧 初始化数据库: host=%s port=%d db=%s\n", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	dsn := cfg.Database.GetDSN() + " TimeZone=Asia/Shanghai"
	db, err := database.Connect(database.Config{
		DSN:             dsn,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.GetConnMaxLifetime(),
		LogLevel:        cfg.Log.Level,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 数据库连接失败: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  获取 sql.DB 失败: %v\n", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  关闭数据库连接失败: %v\n", err)
		}
	}()

	ctx := xcontext.SkipTenantCheck(context.Background())
	db = db.WithContext(ctx)

	fmt.Println("🚀 开始初始化默认数据")

	result, err := SeedAllData(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 初始化失败: %v\n", err)
		os.Exit(1)
	}

	printResult(result)
}

// SeedAllData 执行所有数据初始化
func SeedAllData(db *gorm.DB) (*SeedResult, error) {
	result := &SeedResult{
		Password: seeds.DefaultAdminPassword,
	}

	// 生成所需的ID
	// 5个基础ID (租户、用户、3个角色) + 29个菜单ID + 19个部门ID + 37个岗位ID + 52个字典ID (13个类型+39个项) = 142个ID
	ids, err := idgen.GenerateUUIDs(142)
	if err != nil {
		return nil, fmt.Errorf("生成ID失败: %w", err)
	}
	idIndex := 0

	// 1. 初始化租户
	tenant, err := seeds.SeedTenant(db, ids[idIndex])
	if err != nil {
		return nil, fmt.Errorf("初始化租户失败: %w", err)
	}
	result.Tenant = *tenant
	idIndex++

	// 2. 初始化用户
	user, err := seeds.SeedUser(db, ids[idIndex], tenant.TenantID)
	if err != nil {
		return nil, fmt.Errorf("初始化用户失败: %w", err)
	}
	result.User = *user
	idIndex++

	// 3. 初始化角色
	roleDefs := seeds.DefaultRoleDefinitions(ids[idIndex : idIndex+3])
	roles, err := seeds.SeedRoles(db, roleDefs, tenant.TenantID)
	if err != nil {
		return nil, fmt.Errorf("初始化角色失败: %w", err)
	}
	result.Roles = roles
	idIndex += 3

	// 4. 初始化 Casbin 表（如果不存在）
	if err := seeds.InitCasbinTable(db); err != nil {
		return nil, fmt.Errorf("初始化 Casbin 表失败: %w", err)
	}

	// 5. 初始化用户-角色关系（通过 Casbin）
	if err := seeds.SeedUserRoles(db, user.UserName, roles[0].RoleCode, tenant.TenantCode); err != nil {
		return nil, fmt.Errorf("初始化用户角色关系失败: %w", err)
	}

	// 6. 初始化 Casbin 权限策略
	if err := seeds.SeedPolicies(db, tenant.TenantCode); err != nil {
		return nil, fmt.Errorf("初始化权限策略失败: %w", err)
	}

	// 7. 初始化系统菜单
	menuDefs := seeds.DefaultMenuDefinitions(ids[idIndex : idIndex+29])
	if err := seeds.SeedSystemMenus(db, menuDefs); err != nil {
		return nil, fmt.Errorf("初始化系统菜单失败: %w", err)
	}

	// 7.1. 为角色分配菜单权限
	// 提取所有菜单ID
	menuIDs := ids[idIndex : idIndex+29]
	// 为 super_admin 和 admin 角色分配所有菜单权限
	roleMenus := map[string][]string{
		"super_admin": menuIDs, // super_admin 角色拥有所有菜单权限
		"admin":       menuIDs, // admin 角色拥有所有菜单权限
	}
	if err := seeds.SeedRoleMenus(db, roleMenus, tenant.TenantCode); err != nil {
		return nil, fmt.Errorf("初始化角色菜单权限失败: %w", err)
	}

	idIndex += 29

	// 8. 初始化组织架构 - 部门
	fmt.Println("\n📁 开始初始化组织架构")
	deptDefs := seeds.DefaultDepartmentDefinitions(ids[idIndex : idIndex+19])
	departments, err := seeds.SeedDepartments(db, deptDefs, tenant.TenantID)
	if err != nil {
		return nil, fmt.Errorf("初始化部门失败: %w", err)
	}
	result.Departments = departments
	idIndex += 19

	// 9. 初始化组织架构 - 岗位
	posDefs := seeds.DefaultPositionDefinitions(ids[idIndex:])
	positions, err := seeds.SeedPositions(db, posDefs, tenant.TenantID)
	if err != nil {
		return nil, fmt.Errorf("初始化岗位失败: %w", err)
	}
	result.Positions = positions

	// 10. 初始化系统字典
	fmt.Println("\n📚 开始初始化系统字典")
	dictDefs := seeds.DefaultDictTypeDefinitions()
	dictTypes, err := seeds.SeedDicts(db, dictDefs, tenant.TenantID, ids[idIndex:])
	if err != nil {
		return nil, fmt.Errorf("初始化字典失败: %w", err)
	}
	// 转换 []*model.DictType 为 []model.DictType
	for _, dt := range dictTypes {
		result.DictTypes = append(result.DictTypes, *dt)
	}

	return result, nil
}

// printResult 打印初始化结果
func printResult(result *SeedResult) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎉 初始化完成")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("\n📋 默认管理员账号信息：")
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ 用户名: %-55s │\n", result.User.UserName)
	fmt.Printf("│ 密码:   %-55s │\n", result.Password)
	fmt.Printf("│ 昵称:   %-55s │\n", result.User.Nickname)
	fmt.Printf("│ 邮箱:   %-55s │\n", result.User.Email)
	fmt.Printf("│ 手机:   %-55s │\n", result.User.Phone)
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")
	fmt.Printf("\n🏢 租户信息: %s (%s)\n", result.Tenant.Name, result.Tenant.TenantCode)
	fmt.Printf("🔑 角色: ")
	for i, role := range result.Roles {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s(%s)", role.Name, role.RoleCode)
	}
	fmt.Printf("\n\n📁 组织架构: %d个部门, %d个岗位\n", len(result.Departments), len(result.Positions))

	// 打印部门树结构
	fmt.Println("\n📂 部门结构:")
	printDepartmentTree(result.Departments, "", 0)

	// 打印岗位列表（按职级排序）
	fmt.Println("\n💼 岗位列表（按职级排序）:")
	printPositionList(result.Positions)

	// 打印字典信息
	fmt.Printf("\n📚 系统字典: 共 %d 个字典类型\n", len(result.DictTypes))
	printDictList(result.DictTypes)

	fmt.Println()
}

// printDepartmentTree 打印部门树
func printDepartmentTree(departments []model.Department, prefix string, level int) {
	// 找出根部门或指定父级的部门
	var children []model.Department
	if level == 0 {
		children = filterByParentID(departments, "")
	} else {
		return
	}

	for i, dept := range children {
		isLast := i == len(children)-1
		connector := "├──"
		if isLast {
			connector = "└──"
		}

		fmt.Printf("%s%s %s\n", prefix+connector, dept.DepartmentName, getDepartmentInfo(dept))

		// 获取子部门
		subChildren := filterByParentID(departments, dept.DepartmentID)
		if len(subChildren) > 0 {
			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			printDepartmentTreeRecursive(departments, dept.DepartmentID, newPrefix)
		}
	}
}

// printDepartmentTreeRecursive 递归打印部门树
func printDepartmentTreeRecursive(departments []model.Department, parentID string, prefix string) {
	children := filterByParentID(departments, parentID)

	for i, dept := range children {
		isLast := i == len(children)-1
		connector := "├──"
		if isLast {
			connector = "└──"
		}

		fmt.Printf("%s%s %s\n", prefix+connector, dept.DepartmentName, getDepartmentInfo(dept))

		// 获取子部门
		subChildren := filterByParentID(departments, dept.DepartmentID)
		if len(subChildren) > 0 {
			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			printDepartmentTreeRecursive(departments, dept.DepartmentID, newPrefix)
		}
	}
}

// filterByParentID 按父部门ID过滤
func filterByParentID(departments []model.Department, parentID string) []model.Department {
	var result []model.Department
	for _, dept := range departments {
		if dept.ParentID == parentID {
			result = append(result, dept)
		}
	}
	return result
}

// getDepartmentInfo 获取部门信息字符串
func getDepartmentInfo(dept model.Department) string {
	return fmt.Sprintf("[排序:%d]", dept.Sort)
}

// printPositionList 打印岗位列表
func printPositionList(positions []model.Position) {
	// 按职级降序排序
	sortedPositions := make([]model.Position, len(positions))
	copy(sortedPositions, positions)

	for i := 0; i < len(sortedPositions)-1; i++ {
		for j := i + 1; j < len(sortedPositions); j++ {
			if sortedPositions[i].Level < sortedPositions[j].Level {
				sortedPositions[i], sortedPositions[j] = sortedPositions[j], sortedPositions[i]
			}
		}
	}

	// 按职级分组打印
	currentLevel := sortedPositions[0].Level + 1
	for _, pos := range sortedPositions {
		if pos.Level < currentLevel {
			if currentLevel <= 100 {
				fmt.Printf("\n   职级 %d:\n", pos.Level)
			} else {
				fmt.Printf("\n   管理层:\n")
			}
			currentLevel = pos.Level
		}
		fmt.Printf("   • %s (%s) - L%d\n", pos.PositionName, pos.PositionCode, pos.Level)
	}
}

// printDictList 打印字典列表
func printDictList(dictTypes []model.DictType) {
	for _, dict := range dictTypes {
		fmt.Printf("   • %s (%s) - %s\n", dict.TypeName, dict.TypeCode, dict.Description)
	}
}
