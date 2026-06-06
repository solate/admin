# 业界成熟方案解析：他们是怎么做多租户 RBAC 的

> 分析 Keycloak、Auth0、AWS IAM、Laravel Spatie、PostgreSQL RLS 等主流方案的多租户权限设计。

---

## 1. 三种多租户数据隔离模式

所有方案的底层都逃不开这三种数据隔离模式：

```
┌───────────────────────────────────────────────────────────────────┐
│                    多租户数据隔离模式                               │
├───────────────────┬──────────────────┬────────────────────────────┤
│  模式一            │  模式二           │  模式三                    │
│  共享数据库+鉴别列  │  独立 Schema      │  独立数据库                 │
│  (Discriminator)  │  (Schema-per)     │  (Database-per)            │
├───────────────────┼──────────────────┼────────────────────────────┤
│  所有租户共享表     │  共享数据库        │  每个租户一个完整数据库      │
│  tenant_id 列过滤  │  每个 schema 独立  │  完全物理隔离               │
│                   │                   │                            │
│  ┌───────────┐    │  ┌──────────────┐ │  ┌──────────┐ ┌──────────┐│
│  │ users     │    │  │tenant_a.users│ │  │DB:tenant_a│ │DB:tenant_b││
│  │ tenant_id │    │  │tenant_b.users│ │  │  users   │ │  users   ││
│  │ name      │    │  │tenant_c.users│ │  │  roles   │ │  roles   ││
│  └───────────┘    │  └──────────────┘ │  │  perms   │ │  perms   ││
│                   │                   │  └──────────┘ └──────────┘│
├───────────────────┼──────────────────┼────────────────────────────┤
│  成本最低           │  成本中等          │  成本最高                   │
│  隔离最弱           │  隔离中等          │  隔离最强                   │
│  超管最容易实现     │  超管需跨 schema   │  超管需跨数据库              │
│  适合中小规模       │  适合中等规模       │  适合大型企业/合规要求       │
└───────────────────┴──────────────────┴────────────────────────────┘
```

**本项目使用的是模式一（共享数据库 + 鉴别列）**，这也是大多数 SaaS 产品的选择。

---

## 2. Keycloak — 企业级身份管理平台

### 多租户方案

Keycloak 提供三种多租户实现方式：

#### 方案 A：Realm-per-Tenant（推荐用于强隔离）

```
Keycloak Server
├── Realm: Tenant-A        ← 独立的用户、角色、权限
│   ├── Users
│   ├── Roles (admin, editor, viewer)
│   └── Clients
├── Realm: Tenant-B
│   ├── Users
│   ├── Roles
│   └── Clients
└── Master Realm           ← 超级管理员
    └── 可以管理所有 Realm
```

- **隔离性**：最强（每个 Realm 完全独立）
- **超管**：Master Realm 的 admin 可以管理所有 Realm
- **缺点**：每个 Realm 是独立的用户池，同一用户在不同租户需要不同账号
- **适用**：强合规要求、租户间完全隔离

#### 方案 B：单 Realm + Groups/Organizations（推荐用于共享用户）

Keycloak 26+ 新增了 **Organizations** 功能：

```
Single Realm
├── Organization: Tenant-A
│   ├── Members (user1, user2)
│   └── Roles (org-admin, org-member)
├── Organization: Tenant-B
│   ├── Members (user2, user3)  ← user2 同时在两个组织
│   └── Roles
└── Realm Roles
    └── super-admin             ← 跨组织超级管理员
```

- **隔离性**：中等（通过 Organization 边界）
- **超管**：Realm 级角色，跨所有 Organization
- **优点**：同一用户可以属于多个组织
- **适用**：B2B SaaS、顾问公司等跨组织用户

#### 方案 C：单 Realm + Client Roles

```
Single Realm
├── Client: app-tenant-a
│   └── Roles: admin, editor, viewer
├── Client: app-tenant-b
│   └── Roles: admin, editor, viewer
└── Realm Roles
    └── platform-admin
```

- 通过 Client 区分租户
- 最轻量但隔离最弱

### Keycloak 的超管设计

```
超管身份验证流程：
1. 用户登录 Master Realm → 获得 realm-admin 角色
2. 通过 Admin REST API 管理任意 Realm
3. 或者通过 "身份模拟" (Impersonation) 以特定租户用户身份登录
```

**核心模式**：**平台级身份 → 管理接口 → 切换到租户上下文**

---

## 3. Auth0 — 云端身份平台

### 多租户方案

Auth0 使用 **Organizations** 功能实现多租户：

```
Auth0 Tenant
├── Organization: acme-corp
│   ├── Members
│   ├── Roles (org-admin, org-member)
│   └── Connections (SSO 配置)
├── Organization: stark-industries
│   ├── Members
│   └── Roles
└── Global Roles
    └── platform-admin
```

### 权限模型

Auth0 的 RBAC 是 **Organization-scoped** 的：

