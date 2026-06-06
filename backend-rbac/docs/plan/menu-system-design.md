# 多租户菜单系统设计

## 设计原则

- **超管属于 default 租户**：超管用户和角色模板都在 `default` 租户下
- **角色直接关联菜单**：通过 Casbin p 策略直接配置角色可访问的菜单，无需租户级别边界控制
- **角色继承（实时计算）**：租户角色可继承 `default` 租户的角色模板及其菜单权限，权限实时计算无需复制
- **API 权限自动关联**：为角色分配菜单权限时，自动关联菜单的 API 权限，确保前后端权限一致
- **权限简化**：用户实际菜单 = 角色权限中的菜单（无需取交集）
- **级联删除**：删除菜单/角色/用户时，自动清理相关的 Casbin 策略

---

## 数据模型

### menus 表（菜单定义）

```sql
CREATE TABLE menus (
    menu_id VARCHAR(20) PRIMARY KEY,
    parent_id VARCHAR(20),
    name VARCHAR(100) NOT NULL,
    path VARCHAR(255),
    component VARCHAR(255),
    redirect VARCHAR(255),
    icon VARCHAR(50),
    sort INT DEFAULT 0,
    status SMALLINT DEFAULT 1,
    api_paths TEXT NOT NULL DEFAULT '',  -- 新增：关联的 API 路径（JSON 格式）
    description TEXT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    KEY idx_parent(parent_id, deleted_at)
);
```

**api_paths 字段说明**：
- 类型：TEXT，存储 JSON 数组
- 用途：定义菜单关联的 API 路径，实现前后端权限联动
- 格式：`[{"path": "/api/v1/users", "methods": ["GET", "POST"]}]`
- 作用：为角色分配菜单权限时，自动关联这些 API 路径的访问权限

### roles 表（角色）

```sql
CREATE TABLE roles (
    role_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    role_code VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    status SMALLINT DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY uk_tenant_code(tenant_id, role_code, deleted_at)
);
```

**注意**：`parent_id` 已删除，角色继承关系通过 Casbin `g2` 策略管理。

### permissions 表（权限点定义）

```sql
CREATE TABLE permissions (
    permission_id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(10) NOT NULL,       -- MENU, BUTTON, API
    resource VARCHAR(255) NOT NULL,  -- menu:xxx, btn:xxx, /api/v1/xxx
    action VARCHAR(50),
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);
```

---

## Casbin 设计

### 数据存储分配

| 内容 | MySQL | Casbin | 说明 |
|------|-------|--------|------|
| 菜单定义 | menus | - | 元数据 |
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

### 超管创建角色模板（default 租户）

```go
func (s *RoleService) CreateRoleTemplate(ctx context.Context, req *CreateRoleRequest) error {
    // 验证是超管（user_type = 3）
    // 创建角色模板
    role := &Role{
        RoleID:   uuid.New(),
        TenantID: constants.DefaultTenantID,
        Code:     req.Code,
        Name:     req.Name,
    }
    s.roleRepo.Create(ctx, role)

    // 分配菜单权限到 Casbin
    for _, menuID := range req.MenuIDs {
        s.enforcer.AddPolicy(role.Code, constants.DefaultTenantCode, "menu:"+menuID, "*")
    }

    return nil
}
```

### 租户创建继承角色

```go
func (s *RoleService) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
    tenantID := xcontext.GetTenantID(ctx)

    // 检查角色编码是否已存在（租户内唯一）
    exists, err := s.roleRepo.CheckExists(ctx, tenantID, req.RoleCode)
    if exists {
        return nil, xerr.ErrRoleCodeExists
    }

    // 如果有父角色，验证父角色属于 default 租户的角色模板
    var parentRole *model.Role
    if req.ParentRoleCode != nil {
        defaultTenantID := s.tenantCache.GetDefaultTenantID()
        parentRole, err = s.roleRepo.GetByCodeWithTenant(ctx, defaultTenantID, *req.ParentRoleCode)
        if err != nil {
            return nil, xerr.New(xerr.ErrInvalidParams.Code, "父角色不存在或只能继承 default 租户的角色模板")
        }
    }

    // 创建角色
    role := &model.Role{
        RoleID:   idgen.GenerateUUID(),
        TenantID: tenantID,
        RoleCode: req.RoleCode,
        Name:     req.Name,
    }
    s.roleRepo.Create(ctx, role)

    // 如果有父角色，建立继承关系（不复制权限，实时计算）
    if req.ParentRoleCode != nil && parentRole != nil {
        // 创建 Casbin g2 策略（角色继承，不需要 domain）
        s.enforcer.AddGroupingPolicy(role.RoleCode, *req.ParentRoleCode)
    }

    return s.toRoleResponse(role, req.ParentRoleCode), nil
}
```

