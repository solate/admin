# 权限管理系统设计

## 设计理念

本系统采用**权限分离设计**，将三种权限类型独立管理：
1. **API ACL 权限**：控制后端 API 接口访问
2. **UI 权限**：控制前端菜单和按钮显示
3. **数据权限**：控制数据访问范围

---

## 一、API ACL 权限管理

### 设计说明

API ACL 权限管理用于控制用户对后端 API 接口的访问权限。

**核心要素**：
- **资源（Resource）**：API 端点，如 `/api/users`
- **操作（Action）**：HTTP 方法（GET, POST, PUT, DELETE 等）
- **主体（Subject）**：角色或用户
- **权限规则**：主体 + 资源 + 操作 的组合

**设计架构**：
- 使用 **api_resources 表**存储 API 元数据（供前端管理界面展示）
- 使用 **Casbin** 进行实际的权限控制（casbin_rule 表）
- 前端通过 api_resources 管理界面配置权限，后端自动同步到 Casbin

### 表设计

#### api_resources（API 资源元数据表）

**用途**：存储 API 定义和业务元数据，供前端管理界面展示和使用。

**注意**：此表仅用于元数据管理和前端展示，**不用于权限检查**。权限检查由 Casbin 负责。

| 字段名 | 类型 | 长度 | 允许空 | 说明 |
|--------|------|------|--------|------|
| id | BIGINT | - | NO | 主键 |
| name | VARCHAR | 100 | NO | **前端显示名称**（如：用户列表） |
| path | VARCHAR | 255 | NO | API 路径（如：/api/v1/users） |
| method | VARCHAR | 10 | NO | HTTP 方法（GET/POST/PUT/DELETE/PATCH） |
| description | VARCHAR | 255 | YES | **前端显示描述** |
| module | VARCHAR | 50 | NO | **所属模块**（如：用户管理、订单管理） |
| created_at | TIMESTAMP | - | NO | 创建时间 |
| updated_at | TIMESTAMP | - | NO | 更新时间 |

**索引**：
- PRIMARY KEY (id)
- UNIQUE KEY uk_path_method (path, method)
- INDEX idx_module (module)

**示例数据**：
```
id: 1, name: "用户列表", path: "/api/v1/users", method: "GET", module: "用户管理"
id: 2, name: "创建用户", path: "/api/v1/users", method: "POST", module: "用户管理"
id: 3, name: "删除用户", path: "/api/v1/users/:id", method: "DELETE", module: "用户管理"
```

#### casbin_rule（Casbin 权限策略表）

**用途**：存储实际的权限策略，由 Casbin 自动创建和管理。

**表结构**（Casbin 自动创建）：

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | BIGINT | 主键 |
| ptype | VARCHAR(100) | 策略类型（p=权限策略，g=角色继承） |
| v0 | VARCHAR(255) | 角色 ID（sub） |
| v1 | VARCHAR(255) | 租户 ID（dom） |
| v2 | VARCHAR(255) | API 路径（obj） |
| v3 | VARCHAR(255) | HTTP 方法（act） |
| v4 | VARCHAR(255) | 扩展字段1 |
| v5 | VARCHAR(255) | 扩展字段2 |

**策略格式**：`p, sub, dom, obj, act`
- `sub`：角色 ID
- `dom`：租户 ID（多租户场景）
- `obj`：API 路径（如 `/api/v1/users`）
- `act`：HTTP 方法（如 `GET`）

**示例数据**：
```sql
-- 角色ID=1可以访问用户列表API
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES
('p', '1', '1', '/api/v1/users', 'GET');

-- 角色ID=1可以创建用户
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES
('p', '1', '1', '/api/v1/users', 'POST');

-- 角色ID=2不能删除用户（不添加策略即为拒绝）
```

---

## 二、菜单和按钮权限管理（UI 权限）

### 设计说明

UI 权限管理控制前端界面的显示，包括菜单和按钮。

**核心要素**：
- **菜单**：前端导航菜单项（一级菜单、二级菜单等）
- **按钮**：页面内的操作按钮（新增、编辑、删除、导出等）
- **权限绑定**：角色与菜单/按钮的关联关系

### 表设计

#### menus（菜单资源表 - 树形结构）

**用途**：统一管理目录、菜单、按钮，采用树形结构。