```
1. 在 API 中定义 Permissions（全局）
   POST /api/v1/users:read
   POST /api/v1/users:write

2. 在 Organization 中定义 Roles（租户级）
   Organization: acme-corp
   ├── Role: Admin → users:read, users:write
   └── Role: Viewer → users:read

3. 将 Roles 分配给用户（在 Organization 上下文中）
   Alice → Admin @ acme-corp
   Alice → Viewer @ stark-industries
```

### 超管模式

Auth0 没有原生的"跨组织超管"概念。典型做法：

```javascript
// 在自定义 Claim 中标记超管
const namespace = 'https://your-app.com';
// 规则或 Action 中
if (user.app_metadata.is_platform_admin) {
  context.idToken[namespace + '/roles'] = ['platform_admin'];
  context.idToken[namespace + '/can_switch_org'] = true;
}

// 应用层检查
if (jwt.roles.includes('platform_admin')) {
  // 允许跨组织访问
}
```

**核心模式**：**自定义 Claim + 应用层检查** — 超管不在标准 RBAC 模型中，通过应用层扩展实现。

---

## 4. AWS IAM — 云基础设施权限

### 多租户方案

AWS IAM 使用 **Account（账号）** 作为租户隔离边界：

```
AWS Organization（管理账号）
├── Account: prod-tenant-a
│   ├── IAM Roles: tenant-admin, developer, viewer
│   └── Resources
├── Account: prod-tenant-b
│   ├── IAM Roles
│   └── Resources
└── Account: management (超管)
    └── 可以 AssumeRole 到任何租户账号
```

### 跨账号访问 — AssumeRole 模式

这是 AWS 最经典的多租户超管设计：

```
┌─────────────┐  AssumeRole   ┌──────────────────┐
│ Management   │─────────────▶│ Tenant-A Account  │
│ Account      │              │ Role: cross-admin  │
│ (超管)        │─────────────▶│ Tenant-B Account  │
│              │  AssumeRole   │ Role: cross-admin  │
└─────────────┘              └──────────────────┘

流程：
1. 超管在 Management Account 中拥有 sts:AssumeRole 权限
2. 每个租户账号中预创建一个 cross-admin IAM Role
3. 该 Role 的 Trust Policy 允许 Management Account 假定
4. 超管通过 AssumeRole API 获取临时凭证
5. 使用临时凭证操作租户资源
```

### ABAC 标签策略

AWS 还支持通过 **资源标签** 实现 ABAC：

```json
{
  "Effect": "Allow",
  "Action": ["s3:GetObject"],
  "Resource": "*",
  "Condition": {
    "StringEquals": {
      "s3:ResourceTag/tenant": "${aws:PrincipalTag/tenant}"
    }
  }
}
```

- 用户的 Principal Tag 标记了可访问的租户
- 资源的 Resource Tag 标记了所属租户
- IAM 自动匹配两者

**核心模式**：**角色假定 (AssumeRole) + 临时凭证** — 超管通过 "借用" 目标租户的身份来访问。

---

## 5. Laravel Spatie Permission — PHP 生态最流行的 RBAC

### 核心设计

Spatie Permission 是 Laravel 生态中最广泛使用的 RBAC 包：

```php
// 数据模型
users ──M:N──> roles ──M:N──> permissions

// 表结构
roles:        id, name, guard_name
permissions:  id, name, guard_name
role_has_permissions:  permission_id, role_id
model_has_roles:      role_id, model_type, model_id
model_has_permissions: permission_id, model_type, model_id
```

### 超管模式 — Gate::before

这是 Spatie 官方推荐的超管实现方式，极其简洁：

```php
// AuthServiceProvider.php
public function boot()
{
    // 一行搞定超管绕行
    Gate::before(function ($user, $ability) {
        return $user->hasRole('super-admin') ? true : null;
    });
}
```

**工作原理**：
1. 每次权限检查前先执行 `Gate::before`
2. 如果用户有 `super-admin` 角色，直接返回 `true`（允许）
3. 否则返回 `null`（继续正常的权限检查流程）

### 多租户方案

Spatie 本身不原生支持多租户。社区有三种做法：

#### 方案 A：独立数据库（最常见）

```
主数据库:        users, tenants, subscriptions
租户数据库 A:     roles, permissions, role_has_permissions, ...
租户数据库 B:     roles, permissions, role_has_permissions, ...
```

- 超管在主数据库中，通过 `Gate::before` 绕行
- 切换租户时切换数据库连接
- Laravel 的 `tenancy` 包自动处理连接切换

#### 方案 B：鉴别列 + team_id

```php
// 给角色加 team_id
Schema::table('roles', function ($table) {
    $table->unsignedBigInteger('team_id')->nullable();
});

// 查询时过滤
$roles = Role::where('team_id', currentTenant()->id)->get();

// 分配角色时指定团队
$user->assignRole(['admin'], currentTenant());
```

#### 方案 C：Filament 多租户面板

