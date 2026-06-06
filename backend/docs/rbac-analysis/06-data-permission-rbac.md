# 数据权限 ≠ ABAC：部门/岗位场景下 RBAC 完全够用

> 很多团队把"按部门过滤数据"理解为 ABAC，但实际上这是 RBAC 的标准扩展——给角色加一个 `data_scope` 字段就够了。

---

## 1. 先搞清楚：权限系统的两个维度

任何企业级后台系统的权限，本质上只有两个问题：

```
问题一：用户能做什么操作？  → 功能权限（Feature Permission）
问题二：用户能看到哪些数据？  → 数据权限（Data Permission）
```

```
┌──────────────────────────────────────────────────────────────┐
│                     权限系统的两个维度                         │
├────────────────────────────┬─────────────────────────────────┤
│      功能权限               │         数据权限                 │
│  (Feature Permission)      │    (Data Permission / Scope)    │
├────────────────────────────┼─────────────────────────────────┤
│  "用户能不能访问这个 API?"   │  "用户能看到哪些行?"             │
│  "用户能不能看到这个菜单?"   │  "用户能查到本部门还是全部?"     │
│  "用户能不能点这个按钮?"     │  "用户能操作自己还是所有人的?"   │
├────────────────────────────┼─────────────────────────────────┤
│  控制粒度：操作              │  控制粒度：数据行                │
│  本项目：已实现 ✅           │  本项目：规划中 📋               │
│  实现方式：PermissionCache  │  实现方式：data_scope + GORM     │
│  检查时机：RBAC 中间件       │  检查时机：Repository 查询       │
└────────────────────────────┴─────────────────────────────────┘
```

**关键认知**：这两个维度都属于 RBAC 的范畴。数据权限是 RBAC 的第二维度，不是 ABAC。

---

## 2. 为什么部门/岗位数据权限是 RBAC，不是 ABAC？

### RBAC 的定义

> 用户 → 角色 → 权限（操作 + 数据范围）

数据权限只是在"权限"部分增加了一个维度：**数据范围（data scope）**。

```
传统 RBAC（只有功能权限）：
  用户 Alice → 角色 "HR专员" → 权限 [查看员工列表, 导出工资条]

增强 RBAC（功能权限 + 数据权限）：
  用户 Alice → 角色 "HR专员" → 功能权限 [查看员工列表, 导出工资条]
                              → 数据权限 [本部门及下级]
```

### ABAC 的定义

> 主体属性 + 资源属性 + 环境属性 → 策略评估 → 允许/拒绝

ABAC 的决策依赖于 **请求时的动态属性**，不仅仅是角色。

### 关键区别

| 维度 | 数据权限（RBAC 扩展） | 真正的 ABAC |
|------|----------------------|-------------|
| **决策依据** | 角色的 `data_scope` 字段（静态） | 请求时的属性组合（动态） |
| **配置方式** | 管理员在 UI 上选择下拉框 | 编写策略规则（YAML/代码） |
| **评估时机** | 角色创建/修改时确定 | 每次请求实时评估 |
| **依赖资源内容** | 不依赖（只看 dept_id 列） | 依赖（看 resource.status, resource.amount 等） |
| **SQL 表达** | `WHERE dept_id IN (...)` | 复杂条件组合 |

### 用代码说明区别

```go
// ✅ 数据权限（RBAC 扩展）— 角色上有 data_scope，静态确定
func applyDataScope(query *gorm.DB, role *Role, userDeptID string) *gorm.DB {
    switch role.DataScope {
    case DataScopeAll:
        return query  // 不加过滤
    case DataScopeDept:
        return query.Where("dept_id = ?", userDeptID)
    case DataScopeSelf:
        return query.Where("created_by = ?", userID)
    }
}

// ❌ 真正的 ABAC — 每次请求动态评估资源属性
func evaluateABAC(user User, resource Resource, action string) bool {
    // 检查：金额 < 10000 且 用户是经理 且 资源状态是 "待审批" 且 当前在工作时间
    return resource.Amount < 10000 &&
        user.Level >= Manager &&
        resource.Status == "pending" &&
        time.Now().Hour() >= 9 && time.Now().Hour() <= 18
}
```

**结论**：部门/岗位的数据权限 = 角色带 `data_scope` 字段 + `WHERE dept_id IN (...)`。这就是 RBAC，不是 ABAC。

