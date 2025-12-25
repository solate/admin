# 多租户 SaaS 菜单系统设计文档

## 一、需求描述

### 1.1 业务需求

多租户 SaaS 系统，需要实现灵活的菜单和权限管理：

1. **超管（平台方）**：管理所有菜单和角色模板
2. **租户（客户方）**：继承超管角色，在分配的菜单范围内配置权限
3. **用户（终端用户）**：只能看到租户授权的菜单和按钮

### 1.2 核心流程

```
超管操作：
1. 创建所有系统菜单
2. 创建角色模板（如：销售、财务、管理员）
3. 为角色模板配置菜单和按钮权限
4. 创建租户时，给租户分配可用菜单（子集）

租户操作：
5. 创建角色，继承超管的角色模板
6. （可选）调整角色的菜单权限（不能超出租户已分配范围）
7. 给用户分配角色

用户登录：
8. 获取用户的角色 → 获取权限 → 与租户菜单边界取交集 → 返回菜单树
```

### 1.3 关键约束

- **租户菜单边界**：租户只能访问超管分配的菜单（`tenant_menus` 表）
- **角色继承**：租户角色可继承超管角色模板，但受租户菜单边界限制
- **权限交集**：用户实际权限 = 角色权限 ∩ 租户分配菜单

---

## 二、数据模型设计

### 2.1 表结构

#### 2.1.1 menus 表（菜单定义）

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
    KEY idx_parent(parent_id, deleted_at),
    KEY idx_sort(sort)
);
```

#### 2.1.2 tenant_menus 表（租户菜单边界）

```sql
CREATE TABLE tenant_menus (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    menu_id VARCHAR(36) NOT NULL,
    created_at BIGINT,
    deleted_at BIGINT,
    UNIQUE KEY uk_tenant_menu(tenant_id, menu_id, deleted_at),
    KEY idx_tenant(tenant_id, deleted_at)
);
```

#### 2.1.3 roles 表（角色）

```sql
CREATE TABLE roles (
    role_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    parent_id VARCHAR(36),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50),
    description TEXT,
    status TINYINT DEFAULT 1,
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    KEY idx_tenant(tenant_id, deleted_at),
    KEY idx_parent(parent_id, deleted_at)
);
```

### 2.2 Casbin 策略存储

以下内容用 Casbin 存储，**不需要建表**：

| 内容 | Casbin 策略类型 | 示例 |
|------|-----------------|------|
| 角色菜单权限 | p | `p, role_sales, tenant_a, menu:orders, *` |
| 角色按钮权限 | p | `p, role_sales, tenant_a, btn:order_create, *` |
| 角色API权限 | p | `p, role_sales, tenant_a, /api/v1/orders, GET` |
| 用户角色绑定 | g | `g, user_001, role_sales, tenant_a` |
| 角色继承 | g | `g, role_tenant_sales, role_sales, tenant_a` |

### 2.3 Casbin 模型配置

```conf
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj)
```

---

## 三、数据示例

### 3.1 超管准备数据

```sql
-- 1. 超管创建菜单
INSERT INTO menus VALUES
('m1', NULL, '工作台', '/dashboard', 'Dashboard.vue', 'Dashboard', 1, 1, ...),
('m2', NULL, '用户管理', '/users', 'Users.vue', 'User', 2, 1, ...),
('m3', NULL, '订单管理', '/orders', 'Orders.vue', 'Shopping', 3, 1, ...),
('m4', NULL, '财务管理', '/finance', 'Finance.vue', 'Money', 4, 1, ...);

-- 2. 超管创建角色模板（tenant_id = 'super_admin'）
INSERT INTO roles VALUES
('r_sales', 'super_admin', NULL, '销售角色', 'sales', '销售可访问订单', 1, ...),
('r_finance', 'super_admin', NULL, '财务角色', 'finance', '财务可访问订单和财务', 1, ...);

-- 3. Casbin: 为角色配置菜单权限
p, r_sales, super_admin, menu:dashboard, *
p, r_sales, super_admin, menu:orders, *
p, r_finance, super_admin, menu:dashboard, *
p, r_finance, super_admin, menu:orders, *
p, r_finance, super_admin, menu:finance, *
```

### 3.2 超管给租户分配菜单

```sql
-- 创建租户
INSERT INTO tenants VALUES ('tenant_a', 'tenant_a', '租户A', ...);

