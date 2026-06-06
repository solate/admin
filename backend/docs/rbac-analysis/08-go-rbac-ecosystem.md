# Go 生态的 RBAC 方案：有没有现成的轮子？

> 短答案：Go 没有 Laravel Spatie Permission 那样的 "拿来即用" RBAC 库。大部分团队都是自己实现。

---

## 1. 为什么 Go 没有自己的 "Spatie"？

### Laravel Spatie 成功的原因

```
Laravel 生态：
  ✅ 统一的 ORM（Eloquent）→ Spatie 知道你的数据长什么样
  ✅ 统一的路由/中间件 → Spatie 知道怎么注册检查
  ✅ 统一的认证系统 → Spatie 知道怎么获取当前用户
  ✅ 约定大于配置 → 所有 Laravel 项目结构相似
  ✅ 一行命令安装 → php artisan vendor:publish

结果：Spatie 只需要定义 5 张表 + 2 个 Trait + 1 个 Gate::before
```

### Go 生态的现实

```
Go 生态：
  ❌ ORM 不统一：GORM、ent、sqlx、sqlc、pgx... → 库不知道你的数据层
  ❌ 框架不统一：Gin、Echo、Chi、Fiber、标准库... → 库不知道你的中间件链
  ❌ 认证不统一：JWT、Session、OAuth... → 库不知道怎么获取当前用户
  ❌ 项目结构不统一 → 每个 Go 项目的目录布局都不同
  ❌ Go 的哲学是 "自己写" → 标准库够用就不引入第三方

结果：一个 RBAC 库要兼容所有组合，复杂度爆炸，不如自己写
```

**核心矛盾**：Go 的 "没有框架约束" 既是优点也是缺点——库作者无法假设你的项目结构。

---

## 2. Go 生态现有的 RBAC 相关库

### 库对比

