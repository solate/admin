# Casbin 深度分析：为什么有人选它，痛点在哪

> 本文档基于实际项目经验与行业研究，全面分析 Casbin 的优势与局限性。

---

## 1. Casbin 是什么

Casbin 是一个开源的、轻量级的访问控制库，由 Apache 基金会孵化。它的核心理念是 **PERM 元模型**（Policy Effect Request Matching）：

```
Request  →  Policy Matcher  →  Effect (allow/deny)
```

它不关心权限数据存储在哪里（内存、数据库、文件），只关心 **如何匹配和求值**。开发者通过一个 `.conf` 模型文件定义匹配规则，再通过 Adapter 将策略持久化到数据库。

### 核心架构

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  model.conf  │     │  Adapter     │     │  Enforcer    │
│  (匹配规则)   │────▶│  (策略存储)   │────▶│  (求值引擎)   │
└──────────────┘     └──────────────┘     └──────────────┘
```

支持的模型：
- **ACL**（访问控制列表）
- **RBAC**（基于角色的访问控制）
- **ABAC**（基于属性的访问控制）
- **RBAC with domains**（多租户 RBAC）

---

## 2. Casbin 的真正优势

### 2.1 模型灵活性 — 一套配置切换授权模式

这是 Casbin 最大的卖点。通过修改 `model.conf`，你可以在不改动代码的情况下从 RBAC 切换到 ABAC：

```ini
# RBAC 模型
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```

```ini
# 多租户 RBAC 模型（加了一个 domain 字段）
[request_definition]
r = sub, dom, obj, act