**重要变更**：创建角色时不再复制父角色的权限，而是：
1. 只建立 g2 继承关系
2. 获取用户菜单时，通过应用层实时查询 default domain 的权限
3. 提供可选的"固化权限"功能，将继承的权限复制到当前租户并断开继承关系

### 为角色分配权限（菜单+按钮）

```go
// AssignPermissions 为角色分配权限（菜单+按钮）
// 策略格式: p, role_code, domain, resource, action
func (s *RoleService) AssignPermissions(ctx context.Context, roleID string, req *dto.AssignPermissionsRequest) error {
    // 1. 查询角色信息
    role, err := s.roleRepo.GetByID(ctx, roleID)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return xerr.ErrRoleNotFound
        }
        return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
    }

    // 2. 获取租户代码
    tenantCode := xcontext.GetTenantCode(ctx)
    if tenantCode == "" {
        return xerr.ErrUnauthorized
    }

    // 3. 清除角色的所有菜单权限
    _, err = s.enforcer.RemoveFilteredPolicy(0, role.RoleCode, tenantCode, "menu:", "*")
    if err != nil {
        return xerr.Wrap(xerr.ErrInternal.Code, "清除旧菜单权限失败", err)
    }

    // 4. 清除角色的所有按钮权限
    _, err = s.enforcer.RemoveFilteredPolicy(0, role.RoleCode, tenantCode, "btn:", "*")
    if err != nil {
        return xerr.Wrap(xerr.ErrInternal.Code, "清除旧按钮权限失败", err)
    }

    // 5. 添加新的菜单权限
    for _, menuID := range req.MenuIDs {
        // 策略格式: p, role_code, domain, menu:menu_id, *
        _, err := s.enforcer.AddPolicy(role.RoleCode, tenantCode, "menu:"+menuID, "*")
        if err != nil {
            return xerr.Wrap(xerr.ErrInternal.Code, "添加菜单权限失败", err)
        }
    }

    // 6. 添加新的按钮权限
    for _, buttonID := range req.ButtonIDs {
        // 根据 permission_id 查询 resource
        perm, err := s.permissionRepo.GetByID(ctx, buttonID)
        if err != nil {
            continue // 跳过不存在的权限
        }
        // 策略格式: p, role_code, domain, btn:menuID:action, *
        _, err = s.enforcer.AddPolicy(role.RoleCode, tenantCode, perm.Resource, "*")
        if err != nil {
            return xerr.Wrap(xerr.ErrInternal.Code, "添加按钮权限失败", err)
        }
    }

    // 7. 记录操作日志
    auditlog.RecordUpdate(ctx, constants.ModuleRole, constants.ResourceTypeRole, role.RoleID, role.Name+"-权限", nil, map[string]interface{}{
        "menu_ids":   req.MenuIDs,
        "button_ids": req.ButtonIDs,
    })

    return nil
}

// GetRolePermissions 获取角色的所有权限（菜单+按钮）
func (s *RoleService) GetRolePermissions(ctx context.Context, roleID string) (*dto.RolePermissionsResponse, error) {
    // 1. 查询角色信息
    role, err := s.roleRepo.GetByID(ctx, roleID)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, xerr.ErrRoleNotFound
        }
        return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
    }

    // 2. 获取租户代码
    tenantCode := xcontext.GetTenantCode(ctx)
    if tenantCode == "" {
        return nil, xerr.ErrUnauthorized
    }

    // 3. 从 Casbin 获取角色的所有权限
    policies, _ := s.enforcer.GetFilteredPolicy(0, role.RoleCode, tenantCode)

    var menuIDs []string
    var buttonResources []string

    for _, policy := range policies {
        if len(policy) >= 4 {
            resource := policy[2]
            if strings.HasPrefix(resource, "menu:") {
                menuID := strings.TrimPrefix(resource, "menu:")
                menuIDs = append(menuIDs, menuID)
            } else if strings.HasPrefix(resource, "btn:") {
                buttonResources = append(buttonResources, resource)
            }
        }
    }

    // 4. 根据按钮 resource 查询 permission_id
    var buttonIDs []string
    for _, resource := range buttonResources {
        perm, err := s.permissionRepo.GetByResource(ctx, resource)
        if err == nil {
            buttonIDs = append(buttonIDs, perm.PermissionID)
        }
    }

    return &dto.RolePermissionsResponse{
        MenuIDs:   menuIDs,
        ButtonIDs: buttonIDs,
    }, nil
}
```

