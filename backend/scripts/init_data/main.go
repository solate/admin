package main

import (
	"admin/internal/dal/model"
	"admin/pkg/config"
	"admin/pkg/database"
	"admin/pkg/idgen"
	"admin/scripts/init_data/seeds"
	"context"
	"fmt"
	"os"
	"strings"

	"gorm.io/gorm"
)

// SeedResult 初始化结果
type SeedResult struct {
	Tenant   model.Tenant
	User     model.User
	Roles    []model.Role
	Password string // 仅用于输出，不存储
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
			fmt.Fprintf(os.Stderr, "⚠️ 获取 sql.DB 失败: %v\n", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️ 关闭数据库连接失败: %v\n", err)
		}
	}()

	ctx := database.SkipTenantCheck(context.Background())
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
	ids, err := idgen.GenerateUUIDs(5)
	if err != nil {
		return nil, fmt.Errorf("生成ID失败: %w", err)
	}

	// 1. 初始化租户
	tenant, err := seeds.SeedTenant(db, ids[0])
	if err != nil {
		return nil, fmt.Errorf("初始化租户失败: %w", err)
	}
	result.Tenant = *tenant

	// 2. 初始化用户
	user, err := seeds.SeedUser(db, ids[1], tenant.TenantID)
	if err != nil {
		return nil, fmt.Errorf("初始化用户失败: %w", err)
	}
	result.User = *user

	// 3. 初始化角色
	roleDefs := seeds.DefaultRoleDefinitions(ids[2:5])
	roles, err := seeds.SeedRoles(db, roleDefs, tenant.TenantID)
	if err != nil {
		return nil, fmt.Errorf("初始化角色失败: %w", err)
	}
	result.Roles = roles

	// 4. 初始化 Casbin 表（如果不存在）
	if err := seeds.InitCasbinTable(db); err != nil {
		return nil, fmt.Errorf("初始化 Casbin 表失败: %w", err)
	}

	// 5. 初始化用户-角色关系（通过 Casbin）
	if err := seeds.SeedUserRoles(db, user.UserName, roles[0].RoleCode, tenant.TenantCode); err != nil {
		return nil, fmt.Errorf("初始化用户角色关系失败: %w", err)
	}

	// 6. 初始化系统菜单
	if err := seeds.SeedSystemMenus(db); err != nil {
		return nil, fmt.Errorf("初始化系统菜单失败: %w", err)
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
	fmt.Println()
	fmt.Println()
}
