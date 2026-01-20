package seeds

import (
	"admin/internal/dal/model"
	"admin/pkg/passwordgen"
	"admin/pkg/rsapwd"
	"fmt"

	"gorm.io/gorm"
)

const (
	// DefaultAdminPassword 默认管理员密码
	DefaultAdminPassword = "admin@123"
)

// SeedUser 初始化默认管理员用户
func SeedUser(db *gorm.DB, userID string, tenantID string) (*model.User, error) {
	// 生成密码哈希
	// 前端会对密码进行 SHA256 哈希后再 RSA 加密传输
	// 所以数据库中存储的应该是 SHA256 哈希值的 Argon2 哈希
	sha256Hash := rsapwd.HashPassword(DefaultAdminPassword)

	salt, err := passwordgen.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("生成密码盐值失败: %w", err)
	}
	hashedPassword, err := passwordgen.Argon2Hash(sha256Hash, salt)
	if err != nil {
		return nil, fmt.Errorf("生成密码哈希失败: %w", err)
	}

	fmt.Printf("ℹ️  密码哈希信息:\n")
	fmt.Printf("   原始密码: %s\n", DefaultAdminPassword)
	fmt.Printf("   SHA256: %s\n", sha256Hash)
	fmt.Printf("   Argon2: %s\n", hashedPassword)

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
			Nickname: "默认管理员",
			Email:    email,
			Phone:    phone,
			Status:   1,
		}
		if err := db.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("创建默认管理员失败: %w", err)
		}
		fmt.Printf("✅ 默认管理员创建成功 user_id=%s username=%s\n", user.UserID, user.UserName)
	} else {
		// 用户已存在，更新密码
		user.Password = hashedPassword
		user.Email = email
		user.Phone = phone
		user.Status = 1
		if err := db.Save(&user).Error; err != nil {
			return nil, fmt.Errorf("更新默认管理员失败: %w", err)
		}
		fmt.Printf("ℹ️  默认管理员已存在，已更新密码 user_id=%s username=%s\n", user.UserID, user.UserName)
	}

	return &user, nil
}
