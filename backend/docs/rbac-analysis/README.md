# 多租户 RBAC 权限系统分析

> 从 Casbin 迁移到纯数据库 RBAC 的设计决策分析，以及业界成熟方案对比。

---

## 核心结论

### 对本项目而言，纯数据库 RBAC 是最优解

```
Casbin（已移除）                    纯数据库 RBAC（当前方案）
─────────────────                  ─────────────────────
策略引擎 + 内存缓存                  SQL + 内存缓存
自定义 DSL 配置                     标准 SQL 查询
缓存同步困难                        事件驱动即时刷新
调试需要理解 matcher                 直接查表，一目了然
多租户 domain 有限                   WHERE tenant_id = ? 搞定
超管需要全局策略                     HasRole("super_admin") 一行
学习曲线陡峭                         团队 30 分钟上手
```

### 业界共识

- **90% 的 SaaS 产品** 只需要 RBAC，不需要 ABAC 或 ReBAC
- **超管的最佳实现** 是 `Gate::before` / `HasRole()` 绕行模式（Laravel Spatie、Spring Security 都这么做）
- **租户隔离在数据层** 做（`WHERE tenant_id = ?`），不在权限层做
- **角色模板 + 租户自定义** 是应对角色爆炸的标准做法

---

## 文档目录

| 文档 | 内容 | 阅读时间 |
|------|------|----------|
| [01-casbin-analysis.md](01-casbin-analysis.md) | Casbin 深度分析：为什么有人选它，痛点在哪 | 10 分钟 |
| [02-abac-vs-rbac.md](02-abac-vs-rbac.md) | ABAC vs RBAC 全面对比（含多租户场景） | 10 分钟 |
| [03-industry-patterns.md](03-industry-patterns.md) | 业界成熟方案（Keycloak/Auth0/AWS IAM/Spatie/PostgreSQL RLS） | 15 分钟 |
| [04-super-admin-design.md](04-super-admin-design.md) | 跨租户超管的三种设计模式 | 10 分钟 |
| [05-recommended-design.md](05-recommended-design.md) | 推荐的简单多租户 RBAC 设计 + 本项目评析 | 15 分钟 |
| [06-data-permission-rbac.md](06-data-permission-rbac.md) | 数据权限 ≠ ABAC：部门/岗位场景 RBAC 完全够用 | 12 分钟 |
| [07-abac-rebac-real-world.md](07-abac-rebac-real-world.md) | ABAC/ReBAC/跨域策略的真实产品场景（Google Drive/GitHub/AWS） | 15 分钟 |
| [08-go-rbac-ecosystem.md](08-go-rbac-ecosystem.md) | Go 生态 RBAC 方案调研：有没有现成的轮子？ | 10 分钟 |

---

## 本项目架构概览

```
                        ┌─────────────┐
                        │   Client     │
                        └──────┬──────┘
                               │ HTTP Request
                               ▼
                    ┌─────────────────────┐
                    │    Auth Middleware    │  JWT 验证
                    │    (jwtManager)      │  → Context: tenant_id, user_id, roles, role_ids
                    └──────────┬──────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │    RBAC Middleware    │
                    │  ┌─────────────────┐ │
                    │  │ super_admin?    │ │──Yes──▶ Next()
                    │  │ CheckAPI()      │ │──No───▶ 403
                    │  └─────────────────┘ │
                    └──────────┬──────────┘
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                 ▼
        ┌──────────┐   ┌──────────┐      ┌──────────┐
        │ Handler  │   │ Service  │      │  Repo    │
        │ (Gin)    │──▶│ (业务)    │───▶  │ (Gen)    │
        └──────────┘   └──────────┘      └──────────┘
                                             │
                                             ▼
                                    WHERE tenant_id = ?
                                    (租户数据隔离)
```

### 关键组件

| 组件 | 文件 | 职责 |
|------|------|------|
| PermissionCache | `backend/pkg/cache/permission.go` | 内存缓存角色权限，递归 CTE 解析继承 |
| RBAC Middleware | `backend/internal/middleware/rbac_middleware.go` | 超管绕行 + API 权限检查 |
| JWT Claims | `backend/pkg/utils/jwt/jwt.go` | 携带 tenant_id, roles(code), role_ids |
| xcontext | `backend/pkg/xcontext/` | 请求上下文传播（租户、用户、角色） |
| SwitchTenant | `backend/internal/service/auth/refresh.go` | 超管切换租户上下文 |