| 库 | GitHub Stars | 定位 | 多租户 | 能拿来即用吗 |
|---|---|---|---|---|
| [Casbin](https://github.com/casbin/casbin) | 18k+ | 通用策略引擎 | ✅ Domains | ❌ 需要学 DSL，自己写数据层 |
| [goRBAC](https://github.com/mikespook/gorbac) | 1.5k | 轻量 RBAC 接口 | ❌ | ❌ 纯接口，无持久化 |
| [grbac](https://github.com/storyicon/grbac) | 700+ | HTTP 中间件 RBAC | ❌ | ⚠️ 只做路由匹配，不管数据 |
| [gate](https://github.com/panapol-p/gate) | 100+ | Casbin 多租户封装 | ✅ | ❌ 基于 Casbin，同样的痛点 |
| [Permitta](https://github.com/permitta/permitta) | 新项目 | 细粒度 CRUD 权限 | ❌ | ⚠️ 新项目，未验证 |
| [go-admin](https://github.com/go-admin-team/go-admin) | 12k+ | 完整后台脚手架 | ✅ | ⚠️ 它是整个项目，不是库 |

### 逐个分析

#### Casbin — 最流行，但不是 "RBAC 库"

```
Casbin 的本质：
  它不是 RBAC 库，它是"策略引擎"。
  RBAC 只是它支持的众多模型之一。

拿来即用程度：⭐⭐（需要大量胶水代码）

你需要自己做的事：
  1. 设计 model.conf（学习 DSL）
  2. 选择和配置 Adapter（数据库持久化）
  3. 实现 Watcher（多实例缓存同步）
  4. 写中间件集成到 Gin/Echo
  5. 设计策略管理 API
  6. 处理缓存失效

得到的好处：
  + 灵活切换 ACL/RBAC/ABAC
  + 内置策略匹配算法
  + 多语言 SDK 一致
```

#### goRBAC — 纯接口，无实现

```go
// goRBAC 只定义了接口，不包含任何持久化或中间件
type RBAC interface {
    AddRole(role Role) error
    RemoveRole(role Role) error
    AddPermission(role Role, permission Permission) error
    // ...
}

// 你需要自己：
// 1. 实现 Role 和 Permission 接口
// 2. 实现持久化
// 3. 实现中间件
// 本质上等于自己写
```

#### grbac — 只做路由匹配

```go
// grbac 只解决一个问题：这个 HTTP 请求能不能通过？
// 它基于角色 + 路径 + 方法的匹配规则
// 不关心数据从哪来（数据库？配置文件？）
// 不关心多租户
// 不关心权限管理界面

grbac.New(grbac.WithRoleBindingFunc(func(roles []string) {
    // 你自己返回当前用户的角色
    return []*grbac.Role{{...}}
}))
```

#### go-admin — 完整脚手架，不是库

```
go-admin 是一个完整的后台管理系统项目：
  - Gin + GORM + Casbin + Vue
  - 用户管理、角色管理、菜单管理...
  - 代码生成器
  - 多租户

但它是一个"项目"，不是"库"：
  - 你不能 go get 然后在自己项目里用
  - 你要 fork 它，在它的基础上开发
  - 它的技术选型（Casbin）你可能不想要
```

---

## 3. Go 团队实际上是怎么做 RBAC 的？

### 主流做法：自己实现

搜索 Reddit r/golang、Go Forum、Stack Overflow 上的讨论，共识是：

> **Go 的 RBAC 需求太项目特定了，直接用 GORM 写比引入第三方库更简单。**

典型实现只需要：

```
1. 三张表（roles, permissions, user_roles 或 role_permissions）
2. 一个内存缓存（map[roleID][]Permission）
3. 一个中间件（检查缓存）
4. 一个超管绕行（if HasRole("super_admin") skip）
```

**这大概 200-300 行代码。** 引入 Casbin 需要学的 DSL + Adapter + Watcher 配置，代码量可能更多。

### 真实项目的实现模式

```
模式 A：最简单（适合小项目）
  1. roles 表存角色
  2. JWT 中存 role code
  3. 中间件 switch role 检查路由白名单
  → 代码量：~50 行

模式 B：标准（适合中项目，本项目采用）
  1. roles + permissions + role_permissions + user_roles
  2. 内存 PermissionCache 缓存权限
  3. RBAC 中间件查缓存
  4. 超管 HasRole() 绕行
  → 代码量：~300 行

模式 C：完整（适合大项目）
  模式 B + data_scope 数据权限
  + 角色模板
  + 操作审计
  → 代码量：~500 行
```

### 为什么 "自己写" 在 Go 中比在其他语言更可行

```
Go 的优势：
  ✅ GORM/Gen 让数据层代码极少（自动生成）
  ✅ 中间件模式是语言级原语（func(c *gin.Context)）
  ✅ sync.RWMutex 让内存缓存很简单
  ✅ 接口组合让代码结构清晰
  ✅ 没有魔法，每一行都能理解

对比 Laravel/Spatie：
  Spatie 内部有大量"魔法"（Trait、Service Provider、Migration、Blade Directive）
  你不需要理解它们，只需要调用 API
  但如果出了 bug，你需要钻进它的源码
```

---

## 4. 本项目的实现就是 "Go 标准做法"

本项目的 RBAC 实现和主流 Go 项目完全一致：

| 组件 | 本项目 | 说明 |
|------|--------|------|
| 数据模型 | 5 张表 | users, roles, permissions, user_roles, role_permissions |
| 权限缓存 | PermissionCache | sync.RWMutex + map，30s TTL + 事件驱动刷新 |
| 权限检查 | RBAC Middleware | 超管绕行 + CheckAPI |
| 超管 | HasRole("super_admin") | 一行代码绕行 |
| 角色继承 | 递归 CTE | PostgreSQL 原生支持 |
| 租户隔离 | WHERE tenant_id = ? | Repository 层统一过滤 |

**这就是 Go 社区公认的 "正确做法"。** 不需要感到 "我是否应该用一个库"。

---

## 5. 什么时候才需要引入第三方库？

| 场景 | 自己写 | 用 Casbin |
|------|--------|-----------|
| 标准 RBAC（用户→角色→权限） | ✅ | 不值得 |
| 需要在运行时动态切换 ACL/RBAC/ABAC 模型 | — | ✅ |
| 多语言微服务，策略需要跨语言一致 | — | ✅ |
| 需要策略即代码 + GitOps 工作流 | — | ✅ |
| 需要细粒度的 ABAC 策略评估 | — | ✅ |

---

## 6. 如果重新开始，应该怎么做？

如果今天要从零搭建一个 Go 多租户 RBAC 系统：

### 方案一：用脚手架起步（最快）

```
go-admin（12k stars）
  → fork，改掉 Casbin 为纯数据库方案
  → 或者直接接受 Casbin（如果你不介意它的痛点）
  → 前后端都有，开箱即用
```

### 方案二：自己实现（最灵活，推荐）

```
1. 设计表结构（roles, permissions, role_permissions, user_roles）
2. 用 GORM Gen 生成 model 和 query
3. 写 PermissionCache（~100 行）
4. 写 RBAC 中间件（~50 行）
5. 写超管绕行（~5 行）

总时间：1-2 天
结果：完全可控，完全理解，完全匹配业务需求
```

### 方案三：用外部服务（最省心但最贵）

```
Auth0 / WorkOS / Cerbos
  → 权限检查变成 API 调用
  → 不需要自己维护权限系统
  → 但有网络延迟、费用、供应商锁定
```

---

## 7. 结论

```
Go 生态没有 Laravel Spatie 那样的 "RBAC 即插即用" 方案。

这不是 Go 的缺陷，而是 Go 的设计哲学：
  "简单的需求不需要复杂的库"

标准 RBAC 的核心逻辑只有 200-300 行代码：
  - 几张数据库表
  - 一个内存缓存
  - 一个中间件
  - 一个超管绕行

在 Go 中，自己实现比引入 Casbin 更简单、更可控、更好维护。

本项目的实现就是 Go 社区的标准做法，不需要引入任何第三方 RBAC 库。
```

---

## 参考资料

- [Casbin - GitHub](https://github.com/casbin/casbin) (18k+ stars)
- [goRBAC - GitHub](https://github.com/mikespook/gorbac) (1.5k stars)
- [grbac - GitHub](https://github.com/storyicon/grbac) (700+ stars)
- [go-admin - GitHub](https://github.com/go-admin-team/go-admin) (12k+ stars)
- [gate (多租户 Casbin 封装)](https://github.com/panapol-p/gate)
- [Reddit r/golang - Multi-Tenant Permissions](https://www.reddit.com/r/golang/comments/i91zqp/permissions_in_multitenant_rest_api/)
- [Building RBAC in Go](https://martinyonathann.medium.com/building-a-role-and-scope-based-access-control-rbac-system-in-go-e804a8d4072f)

---

*最后更新：2026-06-06*
