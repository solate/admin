package seeds

import (
	"admin/internal/dal/model"
	"admin/pkg/passwordgen"
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
	salt, err := passwordgen.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("生成密码盐值失败: %w", err)
	}
	hashedPassword, err := passwordgen.Argon2Hash(DefaultAdminPassword, salt)
	if err != nil {
		return nil, fmt.Errorf("生成密码哈希失败: %w", err)
	}

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
