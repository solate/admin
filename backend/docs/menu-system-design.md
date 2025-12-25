# 多租户 SaaS 菜单系统设计文档

## 一、概述

本文档描述了多租户 SaaS 后台管理系统中的菜单系统设计方案。该方案基于现有的 RBAC 权限模型和 Casbin 框架，实现了租户级别的菜单隔离和动态权限控制。

### 1.1 设计目标

- **三层权限控制**：支持菜单级（MENU）、按钮级（BUTTON）、API级（API）三层权限
- **租户隔离**：每个租户拥有独立的菜单树，互不干扰
- **统一管理**：菜单作为权限资源的一种，与按钮、API 权限统一管理
- **灵活配置**：支持菜单层级结构、显示/隐藏、排序等配置
- **权限池管理**：超管定义权限池，租户在可用范围内分配权限
- **自定义菜单**：租户可添加自定义菜单，不影响其他租户

### 1.2 核心场景

```
超管操作流程：
1. 超管在 permissions 表（tenant_id=0）定义完整的权限资源树（菜单 + 按钮）
2. 新增租户时，超管为该租户勾选可用的菜单（如：不能给租户开放"租户管理"菜单）

租户管理员操作流程：
3. 创建角色时，勾选该角色能访问的菜单和按钮权限
   - 例如：销售角色 -> 订单菜单(查看+新增)、商品菜单(仅查看)
   - 例如：财务角色 -> 订单菜单(查看+导出)

用户登录后：
4. 根据用户角色，动态加载可见菜单和按钮
```

### 1.3 方案选择

**方案：基于现有 permissions 表的扩展方案（方案 A+）**

- 超管在 `permissions` 表（tenant_id=0）定义完整的权限资源树
- 新增 `source_type` 字段：SYSTEM（系统菜单）/ CUSTOM（租户自定义）
- 新增 `tenant_permissions` 表：超管给租户勾选可用的菜单
- 租户管理员在可用范围内，通过 Casbin 给角色分配权限
- 租户自定义菜单：在 `permissions` 表插入租户自己的菜单（source_type=CUSTOM）

### 1.4 参考资源

