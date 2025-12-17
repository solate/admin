



func main() {
    db, _ := gorm.Open(...)
    tenant.RegisterCallbacks(db)
}

// 2. 业务代码中设置租户ID
func GetUser(ctx context.Context, userID string) (*User, error) {
    // 自动带上租户隔离条件
    ctx = tenant.WithTenantID(ctx, "tenant-123")
    var user User
    err := db.WithContext(ctx).First(&user, userID).Error
    return &user, err
}

// 3. 特殊场景跳过租户检查
func AdminGetAllUsers(ctx context.Context) ([]User, error) {
    // 显式跳过租户检查
    ctx = tenant.SkipTenantCheck(ctx)
    var users []User
    err := db.WithContext(ctx).Find(&users).Error
    return users, err
}