**树形结构示例**：
```
系统管理（目录）
  └── 用户管理（菜单）
       ├── 新增（按钮）
       ├── 编辑（按钮）
       └── 删除（按钮）
```

| 字段名 | 类型 | 长度 | 允许空 | 说明 |
|--------|------|------|--------|------|
| menu_id | VARCHAR | 20 | NO | 主键 |
| parent_id | VARCHAR | 20 | NO | 父级菜单 ID（空字符串表示根节点） |
| name | VARCHAR | 100 | NO | 菜单名称 |
| type | VARCHAR | 10 | NO | **菜单类型**（dir=目录，menu=菜单，button=按钮） |
| perms | VARCHAR | 200 | YES | 权限标识（按钮类型必填，如：system:user:add） |
| route_name | VARCHAR | 100 | NO | 路由名称（用于 keep-alive 缓存） |
| path | VARCHAR | 255 | NO | 路由路径（前端路由路径） |
| component | VARCHAR | 255 | NO | 组件路径（前端组件路径） |
| redirect | VARCHAR | 255 | NO | 重定向路径（父菜单重定向到第一个子菜单） |
| visible | SMALLINT | - | NO | 是否可见（1=显示，2=隐藏） |
| keep_alive | SMALLINT | - | NO | 缓存页面（1=开启，2=关闭） |
| sort | INT | - | NO | 排序（数字越小越靠前） |
| icon | VARCHAR | 100 | NO | 图标 |
| status | SMALLINT | - | NO | 状态（1=启用，2=禁用） |
| description | TEXT | - | NO | 描述信息 |
| created_at | BIGINT | - | NO | 创建时间戳（毫秒） |
| updated_at | BIGINT | - | NO | 更新时间戳（毫秒） |
| deleted_at | BIGINT | - | YES | 删除时间戳（毫秒，软删除） |

**索引**：
- PRIMARY KEY (menu_id)
- INDEX idx_parent_id (parent_id, deleted_at)
- INDEX idx_type (type, visible, deleted_at)

**示例数据**：
```
# 目录
menu_id: "1", parent_id: "", name: "系统管理", type: "dir", icon: "setting", sort: 1

# 菜单项
menu_id: "2", parent_id: "1", name: "用户管理", type: "menu", route_name: "SystemUser", path: "/system/user", component: "system/user/index", visible: 1, keep_alive: 1, sort: 1

# 按钮
menu_id: "3", parent_id: "2", name: "新增", type: "button", perms: "system:user:add", sort: 1
menu_id: "4", parent_id: "2", name: "编辑", type: "button", perms: "system:user:edit", sort: 2
menu_id: "5", parent_id: "2", name: "删除", type: "button", perms: "system:user:delete", sort: 3
```

#### menu_permissions（菜单资源权限表）

**设计原则**：记录存在即有权限，记录不存在即无权限。

| 字段名 | 类型 | 长度 | 允许空 | 说明 |
|--------|------|------|--------|------|
| id | BIGINT | - | NO | 主键 |
| role_id | VARCHAR | 20 | NO | 角色 ID（关联 roles 表，通过 roles.tenant_id 实现租户隔离） |
| menu_id | VARCHAR | 20 | NO | 菜单 ID（关联 menus 表，menus 是全局的） |
| created_at | BIGINT | - | NO | 创建时间戳（毫秒） |
| updated_at | BIGINT | - | NO | 更新时间戳（毫秒） |

**索引**：
- PRIMARY KEY (id)
- UNIQUE KEY uk_role_menu (role_id, menu_id)
- INDEX idx_menu_id (menu_id) -- 支持反向查询

**说明**：
- `uk_role_menu` 既保证唯一性（同一角色不能重复授权同一菜单），又支持按 `role_id` 的查询（最左前缀原则）
- `idx_menu_id` 支持反向查询（查询某个菜单被哪些角色使用）
- 租户隔离通过 `role_id -> roles.tenant_id` 实现，因此本表不需要 tenant_id 字段

**操作说明**：
- ✅ 授权：`INSERT INTO menu_permissions`
- ✅ 撤销：`DELETE FROM menu_permissions`
- ✅ 检查：`EXISTS (SELECT 1 FROM menu_permissions WHERE role_id = ? AND menu_id = ?)`