---

## 3. 五级数据范围设计

### 设计来源

这个设计源自中国开源后台框架的经典实践（RuoYi、JeecgBoot、若依等），经过数千家企业验证：

```
┌───────────────────────────────────────────────────────────────┐
│                    角色的 data_scope 字段                       │
│                                                               │
│  data_scope = 1 → 全部数据        → 无额外 WHERE              │
│  data_scope = 2 → 自定义部门      → WHERE dept_id IN (自定义)  │
│  data_scope = 3 → 本部门          → WHERE dept_id = ?          │
│  data_scope = 4 → 本部门及下级    → WHERE dept_id IN (递归子)  │
│  data_scope = 5 → 仅本人          → WHERE created_by = ?       │
└───────────────────────────────────────────────────────────────┘
```

### 对应的 SQL

```sql
-- 1. 全部数据（admin 角色）
SELECT * FROM orders WHERE tenant_id = ?;

-- 2. 自定义部门（跨部门协作角色）
SELECT * FROM orders WHERE tenant_id = ? AND dept_id IN (?, ?, ?);
-- 自定义部门列表存在 roles.data_scope_dept_ids (JSONB)

-- 3. 本部门（部门经理）
SELECT * FROM orders WHERE tenant_id = ? AND dept_id = ?;
-- ? = 当前用户的 department_id

-- 4. 本部门及下级（大部门总监）
SELECT * FROM orders WHERE tenant_id = ? AND dept_id IN (?, ?, ?, ?);
-- IN 列表 = 用户部门 + 所有子部门（递归 CTE 或 path 前缀匹配）

-- 5. 仅本人（普通员工）
SELECT * FROM orders WHERE tenant_id = ? AND created_by = ?;
-- ? = 当前用户的 user_id
```

### 数据模型变更

只需在 `roles` 表加两个字段：

```sql
ALTER TABLE roles ADD COLUMN data_scope SMALLINT NOT NULL DEFAULT 5;
-- 默认值 5 = 仅本人（最安全）

ALTER TABLE roles ADD COLUMN data_scope_dept_ids JSONB DEFAULT '[]'::jsonb;
-- 仅 data_scope = 2 时有值，例如 ["dept_001", "dept_002", "dept_003"]
```

**就这么简单**。不需要新表，不需要策略引擎，不需要 ABAC。

---

## 4. 多角色合并策略

当用户有多个角色时，取 **最大权限**（宽松策略）：

```
用户 Alice 的角色：
  角色 A "销售专员" → data_scope = 5 (仅本人)
  角色 B "华东区经理" → data_scope = 4 (本部门及下级)
  角色 C "临时审计" → data_scope = 2 (自定义部门: [华东, 华南])

最终权限：
  取最大范围 → data_scope = 2 (自定义)
  合并部门 → [华东, 华南, 华东子1, 华东子2]
```

合并规则（数值越小权限越大）：

```
1(全部) > 2(自定义) > 4(部门+子) > 3(本部门) > 5(仅本人)
```

特殊情况：多个 `data_scope = 2`（自定义）的角色，合并部门列表并去重。

---

## 5. 实现方式：不是中间件，是 Repository 层

### 为什么不在中间件做？

```
中间件层（Auth/RBAC）：
  ✅ 判断"能不能调这个 API" → 知道路径和方法就够了
  ❌ 不能判断"能看到哪些数据行" → 因为中间件不知道查询的是哪张表

Repository 层（查询时）：
  ✅ 知道查的是 orders 表，可以加 WHERE dept_id IN (...)
  ✅ 知道查的是 login_logs 表，可以加 WHERE user_id = ?
```

### 推荐实现：GORM Scope

