# 跨租户超管设计模式

> 超管的核心矛盾：如何让一个人拥有 "上帝视角" 的同时，不破坏租户隔离的安全承诺。

---

## 1. 超管的本质需求

在多租户系统中，以下角色需要跨租户访问能力：

| 角色 | 需求 | 典型操作 |
|------|------|----------|
| **平台超管 (super_admin)** | 查看和操作所有租户数据 | 系统运维、租户开通、故障排查 |
| **运维 (ops)** | 只读查看所有租户数据 | 监控告警、日志排查、性能分析 |
| **审计 (auditor)** | 只读 + 审计日志 | 合规审计、安全事件调查 |
| **运管 (regulator)** | 查看特定租户数据 | 监管检查、数据抽查 |

### 关键区别

```
普通用户的权限模型：
  "我能做什么" × "我属于哪个租户" = 我的权限范围

超管的权限模型：
  "我能做什么" × "所有租户" = 我的权限范围
```

问题在于：**如果超管绕过了租户隔离，怎么保证超管不会误操作错误的租户数据？**

---

## 2. 模式一：平台级角色（本项目方案）

### 设计

```
┌─────────────────────────────────────────────┐
│              JWT Token Claims                │
├─────────────────────────────────────────────┤
│  user_id: "admin_001"                        │
│  tenant_id: "default"        ← 默认租户      │
│  roles: ["super_admin"]       ← 平台角色      │
│  role_ids: ["role_sa"]                       │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│            RBAC Middleware                    │
├─────────────────────────────────────────────┤
│  if HasRole("super_admin"):                  │
│      c.Next()  // 跳过所有权限检查             │
│      return                                  │
│  // ... 正常权限检查                          │
└─────────────────────────────────────────────┘
```

### 实现代码（本项目）

```go
// rbac_middleware.go
func RBAC(cache *rbac.PermissionCache) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        // 超管绕行 — 一行搞定
        if xcontext.HasRole(ctx, constants.RoleSuperAdmin) {
            c.Next()
            return
        }

        // 正常权限检查
        roleIDs := xcontext.GetRoleIDs(ctx)
        if !cache.CheckAPI(roleIDs, path, method) {
            response.Fail(c, xerr.ErrForbidden)
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 超管如何操作特定租户

本项目通过 **租户切换 (SwitchTenant)** 实现：

```go
// refresh.go - SwitchTenant
func (s *Service) SwitchTenant(ctx context.Context, targetTenantID string) (*TokenPair, error) {
    userID := xcontext.GetUserID(ctx)

    // 超管可以切换到任意租户
    if xcontext.HasRole(ctx, constants.RoleSuperAdmin) {
        return s.generateTokenForTenant(ctx, userID, targetTenantID)
    }

    // 普通用户只能切换到有角色的租户
    roleIDs, err := s.userRoleRepo.GetUserRoleIDs(ctx, userID, targetTenantID)
    if len(roleIDs) == 0 {
        return nil, xerr.ErrForbidden
    }
    return s.generateTokenForTenant(ctx, userID, targetTenantID)
}
```

### 优势

| 优势 | 说明 |
|------|------|
| **极简** | 一行 `HasRole` 搞定，零额外复杂度 |
| **一致** | 超管和普通用户走同一套 JWT 机制 |
| **安全** | 切换租户时重新生成 Token，带目标租户的上下文 |
| **可审计** | 每次操作都有租户上下文，操作日志可追溯 |

### 风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| 超管身份被窃取 | 强制 MFA + 登录告警 + Session 时效短 |
| 超管误操作其他租户 | 操作日志审计 + 敏感操作二次确认 |
| 超管绕过了所有检查 | 代码中明确注释 `// 超管绕行`，审计时容易发现 |

### 适用场景

- **中小型 SaaS**（< 1000 租户）
- 超管数量少（< 10 人）
- 对超管行为有审计要求但不需要实时拦截

---

## 3. 模式二：角色假定 / 身份切换

### 设计灵感

来源于 AWS IAM 的 **AssumeRole** 模式：超管不直接操作租户数据，而是 "借用" 租户管理员身份。

```
┌──────────────┐  Assume Identity  ┌──────────────────┐
│  Platform     │─────────────────▶│  Tenant Context    │
│  Super Admin  │                  │  (as tenant_admin) │
│               │◀─────────────────│                    │
└──────────────┘  Release Identity └──────────────────┘
```

### 实现方式

```go
// 超管登录时，不加载任何租户上下文
// JWT 中只有 platform 级信息
type PlatformClaims struct {
    UserID   string   `json:"user_id"`
    IsSuperAdmin bool `json:"is_super_admin"`
    // 没有 tenant_id、roles 等租户级信息
}

// 超管需要操作租户时，先 "假定" 租户身份
func AssumeTenant(ctx context.Context, targetTenantID string) (*gin.Context, error) {
    if !isPlatformAdmin(ctx) {
        return nil, ErrForbidden
    }

    // 加载目标租户的角色和权限
    tenantCtx := buildTenantContext(ctx, targetTenantID)
    return tenantCtx, nil
}

// 所有操作都在 assumed context 中进行
// 操作日志记录 "谁 (platform_admin_x) 以什么身份 (tenant_a_admin) 做了什么"
```

### 优势