- [Casbin 菜单权限官方文档](https://casbin.org/docs/menu-permissions/)
- [Casbin RBAC with Domains](https://casbin.org/docs/rbac-with-domains/)
- [ContiNew Admin 开源项目](https://github.com/continew-org/continew-admin)

---

## 二、数据模型设计

### 2.1 表结构

#### 2.1.1 permissions 表（权限/菜单表）

菜单数据存储在现有的 `permissions` 表中，通过 `type='MENU'` 标识菜单类型。

```sql
CREATE TABLE permissions (
    permission_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,  -- MENU, BUTTON, API, DATA
    parent_id VARCHAR(255),     -- 父菜单ID（用于构建树形结构）
    resource VARCHAR(255),      -- 资源路径
    action VARCHAR(20),         -- 请求方法
    path VARCHAR(255),          -- 前端路由路径（仅MENU类型）
    component VARCHAR(255),     -- 前端组件路径（仅MENU类型）
    redirect VARCHAR(255),      -- 重定向路径（仅MENU类型）
    icon VARCHAR(100),          -- 图标（仅MENU类型）
    sort SMALLINT,              -- 排序
    status SMALLINT DEFAULT 1,  -- 显示状态：1显示，2隐藏
    source_type VARCHAR(20) DEFAULT 'SYSTEM',  -- SYSTEM（系统定义）/ CUSTOM（租户自定义）
    is_visible SMALLINT DEFAULT 1,  -- 菜单是否在前端显示（按钮/API不需要显示）
    description TEXT,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);
```

#### 2.1.2 tenant_permissions 表（租户启用权限表）

```sql
-- 超管给租户勾选可用的菜单
CREATE TABLE tenant_permissions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id VARCHAR(36) NOT NULL COMMENT '租户ID',
    permission_id VARCHAR(255) NOT NULL COMMENT '权限/菜单ID',
    enabled SMALLINT DEFAULT 1 COMMENT '是否启用：1启用，0禁用',
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY uk_tenant_permission (tenant_id, permission_id),
    KEY idx_tenant_permissions (tenant_id, enabled, deleted_at),
    FOREIGN KEY (permission_id) REFERENCES permissions(permission_id) ON DELETE CASCADE
) COMMENT '租户启用的权限关联表';
```

### 2.2 字段说明

| 字段 | 类型 | 说明 | 适用类型 |
|------|------|------|----------|
| `permission_id` | VARCHAR(255) | 权限/菜单ID | 全部 |
| `tenant_id` | VARCHAR(36) | 所属租户ID（0=超管系统权限） | 全部 |
| `name` | VARCHAR(100) | 名称 | 全部 |
| `type` | VARCHAR(20) | MENU/BUTTON/API/DATA | 全部 |
| `parent_id` | VARCHAR(255) | 父菜单ID | MENU |
| `resource` | VARCHAR(255) | 路由/API地址 | MENU/API |
| `action` | VARCHAR(20) | 请求方法 | API |
| `path` | VARCHAR(255) | 前端路由路径 | MENU |
| `component` | VARCHAR(255) | 前端组件路径 | MENU |
| `redirect` | VARCHAR(255) | 重定向路径 | MENU |
| `icon` | VARCHAR(100) | 图标 | MENU |
| `status` | SMALLINT | 显示状态(1:显示, 2:隐藏) | MENU |
| `sort` | SMALLINT | 排序 | MENU/BUTTON |
| `source_type` | VARCHAR(20) | SYSTEM/CUSTOM | 全部 |
| `is_visible` | SMALLINT | 是否在前端显示 | MENU |
| `description` | TEXT | 描述 | 全部 |

### 2.3 索引设计

```sql
-- permissions 表索引
CREATE INDEX idx_permissions_tenant_type ON permissions(tenant_id, type);
CREATE INDEX idx_permissions_tenant_parent ON permissions(tenant_id, parent_id, deleted_at);
CREATE INDEX idx_permissions_source ON permissions(tenant_id, source_type, deleted_at);
CREATE INDEX idx_permissions_status ON permissions(status) WHERE status IS NOT NULL;
```

---

## 三、API 接口设计

### 3.1 接口列表

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | `/api/v1/menus` | 创建菜单 | menu:create |
| GET | `/api/v1/menus` | 菜单列表（分页） | menu:query |
| GET | `/api/v1/menus/all` | 所有菜单（平铺） | menu:query |
| GET | `/api/v1/menus/tree` | 菜单树 | - |
| GET | `/api/v1/menus/:menu_id` | 菜单详情 | menu:query |
| PUT | `/api/v1/menus/:menu_id` | 更新菜单 | menu:update |
| DELETE | `/api/v1/menus/:menu_id` | 删除菜单 | menu:delete |
| PUT | `/api/v1/menus/:menu_id/status/:status` | 更新菜单状态 | menu:update |

### 3.2 请求/响应示例

#### 3.2.1 创建菜单

**请求：**
```json
POST /api/v1/menus
{
  "name": "用户管理",
  "type": "MENU",
  "parent_id": "menu_system",
  "path": "/system/users",
  "component": "views/system/Users.vue",
  "icon": "User",
  "sort": 101,
  "status": 1
}
```

**响应：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "permission_id": "menu_001",
    "tenant_id": "tenant_001",
    "name": "用户管理",
    "type": "MENU",
    "parent_id": "menu_system",
    "path": "/system/users",
    "component": "views/system/Users.vue",
    "icon": "User",
    "sort": 101,
    "status": 1,
    "created_at": 1703145600000
  }
}
```

#### 3.2.2 获取菜单树

**请求：**
```
GET /api/v1/menus/tree
```

**响应：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "permission_id": "menu_system",
        "name": "系统管理",
        "type": "MENU",
        "path": "/system",
        "icon": "Setting",
        "sort": 100,
        "status": 1,
        "children": [
          {
            "permission_id": "menu_user",
            "name": "用户管理",
            "type": "MENU",
            "path": "/system/users",
            "icon": "User",
            "sort": 101,
            "status": 1,
            "children": []
          }
        ]
      }
    ]
  }
}
```

---

## 四、业务逻辑设计

### 4.1 菜单树构建算法

使用 Map 两次遍历算法，时间复杂度 O(n)：

```go
func (s *MenuService) buildMenuTree(menus []*model.Permission) []*dto.MenuTreeNode {
    // 创建节点映射
    nodeMap := make(map[string]*dto.MenuTreeNode)
    for _, menu := range menus {
        nodeMap[menu.PermissionID] = &dto.MenuTreeNode{
            MenuInfo: s.toMenuInfo(menu),
            Children: []*dto.MenuTreeNode{},
        }
    }

    // 构建树结构
    var roots []*dto.MenuTreeNode
    for _, menu := range menus {
        node := nodeMap[menu.PermissionID]
        if menu.ParentID == nil || *menu.ParentID == "" {
            roots = append(roots, node)
        } else if parent, exists := nodeMap[*menu.ParentID]; exists {
            parent.Children = append(parent.Children, node)
        }
    }

    return roots
}
```

### 4.2 菜单删除逻辑

删除菜单时需要检查是否有子菜单：

```go
// 检查是否有子菜单
childrenCount, err := s.menuRepo.GetChildrenCount(ctx, menuID)
if err != nil {
    return xerr.Wrap(xerr.ErrInternal.Code, "检查子菜单失败", err)
}
if childrenCount > 0 {
    return xerr.ErrMenuHasChildren
}
```

### 4.3 菜单移动逻辑

更新父菜单时需要防止循环引用：

```go
// 不能将菜单移动到自己或其子菜单下
if req.ParentID == menuID {
    return nil, xerr.ErrMenuCannotMoveToSelf
}
```

---

## 五、与 Casbin 集成

### 5.1 策略模型

系统使用 RBAC with Domains 模型（domain = tenant_id）：

```conf
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```

### 5.2 策略示例

#### 5.2.1 超管特殊策略（全局权限）

```
# 超管特殊策略（全局）
p, super_admin, 0, *, *
```

超管拥有所有权限，可以：
- 访问所有 API 接口
- 管理所有租户
- 定义系统权限资源池
- 给租户勾选可用菜单

#### 5.2.2 租户角色权限策略

```
# 策略定义：角色在租户下对资源的访问权限
# 格式: p, role_id, tenant_id, resource_path, action

# 租户A的销售角色权限
p, role_sales, tenant_a, /api/v1/orders, GET
p, role_sales, tenant_a, /api/v1/orders, POST
p, role_sales, tenant_a, /api/v1/users, GET

# 租户A的财务角色权限
p, role_finance, tenant_a, /api/v1/orders, GET
p, role_finance, tenant_a, /api/v1/orders/export, POST
p, role_finance, tenant_a, /api/v1/users, GET
```

#### 5.2.3 用户角色关联

```
# 用户角色关联（g策略）
# 格式: g, user_id, role_id, tenant_id

# 超管用户关联（tenant_id=0）
g, admin_001, super_admin, 0

# 租户A的用户角色关联
g, user_001, role_sales, tenant_a
g, user_002, role_finance, tenant_a
g, user_admin_001, role_admin, tenant_a
```

### 5.3 权限检查流程

#### 5.3.1 前端获取用户菜单

```
1. 用户登录后，前端调用 GET /api/v1/user/menu
2. 后端获取用户的所有角色（通过 Casbin GetRolesForUserInDomain）
3. 后端获取角色的所有 MENU 类型权限
4. 查询权限详情，构建菜单树返回
```

#### 5.3.2 前端获取按钮权限

```
1. 用户进入某个菜单页面
2. 前端调用 GET /api/v1/user/buttons?menu_id=xxx
3. 后端返回该菜单下的所有 BUTTON 类型权限
4. 前端根据权限显示/隐藏按钮
```

#### 5.3.3 后端 API 权限检查

```
1. 请求经过 auth 中间件
2. 从 context 获取 user_id, tenant_id
3. 构造请求：(user_id, tenant_id, api_path, http_method)
4. 调用 Casbin.Enforce() 判断
5. 超管（tenant_id=0 且 role=super_admin）直接通过
```

### 5.4 超管给租户勾选菜单流程

```
1. 超管创建租户时，选择该租户可用的菜单
2. 后端在 tenant_permissions 表插入记录
3. 租户管理员创建角色时，只能从 tenant_permissions 中选择权限
4. 后端校验权限是否在租户可用范围内
5. 校验通过后，写入 Casbin 策略
```

---

## 六、前端集成

### 6.1 API 调用

前端已定义菜单 API 接口（[frontend/src/api/menu.ts](frontend/src/api/menu.ts)）：

```typescript
export const menuApi = {
  // 获取菜单列表
  getList: (params: MenuListParams): Promise<MenuListResponse> => {
    return http.get('/api/v1/menus', { params })
  },

  // 创建菜单
  create: (data: CreateMenuRequest): Promise<CreateMenuResponse> => {
    return http.post('/api/v1/menus', data)
  },

  // 获取菜单树
  getMenuTree: (): Promise<MenuTreeResponse> => {
    return http.get('/api/v1/menus/tree')
  }
}
```

### 6.2 动态菜单加载

在登录后调用 `/api/v1/menus/tree` 接口获取用户可见的菜单树，然后动态渲染导航菜单。

---

## 七、文件结构

### 7.1 后端文件

```
backend/
├── migrations/
│   └── 000002_add_menu_fields.up.sql    # 数据库迁移文件
├── internal/
│   ├── dto/
│   │   └── menu_dto.go                  # 菜单 DTO 定义
│   ├── repository/
│   │   └── menu_repo.go                 # 菜单数据访问层
│   ├── service/
│   │   └── menu_service.go              # 菜单业务逻辑层
│   ├── handler/
│   │   └── menu_handler.go              # 菜单 HTTP 处理器
│   └── router/
│       ├── app.go                       # 注册 MenuHandler
│       └── router.go                    # 菜单路由配置
├── pkg/
│   ├── constants/
│   │   └── system.go                    # 权限类型常量
│   └── xerr/
│       └── codes.go                     # 菜单错误码
└── doc/
    └── menu-system-design.md            # 本文档
```

### 7.2 前端文件

```
frontend/
├── src/
│   ├── api/
│   │   └── menu.ts                      # 菜单 API 接口定义
│   └── views/
│       ├── Layout.vue                   # 主布局（动态菜单加载）
│       └── system/
│           └── Menus.vue                # 菜单管理页面（待实现）
```

---

## 八、常见场景处理

### 8.1 新增租户的默认权限

```go
// 创建租户时的默认权限处理
func (s *TenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) error {
    // 1. 创建租户
    tenant := &model.Tenant{
        TenantID: uuid.New().String(),
        Name:     req.Name,
    }

    if err := s.db.Create(tenant).Error; err != nil {
        return err
    }

    // 2. 如果指定了默认菜单，启用
    if len(req.EnabledMenus) > 0 {
        s.permissionService.EnableTenantMenus(ctx, tenant.TenantID, req.EnabledMenus)
    }

    // 3. 创建默认角色（如：管理员角色）
    defaultRole := &model.Role{
        RoleID:   uuid.New().String(),
        TenantID: tenant.TenantID,
        Name:     "管理员",
    }
    s.db.Create(defaultRole)

    // 4. 给默认角色分配所有启用的菜单权限
    s.permissionService.AssignPermissionsToRole(ctx, tenant.TenantID, defaultRole.RoleID, req.EnabledMenus)

    return nil
}
```

### 8.2 租户升级套餐（增加菜单）

```go
// 租户升级套餐时的菜单处理
func (s *TenantService) UpgradePackage(ctx context.Context, tenantID string, packageID string) error {
    // 1. 查询套餐包含的菜单
    packageMenus, _ := s.getPackageMenus(packageID)

    // 2. 启用新菜单
    s.permissionService.EnableTenantMenus(ctx, tenantID, packageMenus)

    // 3. 通知租户管理员新菜单可用（可选）
    // 可以发送邮件或站内消息

    return nil
}
```

### 8.3 超管修改权限池

使用软同步策略：
1. 超管修改权限池，不影响已有租户
2. 租户下次启用菜单时，使用新的权限定义
3. 对于需要强制升级的场景，提供"批量同步"功能

```go
// 同步权限到租户
func (s *SystemPermissionService) SyncPermissionToTenants(ctx context.Context, permissionID string, tenantIDs []string) error {
    for _, tenantID := range tenantIDs {
        // 检查租户是否已启用该权限
        var tp model.TenantPermission
        err := s.db.Where("tenant_id = ? AND permission_id = ?", tenantID, permissionID).First(&tp).Error
        if err == gorm.ErrRecordNotFound {
            // 未启用，自动启用
            s.permissionService.EnableTenantMenus(ctx, tenantID, []string{permissionID})
        }
    }
    return nil
}
```

### 8.4 租户添加自定义菜单

```go
// 租户添加自定义菜单
func (s *PermissionService) AddCustomMenu(ctx context.Context, tenantID string, menu *model.Permission) error {
    menu.TenantID = tenantID
    menu.SourceType = "CUSTOM"  // 标记为自定义
    menu.CreatedAt = time.Now().Unix()
    menu.UpdatedAt = time.Now().Unix()

    // 同时添加到 tenant_permissions，让租户可以使用
    err := s.db.WithContext(ctx).Create(menu).Error
    if err != nil {
        return err
    }

    // 自动启用该菜单
    tp := model.TenantPermission{
        TenantID:     tenantID,
        PermissionID: menu.PermissionID,
        Enabled:      1,
        CreatedAt:    time.Now().Unix(),
        UpdatedAt:    time.Now().Unix(),
    }
    return s.db.Create(&tp).Error
}
```

### 8.5 权限继承（父菜单权限包含子菜单）

```go
// 分配权限时自动包含父菜单
func (s *PermissionService) AssignPermissionsToRole(ctx context.Context, tenantID, roleID string, permissionIDs []string) error {
    // 获取所有父菜单ID
    allPermissions := make(map[string]bool)
    for _, permID := range permissionIDs {
        allPermissions[permID] = true
        s.addParentPermissions(ctx, permID, allPermissions)
    }

    // 转换为切片
    finalPermissions := make([]string, 0, len(allPermissions))
    for permID := range allPermissions {
        finalPermissions = append(finalPermissions, permID)
    }

    return s.base.AssignPermissionsToRole(ctx, tenantID, roleID, finalPermissions)
}

func (s *PermissionService) addParentPermissions(ctx context.Context, permissionID string, result map[string]bool) {
    var perm model.Permission
    err := s.db.WithContext(ctx).Where("permission_id = ?", permissionID).First(&perm).Error
    if err != nil || perm.ParentID == "" {
        return
    }
    result[perm.ParentID] = true
    s.addParentPermissions(ctx, perm.ParentID, result)
}
```

---

## 九、性能优化

### 9.1 菜单缓存

```go
// 使用缓存减少数据库查询
type CachedPermissionService struct {
    base      *PermissionService
    menuCache *cache.Cache
}

func (s *CachedPermissionService) GetUserMenus(ctx context.Context, userID, tenantID string) ([]map[string]interface{}, error) {
    cacheKey := fmt.Sprintf("menu:%s:%s", userID, tenantID)
    if cached, ok := s.menuCache.Get(cacheKey); ok {
        return cached.([]map[string]interface{}), nil
    }

    menus, err := s.base.GetUserMenus(ctx, userID, tenantID)
    if err == nil {
        s.menuCache.Set(cacheKey, menus, 5*time.Minute)
    }

    return menus, err
}
```

### 9.2 Casbin 策略缓存

Casbin 自带策略缓存，确保启用：

```go
enforcer, _ := casbin.NewEnforcer("./rbac_model.conf", "./policy.csv")
enforcer.EnableCache(true)  // 启用缓存
```

---

## 十、后续扩展

1. **菜单拖拽排序**：支持拖拽调整菜单顺序
2. **菜单图标库**：集成 Element Plus / Ant Design 图标库
3. **数据权限**：结合菜单实现数据范围控制（如：只看自己的数据）
4. **前端菜单管理页面**：实现可视化的菜单管理界面
5. **权限模板**：超管可以创建权限模板，快速分配给租户

---

## 十一、注意事项

### 11.1 权限校验顺序

```
1. 超管（tenant_id=0 且 role=super_admin）-> 直接通过
2. 检查 Casbin 策略
3. 返回权限检查结果
```

### 11.2 菜单与按钮的关联

按钮的 `parent_id` 指向所属菜单的 `permission_id`，这样在获取菜单时可以一并获取该菜单下的所有按钮。

### 11.3 前端权限指令示例

```javascript
// Vue 3 权限指令
import { useUserStore } from '@/stores/user'

export const permission = {
  mounted(el, binding) {
    const { value } = binding
    const userStore = useUserStore()
    const buttons = userStore.buttons // 当前页面的按钮权限列表

    if (value && !buttons.includes(value)) {
      el.parentNode?.removeChild(el)
    }
  }
}

// 使用
<button v-permission="'btn_order_create'">新建订单</button>
```

---

## 十二、参考资料

- [Casbin 菜单权限官方文档](https://casbin.org/docs/menu-permissions/)
- [Casbin RBAC with Domains](https://casbin.org/docs/rbac-with-domains/)
- [多租户租户隔离方案](https://blog.csdn.net/wbx044720/article/details/154946289)
- [ContiNew Admin](https://github.com/continew-org/continew-admin)
