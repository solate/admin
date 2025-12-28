# 角色权限系统设计

## 设计原则

- **Casbin 核心**：权限检查由 Casbin 处理，数据库只存储元数据
- **角色继承**：通过 Casbin `g2` 实现角色继承（无需 domain）
- **租户隔离**：角色定义按租户隔离，权限策略通过 `domain` 隔离
- **数据权限**：通过资源前缀控制（如 `user:mine`、`org:mine`）

---

## 数据模型

### 角色表 (roles)

```sql
CREATE TABLE roles (
    role_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    role_name VARCHAR(50) NOT NULL,
    role_code VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    sort INT DEFAULT 0,
    status SMALLINT DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY uk_tenant_code (tenant_id, role_code, deleted_at),
    INDEX idx_tenant (tenant_id, deleted_at)
);
```

### 权限点表 (permissions)

```sql
CREATE TABLE permissions (
    permission_id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(10) NOT NULL,          -- MENU, BUTTON, API
    resource VARCHAR(255) NOT NULL,     -- menu:xxx, btn:xxx, /api/v1/xxx
    action VARCHAR(50),                 -- GET, POST, * 等
    description VARCHAR(255),
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    INDEX idx_resource (resource, deleted_at)
);
```

**注意**：
- `permissions` 表是全局的，不按租户隔离
- 用于前端权限选择器和反查权限详情
- 实际权限检查只依赖 Casbin 策略

---

## Casbin 策略设计

### 模型配置

```conf
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _  # 用户-角色-租户
g2 = _, _    # 角色继承

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```

### 策略示例

```conf
# g 策略：用户角色绑定
g, user-001, sales, company-a
g, user-002, admin, company-a

# g2 策略：角色继承
g2, senior_sales, sales
g2, manager, sales

# p 策略：角色权限
p, sales, default, menu:orders, *
p, sales, default, btn:order_create, *
p, sales, default, /api/v1/orders, GET
p, admin, default, *, *
```

---

## 业务逻辑

### 1. 创建角色

```go
func (s *RoleService) Create(ctx context.Context, req *CreateRoleRequest) error {
    tenantID := getTenantID(ctx)

    // 验证编码唯一性
    existing, _ := s.roleRepo.GetByCode(ctx, tenantID, req.Code)
    if existing != nil {
        return errors.New("角色编码已存在")
    }

    role := &Role{
        RoleID:      uuid.New().String(),
        TenantID:    tenantID,
        RoleName:    req.Name,
        RoleCode:    req.Code,
        Description: req.Description,
        Status:      1,
    }

    return s.roleRepo.Create(ctx, role)
}
```

### 2. 分配权限

```go
func (s *RoleService) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
    role, _ := s.roleRepo.GetByID(ctx, roleID)
    if role == nil || role.TenantID != getTenantID(ctx) {
        return errors.New("角色不存在")
    }

    tenantCode := getTenantCode(ctx)

    // 清除旧权限
    s.enforcer.DeletePermissionsForUser(role.RoleCode, tenantCode)

    // 添加新权限
    for _, permID := range permissionIDs {
        perm, _ := s.permissionRepo.GetByID(ctx, permID)
        if perm != nil {
            s.enforcer.AddPolicy(role.RoleCode, tenantCode, perm.Resource, perm.Action)
        }
    }

    return nil
}
```

### 3. 给用户分配角色

```go
func (s *UserService) AssignRole(ctx context.Context, userID, roleID string) error {
    tenantCode := getTenantCode(ctx)

    user, _ := s.userRepo.GetByID(ctx, userID)
    if user == nil {
        return errors.New("用户不存在")
    }

    role, _ := s.roleRepo.GetByID(ctx, roleID)
    if role == nil || role.TenantID != getTenantID(ctx) {
        return errors.New("角色不存在")
    }

    // 写入 Casbin g 策略
    _, err := s.enforcer.AddRoleForUserInDomain(user.UserName, role.RoleCode, tenantCode)
    return err
}
```

### 4. 获取用户权限

