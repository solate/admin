# 用户管理系统设计

## 设计原则

- **租户隔离**：每个租户独立管理用户，通过 `tenant_id` 隔离
- **用户名唯一性**：同一租户内用户名唯一 `(tenant_id, username)`
- **关联组织岗位**：用户归属部门和岗位，支持一人多岗
- **状态管理**：用户支持启用/禁用/锁定状态

---

## 数据模型

### 用户表 (users)

```sql
CREATE TABLE users (
    user_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    dept_id VARCHAR(20),                      -- 所属部门
    user_name VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    real_name VARCHAR(50),                    -- 真实姓名
    email VARCHAR(100),
    phone VARCHAR(20),
    avatar VARCHAR(255),                      -- 头像URL
    user_type SMALLINT DEFAULT 1,             -- 1:普通用户 2:租户管理员 3:超级管理员
    status SMALLINT DEFAULT 1,                -- 1:启用 0:禁用 2:锁定
    last_login_at BIGINT,
    last_login_ip VARCHAR(50),
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY uk_tenant_username (tenant_id, user_name, deleted_at),
    INDEX idx_tenant (tenant_id, deleted_at),
    INDEX idx_dept (dept_id, deleted_at)
);
```

### 用户岗位关联表 (user_positions)

```sql
CREATE TABLE user_positions (
    id VARCHAR(20) PRIMARY KEY,
    user_id VARCHAR(20) NOT NULL,
    position_id VARCHAR(20) NOT NULL,
    is_primary BOOLEAN DEFAULT TRUE,          -- 是否主岗位
    created_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY uk_user_position (user_id, position_id, deleted_at)
);
```

---

## 业务逻辑

### 1. 创建用户

```go
func (s *UserService) Create(ctx context.Context, req *CreateUserRequest) error {
    tenantID := getTenantID(ctx)

    // 验证用户名唯一性
    existing, _ := s.userRepo.GetByUsername(ctx, tenantID, req.Username)
    if existing != nil {
        return errors.New("用户名已存在")
    }

    // 密码加密
    hashedPassword, _ := bcrypt.Hash(req.Password)

    user := &User{
        UserID:    uuid.New().String(),
        TenantID:  tenantID,
        DeptID:    req.DeptID,
        UserName:  req.Username,
        Password:  hashedPassword,
        RealName:  req.RealName,
        Email:     req.Email,
        Phone:     req.Phone,
        UserType:  constants.UserTypeUser,
        Status:    constants.UserStatusEnabled,
    }

    return s.userRepo.Create(ctx, user)
}
```

### 2. 更新用户

```go
func (s *UserService) Update(ctx context.Context, userID string, req *UpdateUserRequest) error {
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil || user.TenantID != getTenantID(ctx) {
        return errors.New("用户不存在")
    }

    if req.RealName != nil {
        user.RealName = *req.RealName
    }
    if req.Email != nil {
        user.Email = *req.Email
    }
    if req.Phone != nil {
        user.Phone = *req.Phone
    }
    if req.DeptID != nil {
        user.DeptID = *req.DeptID
    }

    return s.userRepo.Update(ctx, user)
}
```

### 3. 重置密码

```go
func (s *UserService) ResetPassword(ctx context.Context, userID, newPassword string) error {
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return err
    }

    hashedPassword, _ := bcrypt.Hash(newPassword)
    user.Password = hashedPassword
    return s.userRepo.Update(ctx, user)
}
```

### 4. 启用/禁用用户

```go
func (s *UserService) SetStatus(ctx context.Context, userID string, status int) error {
    user, _ := s.userRepo.GetByID(ctx, userID)
    if user == nil {
        return errors.New("用户不存在")
    }

    user.Status = status
    return s.userRepo.Update(ctx, user)
}
```

### 5. 批量导入用户

```go
func (s *UserService) BatchImport(ctx context.Context, users []*ImportUserRequest) (*ImportResult, error) {
    var success, failed []string
    tenantID := getTenantID(ctx)

    for _, req := range users {
        // 验证用户名
        existing, _ := s.userRepo.GetByUsername(ctx, tenantID, req.Username)
        if existing != nil {
            failed = append(failed, req.Username)
            continue
        }

        user := &User{
            UserID:    uuid.New().String(),
            TenantID:  tenantID,
            UserName:  req.Username,
            Password:  mustHashPassword(req.Password), // 默认密码
            RealName:  req.RealName,
            Status:    constants.UserStatusEnabled,
        }

        if err := s.userRepo.Create(ctx, user); err != nil {
            failed = append(failed, req.Username)
        } else {
            success = append(success, req.Username)
        }
    }

    return &ImportResult{Success: success, Failed: failed}, nil
}
```

### 6. 分配岗位

```go
func (s *UserService) AssignPosition(ctx context.Context, userID, positionID string) error {
    user, _ := s.userRepo.GetByID(ctx, userID)
    if user == nil {
        return errors.New("用户不存在")
    }

    position, _ := s.positionRepo.GetByID(ctx, positionID)
    if position == nil || position.TenantID != getTenantID(ctx) {
        return errors.New("岗位不存在")
    }

    // 创建用户岗位关联
    userPos := &UserPosition{
        ID:         uuid.New().String(),
        UserID:     userID,
        PositionID: positionID,
        IsPrimary:  true,
    }

    // 更新用户主岗位
    user.PositionID = positionID
    s.userRepo.Update(ctx, user)

    return s.userPositionRepo.Create(ctx, userPos)
}
```

---

## API 设计

### 用户接口

```
GET    /api/v1/users                        获取用户列表（分页）
POST   /api/v1/users                        创建用户
GET    /api/v1/users/:id                    获取用户详情
PUT    /api/v1/users/:id                    更新用户
DELETE /api/v1/users/:id                    删除用户
POST   /api/v1/users/:id/password           重置密码
PUT    /api/v1/users/:id/status             设置用户状态
POST   /api/v1/users/import                 批量导入
GET    /api/v1/users/export                 导出用户
```

### 岗位接口

```
POST   /api/v1/users/:id/positions          分配岗位
DELETE /api/v1/users/:id/positions/:posId   移除岗位
PUT    /api/v1/users/:id/positions/:posId   设置主岗位
```

---

## Repository 层

```go
// 根据用户名查询
func (r *UserRepo) GetByUsername(ctx context.Context, tenantID, username string) (*User, error)

// 按部门查询用户
func (r *UserRepo) ListByDept(ctx context.Context, deptID string) ([]*User, error)

// 分页查询
func (r *UserRepo) List(ctx context.Context, req *ListRequest) ([]*User, int64, error)

// 搜索用户（姓名/邮箱/手机号）
func (r *UserRepo) Search(ctx context.Context, keyword string) ([]*User, error)
```

---

## 常量定义

```go
package constants

const (
    // 用户状态
    UserStatusEnabled = 1  // 启用
    UserStatusDisabled = 0 // 禁用
    UserStatusLocked = 2   // 锁定

    // 默认密码
    DefaultPassword = "123456"
)
```
