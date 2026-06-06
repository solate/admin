# 数据库设计指南

> 本文档定义项目的核心设计规范，确保 AI 和开发人员遵循统一规则。

---

## 1. 多租户查询规范

### 三种模式

| 模式 | 使用场景 | 实现方式 |
|------|---------|---------|
| **Auto（默认）** | 普通业务查询 | JWT 自动添加 tenant_id |
| **Skip** | 超管查询所有租户 | `SkipTenantCheck(ctx)` |
| **Manual** | 跨租户查询指定租户 | `ManualTenantMode(ctx)` |

### Repo 层方法命名规范

```go
// 默认方法：查询当前租户（自动模式）
func (r *RoleRepo) GetByCode(ctx context.Context, roleCode string) (*model.Role, error) {
    return r.q.Role.WithContext(ctx).
        Where(r.q.Role.RoleCode.Eq(roleCode)).
        First()
}

// 手动方法：指定租户查询（使用 WithTenant 后缀）
func (r *RoleRepo) GetByCodeWithTenant(ctx context.Context, tenantID, roleCode string) (*model.Role, error) {
    ctx = database.ManualTenantMode(ctx)
    return r.q.Role.WithContext(ctx).
        Where(r.q.Role.TenantID.Eq(tenantID)).
        Where(r.q.Role.RoleCode.Eq(roleCode)).
        First()
}
```

**命名规则**：
- 查询当前租户 → 默认方法名：`GetByCode`, `List`, `GetByID`
- 指定租户查询 → `WithTenant` 后缀：`GetByCodeWithTenant`, `ListWithTenant`

**注意事项**：
1. 默认方法**不要**手动添加 `tenant_id` 条件（自动模式会添加）
2. 手动方法**必须**调用 `ManualTenantMode(ctx)`（安全要求）
3. Tenant 表：超管用 Skip 模式，普通用户无法访问

---

## 2. NULL vs NOT NULL 决策规则

### 快速决策表

| 条件 | 设计 | 示例 |
|------|------|------|
| 业务总是需要该字段 | `NOT NULL` | `created_at`, `tenant_id` |
| 空值有明确含义 | `NOT NULL DEFAULT ''` | `phone`, `status` |
| 需区分"未设置"和"空值" | `NULL` | `deleted_at` |
| 大字段且可选 | `NULL` | `request_params` |
| 用于索引/JOIN | `NOT NULL` | `user_id`, `tenant_id` |

### Go 类型映射

| SQL 类型 | NOT NULL | NULL |
|---------|----------|------|
| VARCHAR | `string` | `*string` |
| BIGINT | `int64` | `*int64` |
| SMALLINT | `int16` | `*int16` |
| TEXT | `string` | `*string` |
| TIMESTAMP | `time.Time` | `*time.Time` |

### 核心原则

> **"Use pointer when you want to store zero value as null"** —— GORM

- 默认用值类型：零值（`""`, `0`）是 Go 特性，充分利用
- 指针的代价：每次访问需 `nil` 检查，增加复杂度
- 优先 `NOT NULL`：90% 业务字段应有默认值

### 示例对比

```go
// 推荐：NOT NULL -> string
type User struct {
    Nickname string `gorm:"column:nickname;not null"`
}
if user.Nickname != "" { }

// 必要时：NULL -> *string
type User struct {
    Phone *string `gorm:"column:phone"`  // 未填 vs 无号码
}
if user.Phone != nil { }
```

---

## 3. 日志表字段设计参考

| 字段类型 | 设计 | 理由 |
|---------|------|------|
| `tenant_id` | `NOT NULL` | 租户隔离核心 |
| `user_id` | `NOT NULL` | 审计必需 |
| `user_name` | `NOT NULL DEFAULT ''` | 展示必需 |
| `status` | `NOT NULL DEFAULT 1` | 状态判断必需 |
| `request_params` | `NULL` | 可选大字段 |
| `old_value` | `NULL` | 仅 UPDATE 有 |
| `error_message` | `NULL` | 仅失败时有值 |

---

## 4. 检查清单

建表/修改字段时：

- [ ] 多租户表是否需要 `tenant_id NOT NULL`？
- [ ] 业务总是需要该字段？ → `NOT NULL`
- [ ] 空值有明确含义？ → `NOT NULL DEFAULT ''`
- [ ] 需区分"未设置"和"空值"？ → `NULL`
- [ ] 字段用于索引/JOIN？ → `NOT NULL`