**API 变更**：
- 旧接口：`PUT /api/v1/roles/:role_id/menus` → 已弃用，保留向后兼容
- 新接口：`PUT /api/v1/roles/:role_id/permissions` → 统一管理菜单和按钮权限

**请求体示例**：
```json
{
  "menu_ids": ["menu_dashboard", "menu_users", "menu_orders"],
  "button_ids": ["btn_user_create", "btn_user_delete"]
}
```

### 给用户分配角色

用户角色绑定通过 Casbin g 策略存储，格式：`g, user_id, role_code, domain`

```go
// AssignRole 为用户分配角色
func (s *UserService) AssignRole(ctx context.Context, userID, roleID string) error {
    // 1. 查询角色信息获取 role_code
    role, err := s.roleRepo.GetByID(ctx, roleID)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return xerr.ErrRoleNotFound
        }
        return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
    }

    // 2. 获取租户代码
    tenantCode := xcontext.GetTenantCode(ctx)
    if tenantCode == "" {
        return xerr.ErrUnauthorized
    }

    // 3. 添加 g 策略: g, user_id, role_code, domain
    _, err = s.enforcer.AddRoleForUserInDomain(userID, role.RoleCode, tenantCode)
    if err != nil {
        return xerr.Wrap(xerr.ErrInternal.Code, "分配角色失败", err)
    }

    return nil
}
```

### 获取用户菜单（核心）

获取用户菜单时会自动处理角色继承（通过 g2 策略递归获取父角色权限）

```go
// GetUserMenus 获取用户菜单列表（树形结构）
func (s *UserMenuService) GetUserMenus(ctx context.Context) ([]*dto.MenuTreeNode, error) {
    // 1. 获取用户 ID
    userID := xcontext.GetUserID(ctx)

    // 2. 获取租户代码
    tenantCode := xcontext.GetTenantCode(ctx)

    // 3. 从 Casbin 获取用户的角色（g 策略）
    roles, err := s.enforcer.GetRolesForUserInDomain(userID, tenantCode)
    if err != nil {
        return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取用户角色失败", err)
    }

    // 4. 递归获取所有角色（通过 g2 处理继承）
    allRoleCodes, inheritedRoles := s.getAllRoleCodes(roles)

    // 5. 从 Casbin 获取角色的菜单权限（支持跨 domain 查询）
    menuIDs := s.getMenuPermissionsForRoles(ctx, allRoleCodes, inheritedRoles, tenantCode)

    // 6. 查询菜单详情
    menus, err := s.menuRepo.GetByIDs(ctx, menuIDs)
    if err != nil {
        return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单详情失败", err)
    }

    // 7. 构建菜单树
    return s.buildMenuTree(menus), nil
}

// getAllRoleCodes 递归获取所有角色（通过 g2 处理继承）
// 返回值: (所有角色编码, 继承角色集合)
func (s *UserMenuService) getAllRoleCodes(roleCodes []string) ([]string, map[string]bool) {
    var allCodes []string
    visited := make(map[string]bool)
    inheritedRoles := make(map[string]bool)

    var dfs func(roleCode string, isInherited bool)
    dfs = func(roleCode string, isInherited bool) {
        if visited[roleCode] {
            return
        }
        visited[roleCode] = true
        allCodes = append(allCodes, roleCode)

        // 如果是继承角色，标记为继承
        if isInherited {
            inheritedRoles[roleCode] = true
        }

        // 通过 g2 获取继承的父角色
        // GetRolesForUser 在 g2 中返回 (child -> parent)
        parents, _ := s.enforcer.GetRolesForUser(roleCode)
        for _, parentCode := range parents {
            dfs(parentCode, true) // 父角色都是继承角色
        }
    }

    for _, code := range roleCodes {
        dfs(code, false)
    }

    return allCodes, inheritedRoles
}

// getMenuPermissionsForRoles 获取角色的 MENU 类型权限ID列表
// 支持角色继承：对于继承角色，同时查询 default domain 的权限
func (s *UserMenuService) getMenuPermissionsForRoles(ctx context.Context, roles []string, inheritedRoles map[string]bool, tenantCode string) []string {
    menuIDSet := make(map[string]bool)

    for _, role := range roles {
        // 确定要查询的 domain 列表
        domains := []string{tenantCode}
        if inheritedRoles[role] {
            // 继承角色：同时查询 default domain
            domains = append(domains, constants.DefaultTenantCode)
        }

        // 从所有相关 domain 查询权限
        for _, domain := range domains {
            policies, _ := s.enforcer.GetFilteredPolicy(0, role, domain)
            for _, policy := range policies {
                if len(policy) >= 4 {
                    resource := policy[2]
                    action := policy[3]
                    // 对于 MENU 类型，resource 格式为 "menu:menuID"，action 是 "*"
                    if (action == "*" || action == "") && strings.HasPrefix(resource, "menu:") {
                        menuID := strings.TrimPrefix(resource, "menu:")
                        menuIDSet[menuID] = true
                    }
                }
            }
        }
    }

    // 转换为切片
    menuIDs := make([]string, 0, len(menuIDSet))
    for menuID := range menuIDSet {
        menuIDs = append(menuIDs, menuID)
    }

    return menuIDs
}
```

