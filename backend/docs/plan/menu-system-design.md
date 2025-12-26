# 多租户菜单系统设计

## 设计原则

- **超管属于 default 租户**：超管用户和角色模板都在 `default` 租户下
- **租户菜单边界**：通过 `tenant_menus` 表限制租户可访问的菜单范围
- **角色继承**：租户角色可继承 `default` 租户的角色模板，但受菜单边界限制
- **权限交集**：用户实际菜单 = 角色权限中的菜单 ∩ 租户分配的菜单

---

## 数据模型

### menus 表（菜单定义）

```sql
CREATE TABLE menus (
    menu_id VARCHAR(36) PRIMARY KEY,
    parent_id VARCHAR(36),
    name VARCHAR(100) NOT NULL,
    path VARCHAR(255),
    component VARCHAR(255),
    icon VARCHAR(50),
    sort INT DEFAULT 0,
    status TINYINT DEFAULT 1,
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    KEY idx_parent(parent_id, deleted_at)
);
```

### tenant_menus 表（租户菜单边界）

```sql
CREATE TABLE tenant_menus (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    menu_id VARCHAR(36) NOT NULL,
    created_at BIGINT,
    deleted_at BIGINT,
    UNIQUE KEY uk_tenant_menu(tenant_id, menu_id, deleted_at)
);
```

### roles 表（角色）

```sql
CREATE TABLE roles (
    role_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    status TINYINT DEFAULT 1,
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    UNIQUE KEY uk_tenant_code(tenant_id, code, deleted_at)
);
```

**注意**：`parent_id` 已删除，角色继承关系通过 Casbin `g2` 策略管理。

### permissions 表（权限点定义）

```sql
CREATE TABLE permissions (
    permission_id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,       -- MENU, BUTTON, API
    resource VARCHAR(255) NOT NULL,  -- menu:xxx, btn:xxx, /api/v1/xxx
    action VARCHAR(50),
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT
);
```

---

## Casbin 设计

### 数据存储分配

| 内容 | MySQL | Casbin | 说明 |
|------|-------|--------|------|
| 菜单定义 | menus | - | 元数据 |
| 租户菜单边界 | tenant_menus | - | 租户隔离 |
| 角色定义 | roles | - | 元数据 |
| 权限点定义 | permissions | - | 元数据，供前端选择和反查详情 |
| 角色权限关联 | - | p 策略 | role, domain, resource, action |
| 用户角色绑定 | - | g 策略 | user, role, domain |
| 角色继承 | - | g2 策略 | child_role, parent_role (无 domain) |

### permissions 表的用途

| 场景 | 使用方式 |
|------|----------|
| 前端权限选择器 | `List()` 返回所有权限点供用户选择 |
| 查询角色权限详情 | 从 Casbin 获取 `resource`，反查 permissions 表获取名称/描述 |
| 权限检查 | 不需要，直接用 Casbin `Enforce()` |

### Casbin 模型

```conf
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _  # 用户-角色-租户: (user, role, domain)
g2 = _, _    # 角色继承: (child_role, parent_role)

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```

### 策略示例

```conf
# g 策略：用户角色绑定 (user, role, domain)
g, user-001, tenant-a-sales, tenant-a
g, admin, super_admin, default

# g2 策略：角色继承 (child, parent) - 不需要 domain
g2, tenant-a-sales, sales

# p 策略：角色权限 (role, domain, resource, action)
p, sales, default, menu:orders, *
p, sales, default, btn:order_create, *
```

---

## 核心业务逻辑

### 超管给租户分配菜单

```go
func (s *TenantService) AssignMenus(ctx context.Context, tenantID string, menuIDs []string) error {
    // 验证是超管（user_type = 3）
    // 更新 tenant_menus 表
    s.tenantMenuRepo.DeleteByTenant(ctx, tenantID)
    for _, menuID := range menuIDs {
        s.tenantMenuRepo.Create(ctx, &TenantMenu{
            ID:       uuid.New(),
            TenantID: tenantID,
            MenuID:   menuID,
        })
    }
    return nil
}
```

### 租户创建继承角色

