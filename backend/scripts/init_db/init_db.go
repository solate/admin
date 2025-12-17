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

	ids, err := idgen.GenerateUUIDs(2)
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

	tenant := model.Tenant{
		TenantID: ids[0],
		Code:     constants.DefaultTenant,
		Name:     "默认租户",
		Status:   1,
	}
	if err := db.Create(&tenant).Error; err != nil {
		fmt.Fprintf(os.Stderr, "❌ 创建默认租户失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 默认租户创建成功 tenant_id=%s code=%s\n", tenant.TenantID, tenant.Code)

	email := "admin@example.com"
	phone := "13800000000"
	user := model.User{
		UserID:   ids[1],
		TenantID: tenant.TenantID,
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

	fmt.Println("🎉 初始化完成")

}