```go
// pkg/xcontext/datascope.go
type DataScope struct {
    Scope        int32    // 1-5
    DeptIDs      []string // 有效的部门 ID 列表（已展开子部门）
    UserID       string   // 当前用户 ID
    CustomDepts  []string // 自定义部门列表（仅 scope=2）
}

// repository/datascope.go
func ApplyDataScope(query *gorm.DB, scope *xcontext.DataScope, deptCol string, creatorCol string) *gorm.DB {
    if scope == nil {
        return query // 超管或不需要数据权限
    }
    switch scope.Scope {
    case constants.DataScopeAll:
        // 不加任何过滤
    case constants.DataScopeCustom:
        allDepts := append(scope.DeptIDs, scope.CustomDepts...)
        query = query.Where(deptCol+" IN ?", allDepts)
    case constants.DataScopeDept:
        query = query.Where(deptCol+" IN ?", scope.DeptIDs)
    case constants.DataScopeDeptAndSub:
        query = query.Where(deptCol+" IN ?", scope.DeptIDs) // 已展开子部门
    case constants.DataScopeSelf:
        query = query.Where(creatorCol+" = ?", scope.UserID)
    }
    return query
}

// 在 Repository 中使用
func (r *OrderRepo) List(ctx context.Context, ...) ([]*Order, int64, error) {
    query := r.q.Order.WithContext(ctx).
        Where(r.q.Order.TenantID.Eq(xcontext.GetTenantID(ctx)))

    // 加数据权限过滤
    scope := xcontext.GetDataScope(ctx)
    query = ApplyDataScope(query, scope, "dept_id", "created_by")

    return query.Find()
}
```

### 请求完整流程

```
1. 用户请求 GET /api/v1/orders
         │
         ▼
2. Auth 中间件 → 解析 JWT → Context 存入 user_id, tenant_id, role_ids
         │
         ▼
3. RBAC 中间件 → 检查功能权限（能调这个 API 吗？）
         │
         ▼
4. Handler → Service → Repo.List(ctx)
         │
         ▼
5. Repo 从 Context 取 DataScope
   - 如果用户有超管角色 → 不加过滤
   - 如果角色 data_scope = 4 → 展开部门树，WHERE dept_id IN (...)
   - 如果角色 data_scope = 5 → WHERE created_by = userID
         │
         ▼
6. 返回过滤后的数据
```

---

## 6. 真正需要 ABAC 的场景（本项目不需要）

为了对比，列举一些 **真正需要 ABAC** 的场景：

### 场景 A：审批金额阈值

```
规则：经理可以审批 < 10000 的报销，VP 可以审批 < 100000 的，CEO 无上限

这不是"角色有什么权限"的问题，而是"资源的金额属性 × 角色等级"的组合判断。
```

→ 在 RBAC 中如何处理？在 **Service 层加业务规则检查**，不需要 ABAC 引擎：

```go
func (s *Service) ApproveExpense(ctx context.Context, id string) error {
    expense, _ := s.repo.GetByID(ctx, id)
    userLevel := xcontext.GetUserLevel(ctx) // 从角色推导

    if expense.Amount > userLevel.ApprovalLimit {
        return xerr.ErrExceedApprovalLimit
    }
    // ...
}
```

### 场景 B：时间/地点限制

```
规则：只能在工作时间（9:00-18:00）和公司 IP 段内操作敏感功能

这是环境属性，不是角色属性。
```

→ 在 RBAC 中如何处理？在 **中间件或 Service 层加时间检查**：

```go
func (s *Service) SensitiveOperation(ctx context.Context) error {
    if !isBusinessHours() {
        return xerr.ErrOutsideBusinessHours
    }
    // ...
}
```

### 场景 C：资源状态依赖

```
规则：只能编辑"草稿"状态的文档，已发布的不能改
```

→ 在 RBAC 中如何处理？在 **Service 层检查资源状态**：

```go
func (s *Service) UpdateDocument(ctx context.Context, id string) error {
    doc, _ := s.repo.GetByID(ctx, id)
    if doc.Status != "draft" {
        return xerr.ErrDocumentNotEditable
    }
    // ...
}
```

### 核心洞察

> **上面这些"ABAC 场景"，在 99% 的项目中都是在 Service 层用 if/else 处理的业务规则。**
> 不需要 ABAC 策略引擎，不需要 Casbin，不需要 Cerbos。
> **一个 if 判断就是最简单、最可读、最可维护的"ABAC"。**

只有当你的系统需要 **非开发人员（业务管理员）动态配置** 这些规则时，才需要引入 ABAC 策略引擎。对于后台管理系统，规则通常是固定的，写代码比写策略文件更简单。

---

## 7. 本项目现状与实现路径

### 已有的基础设施