```go
func (s *RoleService) CreateRole(ctx context.Context, req *CreateRoleRequest) error {
    // 如果有父角色，验证父角色属于 default 租户的角色模板
    if req.ParentRoleCode != nil {
        parent, _ := s.roleRepo.GetByCode(ctx, *req.ParentRoleCode)
        if parent.TenantID != constants.DefaultTenantID {
            return errors.New("只能继承平台角色模板")
        }
    }

    // 创建角色
    role := &Role{
        RoleID:   uuid.New(),
        TenantID: getTenantID(ctx),
        Code:     req.Code,
        Name:     req.Name,
    }
    s.roleRepo.Create(ctx, role)

    // Casbin g2: 创建角色继承关系 (child, parent) - 注意 g2 不需要 domain
    if req.ParentRoleCode != nil {
        s.enforcer.AddGroupingPolicy(role.Code, *req.ParentRoleCode)

        // 复制父角色权限（受租户菜单边界限制）
        parentPolicies := s.enforcer.GetFilteredPolicy(0, *req.ParentRoleCode, constants.DefaultTenantCode)
        tenantMenuIDs, _ := s.tenantMenuRepo.GetMenuIDsByTenant(ctx, getTenantID(ctx))

        for _, policy := range parentPolicies {
            resource := policy[3]
            if strings.HasPrefix(resource, "menu:") {
                menuID := strings.TrimPrefix(resource, "menu:")
                if !contains(tenantMenuIDs, menuID) {
                    continue // 跳过未分配的菜单
                }
            }
            s.enforcer.AddPolicy(role.Code, getTenantCode(ctx), resource, "*")
        }
    }

    return nil
}
```

### 为角色分配权限

```go
func (s *RoleService) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
    role, _ := s.roleRepo.GetByID(ctx, roleID)
    tenantID := getTenantID(ctx)

    for _, permID := range permissionIDs {
        perm, _ := s.permissionRepo.GetByID(ctx, permID)

        // 菜单权限必须在租户分配范围内
        if perm.Type == "MENU" {
            menuID := strings.TrimPrefix(perm.Resource, "menu:")
            granted, _ := s.tenantMenuRepo.Exists(ctx, tenantID, menuID)
            if !granted {
                return fmt.Errorf("菜单未分配给租户")
            }
        }

        s.enforcer.AddPolicy(roleID, getTenantCode(ctx), perm.Resource, "*")
    }

    return nil
}
```

### 给用户分配角色

```go
func (s *UserService) AssignRole(ctx context.Context, userID, tenantCode, roleID string) error {
    // 直接写 Casbin
    _, err := s.enforcer.AddRoleForUserInDomain(userID, roleID, tenantCode)
    return err
}
```

### 获取用户菜单（核心）

```go
func (s *UserService) GetUserMenus(ctx context.Context, userID, tenantCode string) ([]*Menu, error) {
    // 1. 从 Casbin 获取用户的角色（g 策略）
    roles := s.enforcer.GetRolesForUserInDomain(userID, tenantCode)

    // 2. 递归获取所有角色（通过 g2 处理继承）
    allRoleCodes := s.getAllRoleCodes(roles)

    // 3. 从 Casbin 获取角色的菜单权限
    var menuIDs []string
    for _, roleCode := range allRoleCodes {
        policies := s.enforcer.GetFilteredPolicy(0, roleCode, tenantCode)
        for _, policy := range policies {
            resource := policy[3]
            if strings.HasPrefix(resource, "menu:") {
                menuIDs = append(menuIDs, strings.TrimPrefix(resource, "menu:"))
            }
        }
    }

    // 4. 【关键】与租户菜单边界取交集
    tenantMenuIDs, _ := s.tenantMenuRepo.GetMenuIDsByTenant(ctx, getTenantID(ctx))
    var validMenuIDs []string
    for _, menuID := range unique(menuIDs) {
        if contains(tenantMenuIDs, menuID) {
            validMenuIDs = append(validMenuIDs, menuID)
        }
    }

    // 5. 查询菜单详情并构建树
    menus, _ := s.menuRepo.GetByIDs(ctx, validMenuIDs)
    return buildMenuTree(menus), nil
}

// 递归获取所有角色（通过 g2 处理继承）
func (s *UserService) getAllRoleCodes(roleCodes []string) []string {
    var allCodes []string
    visited := make(map[string]bool)

    var dfs func(roleCode string)
    dfs = func(roleCode string) {
        if visited[roleCode] {
            return
        }
        visited[roleCode] = true
        allCodes = append(allCodes, roleCode)

        // 通过 g2 获取继承的父角色
        // GetRolesForUser 在 g2 中返回 (child -> parent)
        parents := s.enforcer.GetRolesForUser(roleCode)
        for _, parentCode := range parents {
            dfs(parentCode)
        }
    }

    for _, code := range roleCodes {
        dfs(code)
    }

    return allCodes
}
```

---

## 常量定义

```go
package constants

const (
    // 租户
    DefaultTenantID   = "tenant-default"
    DefaultTenantCode = "default"

    // 用户类型
    UserTypeUser        = 1 // 普通用户
    UserTypeTenantAdmin = 2 // 租户管理员
    UserTypeSuperAdmin  = 3 // 超级管理员

    // 权限类型
    PermissionTypeMenu   = "MENU"
    PermissionTypeButton = "BUTTON"
    PermissionTypeAPI    = "API"
)
```
