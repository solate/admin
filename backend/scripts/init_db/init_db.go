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
)

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

	fmt.Println("🚀 开始初始化默认租户与默认超管")

	// 生成所需的ID：租户ID、用户ID、角色ID
	ids, err := idgen.GenerateUUIDs(3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 生成ID失败: %v\n", err)
		os.Exit(1)
	}

	rawPassword, err := passwordgen.GeneratePassword(16)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 生成初始密码失败: %v\n", err)
		os.Exit(1)
	}
	salt, err := passwordgen.GenerateSalt()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 生成密码盐值失败: %v\n", err)
		os.Exit(1)
	}
	hashedPassword, err := passwordgen.Argon2Hash(rawPassword, salt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 生成密码哈希失败: %v\n", err)
		os.Exit(1)
	}

	// 创建或获取默认租户
	var tenant model.Tenant
	tenantID := ids[0]
	if err := db.Where("tenant_code = ?", constants.DefaultTenant).First(&tenant).Error; err != nil {
		// 租户不存在，创建新租户
		tenant = model.Tenant{
			TenantID:   tenantID,
			TenantCode: constants.DefaultTenant,
			Name:       "默认租户",
			Status:     1,
		}
		if err := db.Create(&tenant).Error; err != nil {
			fmt.Fprintf(os.Stderr, "❌ 创建默认租户失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ 默认租户创建成功 tenant_id=%s code=%s name=%s\n", tenant.TenantID, tenant.TenantCode, tenant.Name)
	} else {
		// 租户已存在，使用现有租户ID
		tenantID = tenant.TenantID
		fmt.Printf("ℹ️  默认租户已存在 tenant_id=%s code=%s name=%s\n", tenant.TenantID, tenant.TenantCode, tenant.Name)
	}

	// 创建或更新默认管理员用户
	userID := ids[1]
	email := "admin@example.com"
	phone := "13800000000"
	var user model.User
	if err := db.Where("user_name = ?", "admin").First(&user).Error; err != nil {
		// 用户不存在，创建新用户
		user = model.User{
			UserID:   userID,
			TenantID: tenantID,
			UserName: "admin",
			Password: hashedPassword,
			Name:     "默认管理员",
			Email:    &email,
			Phone:    &phone,
			Status:   1,
			RoleType: constants.RoleTypeSuperAdmin,
		}
		if err := db.Create(&user).Error; err != nil {
			fmt.Fprintf(os.Stderr, "❌ 创建默认超管失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ 默认超管创建成功 user_id=%s tenant_id=%s username=%s role_type=%d\n", user.UserID, user.TenantID, user.UserName, user.RoleType)
		fmt.Printf("🔑 初始密码（仅本次输出）: %s\n", rawPassword)
	} else {
		// 用户已存在，更新密码
		userID = user.UserID
		user.Password = hashedPassword
		user.Email = &email
		user.Phone = &phone
		user.Status = 1
		user.RoleType = constants.RoleTypeSuperAdmin
		if err := db.Save(&user).Error; err != nil {
			fmt.Fprintf(os.Stderr, "❌ 更新默认超管失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("ℹ️  默认超管已存在，已更新密码 user_id=%s tenant_id=%s username=%s\n", user.UserID, user.TenantID, user.UserName)
		fmt.Printf("🔑 更新后密码（仅本次输出）: %s\n", rawPassword)
	}

	// 创建超级管理员角色
	roleID := ids[2]
	var role model.Role
	if err := db.Where("role_code = ? AND tenant_id = ?", "super_admin", tenantID).First(&role).Error; err != nil {
		// 角色不存在，创建新角色
		role = model.Role{
			RoleID:     roleID,
			TenantID:   tenantID,
			RoleCode:   "super_admin",
			Name:       "超级管理员",
			Status:     1,
		}
		if err := db.Create(&role).Error; err != nil {
			fmt.Fprintf(os.Stderr, "❌ 创建超级管理员角色失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ 超级管理员角色创建成功 role_id=%s role_code=%s name=%s\n", role.RoleID, role.RoleCode, role.Name)
	} else {
		roleID = role.RoleID
		fmt.Printf("ℹ️  超级管理员角色已存在 role_id=%s role_code=%s\n", role.RoleID, role.RoleCode)
	}

	// 创建用户-租户-角色关系
	utrID, err := idgen.GenerateUUID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 生成用户租户角色ID失败: %v\n", err)
		os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "❌ 创建用户租户角色关系失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ 用户租户角色关系创建成功 utr_id=%s user_id=%s tenant_id=%s role_id=%s\n", userTenantRole.UserTenantRoleID, userTenantRole.UserID, userTenantRole.TenantID, userTenantRole.RoleID)
	} else {
		fmt.Printf("ℹ️  用户租户角色关系已存在 utr_id=%s\n", userTenantRole.UserTenantRoleID)
	}

	fmt.Println("🎉 初始化完成")

}
