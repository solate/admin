# 推荐设计：最简单的多租户 RBAC

> 如果你能用一张表画清楚权限模型，说明你的设计足够简单。

---

## 1. 设计原则

```
KISS 原则在权限系统中的体现：

1. 用 SQL 能查出来的东西，不要用策略引擎
2. 用 WHERE tenant_id = ? 能隔离的，不要用 RLS
3. 用 HasRole("super_admin") 能绕行的，不要用 AssumeRole
4. 用一行代码能检查的，不要用十行
5. 团队新成员 30 分钟能理解的，才是好设计
```

---

## 2. 数据模型

### ER 图

```
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│   tenants    │      │    users     │      │    roles     │
├──────────────┤      ├──────────────┤      ├──────────────┤
│ tenant_id PK │◀─────│ tenant_id FK │      │ role_id PK   │
│ tenant_code  │      │ user_id PK   │      │ tenant_id FK │
│ name         │      │ email        │      │ role_code    │
│ status       │      │ password     │      │ name         │
│ type         │      │ status       │      │ parent_role_id│◀──┐
└──────────────┘      └──────────────┘      │ status       │   │
                                              └──────────────┘   │
                                                                 │ 角色继承
┌──────────────┐      ┌──────────────────┐                      │
│ permissions  │      │ role_permissions │                      │
├──────────────┤      ├──────────────────┤                      │
│permission_id │◀─────│ role_id FK ──────────────────────────────┘
│ name         │      │ permission_id FK │
│ type         │      │ tenant_id FK     │  ← 租户级权限分配
│ resource     │      └──────────────────┘
│ action       │
│ status       │      ┌──────────────┐
└──────────────┘      │  user_roles  │
  ↑ 全局表             ├──────────────┤
  (无 tenant_id)       │ user_id FK   │
                       │ role_id FK   │
                       │ tenant_id FK │  ← 租户级角色分配
                       └──────────────┘
```

### 关键设计决策

| 决策 | 选择 | 原因 |
|------|------|------|
| **permissions 全局还是租户级** | 全局（无 tenant_id） | API 路径和菜单是产品定义的，不是租户自定义的 |
| **roles 全局还是租户级** | 租户级（有 tenant_id） | 每个租户可能有不同的角色体系 |
| **user_roles 是否带 tenant_id** | 是 | 同一用户在不同租户可以有不同角色 |
| **role_permissions 是否带 tenant_id** | 是 | 权限分配是租户级的 |

### 为什么 permissions 是全局的？

```
权限（API路径、菜单、按钮）是产品功能的一部分：

/api/v1/users        → 这是产品定义的，所有租户共享
菜单: 用户管理        → 这是产品定义的，所有租户共享
按钮: user:create    → 这是产品定义的，所有租户共享

但 "谁能用" 是租户级的：
角色 admin 在租户 A → 拥有 user:create
角色 admin 在租户 B → 没有 user:create
```

这就是 `permissions` 全局 + `role_permissions` 租户级的原因。

---

## 3. 权限检查流程

### 完整请求流程

```
                    请求进入
                       │
                       ▼
              ┌─────────────────┐
              │  Auth Middleware │  提取 JWT → 填充 Context
              │  (JWT 验证)      │  {tenant_id, user_id, roles, role_ids}
              └────────┬────────┘
                       │
                       ▼
              ┌─────────────────┐
              │  RBAC Middleware │
              └────────┬────────┘
                       │
              ┌────────▼────────┐
              │ HasRole(super_   │───Yes──▶ c.Next() 跳过检查
              │ admin)?         │
              └────────┬────────┘
                       │ No
                       ▼
              ┌─────────────────┐
              │ PermissionCache │  内存查找: roleIDs → {path, method}
              │ .CheckAPI()     │
              └────────┬────────┘
                       │
              ┌────Yes─┴──No────┐
              ▼                  ▼
         c.Next()          403 Forbidden
```

### 权限检查的本质

权限检查就是回答一个问题：

> **"这个用户在当前租户中，有没有角色拥有对这个资源的这个操作权限？"**

翻译成 SQL：

```sql
SELECT 1
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id AND rp.tenant_id = ur.tenant_id
JOIN permissions p ON p.permission_id = rp.permission_id
WHERE ur.user_id = ?
  AND ur.tenant_id = ?
  AND p.type = 'API'
  AND p.resource = '/api/v1/users'
  AND p.action = 'GET'
  AND p.status = 1
LIMIT 1;
```