```php
// Filament + Spatie 集成
class AdminPanel extends Panel
{
    public function isTenantSubscriptionRequired(Tenant $tenant): bool
    {
        return true;
    }
}

// 超管面板（独立面板）
class SuperAdminPanel extends Panel
{
    // 独立路由，独立权限检查
    protected static ?string $path = 'admin';
}
```

**核心模式**：**`Gate::before` 一行绕行 + 独立数据库或鉴别列隔离**

---

## 6. PostgreSQL RLS — 数据库层的租户隔离

### 行级安全策略

PostgreSQL 的 Row-Level Security (RLS) 可以在数据库层强制租户隔离：

```sql
-- 1. 启用 RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- 2. 创建策略：只能看到自己租户的数据
CREATE POLICY tenant_isolation ON users
  USING (tenant_id = current_setting('app.current_tenant')::VARCHAR);

-- 3. 超管策略：绕过租户隔离
CREATE POLICY super_admin_bypass ON users
  USING (current_setting('app.is_super_admin', true)::BOOLEAN = true);

-- 4. 应用层设置上下文
SET app.current_tenant = 'tenant_a';
SET app.is_super_admin = 'false';
```

### 优势与局限

| 优势 | 局限 |
|------|------|
| **最强的隔离保证** — 即使应用代码有 bug 也不会泄漏 | **性能影响** — 每条 SQL 都要额外过滤 |
| **零信任模型** — 不依赖应用层正确性 | **调试困难** — 查询慢时不容易定位 RLS 问题 |
| **与应用语言无关** | **连接池兼容性** — 需要确保 SET 在连接归还时重置 |
| | **超管需要额外策略** — 增加 IS SUPERUSER 检查 |

**适用场景**：合规要求极高的金融、医疗系统，需要 "即使开发者写错代码也不会泄漏数据" 的保证。

---

## 7. 各方案的超管设计对比

| 方案 | 超管实现方式 | 复杂度 | 安全性 |
|------|-------------|--------|--------|
| **Keycloak** | Master Realm + Admin API 或 Impersonation | 中 | 高 |
| **Auth0** | 自定义 Claim + 应用层检查 | 中 | 中 |
| **AWS IAM** | AssumeRole 跨账号 + 临时凭证 | 高 | 极高 |
| **Spatie** | `Gate::before` 一行代码 | 极低 | 中 |
| **PostgreSQL RLS** | 独立 RLS 策略 | 中 | 极高 |
| **本项目** | `HasRole("super_admin")` 一行代码 | 极低 | 中 |

---

## 8. 核心洞察

从这些成熟方案中，可以总结出以下共性：

### 洞察 1：超管几乎都是"绕行"模式

无论是 Keycloak 的 Master Realm、AWS 的 AssumeRole、还是 Spatie 的 `Gate::before`，超管的核心逻辑都是：

```
if (is_super_admin) skip_all_checks();
```

而不是给超管分配所有权限。**这比给超管分配所有权限更安全、更简单。**

### 洞察 2：租户隔离在数据层，不在权限层

所有方案都在 **数据访问层** 做租户隔离（`WHERE tenant_id = ?`），而不是在权限系统中为每个租户定义独立的权限集。权限系统只负责 "用户能做什么"，租户隔离负责 "用户能看到哪些数据"。

### 洞察 3：角色模板是应对"角色爆炸"的标准做法

WorkOS、Cerbos、Keycloak 都采用 "全局模板 + 租户自定义" 的混合模式：
- 平台提供默认角色模板（Admin、Editor、Viewer）
- 租户可以基于模板克隆、修改
- 租户也可以创建完全自定义的角色

### 洞察 4：简单的方案被用得最多

Laravel Spatie 的 `Gate::before` 是最简单的超管实现，也是 Laravel 生态中被采用最广泛的方案。**在安全和简单之间，业界倾向于选择简单。**

---

## 参考资料

- [Keycloak Organizations 公告](https://www.keycloak.org/2024/06/announcement-keycloak-organizations)
- [Auth0 - Multi-Tenant SaaS Authorization](https://auth0.com/blog/how-to-choose-the-right-authorization-model-for-your-multi-tenant-saas-application/)
- [AWS - SaaS Tenant Isolation with ABAC](https://aws.amazon.com/blogs/security/how-to-implement-saas-tenant-isolation-with-abac-and-aws-iam/)
- [AWS - Cross-Account Access for SaaS](https://aws.amazon.com/blogs/security/how-to-improve-cross-account-access-for-saas-applications-accessing-customer-accounts/)
- [Spatie - Defining a Super-Admin](https://spatie.be/docs/laravel-permission/v7/basic-usage/super-admin)
- [AWS - PostgreSQL RLS for Multi-Tenant](https://aws.amazon.com/blogs/database/multi-tenant-data-isolation-with-postgresql-row-level-security/)
- [Crunchy Data - Designing Postgres for Multi-Tenancy](https://www.crunchydata.com/blog/designing-your-postgres-database-for-multi-tenancy)
- [WorkOS - How to Design Multi-Tenant RBAC](https://workos.com/blog/how-to-design-multi-tenant-rbac-saas)

---

*最后更新：2026-06-06*