### 数据模型

```
全局表（无 tenant_id）:  permissions, menus
租户表（有 tenant_id）:  users, roles, departments, positions, ...
关联表（有 tenant_id）:  user_roles, role_permissions
平台表（单例）:          tenants
```

---

## 快速 FAQ

### Q: 为什么不用 Casbin？

权限类型固定（API/MENU/BUTTON），不需要策略引擎的灵活性。Casbin 的缓存同步问题让权限无法即时生效，而纯数据库方案用事件驱动刷新可以做到秒级生效。详见 [01-casbin-analysis.md](01-casbin-analysis.md)。

### Q: 为什么需要 role_code？

JWT 中需要存人类可读、跨环境稳定的标识符。`HasRole("super_admin")` 比 `HasRoleID("550e8400-...")` 更可读、更安全。所有成熟系统（Keycloak、Auth0、Spatie、AWS IAM）都这么做。详见 [05-recommended-design.md](05-recommended-design.md#5-role-code-的必要性)。

### Q: 超管是怎么实现的？

一行代码：`if HasRole("super_admin") { c.Next(); return }`。和 Laravel Spatie 的 `Gate::before` 模式完全一致。超管切换租户通过 `SwitchTenant` API 生成新 Token。详见 [04-super-admin-design.md](04-super-admin-design.md)。

### Q: 租户隔离是怎么保证的？

两层保证：
1. **结构层**：所有业务表有 `tenant_id` 列，Repository 层强制 `WHERE tenant_id = ?`
2. **运行时层**：JWT 中携带 `tenant_id`，Auth Middleware 注入 Context，Service 层从 Context 读取

详见 [05-recommended-design.md](05-recommended-design.md#2-数据模型)。

### Q: 什么时候需要 ABAC？

真正的 ABAC 场景：审批金额阈值、时间/地点限制、资源状态依赖等 **动态属性组合**。但这类需求在 99% 的项目中都是在 Service 层用 if/else 处理的业务规则，不需要 ABAC 策略引擎。详见 [02-abac-vs-rbac.md](02-abac-vs-rbac.md)。

### Q: 部门/岗位的数据权限是不是 ABAC？

**不是。** 这是 RBAC 的标准扩展——给角色加一个 `data_scope` 字段（全部/自定义/本部门/本部门及下级/仅本人），在查询时加 `WHERE dept_id IN (...)` 过滤。中国开源后台框架（RuoYi、JeecgBoot）已验证数千次。详见 [06-data-permission-rbac.md](06-data-permission-rbac.md)。

---

## 参考资料

- [WorkOS - How to Design Multi-Tenant RBAC](https://workos.com/blog/how-to-design-multi-tenant-rbac-saas)
- [Authzed - Comparing Casbin and SpiceDB](https://authzed.com/blog/casbin)
- [Cerbos - Scalable Multitenant Authorization](https://www.cerbos.dev/blog/how-to-implement-scalable-multitenant-authorization)
- [AWS - SaaS Tenant Isolation with ABAC](https://aws.amazon.com/blogs/security/how-to-implement-saas-tenant-isolation-with-abac-and-aws-iam/)
- [Spatie - Defining a Super-Admin](https://spatie.be/docs/laravel-permission/v7/basic-usage/super-admin)
- [Keycloak Organizations](https://www.keycloak.org/2024/06/announcement-keycloak-organizations)
- [Permit.io - RBAC vs ABAC vs ReBAC](https://www.permit.io/blog/rbac-vs-abac-and-rebac-choosing-the-right-authorization-model)
- [AWS - Cross-Account Access for SaaS](https://aws.amazon.com/blogs/security/how-to-improve-cross-account-access-for-saas-applications-accessing-customer-accounts/)
- [Crunchy Data - Designing Postgres for Multi-Tenancy](https://www.crunchydata.com/blog/designing-your-postgres-database-for-multi-tenancy)

---

*最后更新：2026-06-06*