但为了性能，我们把结果缓存到内存中。

---

## 4. PermissionCache 设计

### 缓存结构

```go
type PermissionCache struct {
    mu        sync.RWMutex
    apiPerms   map[string][]APIPermission  // roleID → [{path, method}, ...]
    menuPerms  map[string][]string         // roleID → [menuID, ...]
    buttonPerms map[string][]string        // roleID → [permissionID, ...]
    db         *gorm.DB
    ttl        time.Duration
}
```

### 刷新机制（三重保障）

```
┌─────────────────────────────────────────────────────────────┐
│                   PermissionCache 刷新                        │
├──────────────────┬──────────────────┬───────────────────────┤
│  1. 启动时加载    │  2. TTL 定时刷新  │  3. 事件驱动刷新       │
│                  │                  │                       │
│  watchRefresh()  │  每 30 秒        │  NotifyRefresh()      │
│  协程首次运行     │  ticker 触发     │  在权限变更时调用       │
│                  │                  │                       │
│  适合: 冷启动    │  适合: 兜底       │  适合: 即时生效        │
└──────────────────┴──────────────────┴───────────────────────┘
```

### 角色继承 — 递归 CTE

```sql
-- PostgreSQL 递归 CTE：解析角色祖先链（最深 10 层防循环）
WITH RECURSIVE role_ancestors AS (
    -- 基础：每个角色是自己的祖先
    SELECT role_id, role_id AS ancestor_role_id, 0 AS depth
    FROM roles
    WHERE tenant_id = ? AND deleted_at = 0

    UNION ALL

    -- 递归：沿 parent_role_id 向上查找
    SELECT r.role_id, ra.ancestor_role_id, ra.depth + 1
    FROM roles r
    JOIN role_ancestors ra ON r.parent_role_id = ra.role_id
    WHERE ra.depth < 10 AND r.deleted_at = 0
)
-- 查询：角色 + 所有祖先角色的权限
SELECT DISTINCT p.resource, p.action
FROM role_ancestors ra
JOIN role_permissions rp ON rp.role_id = ra.ancestor_role_id
JOIN permissions p ON p.permission_id = rp.permission_id
WHERE ra.role_id = ANY(?)
  AND p.type = 'API'
  AND p.status = 1;
```

---

## 5. Role Code 的必要性

### 为什么需要 role_code？

这是一个常见的困惑："我已经有 role_id 了，为什么还需要 role_code？"

**答案：因为 JWT 中需要存一个人类可读、跨环境稳定的标识符。**

```
role_id: "550e8400-e29b-41d4-a716-446655440000"  ← UUID，不同环境不同
role_code: "super_admin"                          ← 人类可读，跨环境一致
```

### JWT 中存 Code 的原因

```go
// JWT Claims
type Claims struct {
    Roles   []string `json:"roles"`     // ["super_admin", "admin"] ← 这是 role_code
    RoleIDs []string `json:"role_ids"`  // ["uuid1", "uuid2"]      ← 这是 role_id
}
```

| 字段 | 用途 | 原因 |
|------|------|------|
| `Roles` (code) | **超管判断** | `HasRole("super_admin")` — 硬编码字符串，不会变 |
| `RoleIDs` (ID) | **权限查询** | PermissionCache 用 roleID 做 key |

### 为什么不只用 role_id？

```
如果只用 role_id：

1. 超管判断：HasRoleID("550e8400-e29b-41d4-a716-446655440000")
   → 开发环境、测试环境、生产环境的超管 role_id 不同
   → 硬编码 ID 是反模式

2. 日志可读性：用户角色 [uuid1, uuid2, uuid3]
   → 看不懂，需要额外查表

3. 前端权限：v-if="hasRole('admin')"
   → 前端用 code 更自然
```

### 业界做法

| 系统 | 用 code 吗 | 示例 |
|------|-----------|------|
| **Keycloak** | 是 | Realm Role name: `admin`, `user` |
| **Auth0** | 是 | Role name: `admin`, `editor` |
| **AWS IAM** | 是 | Role name: `S3Admin`, `ReadOnly` |
| **Laravel Spatie** | 是 | Role name: `super-admin`, `admin` |
| **Spring Security** | 是 | `ROLE_ADMIN`, `ROLE_USER` |