-- 给租户A分配菜单：工作台 + 订单（不给用户管理和财务）
INSERT INTO tenant_menus VALUES
('tm1', 'tenant_a', 'm1', ...),  -- 工作台
('tm2', 'tenant_a', 'm3', ...);  -- 订单管理
```

### 3.3 租户创建继承角色

```sql
-- 租户A创建继承销售的角色
INSERT INTO roles VALUES
('r_tenant_a_sales', 'tenant_a', 'r_sales', '租户A销售', 'tenant_a_sales', ...);

-- Casbin: 创建角色继承关系
g, r_tenant_a_sales, r_sales, tenant_a
```

### 3.4 租户给用户分配角色

```sql
-- Casbin: 用户角色绑定
g, user_001, r_tenant_a_sales, tenant_a
```

---

## 四、核心业务逻辑

### 4.1 超管给租户分配菜单

```go
func (s *TenantService) AssignMenus(ctx context.Context, tenantID string, menuIDs []string) error {
    // 1. 验证是超管
    if !isSuperAdmin(ctx) {
        return ErrPermissionDenied
    }

    // 2. 验证菜单存在
    menus, _ := s.menuRepo.GetByIDs(ctx, menuIDs)
    if len(menus) != len(menuIDs) {
        return ErrMenuNotFound
    }

    // 3. 更新 tenant_menus 表（先删后增）
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

### 4.2 租户创建继承角色

```go
func (s *RoleService) CreateRole(ctx context.Context, req *CreateRoleRequest) error {
    // 1. 验证父角色存在且属于超管租户
    if req.ParentID != nil {
        parent, _ := s.roleRepo.GetByID(ctx, *req.ParentID)
        if parent == nil || parent.TenantID != "super_admin" {
            return ErrInvalidParentRole
        }
    }

    // 2. 创建角色
    role := &Role{
        RoleID:   uuid.New(),
        TenantID: getTenantID(ctx),
        ParentID: req.ParentID,
        Name:     req.Name,
        Code:     req.Code,
    }
    s.roleRepo.Create(ctx, role)

    // 3. Casbin: 创建角色继承关系
    if req.ParentID != nil {
        s.enforcer.AddGroupingPolicy(role.RoleID, *req.ParentID, getTenantID(ctx))
    }

    // 4. 复制父角色权限（但受租户菜单边界限制）
    if req.ParentID != nil {
        parentPerms := s.enforcer.GetFilteredPolicy(1, *req.ParentID, getTenantID(ctx))
        tenantMenuIDs, _ := s.tenantMenuRepo.GetMenuIDsByTenant(ctx, getTenantID(ctx))

        for _, perm := range parentPerms {
            // 提取menu_id并验证在租户分配范围内
            menuID := extractMenuID(perm[2])
            if contains(tenantMenuIDs, menuID) {
                s.enforcer.AddPolicy(role.RoleID, getTenantID(ctx), perm[2], perm[3])
            }
        }
    }

    return nil
}

func extractMenuID(obj string) string {
    // obj格式: menu:xxx 或 btn:xxx
    parts := strings.Split(obj, ":")
    if len(parts) >= 2 && parts[0] == "menu" {
        return parts[1]
    }
    return ""
}
```

### 4.3 获取用户菜单（核心）

```go
func (s *UserService) GetUserMenus(ctx context.Context, userID, tenantID string) ([]*Menu, error) {
    // 1. 通过 Casbin 获取用户的所有角色（包括继承的）
    roles := s.enforcer.GetRolesForUserInDomain(userID, tenantID)

    // 2. 递归获取所有角色ID（处理角色继承）
    allRoleIDs := s.getAllRoleIDs(roles, tenantID)

    // 3. 通过 Casbin 获取角色的菜单权限
    var menuIDs []string
    for _, roleID := range allRoleIDs {
        perms := s.enforcer.GetFilteredPolicy(0, roleID, tenantID)
        for _, perm := range perms {
            obj := perm[2]
            if strings.HasPrefix(obj, "menu:") {
                menuID := strings.TrimPrefix(obj, "menu:")
                menuIDs = append(menuIDs, menuID)
            }
        }
    }

    // 4. 【关键】从数据库获取租户菜单边界
    tenantMenuIDs, _ := s.tenantMenuRepo.GetMenuIDsByTenant(ctx, tenantID)

    // 5. 取交集：权限中的菜单 ∩ 租户分配的菜单
    var validMenuIDs []string
    for _, menuID := range unique(menuIDs) {
        if contains(tenantMenuIDs, menuID) {
            validMenuIDs = append(validMenuIDs, menuID)
        }
    }

    // 6. 查询菜单详情并构建树
    menus, _ := s.menuRepo.GetByIDs(ctx, validMenuIDs)
    return s.buildMenuTree(menus), nil
}

// 递归获取所有角色ID（处理继承）
func (s *UserService) getAllRoleIDs(roles []string, tenantID string) []string {
    var allRoleIDs []string
    visited := make(map[string]bool)

    var dfs func(roleID string)
    dfs = func(roleID string) {
        if visited[roleID] {
            return
        }
        visited[roleID] = true
        allRoleIDs = append(allRoleIDs, roleID)

        // 获取该角色继承的角色
        inherited := s.enforcer.GetRolesForUserInDomain(roleID, tenantID)
        for _, childRoleID := range inherited {
            dfs(childRoleID)
        }
    }

    for _, roleID := range roles {
        dfs(roleID)
    }

    return allRoleIDs
}
```

### 4.4 检查按钮权限

```go
func (s *UserService) HasButtonPermission(ctx context.Context, userID, tenantID, buttonCode string) bool {
    // 直接用 Casbin 检查
    return s.enforcer.Enforce(userID, tenantID, "btn:"+buttonCode, "*")
}
```

### 4.5 菜单树构建

```go
func (s *UserService) buildMenuTree(menus []*Menu) []*MenuNode {
    nodeMap := make(map[string]*MenuNode)
    for _, m := range menus {
        nodeMap[m.MenuID] = &MenuNode{
            Menu:    m,
            Children: []*MenuNode{},
        }
    }

    var roots []*MenuNode
    for _, m := range menus {
        node := nodeMap[m.MenuID]
        if m.ParentID == nil || *m.ParentID == "" {
            roots = append(roots, node)
        } else if parent, exists := nodeMap[*m.ParentID]; exists {
            parent.Children = append(parent.Children, node)
        }
    }

    return roots
}
```

---

## 五、API 接口设计

### 5.1 超管接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/system/tenants/:id/menus` | 给租户分配菜单 |
| GET | `/api/v1/system/tenants/:id/menus` | 查看租户已分配菜单 |

### 5.2 租户接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/roles` | 创建角色（支持继承） |
| GET | `/api/v1/roles` | 角色列表 |
| PUT | `/api/v1/roles/:id/menus` | 配置角色菜单 |
| GET | `/api/v1/users/:id/menus` | 获取用户菜单树 |
| GET | `/api/v1/users/:id/buttons` | 获取用户按钮权限 |

---

## 六、常量定义

```go
package constants

const (
    // 租户
    DefaultTenantCode = "default"  // 超管租户编码
    SuperAdminTenantID = "super_admin"  // 超管租户ID

    // 菜单状态
    MenuStatusEnabled  = 1
    MenuStatusDisabled = 2
)
```

---

## 七、实现要点

### 7.1 权限存储分配

| 内容 | 存储方式 |
|------|---------|
| menus | MySQL 表 |
| tenant_menus | MySQL 表 |
| roles | MySQL 表 |
| 角色菜单权限 | Casbin (p 策略) |
| 角色按钮权限 | Casbin (p 策略) |
| 用户角色绑定 | Casbin (g 策略) |
| 角色继承 | Casbin (g 策略) |

### 7.2 核心公式

```
用户实际菜单 = Casbin权限中的菜单 ∩ 数据库tenant_menus中的菜单
```

### 7.3 关键校验

1. 创建角色时：父角色必须属于超管租户
2. 分配权限时：菜单必须在租户的 `tenant_menus` 范围内
3. 获取菜单时：必须取角色权限和租户分配的交集
