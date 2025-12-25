package main

import (
	"admin/internal/dal/model"
	"admin/pkg/config"
	"admin/pkg/constants"
	"admin/pkg/database"
	"admin/pkg/idgen"
	"admin/pkg/passwordgen"
	"context"
	"fmt"
	"os"
	"strings"

	"gorm.io/gorm"
)

const (
	// DefaultAdminPassword 默认管理员密码
	DefaultAdminPassword = "admin@123"
)

// SeedResult 初始化结果
type SeedResult struct {
	Tenant         model.Tenant
	User           model.User
	Roles          []model.Role
	UserTenantRole model.UserTenantRole
	Password       string // 仅用于输出，不存储
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
		Password: DefaultAdminPassword,
	}

	// 生成所需的ID
	ids, err := idgen.GenerateUUIDs(5)
	if err != nil {
		return nil, fmt.Errorf("生成ID失败: %w", err)
	}

	// 1. 初始化租户
	if err := seedTenant(db, ids[0], result); err != nil {
		return nil, fmt.Errorf("初始化租户失败: %w", err)
	}

	// 2. 初始化用户
	if err := seedUser(db, ids[1], result); err != nil {
		return nil, fmt.Errorf("初始化用户失败: %w", err)
	}

	// 3. 初始化角色
	if err := seedRoles(db, ids[2:5], result); err != nil {
		return nil, fmt.Errorf("初始化角色失败: %w", err)
	}

	// 4. 初始化用户-租户-角色关系
	if err := seedUserTenantRole(db, ids[1], ids[0], ids[2], result); err != nil {
		return nil, fmt.Errorf("初始化用户租户角色关系失败: %w", err)
	}

	// 5. 初始化系统菜单
	if err := SeedSystemMenus(db); err != nil {
		return nil, fmt.Errorf("初始化系统菜单失败: %w", err)
	}

	return result, nil
}

// seedTenant 初始化租户
func seedTenant(db *gorm.DB, tenantID string, result *SeedResult) error {
	var tenant model.Tenant
	if err := db.Where("tenant_code = ?", constants.DefaultTenant).First(&tenant).Error; err != nil {
		// 租户不存在，创建新租户
		tenant = model.Tenant{
			TenantID:   tenantID,
			TenantCode: constants.DefaultTenant,
			Name:       "默认租户",
			Status:     1,
		}
		if err := db.Create(&tenant).Error; err != nil {
			return fmt.Errorf("创建默认租户失败: %w", err)
		}
		fmt.Printf("✅ 默认租户创建成功 tenant_id=%s code=%s name=%s\n", tenant.TenantID, tenant.TenantCode, tenant.Name)
	} else {
		fmt.Printf("ℹ️  默认租户已存在 tenant_id=%s code=%s name=%s\n", tenant.TenantID, tenant.TenantCode, tenant.Name)
	}

	result.Tenant = tenant
	return nil
}

// seedUser 初始化用户
func seedUser(db *gorm.DB, userID string, result *SeedResult) error {
	// 生成密码哈希
	salt, err := passwordgen.GenerateSalt()
	if err != nil {
		return fmt.Errorf("生成密码盐值失败: %w", err)
	}
	hashedPassword, err := passwordgen.Argon2Hash(DefaultAdminPassword, salt)
	if err != nil {
		return fmt.Errorf("生成密码哈希失败: %w", err)
	}

	email := "admin@example.com"
	phone := "13800000000"
	var user model.User
	if err := db.Where("user_name = ?", "admin").First(&user).Error; err != nil {
		// 用户不存在，创建新用户
		user = model.User{
			UserID:   userID,
			UserName: "admin",
			Password: hashedPassword,
			Name:     "默认管理员",
			Email:    &email,
			Phone:    &phone,
			Status:   1,
		}
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("创建默认管理员失败: %w", err)
		}
		fmt.Printf("✅ 默认管理员创建成功 user_id=%s username=%s\n", user.UserID, user.UserName)
	} else {
		// 用户已存在，更新密码
		user.Password = hashedPassword
		user.Email = &email
		user.Phone = &phone
		user.Status = 1
		if err := db.Save(&user).Error; err != nil {
			return fmt.Errorf("更新默认管理员失败: %w", err)
		}
		fmt.Printf("ℹ️  默认管理员已存在，已更新密码 user_id=%s username=%s\n", user.UserID, user.UserName)
	}

	result.User = user
	return nil
}

// seedRoles 初始化角色
func seedRoles(db *gorm.DB, roleIDs []string, result *SeedResult) error {
	roleDefinitions := []struct {
		roleID   string
		roleCode string
		name     string
	}{
		{roleIDs[0], "super_admin", "超级管理员"},
		{roleIDs[1], "admin", "租户管理员"},
		{roleIDs[2], "user", "普通用户"},
	}

	roles := make([]model.Role, 0, len(roleDefinitions))
	for _, def := range roleDefinitions {
		var role model.Role
		if err := db.Where("role_code = ? AND tenant_id = ?", def.roleCode, result.Tenant.TenantID).First(&role).Error; err != nil {
			// 角色不存在，创建新角色
			role = model.Role{
				RoleID:   def.roleID,
				TenantID: result.Tenant.TenantID,
				RoleCode: def.roleCode,
				Name:     def.name,
				Status:   1,
			}
			if err := db.Create(&role).Error; err != nil {
				return fmt.Errorf("创建角色 %s 失败: %w", def.name, err)
			}
			fmt.Printf("✅ 角色创建成功 role_id=%s role_code=%s name=%s\n", role.RoleID, role.RoleCode, role.Name)
		} else {
			fmt.Printf("ℹ️  角色已存在 role_id=%s role_code=%s name=%s\n", role.RoleID, role.RoleCode, role.Name)
		}
		roles = append(roles, role)
	}

	result.Roles = roles
	return nil
}

// seedUserTenantRole 初始化用户-租户-角色关系
func seedUserTenantRole(db *gorm.DB, userID, tenantID, roleID string, result *SeedResult) error {
	utrID, err := idgen.GenerateUUID()
	if err != nil {
		return fmt.Errorf("生成用户租户角色ID失败: %w", err)
	}

	var userTenantRole model.UserTenantRole
	if err := db.Where("user_id = ? AND tenant_id = ? AND role_id = ?", userID, tenantID, roleID).First(&userTenantRole).Error; err != nil {
		// 关系不存在，创建新关系
		userTenantRole = model.UserTenantRole{
			UserTenantRoleID: utrID,
			UserID:           userID,
			TenantID:         tenantID,
			RoleID:           roleID,
		}
		if err := db.Create(&userTenantRole).Error; err != nil {
			return fmt.Errorf("创建用户租户角色关系失败: %w", err)
		}
		fmt.Printf("✅ 用户租户角色关系创建成功 utr_id=%s user_id=%s tenant_id=%s role_id=%s\n",
			userTenantRole.UserTenantRoleID, userTenantRole.UserID, userTenantRole.TenantID, userTenantRole.RoleID)
	} else {
		fmt.Printf("ℹ️  用户租户角色关系已存在 utr_id=%s\n", userTenantRole.UserTenantRoleID)
	}

	result.UserTenantRole = userTenantRole
	return nil
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
	fmt.Printf("│ 姓名:   %-55s │\n", result.User.Name)
	fmt.Printf("│ 邮箱:   %-55s │\n", *result.User.Email)
	fmt.Printf("│ 手机:   %-55s │\n", *result.User.Phone)
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
