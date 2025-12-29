# 数据库设计指南：NULL vs NOT NULL 决策框架

> 本文档基于 Go 实践、GORM 使用规范及项目实际经验，总结数据库字段设计的决策原则。

## 目录

1. [核心设计原则](#核心设计原则)
2. [Go 社区实践](#go-社区实践)
3. [NULL vs NOT NULL 对比](#null-vs-not-null-对比)
4. [决策矩阵](#决策矩阵)
5. [项目规范](#项目规范)
6. [代码模式对比](#代码模式对比)

---

## 核心设计原则

### GORM 官方建议

> **"Use pointer when you want to store zero value as null"**
> —— GORM Documentation

翻译：当你需要将零值存储为 NULL 时，才使用指针类型。

### 实践导向原则

| 设计选择 | 适用场景 |
|---------|---------|
| `NOT NULL DEFAULT ''` | 业务总是需要该字段，空值有意义 |
| `NOT NULL DEFAULT 0`  | 数值类型，零值是有效状态 |
| `NULL` (Go pointer)   | 真正可选，需区分"未设置"与"空值" |
| `NULL` (Go value)     | 使用 sql.NullString 等类型（不推荐） |

---

## Go 社区实践

### 主流项目的选择

| 项目 | 指针使用倾向 | 典型场景 |
|------|------------|---------|
| **Gin** | 极少使用指针 | 配置结构体全部用值类型 |
| **Kubernetes** | 严格区分 | API 对象用指针，内部处理用值 |
| **Docker** | 避免指针 | 配置和状态都用值类型 + zero value |
| **GORM** | 按需使用 | 仅在需要 NULL 语义时用指针 |

### 社区共识

1. **默认使用值类型**：零值（`""`, `0`, `false`）是 Go 的特性，充分利用
2. **指针的代价**：每次访问需要 `nil` 检查，增加代码复杂度
3. **明确语义**：值类型表示"有默认值"，指针表示"可能不存在"

---

## NULL vs NOT NULL 对比

### 数据库层面

| 特性 | NOT NULL DEFAULT '' | NULL |
|------|-------------------|------|
| 存储空间 | 相同 | 可能略少（NULL 位图） |
| 索引效率 | 更高 | NULL 值不参与普通索引 |
| 查询性能 | 更稳定 | 需额外 IS NULL 判断 |
| 数据完整性 | 强制保证 | 需应用层保证 |

### Go 代码层面

```go
// NOT NULL -> string
name := user.Name        // 直接使用
if name == "" {          // 零值判断
    // 处理空值
}

// NULL -> *string
name := user.Name        // 可能 panic
if user.Name != nil {    // 必须先检查
    name := *user.Name   // 再解引用
}
```

**复杂度对比：**
- 值类型：1 次判断（`== ""`）
- 指针类型：2 次判断（`!= nil` + 解引用）

---

## 决策矩阵

### 何时使用 NOT NULL DEFAULT ''

满足以下**任一条件**：

| 条件 | 说明 | 示例 |
|------|------|------|
| 业务总是需要该字段 | 查询、展示必然用到 | `user_name`, `created_at` |
| 空值有明确业务含义 | `""` 和 `NULL` 语义相同 | `phone`（未填 vs 无号码） |
| 避免频繁 nil 检查 | 代码简洁性优先 | `status`, `module` |
| 字段用于索引/JOIN | NULL 影响查询优化 | `tenant_id`, `user_id` |

### 何时使用 NULL

满足以下**任一条件**：

| 条件 | 说明 | 示例 |
|------|------|------|
| 需区分"未设置"和"空值" | 三态逻辑（未填/空/有值） | `deleted_at`（未删除/已删除） |
| 大字段可选节省空间 | TEXT/JSON/BLOB | `request_params`, `old_value` |
| 外键可选关联 | `NULL` 表示无关联 | `department_id`（无部门） |
| 历史遗留或兼容性 | 已有数据无法迁移 | - |

---

## 项目规范

### 日志表字段设计原则

```sql
-- 核心必填字段：NOT NULL
tenant_id        VARCHAR(20) NOT NULL,        -- 租户隔离必需
user_id          VARCHAR(20) NOT NULL,        -- 用户追踪必需
user_name        VARCHAR(100) NOT NULL DEFAULT '', -- 审计展示必需
status           SMALLINT NOT NULL DEFAULT 1,  -- 状态判断必需
created_at       BIGINT NOT NULL,              -- 时间排序必需

-- 可选但常用：NOT NULL DEFAULT ''
user_display_name VARCHAR(100) DEFAULT '',     -- 有默认显示逻辑
module           VARCHAR(50) DEFAULT '',       -- 大部分操作有模块
operation_type   VARCHAR(20) DEFAULT '',       -- 枚举值，空表示未知

-- 真正可选：NULL
login_ip         VARCHAR(50),                  -- 可能获取失败
login_location   VARCHAR(100),                 -- 依赖 IP 解析服务
fail_reason      VARCHAR(255),                 -- 仅失败时有值
request_params   TEXT,                         -- 可选，大字段
old_value        TEXT,                         -- 仅 UPDATE 操作有
new_value        TEXT,                         -- 仅 CREATE/UPDATE 有
```

### 字段类型映射表

| SQL 类型 | Go 类型 (NOT NULL) | Go 类型 (NULL) |
|---------|-------------------|----------------|
| VARCHAR(N) | `string` | `*string` |
| BIGINT | `int64` | `*int64` |
| SMALLINT | `int16` | `*int16` |
| TEXT | `string` | `*string` |
| BOOLEAN | `bool` | `*bool` |
| TIMESTAMP | `time.Time` | `*time.Time` |

---

## 代码模式对比

### 场景 1：字符串字段

```go
// NOT NULL -> string (推荐)
type User struct {
    Name string `gorm:"column:name;not null"`
}

if user.Name != "" {
    fmt.Printf("用户名: %s", user.Name)
}

// NULL -> *string (仅当需要区分"未设置")
type User struct {
    Name *string `gorm:"column:name"`
}

if user.Name != nil && *user.Name != "" {
    fmt.Printf("用户名: %s", *user.Name)
}
```

### 场景 2：数值状态字段

```go
// NOT NULL -> int16 (推荐)
type LoginLog struct {
    Status int16 `gorm:"column:status;not null;default:1"`
}

if log.Status == 1 { // 成功
}

// NULL -> *int16 (通常不必要)
type LoginLog struct {
    Status *int16
}

if log.Status != nil && *log.Status == 1 {
}
```

### 场景 3：时间字段

```go
// 软删除：NULL 有明确语义
type User struct {
    DeletedAt *time.Time `gorm:"column:deleted_at"`
}

if user.DeletedAt == nil {
    // 未删除
}

// 创建时间：NOT NULL
type User struct {
    CreatedAt time.Time `gorm:"column:created_at;not null"`
}
```

---

## 实际应用案例

### 登录日志表 (login_logs)

| 字段 | 类型 | 决策理由 |
|------|------|---------|
| `log_id` | VARCHAR NOT NULL | 主键，必须存在 |
| `tenant_id` | VARCHAR NOT NULL | 多租户隔离核心 |
| `user_id` | VARCHAR NOT NULL | 审计必需 |
| `user_name` | VARCHAR NOT NULL | 展示必需，fallback |
| `user_display_name` | VARCHAR DEFAULT '' | 有默认逻辑（显示 user_name） |
| `login_type` | VARCHAR DEFAULT '' | 枚举，空表示未知 |
| `login_ip` | VARCHAR NULL | 可能获取失败 |
| `login_location` | VARCHAR NULL | 依赖外部服务 |
| `status` | SMALLINT NOT NULL | 业务核心判断 |
| `fail_reason` | VARCHAR NULL | 仅失败时有值 |

### 操作日志表 (operation_logs)

| 字段 | 类型 | 决策理由 |
|------|------|---------|
| `request_params` | TEXT NULL | 可选，大字段 |
| `old_value` | TEXT NULL | 仅 UPDATE 操作有 |
| `new_value` | TEXT NULL | 仅 CREATE/UPDATE 有 |
| `error_message` | TEXT NULL | 仅失败时有值 |

---

## 检查清单

在建表或修改字段时，使用以下清单：

- [ ] 业务是否**总是需要**该字段？ → `NOT NULL`
- [ ] 空 `""` 是否有明确业务含义？ → `NOT NULL DEFAULT ''`
- [ ] 是否需要区分"未设置"和"空值"？ → `NULL`
- [ ] 是否为大字段（TEXT/JSON/BLOB）且可选？ → `NULL`
- [ ] 字段是否用于索引、JOIN 或高频查询？ → `NOT NULL`
- [ ] 是否愿意为简洁代码牺牲一点灵活性？ → `NOT NULL DEFAULT ''`

---

## 总结

### 推荐策略

1. **优先使用 NOT NULL**：90% 的业务字段应该有默认值
2. **少用指针**：仅在确实需要 NULL 语义时使用
3. **一致性优先**：同类字段保持一致（如所有 `*_name` 都是 NOT NULL）
4. **文档先行**：在 Schema 注释中说明设计意图

### 反模式警示

❌ **过度使用 NULL**：
```sql
-- 避免这种设计
CREATE TABLE bad_example (
    user_name VARCHAR(100),        -- 为何可空？
    status SMALLINT,               -- 状态应该明确
    created_at BIGINT              -- 时间必须有
);
```

✅ **推荐设计**：
```sql
CREATE TABLE good_example (
    user_name VARCHAR(100) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL
);
```

---

## 参考资源

- [GORM 官方文档 - Fields](https://gorm.io/docs/models.html#Fields)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [PostgreSQL NULL 处理](https://www.postgresql.org/docs/current/ddl-constraints.html)
- 项目内部：[audit-log-system-design.md](./audit-log-system-design.md)