**关键点**：
1. **自动处理角色继承**：通过 `getAllRoleCodes` 递归获取所有父角色，并标记哪些是继承角色
2. **跨 domain 权限查询**：对于继承角色，同时查询当前租户 domain 和 default domain 的权限
3. **权限实时计算**：子角色自动继承父角色模板的最新权限，无需手动同步
4. **菜单全局共享**：menus 表无 tenant_id 字段，所有租户共享菜单定义

---

## API 接口定义

### 角色相关 DTO

```go
// CreateRoleRequest 创建角色请求（支持继承）
type CreateRoleRequest struct {
    RoleCode       string `json:"role_code" binding:"required"`         // 角色编码（租户内唯一）
    Name           string `json:"name" binding:"required"`              // 角色名称
    Description    string `json:"description" binding:"omitempty"`      // 角色描述
    Status         int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:启用 2:禁用
    ParentRoleCode *string `json:"parent_role_code" binding:"omitempty"` // 父角色编码（继承 default 租户的角色模板）
}

// RoleResponse 角色响应
type RoleResponse struct {
    RoleID         string  `json:"role_id"`         // 角色ID
    TenantID       string  `json:"tenant_id"`       // 租户ID
    RoleCode       string  `json:"role_code"`       // 角色编码
    Name           string  `json:"name"`            // 角色名称
    Description    string  `json:"description"`     // 角色描述
    Status         int     `json:"status"`          // 状态 1:启用 2:禁用
    ParentRoleCode *string `json:"parent_role_code"` // 父角色编码
    CreatedAt      int64   `json:"created_at"`      // 创建时间
    UpdatedAt      int64   `json:"updated_at"`      // 更新时间
}

// AssignPermissionsRequest 分配权限请求（菜单+按钮）
type AssignPermissionsRequest struct {
    MenuIDs   []string `json:"menu_ids" binding:"required"`    // 菜单ID列表
    ButtonIDs []string `json:"button_ids" binding:"omitempty"` // 按钮权限ID列表
}

// RolePermissionsResponse 角色权限响应
type RolePermissionsResponse struct {
    MenuIDs   []string `json:"menu_ids"`   // 菜单ID列表
    ButtonIDs []string `json:"button_ids"` // 按钮权限ID列表
}
```

### 菜单相关 DTO

```go
// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
    Name      string `json:"name" binding:"required"`               // 菜单名称
    ParentID  string `json:"parent_id" binding:"omitempty"`        // 父菜单ID
    Path      string `json:"path" binding:"omitempty"`             // 前端路由路径
    Component string `json:"component" binding:"omitempty"`        // 前端组件路径
    Redirect  string `json:"redirect" binding:"omitempty"`         // 重定向路径
    Icon      string `json:"icon" binding:"omitempty"`             // 图标
    Sort      *int16 `json:"sort" binding:"omitempty"`             // 排序
    Status    int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:显示 2:隐藏
}

// MenuInfo 菜单信息
type MenuInfo struct {
    MenuID      string  `json:"menu_id"`      // 菜单ID
    Name        string  `json:"name"`         // 菜单名称
    Type        string  `json:"type"`         // 类型（固定为 "MENU"）
    ParentID    *string `json:"parent_id"`    // 父菜单ID
    Resource    *string `json:"resource"`     // 资源路径（menu:menu_id）
    Action      *string `json:"action"`       // 请求方法（固定为 "*"）
    Path        *string `json:"path"`         // 前端路由路径
    Component   *string `json:"component"`    // 前端组件路径
    Redirect    *string `json:"redirect"`     // 重定向路径
    Icon        *string `json:"icon"`         // 图标
    Sort        *int16  `json:"sort"`         // 排序
    Status      int16   `json:"status"`       // 状态
    Description *string `json:"description"`  // 描述
    CreatedAt   int64   `json:"created_at"`   // 创建时间
    UpdatedAt   int64   `json:"updated_at"`   // 更新时间
}

// MenuTreeNode 菜单树节点
type MenuTreeNode struct {
    *MenuInfo
    Children []*MenuTreeNode `json:"children"` // 子菜单
}
```