```go
func (s *UserService) GetPermissions(ctx context.Context, userID string) ([]*Permission, error) {
    user, _ := s.userRepo.GetByID(ctx, userID)
    tenantCode := getTenantCode(ctx)

    // 获取用户所有角色（含继承）
    roles := s.enforcer.GetRolesForUserInDomain(user.UserName, tenantCode)
    allRoles := s.getAllRolesWithInheritance(roles)

    // 获取权限资源
    resources := make(map[string]string)
    for _, role := range allRoles {
        policies := s.enforcer.GetFilteredPolicy(0, role, tenantCode)
        for _, p := range policies {
            resources[p[2]] = p[3]
        }
    }

    // 反查权限详情
    var permissions []*Permission
    for resource, action := range resources {
        perm, _ := s.permissionRepo.GetByResource(ctx, resource)
        if perm != nil {
            permissions = append(permissions, perm)
        }
    }

    return permissions, nil
}

// 递归获取所有角色（含 g2 继承）
func (s *UserService) getAllRolesWithInheritance(roles []string) []string {
    var all []string
    visited := make(map[string]bool)

    var dfs func(role string)
    dfs = func(role string) {
        if visited[role] {
            return
        }
        visited[role] = true
        all = append(all, role)

        parents := s.enforcer.GetRolesForUser(role)
        for _, p := range parents {
            dfs(p)
        }
    }

    for _, role := range roles {
        dfs(role)
    }

    return all
}
```

### 5. 数据权限检查

```go
// 数据权限范围定义
const (
    DataScopeAll     = "all"      // 全部数据
    DataScopeMine    = "mine"     // 自己的数据
    DataScopeDept    = "dept"     // 本部门
    DataScopeDeptAndSub = "dept_sub" // 本部门及子部门
)

// 检查数据权限
func (s *UserService) CheckDataPermission(ctx context.Context, resource, targetUserID string) bool {
    user := getUserFromContext(ctx)
    tenantCode := getTenantCode(ctx)

    roles := s.enforcer.GetRolesForUserInDomain(user.UserName, tenantCode)

    // 检查是否有全部数据权限
    for _, role := range roles {
        allowed, _ := s.enforcer.Enforce(role, tenantCode, resource+":all", "*")
        if allowed {
            return true
        }
    }

    // 检查是否是自己
    targetUser, _ := s.userRepo.GetByID(ctx, targetUserID)
    if targetUser.UserID == user.UserID {
        return true
    }

    // 检查部门权限
    // ...

    return false
}
```

---

## API 设计

### 角色接口

```
GET    /api/v1/roles                    获取角色列表
POST   /api/v1/roles                    创建角色
GET    /api/v1/roles/:id                获取角色详情
PUT    /api/v1/roles/:id                更新角色
DELETE /api/v1/roles/:id                删除角色
POST   /api/v1/roles/:id/permissions    分配权限
GET    /api/v1/roles/:id/permissions    获取角色权限
```

### 权限接口

```
GET    /api/v1/permissions              获取所有权限点
POST   /api/v1/permissions              创建权限点
```

### 用户角色接口

```
POST   /api/v1/users/:id/roles          分配角色
DELETE /api/v1/users/:id/roles/:roleId  移除角色
GET    /api/v1/users/:id/roles          获取用户角色
GET    /api/v1/users/:id/permissions    获取用户权限
```

---

## Repository 层

```go
// 角色相关
func (r *RoleRepo) GetByCode(ctx context.Context, tenantID, code string) (*Role, error)
func (r *RoleRepo) ListByTenant(ctx context.Context, tenantID string) ([]*Role, error)

// 权限相关
func (r *PermissionRepo) GetByResource(ctx context.Context, resource string) (*Permission, error)
func (r *PermissionRepo) ListByType(ctx context.Context, permType string) ([]*Permission, error)
```

---

## 常量定义

```go
package constants

const (
    // 权限类型
    PermissionTypeMenu   = "MENU"
    PermissionTypeButton = "BUTTON"
    PermissionTypeAPI    = "API"

    // 数据权限范围
    DataScopeAll         = "all"
    DataScopeMine        = "mine"
    DataScopeDept        = "dept"
    DataScopeDeptAndSub  = "dept_sub"
)
```