**示例数据**：
```
role_id: 1, menu_id: 2   # 角色可以访问用户管理菜单
role_id: 1, menu_id: 3   # 角色可以看到新增按钮
# 角色 menu_id: 5 无记录，说明该角色不能看到删除按钮
```

---

## 三、数据权限管理

### 设计说明

数据权限管理控制用户可以访问的数据范围。

**核心要素**：
- **数据权限规则**：定义数据过滤条件
- **权限范围**：全部数据、本部门数据、仅本人数据等
- **规则类型**：部门维度、用户维度、自定义维度

### 表设计

#### data_permission_rules（数据权限规则表）

| 字段名 | 类型 | 长度 | 允许空 | 说明 |
|--------|------|------|--------|------|
| id | BIGINT | - | NO | 主键 |
| name | VARCHAR | 100 | NO | 规则名称 |
| code | VARCHAR | 50 | NO | 规则代码（如：all, dept, dept_and_sub, self, custom） |
| scope_type | VARCHAR | 20 | NO | 范围类型（all=全部，dept=部门，dept_and_sub=部门及下级，self=本人，custom=自定义） |
| description | VARCHAR | 255 | YES | 规则描述 |
| created_at | TIMESTAMP | - | NO | 创建时间 |
| updated_at | TIMESTAMP | - | NO | 更新时间 |

**索引**：
- PRIMARY KEY (id)
- UNIQUE KEY uk_code (code)
- INDEX idx_scope_type (scope_type)

**数据权限分类**（对应 roles 表的 data_scope 字段）：

| data_scope | 名称 | 说明 | 适用场景 |
|------------|------|------|----------|
| 1 | 全部数据 | 可以查看所有数据 | 超级管理员 |
| 2 | 自定义数据 | 只能查看指定部门的数据 | 需要跨部门但非全部权限 |
| 3 | 本部门数据 | 只能查看本部门的数据 | 部门经理 |
| 4 | 本部门及以下 | 查看本部门及所有下级部门的数据 | 总经理/多级部门负责人 |
| 5 | 仅本人数据 | 只能查看自己的数据 | 普通员工 |

**示例数据**：
```
id: 1, name: "全部数据", code: "all", scope_type: "all"
id: 2, name: "自定义数据", code: "custom", scope_type: "custom"
id: 3, name: "本部门数据", code: "dept", scope_type: "dept"
id: 4, name: "本部门及以下", code: "dept_and_sub", scope_type: "dept_and_sub"
id: 5, name: "仅本人数据", code: "self", scope_type: "self"
```

#### data_permission_bindings（数据权限绑定表）

| 字段名 | 类型 | 长度 | 允许空 | 说明 |
|--------|------|------|--------|------|
| id | BIGINT | - | NO | 主键 |
| role_id | BIGINT | - | NO | 角色 ID |
| rule_id | BIGINT | - | NO | 规则 ID |
| resource_type | VARCHAR | 50 | NO | 资源类型/业务实体（如：user, order, product） |
| created_at | TIMESTAMP | - | NO | 创建时间 |
| updated_at | TIMESTAMP | - | NO | 更新时间 |

**索引**：
- PRIMARY KEY (id)
- UNIQUE KEY uk_role_resource (role_id, resource_type)
- INDEX idx_role_id (role_id)
- INDEX idx_rule_id (rule_id)
- INDEX idx_resource_type (resource_type)

**示例数据**：
```
role_id: 1, rule_id: 1, resource_type: "user"    # 管理员可以查看所有用户
role_id: 2, rule_id: 2, resource_type: "order"   # 部门经理可以查看本部门的订单
role_id: 3, rule_id: 3, resource_type: "order"   # 普通员工只能查看自己的订单
```

#### data_permission_custom_rules（自定义数据权限规则表）

| 字段名 | 类型 | 长度 | 允许空 | 说明 |
|--------|------|------|--------|------|
| id | BIGINT | - | NO | 主键 |
| binding_id | BIGINT | - | NO | 绑定 ID（关联 data_permission_bindings） |
| field_name | VARCHAR | 50 | NO | 字段名称（如：dept_id, region_id） |
| operator | VARCHAR | 20 | NO | 操作符（eq=等于，in=在列表中，custom=自定义表达式） |
| field_value | VARCHAR | 255 | YES | 字段值（JSON 格式） |
| custom_expression | VARCHAR | 500 | YES | 自定义表达式（如：dept_id IN (SELECT id FROM dept WHERE path LIKE '1/%')） |
| sort | INT | - | NO | 排序序号 |
| created_at | TIMESTAMP | - | NO | 创建时间 |
| updated_at | TIMESTAMP | - | NO | 更新时间 |