| 优势 | 说明 |
|------|------|
| **最小权限** | 超管每次只获取需要的租户权限 |
| **可追溯** | 日志中明确记录 "平台管理员以租户身份操作" |
| **安全** | assumed identity 有时效，自动过期 |
| **符合合规** | 可以证明 "超管在 X 时间以 Y 身份做了 Z" |

### 劣势

| 劣势 | 说明 |
|------|------|
| **实现复杂** | 需要维护两套 Context（平台级 + 租户级） |
| **用户体验** | 每次切换租户需要额外的 API 调用 |
| **Token 管理** | 需要管理平台 Token + 假定 Token 两套 Token |

### 适用场景

- **大型 SaaS**（> 1000 租户）
- 合规要求严格（金融、医疗）
- 需要精确审计 "谁以什么身份做了什么"
- 超管团队较大（> 10 人），需要权限细分

---

## 4. 模式三：独立的管控平面

### 设计

超管不通过应用 API 操作租户数据，而是通过 **独立的管理后台**：

```
┌─────────────────────────────────────────────────────┐
│                     应用架构                          │
├─────────────────┬──────────────────┬────────────────┤
│  租户应用 API    │  管控平面 API      │  租户数据库     │
│  (Tenant API)   │  (Admin API)     │  (Shared DB)   │
├─────────────────┼──────────────────┤                │
│  JWT Auth       │  MFA + IP 白名单  │                │
│  Tenant 隔离    │  跨租户查询       │                │
│  RBAC 检查      │  独立审计日志     │                │
│  短期 Session   │  操作审批流程     │                │
└─────────────────┴──────────────────┴────────────────┘
```

### 特点

```
租户 API:
  - 路径: /api/v1/users, /api/v1/roles, ...
  - 认证: JWT (Bearer Token)
  - 隔离: 强制 tenant_id 过滤
  - 权限: RBAC PermissionCache 检查

管控 API:
  - 路径: /admin/api/v1/tenants, /admin/api/v1/audit-logs, ...
  - 认证: MFA + IP 白名单 + 独立认证
  - 隔离: 可选择特定租户操作
  - 权限: 独立的平台级角色体系
  - 额外: 操作审批、双人确认、告警通知
```

### 优势

| 优势 | 说明 |
|------|------|
| **最强安全** | 管控平面完全独立，攻击面小 |
| **独立部署** | 管控平面可以部署在内网 |
| **精细控制** | 每个操作都可以有审批流程 |
| **合规友好** | 独立的审计日志、操作记录 |

### 劣势

| 劣势 | 说明 |
|------|------|
| **开发成本高** | 需要维护两套 API |
| **运维成本高** | 独立部署和监控 |
| **不适合小团队** | 只有超大型 SaaS 才需要 |

### 适用场景

- **超大型 SaaS**（> 10000 租户）
- 严格合规（SOC2、ISO 27001、PCI-DSS）
- 超管需要操作审批流程
- 需要物理隔离管控平面（金融、政务）

---

## 5. 三种模式对比

```
复杂度和安全性的权衡：

安全性 ◀────────────────────────────────────▶ 简单性

模式三          模式二           模式一
管控平面      角色假定         平台角色
(AWS风格)   (AssumeRole)    (Spatie风格)

适用于:      适用于:          适用于:
超大型SaaS   大型SaaS         中小型SaaS
金融/政务     合规要求高       通用场景
```

| 维度 | 模式一：平台角色 | 模式二：角色假定 | 模式三：管控平面 |
|------|----------------|----------------|----------------|
| **实现复杂度** | ⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **安全性** | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **审计能力** | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **开发成本** | 低 | 中 | 高 |
| **超管体验** | 好（无缝切换） | 中（需要假定） | 差（独立后台） |
| **团队要求** | 1-2 人 | 2-4 人 | 5+ 人 |

---

## 6. 对本项目的建议

**当前方案（模式一）完全合适**，原因：

1. **项目规模**：后台管理系统，租户数量可控
2. **超管数量**：通常 1-5 人
3. **合规要求**：中等（有审计日志即可）
4. **团队能力**：模式二和三的实现和维护成本不值得

**如果未来需要升级**，建议的演进路径：

```
模式一（当前）
  ↓ 当超管团队 > 5 人，或需要审计 "谁以什么身份操作"
模式二（加 AssumeIdentity 接口）
  ↓ 当租户 > 1000，或需要 SOC2 合规
模式三（独立管控平面）
```

**不需要一步到位**。模式一在 99% 的场景下已经够用。

---

## 参考资料

- [AWS - Cross-Account Access for SaaS](https://aws.amazon.com/blogs/security/how-to-improve-cross-account-access-for-saas-applications-accessing-customer-accounts/)
- [Pronetx - Cross-Account IAM Roles for Multi-Tenant Products](https://pronetx.com/blog/saas-on-top-of-your-customers-aws-accounts-how-we-use-cross-account-iam-roles-to-build-a-multi-tenant-product)
- [Presidio - Multitenancy in AWS with Account per Tenant](https://www.presidio.com/technical-blog/multitenancy-in-aws-with-account-per-tenant-part-2/)
- [Spatie - Defining a Super-Admin](https://spatie.be/docs/laravel-permission/v7/basic-usage/super-admin)
- [Keycloak - Impersonation API](https://www.keycloak.org/docs/latest/server_admin/index.html#impersonation)

---

*最后更新：2026-06-06*