[matchers]
m = g(r.sub, p.sub, r.dom) && r.obj == p.obj && r.act == p.act
```

**这意味着**：如果产品需要从简单的 ACL 演进到复杂的 ABAC，理论上只需改配置文件，不需要重写代码。

### 2.2 多语言 SDK — 跨服务一致性

Casbin 有 Go、Java、Python、Node.js、PHP、.NET、Rust 等十多种语言的实现。对于 **微服务架构中使用不同语言的团队**，这意味着授权逻辑可以在不同服务间保持一致的语义。

### 2.3 策略即代码 — Policy as Code

权限策略可以定义在 `.conf` 文件中，纳入 Git 版本控制。这在合规要求严格的场景下（金融、医疗）很有价值：
- 策略变更有 code review 记录
- 可以写单元测试验证策略正确性
- 审计时可以追溯 "谁在什么时候改了什么规则"

### 2.4 内置的多租户支持（RBAC with Domains）

Casbin 的 RBAC 模型原生支持 `domain` 概念：

```
p, admin, tenant_a, data1, read
p, admin, tenant_b, data2, write
g, alice, admin, tenant_a
g, bob, admin, tenant_b
```

这种声明式语法让多租户策略看起来很直观。

### 2.5 纯内存求值 — 高性能

Casbin 的 `Enforcer` 将策略加载到内存后，权限检查是纯内存操作，通常在 **微秒级** 完成。对于高并发场景，这比每次查数据库快得多。

---

## 3. 痛点实录：实际使用中的问题

### 3.1 缓存同步 — "数据库改了还要重启"

这是最常见也最让人头疼的问题。Casbin 的策略存在内存中，数据库的变更不会自动反映：

```
数据库 INSERT 一条策略  →  Casbin 内存中还是旧数据  →  权限检查失败
```

**解决方案及问题**：
| 方案 | 问题 |
|------|------|
| 手动调用 `e.LoadPolicy()` | 需要在每个修改权限的地方记得调用 |
| 定时全量刷新 | 刷新间隔内权限不一致 |
| Watcher 机制（Redis Pub/Sub） | 增加基础设施复杂度，多实例同步仍有延迟 |
| Adapter 回调 | 不是所有 Adapter 都支持 |

**根本原因**：Casbin 的设计假设策略是相对静态的，变更频率低。但实际 SaaS 产品中，管理员可能随时调整角色权限，用户期望 **立即生效**。

### 3.2 必须给 User 和 Role 加 Code

Casbin 的策略匹配基于字符串比较。在 RBAC 模型中：

```
g, alice, admin     ← 这是字符串 "alice" 和 "admin"
p, admin, /api/users, GET  ← 这是字符串匹配
```

这意味着你 **不能用数据库的自增 ID** 做匹配，必须给每个用户和角色一个 **唯一的字符串标识符**（code）。这导致：
- 数据库要额外维护 `role_code` 字段
- 用户如果用数字 ID，需要转换
- 如果 code 变了，所有策略都要同步更新

### 3.3 调试困难 — 策略不生效时无从下手

当一条权限检查返回 `false`，Casbin 只告诉你结果，不告诉你 **为什么**：
- 是策略不存在？
- 是角色没有关联？
- 是匹配器写错了？
- 是 domain 不对？

你需要手动检查：
1. `e.GetPolicy()` 查看所有策略
2. `e.GetRolesForUser()` 查看用户角色
3. `e.GetFilteredPolicy()` 按条件过滤
4. 甚至读 `model.conf` 检查 matcher 表达式

**对比**：纯数据库 RBAC 可以直接 `SELECT * FROM role_permissions WHERE role_id = ? AND resource = ?`，一目了然。

### 3.4 多租户 Domain 模型的局限性

Casbin 的 `domain` 概念看似支持多租户，但实际上：
- Domain 只是一个策略维度，不提供 **数据隔离**
- 你仍然需要在应用层确保查询带 `tenant_id`
- 跨租户的超管需要在每个 domain 下都创建策略，或者在 matcher 中加特殊逻辑
- **角色继承跨 domain 是个难题** — 需要自定义函数

### 3.5 学习曲线 — 非标准化的 DSL

Casbin 的 `.conf` 文件是一种自创的 DSL（领域特定语言）。团队新成员需要学习：
- `r.sub`, `p.eft`, `e` 这些隐含变量的含义
- `some(where (...))` 这种类 SQL 但又不是 SQL 的语法
- `g(r.sub, p.sub, r.dom)` 中的 `g` 函数是什么
- `_` 通配符、`*` 通配符的区别

**对比**：纯数据库 RBAC 用的是标准 SQL，任何后端开发者都能直接理解。

### 3.6 分布式环境的问题

在多实例部署中：
- 每个实例有自己的 `Enforcer` 内存副本
- 策略变更需要广播到所有实例
- Watcher 机制增加了 Redis 等外部依赖
- 如果 Watcher 消息丢失，实例间权限不一致

---

## 4. 什么时候值得用 Casbin？

### ✅ 适合用 Casbin 的场景

| 场景 | 原因 |
|------|------|
| 策略极其复杂，需要 ABAC + RBAC 混合 | Casbin 的 DSL 比 SQL 更适合表达复杂规则 |
| 策略变更频率很低（天/周级别） | 缓存同步问题不明显 |
| 需要策略可审计、可版本控制 | Policy as Code 是真优势 |
| 多语言微服务架构 | 统一的策略语义跨语言 |
| IoT / 边缘计算场景 | 纯库依赖，不需要外部服务 |

### ❌ 不适合用 Casbin 的场景

| 场景 | 原因 |
|------|------|
| 标准 RBAC（用户→角色→权限） | 杀鸡用牛刀，SQL + 内存缓存更简单 |
| 权限需要实时生效 | 缓存同步是硬伤 |
| 多租户 SaaS | domain 模型不够用，应用层还是要做隔离 |
| 团队小、迭代快 | 学习成本 > 收益 |
| 权限数据量巨大（百万级策略） | 全量加载到内存不可行 |

---

## 5. 替代方案对比

| 方案 | 类型 | 适用场景 | 与 Casbin 的区别 |
|------|------|----------|-----------------|
| **纯数据库 RBAC** | 内嵌 | 标准 RBAC + 多租户 | 最简单，SQL 直接查，没有同步问题 |
| **SpiceDB** | 外部服务（Zanzibar） | 关系型权限（ReBAC） | 有状态服务，支持关系推理，更重但更正确 |
| **Cerbos** | 外部服务（PDP） | ABAC + 策略即代码 | 支持 YAML 策略定义，审计友好 |
| **OPA** | 策略引擎 | 通用策略（不只是权限） | 更通用但更复杂，适合云原生 |
| **Keycloak** | 身份平台 | SSO + RBAC + 多租户 | 完整的身份管理，权限只是其中一部分 |

---

## 6. 核心结论

> **Casbin 是一个强大的权限库，但强大不等于适合。**

对于 **标准多租户 RBAC** 场景（90% 的 SaaS 产品）：

```
纯数据库 RBAC + 内存缓存 > Casbin
```

原因很简单：
1. **SQL 是团队已经会的东西** — 不需要学新 DSL
2. **数据库就是 source of truth** — 不存在缓存同步问题
3. **调试就是查表** — 不需要理解 matcher 表达式
4. **多租户天然支持** — `WHERE tenant_id = ?` 搞定
5. **超管只需要一个 `HasRole("super_admin")`** — 不需要全局策略

Casbin 的价值在于 **非标准场景**：当你需要 ABAC、ReBAC、跨域策略组合等复杂逻辑时，它的 DSL 才真正发挥作用。对于"用户属于角色，角色有权限"这种标准需求，纯数据库方案永远是更简单的选择。

---

## 参考资料

- [Authzed - Comparing Casbin and SpiceDB](https://authzed.com/blog/casbin)
- [Casbin 官方文档](https://casbin.org/)
- [Casbin Model Gallery - Multi-Tenant RBAC](https://editor.casbin.org/gallery)
- [Cerbos - Scalable Multitenant Authorization](https://www.cerbos.dev/blog/how-to-implement-scalable-multitenant-authorization)
- [WorkOS - Top RBAC Providers for Multi-Tenant SaaS](https://workos.com/blog/top-rbac-providers-for-multi-tenant-saas-2025)

---

*最后更新：2026-06-06*