**索引**：
- PRIMARY KEY (id)
- INDEX idx_binding_id (binding_id)
- INDEX idx_field_name (field_name)

**示例数据**：
```
binding_id: 1, field_name: "dept_id", operator: "eq", field_value: "user.dept_id"
binding_id: 2, field_name: "dept_id", operator: "custom", custom_expression: "dept_id IN (SELECT id FROM dept WHERE path LIKE CONCAT((SELECT path FROM dept WHERE id = ?), '%'))"
```

---

## 四、ER 关系图

```
roles (角色表，假设已存在)
  ├── casbin_rule (Casbin 权限策略)
  │     └── api_resources (API 元数据，仅展示用)
  ├── menu_permissions (菜单资源权限)
  │     └── menus (目录/菜单/按钮树形结构)
  └── data_permission_bindings (数据权限绑定)
        ├── data_permission_rules (数据权限规则)
        └── data_permission_custom_rules (自定义规则)
```

---

## 五、多租户支持（可选）

如果系统需要支持多租户（SaaS），在相关表中添加 tenant_id 字段：

| 字段名 | 类型 | 说明 |
|--------|------|------|
| tenant_id | BIGINT | 租户 ID |

**需要添加的表**：
- api_resources
- menus
- menu_permissions
- data_permission_rules
- data_permission_bindings
- data_permission_custom_rules

**索引调整**：
- 所有唯一索引需要包含 tenant_id
- 所有查询条件需要包含 tenant_id 过滤





-- ========================================
-- 3. 角色表 (Roles)
-- 租户自定义角色，继承关系由 Casbin g2 策略管理
-- 支持数据权限控制（若依5级数据范围）
-- ========================================
CREATE TABLE roles (
    role_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,               -- [多租户核心] 角色属于特定租户
    role_code VARCHAR(50) NOT NULL,               -- 角色编码 (如: sales, manager)
    name VARCHAR(100) NOT NULL,                   -- 角色名称 (如: 销售角色)
    description TEXT NOT NULL DEFAULT '',         -- 角色描述

    -- 数据权限字段
    data_scope SMALLINT NOT NULL DEFAULT 1,       -- 数据范围(1:全部 2:自定义 3:本部门 4:本部门及以下 5:仅本人)
    data_scope_custom VARCHAR(500) NOT NULL DEFAULT '',  -- 自定义数据权限(JSON数组，data_scope=2时使用)

    status SMALLINT NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0
);

-- 租户内角色编码唯一约束
CREATE UNIQUE INDEX idx_roles_tenant_code ON roles(tenant_id, role_code) WHERE deleted_at = 0;
CREATE INDEX idx_roles_data_scope ON roles(data_scope);

COMMENT ON TABLE roles IS '角色表(租户隔离，继承关系由Casbin g2策略管理，支持数据权限)';
COMMENT ON COLUMN roles.role_id IS '角色ID(18位字符串)';
COMMENT ON COLUMN roles.tenant_id IS '所属租户ID';
COMMENT ON COLUMN roles.role_code IS '角色编码(租户内唯一，用于Casbin策略)';
COMMENT ON COLUMN roles.name IS '角色名称';
COMMENT ON COLUMN roles.description IS '角色描述';
COMMENT ON COLUMN roles.data_scope IS '数据范围(1:全部数据 2:自定义数据 3:本部门数据 4:本部门及以下 5:仅本人数据)';
COMMENT ON COLUMN roles.data_scope_custom IS '自定义数据权限(JSON数组，data_scope=2时使用，示例:["dept_001","dept_002"])';
COMMENT ON COLUMN roles.status IS '状态(1:启用, 2:禁用)';
COMMENT ON COLUMN roles.created_at IS '创建时间戳(毫秒)';
COMMENT ON COLUMN roles.updated_at IS '更新时间戳(毫秒)';
COMMENT ON COLUMN roles.deleted_at IS '删除时间戳(毫秒,软删除)';