| 组件 | 状态 | 位置 |
|------|------|------|
| `TypeData = "DATA"` 常量 | ✅ 已定义 | `pkg/constants/system.go` |
| `GetDescendantIDs()` | ✅ 已实现 | `repository/department_repo.go` |
| `ListByDeptWithChildren()` | ✅ 已实现 | `repository/user_repo.go` |
| 数据权限规划文档 | ✅ 已完成 | `docs/plan2/data-permission-implementation.md` |
| 前端空壳页面 | ✅ 已创建 | `frontend/src/views/access/DataPermissions.vue` |
| 菜单注册 | ✅ 已注册 | `scripts/init_data/seeds/menu.go` |
| 角色表 `data_scope` 字段 | ❌ 未添加 | 需要迁移 |
| GORM Scope / Hook | ❌ 未实现 | 需要开发 |
| 中间件解析 DataScope | ❌ 未实现 | 需要开发 |
| 多角色合并逻辑 | ❌ 未实现 | 需要开发 |

### 推荐实现顺序

```
Phase 1：基础（1-2 天）
  1. 角色表加 data_scope + data_scope_dept_ids 字段（迁移文件）
  2. 定义 DataScope 常量和 xcontext 存储
  3. 登录时加载角色 data_scope → 存入 JWT 或 Context

Phase 2：最常用的两级（半天）
  4. DataScopeAll（全部）— 超管/admin，不加过滤
  5. DataScopeSelf（仅本人）— 普通员工，WHERE created_by = ?

Phase 3：部门级（1 天）
  6. DataScopeDept（本部门）— WHERE dept_id = ?
  7. DataScopeDeptAndSub（本部门及下级）— 递归 CTE 展开子部门
  8. 复用已有的 GetDescendantIDs()

Phase 4：自定义部门（半天）
  9. DataScopeCustom（自定义）— WHERE dept_id IN (JSONB 解析结果)

Phase 5：多角色合并 + 缓存（1 天）
  10. 多角色合并策略
  11. 角色 DataScope 缓存到 PermissionCache
  12. 前端数据权限配置页面
```

### 与 PermissionCache 的关系

```
当前 PermissionCache：
  apiPerms    map[string][]APIPermission   // roleID → API 权限
  menuPerms   map[string][]string          // roleID → 菜单权限
  buttonPerms map[string][]string          // roleID → 按钮权限

扩展后：
  apiPerms    map[string][]APIPermission   // roleID → API 权限
  menuPerms   map[string][]string          // roleID → 菜单权限
  buttonPerms map[string][]string          // roleID → 按钮权限
  dataScopes  map[string]*DataScopeInfo    // roleID → {scope, deptIDs}  ← 新增
```

只需在 `PermissionCache` 的 Refresh 方法中多查一个字段，零额外复杂度。

---

## 8. 总结

```
常见误区：
  "部门数据权限是 ABAC"          → 错，它是 RBAC 的第二维度
  "按条件过滤数据需要策略引擎"    → 错，WHERE dept_id IN (?) 就够了
  "RBAC 只能控制功能不能控制数据" → 错，data_scope 字段扩展即可

正确理解：
  数据权限 = 角色.data_scope 字段 + Repository 层 WHERE 过滤
  这是 RBAC 的标准扩展，中国开源社区已验证数千次（RuoYi、JeecgBoot 等）
  实现成本：2-4 天，不需要任何策略引擎
```

### 一句话结论

> **如果你只是需要"不同角色看到不同范围的数据"，给角色加一个 `data_scope` 字段，在查询时加 `WHERE` 条件，就是最简单、最正确、最可维护的方案。这就是 RBAC，不需要 ABAC。**

---

## 参考资料

- [WorkOS - How to Design Multi-Tenant RBAC](https://workos.com/blog/how-to-design-multi-tenant-rbac-saas) — 三种角色模式（Global/Tenant-scoped/Hybrid）
- [RuoYi 数据权限设计](https://ruoyi.vip/) — 中国最流行的开源后台框架，五级数据权限的原型
- [JeecgBoot 数据权限](https://www.jeecg.com/) — 另一个流行的中国开源后台框架
- [Cerbos - Scalable Multitenant Authorization](https://www.cerbos.dev/blog/how-to-implement-scalable-multitenant-authorization) — 数据权限 vs ABAC 的边界

---

*最后更新：2026-06-06*