**结论**：所有成熟系统都用 code（name）作为角色的业务标识，用 ID 作为数据库主键。这不是多余的设计，而是必要的。

---

## 6. 本项目当前架构评析

### 做对了什么

| 决策 | 评价 |
|------|------|
| ✅ 纯数据库 RBAC（不用 Casbin） | 完全正确，简单且够用 |
| ✅ PermissionCache 内存缓存 | 性能好，O(1) 权限检查 |
| ✅ 超管 `HasRole("super_admin")` 绕行 | 和 Spatie 的 `Gate::before` 一致 |
| ✅ 递归 CTE 角色继承 | 正确、高效、防循环 |
| ✅ 事件驱动缓存刷新 (`NotifyRefresh`) | 权限变更即时生效 |
| ✅ JWT 中同时存 Roles (code) + RoleIDs | 超管判断 + 权限查询各取所需 |
| ✅ 共享数据库 + 鉴别列模式 | 成本低、超管实现简单 |
| ✅ permissions 全局表 | 产品定义的功能，不应该租户级 |

### 可以改进的地方

| 方面 | 现状 | 建议 |
|------|------|------|
| 角色模板 | 没有模板机制，每个租户从零开始 | 加 role_templates，新租户自动克隆默认角色 |
| 权限版本号 | 没有 acl_version | 加 tenant.acl_version，权限变更时自增，缓存 key 包含版本号 |
| 操作审计 | 有操作日志，但没有关联权限决策 | 超管操作的日志中记录当时的角色和权限上下文 |

---

## 7. 与业界方案的对照

```
本项目设计                    业界对应
────────────────────────────────────────────────────
HasRole("super_admin")    →  Spatie Gate::before
PermissionCache (内存)     →  Keycloak 的 Realm Cache
递归 CTE 角色继承          →  Keycloak 的 Composite Roles
NotifyRefresh()           →  Keycloak 的 Cache Invalidation Event
JWT {roles, role_ids}     →  Auth0 的 ID Token Claims
WHERE tenant_id = ?       →  PostgreSQL RLS (简化版)
SwitchTenant              →  AWS IAM AssumeRole (简化版)
permissions 全局表         →  Auth0 的 API Permissions (全局)
roles 租户级               →  Keycloak 的 Organization Roles
```

**结论**：本项目的 RBAC 设计与业界最佳实践高度一致，只是简化了实现（不需要 AWS 级别的 AssumeRole，不需要 Keycloak 级别的 Identity Broker）。这种简化是 **正确的工程决策**。

---

## 8. 最终建议

### 保持现状

当前架构不需要大的改动。它已经是一个 **简单、正确、可维护** 的多租户 RBAC 系统。

### 可选的增强（按优先级排序）

| 优先级 | 增强 | 工作量 | 收益 |
|--------|------|--------|------|
| P1 | 角色模板（新租户自动创建默认角色） | 1-2 天 | 减少租户初始化成本 |
| P2 | ACL 版本号（缓存 key 含版本） | 半天 | 消除缓存不一致的极端情况 |
| P3 | 超管操作审计增强 | 1 天 | 合规加分 |
| P4 | 租户级权限套餐（不同套餐不同权限上限） | 2-3 天 | 商业化变现 |

### 不要做的事

| 不要做 | 原因 |
|--------|------|
| ❌ 引入 Casbin | 已经证明了不合适，纯数据库方案更好 |
| ❌ 实现 AssumeRole | 过度设计，当前 `HasRole` 绕行足够 |
| ❌ 加 ABAC 策略引擎 | 后台管理系统不需要动态属性评估 |
| ❌ 用 PostgreSQL RLS | 增加调试难度，应用层隔离已经够用 |
| ❌ 给 permissions 加 tenant_id | API 路径和菜单是产品定义的，不应该租户级 |

---

## 参考资料

- [WorkOS - How to Design Multi-Tenant RBAC](https://workos.com/blog/how-to-design-multi-tenant-rbac-saas)
- [Spatie - Defining a Super-Admin](https://spatie.be/docs/laravel-permission/v7/basic-usage/super-admin)
- [Cerbos - Scalable Multitenant Authorization](https://www.cerbos.dev/blog/how-to-implement-scalable-multitenant-authorization)
- [Keycloak Organizations](https://www.keycloak.org/2024/06/announcement-keycloak-organizations)

---

*最后更新：2026-06-06*