### API 路由

```go
// 角色管理路由
POST   /api/v1/roles                      // 创建角色（支持继承）
GET    /api/v1/roles                      // 角色列表
GET    /api/v1/roles/:role_id             // 角色详情
PUT    /api/v1/roles/:role_id             // 更新角色
DELETE /api/v1/roles/:role_id             // 删除角色
PUT    /api/v1/roles/:role_id/status      // 更新角色状态

// 权限管理路由（新接口，推荐使用）
PUT    /api/v1/roles/:role_id/permissions // 分配权限（菜单+按钮）
GET    /api/v1/roles/:role_id/permissions // 获取角色权限

// 向后兼容接口（已弃用）
PUT    /api/v1/roles/:role_id/menus       // 分配菜单（调用 AssignPermissions）
GET    /api/v1/roles/:role_id/menus       // 获取角色菜单（调用 GetRolePermissions）

// 用户菜单路由
GET    /api/v1/user/menus                 // 获取当前用户菜单（树形结构）
GET    /api/v1/user/buttons/:menu_id      // 获取指定菜单的按钮权限
```

---

## 新增功能

### 1. API 权限自动关联

为解决"菜单权限只控制前端显示，不控制后端 API 访问"的安全问题，实现了 API 权限自动关联：

**菜单表新增字段**：
```sql
ALTER TABLE menus ADD COLUMN api_paths TEXT NOT NULL DEFAULT '';
```

**字段格式**：JSON 数组，存储菜单关联的 API 路径
```json
[
  {"path": "/api/v1/users", "methods": ["GET", "POST"]},
  {"path": "/api/v1/users/:id", "methods": ["GET", "PUT", "DELETE"]}
]
```

**自动关联逻辑**：
- 为角色分配菜单权限时，自动关联菜单的 API 权限
- 前端隐藏菜单时，后端 API 也会被拦截
- 详见 `RoleService.AssignPermissions` 实现

### 2. 级联删除

为维护数据一致性，实现了删除操作的级联处理：

**删除菜单时**：
- 自动清理所有租户中该菜单的权限策略（`p, role_code, domain, menu:menu_id, *`）
- 详见 `MenuService.DeleteMenu`

**删除角色时**：
- 自动清理角色的所有权限策略
- 自动清理角色继承关系（g2）
- 自动清理用户-角色绑定关系（g）
- 详见 `RoleService.DeleteRole`

**删除用户时**：
- 自动清理用户的所有角色绑定关系
- 详见 `UserService.DeleteUser`

### 3. 固化权限

提供"固化继承权限"功能，将继承的权限复制到当前租户并断开继承关系：

```go
// FreezeInheritedPermissions 固化继承的权限
// 用途：将继承自 default 租户角色模板的权限"固化"到当前租户
// 效果：固化后，角色不再跟随模板更新，而是拥有独立的权限副本
func (s *RoleService) FreezeInheritedPermissions(ctx context.Context, roleID string) error
```

### 4. 租户管理员权限边界

定义了租户管理员（Admin 角色）的权限范围常量：

```go
// constants/system.go
const (
    // 角色管理
    TenantAdminCanDeleteRoles = false // 不能删除角色

    // 权限管理
    TenantAdminCanModifyInheritedPermissions = false // 不能修改继承的权限
    TenantAdminCanAddCustomPermissions       = true  // 可以添加额外权限

    // 菜单管理
    TenantAdminCannotModifySystemMenus = true // 不能修改系统菜单
)
```

同时提供了权限边界检查方法：
```go
func (s *RoleService) CheckTenantAdminPermission(ctx context.Context, operation string, targetID string) error
```

---

## 常量定义

```go
package constants

const (
    // 租户
    DefaultTenantID   = "000000000000000000" // 默认租户ID（18个零）
    DefaultTenantCode = "default"            // 默认租户code

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
