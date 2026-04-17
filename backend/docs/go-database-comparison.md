# Go 数据库访问方案全面对比分析

> 本文档对 Go 语言主流数据库访问方案进行深度对比，重点关注 **AI 辅助编程**场景下的开发效率与代码质量，并针对本项目的多租户管理后台给出最终建议。

---

## 目录

1. [方案概览](#1-方案概览)
2. [逐维度深度对比](#2-逐维度深度对比)
   - 2.1 基础功能
   - 2.2 结果映射
   - 2.3 查询构建
   - 2.4 性能
   - 2.5 类型安全
   - 2.6 迁移与 Schema 管理
   - 2.7 学习曲线
   - 2.8 调试与可观测性
   - 2.9 AI 辅助编程友好度
   - 2.10 中间件集成（JWT / Casbin / 定时任务）
   - 2.11 适用场景
3. [评分总览](#3-评分总览)
4. [场景化推荐](#4-场景化推荐)
5. [完整代码示例](#5-完整代码示例)
6. [最终建议](#6-最终建议)

---

## 1. 方案概览

### 1.1 六种方案简介

| 方案 | 版本 | 核心理念 | GitHub Stars | 设计哲学 |
|------|------|---------|-------------|---------|
| **database/sql** | Go 标准库 | 手写 SQL + 手动 Scan | — | 最底层，完全控制 |
| **sqlx** | jmoiron/sqlx | 扩展 database/sql | ~16k | 薄封装，保留 SQL 原味 |
| **GORM + Gen** | gorm.io/gorm | ORM + 代码生成 | ~37k | 面向对象操作数据库 |
| **ent** | entgo.io/ent | Schema-first 代码生成 | ~15k | Go 代码定义 Schema，生成一切 |
| **sqlc** | sqlc-dev/sqlc | SQL-first 代码生成 | ~14k | 写 SQL，生成类型安全 Go 代码 |
| **upper/db** | upperio/v3 | 简化数据访问层 | ~3.5k | 简洁 API，多后端适配 |

### 1.2 快速对比矩阵

| 维度 | database/sql | sqlx | GORM+Gen | ent | sqlc | upper/db |
|------|:---:|:---:|:---:|:---:|:---:|:---:|
| 代码量 | ★ | ★★ | ★★★★ | ★★★ | ★★★★ | ★★★ |
| 类型安全 | ★ | ★★ | ★★★ | ★★★★★ | ★★★★★ | ★★ |
| 性能 | ★★★★★ | ★★★★★ | ★★★ | ★★★ | ★★★★★ | ★★★ |
| AI 友好度 | ★★★ | ★★★★ | ★★★ | ★★ | ★★★★★ | ★★ |
| 复杂查询 | ★★★★★ | ★★★★★ | ★★★ | ★★★ | ★★★★★ | ★★ |
| 学习曲线 | ★★★ | ★★★★ | ★★★ | ★★ | ★★★★ | ★★★ |

> 星级越高越好（类型安全和复杂查询除外，★ 越多 = 越简单/越好）

---

## 2. 逐维度深度对比

### 2.1 基础功能

**对比项**：连接池、事务、批量操作、命名参数、预编译语句

| 功能 | database/sql | sqlx | GORM+Gen | ent | sqlc | upper/db |
|------|:---:|:---:|:---:|:---:|:---:|:---:|
| 连接池 | 内置 | 继承 sql | 内置(sql) | 内置(sql) | 使用 sql | 内置(sql) |
| 事务 | `db.Begin()` | 继承 sql | `db.Transaction()` | `tx.Commit()` | 使用 sql | `sess.Collection()` |
| 批量插入 | 手写 | 手写 | `CreateInBatches()` | `BulkCreate()` | 手写 SQL | 支持 |
| 命名参数 | `sql.Named` | `Named()` | 支持 | 支持 | `sqlc.arg()` | 支持 |
| 预编译语句 | `Prepare()` | 继承 sql | `PrepareStmt:true` | 内置 | 自动生成 | 支持 |

**关键差异**：

- **GORM** 开箱即用最多：批量操作、预编译、钩子回调（Hook）全内置
- **sqlc** 本身不管理连接池和事务，通过 `*sql.DB` / `*sql.Tx` 传入，依赖标准库
- **ent** 事务支持最完善：自动重试、乐观锁、WSDL 风格的事务闭包

```go
// GORM: 批量插入
db.CreateInBatches(users, 100)

// sqlx: 批量插入需要手写
query := `INSERT INTO users (id, name) VALUES (:id, :name)`
result, err := db.NamedExec(query, users)

// sqlc: 写 SQL，自动生成批量插入方法
// -- name: CreateUsers :one
// INSERT INTO users (id, name) VALUES ($1, $2) RETURNING *;
```

### 2.2 结果映射

**对比项**：查询结果映射到结构体的方式和代码量

**database/sql**（手动 Scan，代码最多）：
```go
rows, err := db.Query("SELECT id, name, email FROM users WHERE tenant_id = $1", tenantID)
for rows.Next() {
    var u User
    rows.Scan(&u.ID, &u.Name, &u.Email) // 必须一一对应
}
```

**sqlx**（自动反射映射，代码适中）：
```go
var users []User
db.Select(&users, "SELECT * FROM users WHERE tenant_id = $1", tenantID)
// 自动按字段名/db tag 映射，无需手动 Scan
```

**GORM + Gen**（代码生成，代码最少）：
```go
// Gen 生成的代码，类型安全
users, err := query.User.WithContext(ctx).Where(query.User.TenantID.Eq(tenantID)).Find()
```

**ent**（代码生成，类型极强）：
```go
users, err := client.User.Query().Where(user.TenantID(tenantID)).All(ctx)
```

**sqlc**（SQL-first 代码生成，代码最少）：
```go
// SQL: -- name: GetUsersByTenant :many
//      SELECT * FROM users WHERE tenant_id = $1;
users, err := queries.GetUsersByTenant(ctx, tenantID)
```

| 方案 | 映射方式 | 代码量 | 编译期检查 |
|------|---------|--------|----------|
| database/sql | 手动 `Scan()` | ★ (最多) | 无 |
| sqlx | 反射 + struct tag | ★★★ | 无 |
| GORM+Gen | 反射 + 代码生成 | ★★★★ | 部分 |
| ent | 代码生成 | ★★★★ | 完全 |
| sqlc | 代码生成 | ★★★★★ | 完全 |

### 2.3 查询构建

**对比项**：链式 API、手写 SQL、复杂查询能力

#### 简单 CRUD

所有方案都能轻松完成简单 CRUD，差异主要在表达方式。

#### 复杂查询（多表 JOIN、子查询、窗口函数）

**database/sql / sqlx**：原生 SQL，无任何限制
```sql
SELECT u.*, d.name as dept_name,
  ROW_NUMBER() OVER (PARTITION BY u.tenant_id ORDER BY u.created_at DESC) as rn
FROM users u
JOIN departments d ON u.department_id = d.id
WHERE u.tenant_id = $1 AND u.status = 1
```
> 任何 SQL 都能写，完全自由，可读性取决于 SQL 本身。

**GORM + Gen**：复杂查询需要回退到原生 SQL
```go
// 简单查询用 Gen 链式 API
query.User.Where(query.User.Status.Eq(1)).Find()

// 复杂 JOIN 需要回退到原生 GORM 或手写 SQL
db.Table("users u").
    Select("u.*, d.name as dept_name").
    Joins("JOIN departments d ON u.department_id = d.id").
    Where("u.tenant_id = ? AND u.status = ?", tenantID, 1).
    Find(&results)

// 窗口函数需要完全手写 SQL
db.Raw(`SELECT *, ROW_NUMBER() OVER (...) FROM users`).Scan(&results)
```
> GORM 的链式 API 对复杂查询表达力不足，最终经常回退到原生 SQL。

**ent**：用 Go 代码表达复杂关系，有学习曲线
```go
users, err := client.User.Query().
    Where(user.TenantID(tenantID)).
    WithDepartment().           // 自动 JOIN
    Order(ent.Desc(user.FieldCreatedAt)).
    All(ctx)
// 窗口函数：需要手写或使用 ent 的 raw query
```
> 关系查询（eager loading）非常优雅，但窗口函数、CTE 等高级 SQL 特性支持有限。

**sqlc**：直接写 SQL，代码生成封装
```sql
-- name: GetUsersWithDept :many
SELECT u.*, d.name as dept_name
FROM users u
JOIN departments d ON u.department_id = d.id
WHERE u.tenant_id = @tenant_id
  AND u.status = @status
ORDER BY u.created_at DESC;
```
```go
// 自动生成类型安全的 Go 函数
users, err := q.GetUsersWithDept(ctx, GetUsersWithDeptParams{
    TenantID: tenantID,
    Status:   1,
})
```
> SQL 完全自由，AI 生成 SQL 非常准确，代码生成提供类型安全。

### 2.4 性能

**量化对比**（相对于原生 database/sql 的开销）：

| 方案 | 简单查询 | 复杂查询 | 批量操作 | 内存开销 |
|------|---------|---------|---------|---------|
| database/sql | 基准 (0%) | 基准 (0%) | 基准 (0%) | 最低 |
| sqlx | +0-3% | +0-3% | +0-3% | 极低 |
| GORM | +10-30% | +15-40% | +10-20% | 中等 |
| GORM Gen | +5-15% | +10-25% | +5-15% | 中等 |
| ent | +5-15% | +10-25% | +5-10% | 中等 |
| sqlc | +0-3% | +0-3% | +0-3% | 极低 |

**性能分析**：

- **GORM 开销来源**：反射（struct → SQL 字段映射）、回调链（callback chain）、默认事务包装（可通过 `SkipDefaultTransaction:true` 优化）
- **sqlc 接近零开销**：生成代码本质就是 `db.Query()` + `rows.Scan()`，和手写完全一致
- **ent 性能良好**：代码生成避免了运行时反射，但图遍历可能生成多条 SQL
- **sqlx 开销极小**：仅在 Scan 时用反射，查询本身无额外开销

> 对于本项目（管理后台，非高并发 OLTP），性能差异不构成主要考量因素。GORM 的 10-30% 开销在实际业务中几乎感知不到。

### 2.5 类型安全

**对比项**：编译期能捕获哪些错误

| 错误类型 | database/sql | sqlx | GORM+Gen | ent | sqlc |
|---------|:---:|:---:|:---:|:---:|:---:|
| 列名拼写错误 | 运行时 | 运行时 | Gen: 编译时 | 编译时 | 编译时 |
| 类型不匹配 | 运行时 | 运行时 | Gen: 编译时 | 编译时 | 编译时 |
| 缺少参数 | 运行时 | 运行时 | 编译时 | 编译时 | 编译时 |
| SQL 语法错误 | 运行时 | 运行时 | 运行时 | 运行时 | **生成时** |
| NULL 处理错误 | 运行时 | 运行时 | 部分 | 编译时 | 编译时 |

**关键差异**：

- **sqlc 独有优势**：在代码生成阶段就会验证 SQL 语法（通过 PostgreSQL parser），SQL 写错直接报错
- **ent 类型安全最强**：整个 API 都是生成的，几乎不可能写出不合法的查询
- **GORM Gen** 提供了不错的类型安全（字段名是生成的常量），但字符串 `Where("status = ?", ...)` 仍可能运行时出错
- **sqlx / database/sql** 几乎没有编译期保障，全靠测试和代码审查

### 2.6 迁移与 Schema 管理

| 方案 | 内置迁移 | 外部工具推荐 | 管理方式 |
|------|:---:|---------|---------|
| database/sql | 无 | golang-migrate, goose | SQL 文件版本化 |
| sqlx | 无 | golang-migrate, goose | SQL 文件版本化 |
| GORM | AutoMigrate（简单） | golang-migrate, goose | 代码或 SQL 文件 |
| ent | 内置（自动生成） | 内置 | Go Schema → SQL |
| sqlc | 无 | golang-migrate, goose | SQL 文件版本化 |
| upper/db | 无 | golang-migrate, goose | SQL 文件版本化 |

**本项目现状**：使用 golang-migrate 管理 SQL 迁移文件，这是数据库优先方式的最佳实践，与任何方案都兼容。

> sqlc 和 golang-migrate 是天然搭档：SQL 迁移文件定义 schema → sqlc 读取 schema 生成代码。这正好契合本项目的数据库优先工作流。

### 2.7 学习曲线

| 方案 | 上手时间 | 文档质量 | 社区生态 | 概念复杂度 |
|------|---------|---------|---------|----------|
| database/sql | 1 天 | ★★★★★ (官方) | ★★★★★ | 低 |
| sqlx | 1-2 天 | ★★★★ | ★★★★ | 低 |
| GORM+Gen | 3-5 天 | ★★★★ | ★★★★★ | 中 |
| ent | 1-2 周 | ★★★ | ★★★ | 高 |
| sqlc | 1-2 天 | ★★★★ | ★★★★ | 低 |
| upper/db | 2-3 天 | ★★★ | ★★ | 中 |

**关键差异**：

- **sqlc 最简单**：会 SQL 就会用 sqlc，额外的学习只是 `sqlc.yaml` 配置和注释语法
- **ent 学习曲线最陡**：需要理解 Schema 定义、Edge（关系）、Privacy（权限）等独特概念
- **GORM Gen** 在 GORM 基础上增加了代码生成概念，但本项目已在使用，团队已熟悉

### 2.8 调试与可观测性

**对比项**：查看实际 SQL、慢查询排查、日志记录

| 方案 | SQL 可见性 | 日志集成 | 排查难度 |
|------|----------|---------|---------|
| database/sql | 直接看到 SQL | 需手动包装 | 最简单 |
| sqlx | 直接看到 SQL | 需手动包装 | 最简单 |
| GORM | `logger.Info` 模式 | 内置日志级别 | 简单 |
| ent | 可配置 logger | 内置 | 中等 |
| sqlc | 直接看到 SQL | 标准 log 包 | 最简单 |

**GORM 调试**：
```go
// GORM 内置日志，可直接看到生成的 SQL
db.Debug().Where("tenant_id = ?", tid).Find(&users)
// 输出: [0.123ms] [rows:10] SELECT * FROM `users` WHERE tenant_id = 'xxx' AND deleted_at = 0
```

**sqlc 调试**：
```go
// sqlc 生成的代码就是标准 SQL 调用，可以直接打印
// 或使用 database/sql 的钩子
func logQuery(ctx context.Context, query string, args ...any) {
    log.Printf("SQL: %s, Args: %v", query, args)
}
```

> GORM 的 Debug 模式很方便，但有时生成的 SQL 比预期复杂（如自动 JOIN、预加载），排查需要理解 GORM 的 SQL 生成逻辑。sqlc / sqlx 的 SQL 就是手写的，排查最直观。

### 2.9 AI 辅助编程友好度

> **这是本次对比最重要的维度**。在 AI 编程为主的工作流中，不同方案的 AI 生成准确率差异显著。

#### 评分

| 方案 | AI 生成准确率 | 幻觉风险 | 自动修正能力 | 综合评分 |
|------|:---:|:---:|:---:|:---:|
| database/sql | ★★★ | ★★★ | ★★ | ★★★ |
| sqlx | ★★★★ | ★★ | ★★★ | ★★★★ |
| GORM | ★★★ | ★★★★ | ★★★ | ★★★ |
| GORM Gen | ★★ | ★★★★★ | ★★ | ★★ |
| ent | ★★ | ★★★★★ | ★★ | ★★ |
| sqlc | ★★★★★ | ★ | ★★★★★ | ★★★★★ |

#### 详细分析

**sqlc — AI 生成最准确（★★★★★）**

AI 对 SQL 语法极度熟悉，生成准确率接近 100%。sqlc 的模式是：
1. AI 写 SQL（最擅长的领域）
2. `sqlc generate` 验证 SQL 语法（编译时检查）
3. 生成类型安全的 Go 代码

```
提示词: "写一个查询，获取某租户下状态为启用的用户列表，支持分页"
AI 生成:
-- name: ListActiveUsers :many
SELECT * FROM users
WHERE tenant_id = @tenant_id AND status = 1 AND deleted_at = 0
ORDER BY created_at DESC
LIMIT @limit OFFSET @offset;
✅ sqlc generate → 自动生成 ListActiveUsers() 函数
```

**GORM Gen — AI 幻觉最多（★★）**

GORM Gen 的 API 比较独特，AI 训练数据中 Gen 相关代码较少：
- AI 经常混淆 GORM 原生 API 和 Gen API（如 `db.Where()` vs `query.User.Where()`）
- Gen 的字段表达式（`query.User.UserID.Eq()`）AI 经常生成错误的链式调用
- AI 倾向生成 `db.Model(&User{}).Where("field = ?", val)` 而非 Gen 风格

```go
// AI 实际生成（错误）:
db.Where(query.User.Name.Eq("test")) // Gen 不这样用

// 正确写法:
query.User.WithContext(ctx).Where(query.User.UserName.Eq("test")).First()
```

**GORM 原生 — AI 中等准确（★★★）**

AI 对 GORM 原生 API 非常熟悉（训练数据多），但：
- 经常生成不存在的 API 或过时用法
- 链式调用顺序有时不正确
- 复杂查询（JOIN、子查询）的 GORM 表达经常有误

**ent — AI 最容易出错（★★）**

ent 的独特概念（Schema、Edge、Predicate）在训练数据中相对少：
- AI 经常生成不存在的 ent 方法
- Privacy / PrivacyTensors 等 ent 特有概念 AI 基本无法正确生成
- 关系查询的 API 风格 AI 经常搞混

**sqlx — AI 表现良好（★★★★）**

AI 熟悉 SQL + struct tag 的模式：
```go
// AI 生成（准确率高）
var users []User
err := db.Select(&users,
    "SELECT * FROM users WHERE tenant_id = $1 AND status = $2",
    tenantID, 1)
```

#### AI 编程工作流对比

```
sqlc 工作流（最佳）:
  写 SQL → sqlc generate → 编译检查 → AI 写业务代码调用生成的函数
  ↓
  SQL 语法错误 → 生成时立即发现
  字段类型错误 → 生成时立即发现
  业务代码调用错误 → Go 编译时发现

GORM Gen 工作流（当前）:
  写迁移 → make gen-db → AI 写 Repo/Service 代码
  ↓
  Gen API 使用错误 → Go 编译时发现（部分）
  查询逻辑错误 → 运行时发现
  回退到原生 GORM → 完全无编译保障
```

### 2.10 中间件集成（JWT / Casbin / 定时任务 / pkg 封装）

#### JWT 中间件

所有方案在 JWT 层面无差异 — JWT 是 HTTP 中间件的概念，与数据库访问无关。

#### Casbin 集成

| 方案 | Casbin Adapter | 集成难度 |
|------|---------------|---------|
| GORM | gorm-adapter（官方，成熟） | ★★★★★ 零成本 |
| sqlx / sqlx / sqlc | 需自建 adapter 或用 standard adapter | ★★★ |
| ent | 需自建 adapter | ★★ |

> **这是本项目保留 GORM 的最强理由之一**。`casbin/gorm-adapter/v3` 是官方维护的适配器，开箱即用。切换到 sqlc / sqlx 后，需要自己实现 Casbin 的 Adapter 接口，或使用 file-based adapter。

#### 多租户过滤

这是最关键的对比项：

**GORM Callback 方案（当前方案，最优雅）**：
```go
// pkg/database/scopes.go — 自动注入 WHERE tenant_id = ?
func RegisterCallbacks(db *gorm.DB) error {
    callbacks.Create().Before("gorm:create").Register("tenant:create", tenantCreateCallback)
    callbacks.Query().Before("gorm:query").Register("tenant:query", tenantQueryCallback)
    // ... 所有 CRUD 操作自动添加租户过滤
}
// 业务代码无需关心 tenant_id
query.User.WithContext(ctx).Where(query.User.UserID.Eq(userID)).First()
// 实际执行: SELECT * FROM users WHERE tenant_id = 'xxx' AND user_id = 'yyy'
```

**sqlc 方案（需改造）**：
```sql
-- 方式1：每个查询手动添加 tenant_id 参数
-- name: GetUserByID :one
SELECT * FROM users WHERE tenant_id = @tenant_id AND user_id = @user_id AND deleted_at = 0;

-- 方式2：使用 sqlc 的 sqlc.embed() 减少重复
-- name: GetUserByID :one
SELECT * FROM users
WHERE tenant_id = @tenant_id AND user_id = @user_id AND deleted_at = 0;
```
```go
// 每个 Repo 方法都需要传 tenantID
user, err := q.GetUserByID(ctx, GetUserByIDParams{
    TenantID: xcontext.GetTenantID(ctx),
    UserID:   userID,
})
```

> **对比结论**：GORM Callback 的自动注入在开发体验上远优于 sqlc 的手动传参。但 sqlc 方案更明确、更可控 — 不会出现"忘记加 tenant_id"的隐患，因为每个查询都显式声明。

**sqlx 方案（类似 sqlc，但无代码生成）**：
```go
func (r *UserRepo) GetByID(ctx context.Context, userID string) (*User, error) {
    tenantID := xcontext.GetTenantID(ctx)
    var user User
    err := r.db.GetContext(ctx, &user,
        "SELECT * FROM users WHERE tenant_id = $1 AND user_id = $2 AND deleted_at = 0",
        tenantID, userID)
    return &user, err
}
```

#### 定时任务（Cron）

所有方案与定时任务（robfig/cron）的集成方式一致 — 定时任务调用 Service 层，Service 层调用 Repo 层，与具体数据库访问方案无关。

#### pkg 封装影响

| 方案 | pkg/database 需要 | pkg/xcontext 需要 |
|------|-----------------|-----------------|
| GORM | Connect + Callbacks + Tx | 不变 |
| sqlc | Connect（标准 sql.DB）+ Tx | 不变 + 新增 GetDB/GetTx |
| sqlx | Connect（sqlx.DB）+ Tx | 不变 + 新增 GetDB |

### 2.11 适用场景

| 方案 | 最佳场景 | 不适合场景 |
|------|---------|----------|
| database/sql | 需要极致控制、超高性能 | 快速开发、大量 CRUD |
| sqlx | 中小型项目、注重 SQL 控制 | 需要类型安全的大型项目 |
| GORM+Gen | 中大型项目、团队对 ORM 熟悉 | 需要极致性能、复杂 SQL 多 |
| ent | 大型项目、强类型需求、关系复杂 | 小项目、团队不熟悉 |
| sqlc | **AI 编程为主**、SQL 为主的开发流 | 需要 ORM 风格的链式 API |
| upper/db | 快速原型、多数据库切换 | 生产级大型项目 |

---

## 3. 评分总览

| 维度 | database/sql | sqlx | GORM+Gen | ent | sqlc |
|------|:---:|:---:|:---:|:---:|:---:|
| 基础功能 | 3 | 4 | 5 | 5 | 3 |
| 结果映射 | 2 | 4 | 5 | 5 | 5 |
| 查询构建 | 5 | 5 | 3 | 4 | 5 |
| 性能 | 5 | 5 | 3 | 4 | 5 |
| 类型安全 | 1 | 2 | 4 | 5 | 5 |
| 迁移管理 | 2 | 2 | 3 | 5 | 2 |
| 学习曲线 | 4 | 4 | 3 | 2 | 5 |
| 调试 | 4 | 4 | 4 | 3 | 4 |
| **AI 友好度** | 3 | 4 | 2 | 2 | **5** |
| 中间件集成 | 3 | 3 | **5** | 3 | 3 |
| **总分** | **32** | **37** | **37** | **38** | **42** |

> sqlc 在 AI 友好度上领先明显，综合得分最高。GORM+Gen 在中间件集成（多租户 Callback）上有独特优势。

---

## 4. 场景化推荐

### 场景 1：极致性能 + 完全掌控 SQL

**推荐：sqlc 或 sqlx**

直接写 SQL，零额外开销。sqlc 更优因为提供了类型安全的代码生成，避免了手动 Scan 的错误。

### 场景 2：中等 CRUD（20-100 张表，复杂查询中等）

**推荐：sqlc**

- SQL 文件组织清晰，每个查询一目了然
- 代码生成提供完整的类型安全
- 复杂查询直接写 SQL，无任何限制
- AI 生成 SQL 准确率最高

### 场景 3：快速原型 / 简单管理后台

**推荐：GORM 或 sqlx**

- GORM 的 AutoMigrate + 链式 API 让原型开发极快
- sqlx 如果对 SQL 熟悉也能快速上手

### 场景 4：大型团队 + 强类型 + 自动化迁移

**推荐：ent 或 sqlc**

- ent 提供完整的 Schema 管理、代码生成、GraphQL 集成
- sqlc 更轻量，SQL-first 思维，团队只需懂 SQL

### 场景 5：AI 编程为主（★★★ 最重要 ★★★）

**推荐：sqlc**

核心原因：

1. **AI 生成 SQL 最准确** — SQL 是 AI 最擅长的语言之一，训练数据极其丰富
2. **编译时验证** — `sqlc generate` 会验证 SQL 语法，AI 生成的 SQL 错误立即暴露
3. **最小幻觉风险** — 不需要 AI 理解复杂的 ORM API，只需写 SQL
4. **确定性输出** — 同一段 SQL 生成的 Go 代码是固定的，没有歧义
5. **简单直觉** — "写 SQL → 生成代码 → 调用代码" 的流程对 AI 来说最简单

```
AI 生成流程对比:

sqlc:   提示 → SQL → sqlc generate → 类型安全 Go 代码 ✅ (确定性高)
GORM:   提示 → Go 代码 (链式 API) → 编译 → 可能运行时出错 ⚠️ (幻觉多)
ent:    提示 → Schema 代码 → go generate → 复杂 API 调用 ❌ (幻觉最多)
```

---

## 5. 完整代码示例

### 5.1 sqlc 完整示例（推荐方案）

#### 项目配置 `sqlc.yaml`

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries/"
    schema: "migrations/"
    gen:
      go:
        package: "db"
        out: "internal/dal/db"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_result_struct_pointers: false
        emit_params_struct_pointers: false
        emit_empty_slices: true
        overrides:
          - db_type: "varchar"
            go_type: "string"
          - db_type: "smallint"
            go_type: "int16"
```

#### SQL 查询文件 `queries/users.sql`

```sql
-- name: GetUser :one
SELECT * FROM users
WHERE tenant_id = @tenant_id AND user_id = @user_id AND deleted_at = 0;

-- name: ListUsers :many
SELECT * FROM users
WHERE tenant_id = @tenant_id AND deleted_at = 0
ORDER BY created_at DESC
LIMIT @limit OFFSET @offset;

-- name: CreateUser :one
INSERT INTO users (
    user_id, tenant_id, user_name, password, nickname,
    avatar, phone, email, description, department_id,
    position_id, status, remark
) VALUES (
    @user_id, @tenant_id, @user_name, @password, @nickname,
    @avatar, @phone, @email, @description, @department_id,
    @position_id, @status, @remark
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET nickname = @nickname,
    email = @email,
    phone = @phone,
    department_id = @department_id,
    position_id = @position_id,
    status = @status,
    remark = @remark,
    updated_at = EXTRACT(EPOCH FROM NOW()) * 1000
WHERE tenant_id = @tenant_id AND user_id = @user_id AND deleted_at = 0
RETURNING *;

-- name: DeleteUser :exec
UPDATE users SET deleted_at = EXTRACT(EPOCH FROM NOW()) * 1000
WHERE tenant_id = @tenant_id AND user_id = @user_id AND deleted_at = 0;

-- name: CountUsersByTenant :one
SELECT COUNT(*) FROM users
WHERE tenant_id = @tenant_id AND deleted_at = 0;

-- name: GetUserByEmail :one
-- 登录场景，跨租户查询
SELECT * FROM users
WHERE email = @email AND deleted_at = 0;

-- name: GetUsersWithDepartment :many
-- 复杂 JOIN 查询
SELECT
    u.user_id, u.tenant_id, u.user_name, u.nickname, u.email,
    u.phone, u.status, u.created_at,
    d.name AS department_name,
    p.name AS position_name
FROM users u
LEFT JOIN departments d ON u.department_id = d.id AND d.deleted_at = 0
LEFT JOIN positions p ON u.position_id = p.id AND p.deleted_at = 0
WHERE u.tenant_id = @tenant_id
  AND u.deleted_at = 0
  AND (u.nickname ILIKE '%' || @keyword || '%' OR @keyword = '')
ORDER BY u.created_at DESC
LIMIT @limit OFFSET @offset;

-- name: BatchCountUsersByTenants :many
-- 批量统计多个租户的用户数
SELECT tenant_id, COUNT(*) AS user_count
FROM users
WHERE tenant_id = ANY(@tenant_ids::varchar[]) AND deleted_at = 0
GROUP BY tenant_id;
```

#### 数据库连接初始化 `pkg/database/postgres_sqlc.go`

```go
package database

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
    Host            string
    Port            int
    User            string
    Password        string
    DBName          string
    SSLMode         string
    MaxConns        int32
    MinConns        int32
    MaxConnLifetime time.Duration
}

func ConnectPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
    )

    poolCfg, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to parse pool config: %w", err)
    }

    poolCfg.MaxConns = cfg.MaxConns
    poolCfg.MinConns = cfg.MinConns
    poolCfg.MaxConnLifetime = cfg.MaxConnLifetime

    pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
    if err != nil {
        return nil, fmt.Errorf("failed to create pool: %w", err)
    }

    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return pool, nil
}
```

#### Repository 层 `internal/repository/user_repo.go`

```go
package repository

import (
    "context"
    "fmt"

    "admin/internal/dal/db"
    "admin/pkg/xcontext"

    "github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
    pool *pgxpool.Pool
    q    *db.Queries
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
    return &UserRepo{
        pool: pool,
        q:    db.New(pool),
    }
}

// Create 创建用户
func (r *UserRepo) Create(ctx context.Context, user db.CreateUserParams) (*db.User, error) {
    return r.q.CreateUser(ctx, user)
}

// GetByID 根据ID获取用户（自动注入租户过滤）
func (r *UserRepo) GetByID(ctx context.Context, userID string) (*db.User, error) {
    tenantID := xcontext.GetTenantID(ctx)
    return r.q.GetUser(ctx, db.GetUserParams{
        TenantID: tenantID,
        UserID:   userID,
    })
}

// GetByEmail 根据邮箱获取用户（登录场景，跨租户）
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*db.User, error) {
    return r.q.GetUserByEmail(ctx, email)
}

// List 分页获取用户列表
func (r *UserRepo) List(ctx context.Context, offset, limit int32) ([]db.User, error) {
    tenantID := xcontext.GetTenantID(ctx)
    return r.q.ListUsers(ctx, db.ListUsersParams{
        TenantID: tenantID,
        Limit:    limit,
        Offset:   offset,
    })
}

// Update 更新用户
func (r *UserRepo) Update(ctx context.Context, params db.UpdateUserParams) (*db.User, error) {
    params.TenantID = xcontext.GetTenantID(ctx)
    return r.q.UpdateUser(ctx, params)
}

// Delete 删除用户（软删除）
func (r *UserRepo) Delete(ctx context.Context, userID string) error {
    tenantID := xcontext.GetTenantID(ctx)
    return r.q.DeleteUser(ctx, db.DeleteUserParams{
        TenantID: tenantID,
        UserID:   userID,
    })
}

// CountByTenant 统计租户用户数
func (r *UserRepo) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
    count, err := r.q.CountUsersByTenant(ctx, tenantID)
    return count, err
}

// GetUsersWithDepartment 复杂 JOIN 查询
func (r *UserRepo) GetUsersWithDepartment(ctx context.Context, keyword string, offset, limit int32) ([]db.GetUsersWithDepartmentRow, error) {
    tenantID := xcontext.GetTenantID(ctx)
    return r.q.GetUsersWithDepartment(ctx, db.GetUsersWithDepartmentParams{
        TenantID: tenantID,
        Keyword:  keyword,
        Limit:    limit,
        Offset:   offset,
    })
}

// InTransaction 在事务中执行
func (r *UserRepo) InTransaction(ctx context.Context, fn func(q *db.Queries) error) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx)

    if err := fn(db.New(tx)); err != nil {
        return err
    }
    return tx.Commit(ctx)
}

// BatchCountUsersByTenants 批量统计多个租户用户数
func (r *UserRepo) BatchCountUsersByTenants(ctx context.Context, tenantIDs []string) (map[string]int64, error) {
    results, err := r.q.BatchCountUsersByTenants(ctx, tenantIDs)
    if err != nil {
        return nil, err
    }
    countMap := make(map[string]int64, len(results))
    for _, r := range results {
        countMap[r.TenantID] = r.UserCount
    }
    return countMap, nil
}
```

#### 事务使用示例

```go
func (s *UserService) CreateUserWithRole(ctx context.Context, userParams db.CreateUserParams, roleCodes []string) error {
    return s.userRepo.InTransaction(ctx, func(q *db.Queries) error {
        // 创建用户
        user, err := q.CreateUser(ctx, userParams)
        if err != nil {
            return fmt.Errorf("create user: %w", err)
        }

        // 分配角色
        for _, code := range roleCodes {
            err := q.CreateUserRole(ctx, db.CreateUserRoleParams{
                UserID: user.UserID,
                RoleCode: code,
            })
            if err != nil {
                return fmt.Errorf("assign role %s: %w", code, err)
            }
        }
        return nil
    })
}
```

### 5.2 GORM + Gen 完整示例（当前方案）

> 以下代码直接来自本项目实际实现。

#### 数据库连接 `pkg/database/postgres.go`

```go
package database

import (
    "fmt"
    "strings"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    gormlogger "gorm.io/gorm/logger"
)

type Config struct {
    DSN             string
    MaxIdleConns    int
    MaxOpenConns    int
    ConnMaxLifetime time.Duration
    LogLevel        string
}

var globalDB *gorm.DB

func Connect(cfg Config) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
        Logger: gormlogger.Default.LogMode(parseLogLevel(cfg.LogLevel)),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect database: %w", err)
    }

    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    sqlDB.Ping()

    RegisterCallbacks(db) // 注册多租户回调
    globalDB = db
    return db, nil
}
```

#### 模型（自动生成）`internal/dal/model/users.gen.go`

```go
// Code generated by gorm.io/gen. DO NOT EDIT.
package model

import "gorm.io/plugin/soft_delete"

type User struct {
    UserID             string                `gorm:"column:user_id;type:varchar(20);primaryKey"`
    TenantID           string                `gorm:"column:tenant_id;type:varchar(20);not null"`
    UserName           string                `gorm:"column:user_name;type:varchar(100);not null"`
    Password           string                `gorm:"column:password;type:varchar(100);not null"`
    Nickname           string                `gorm:"column:nickname;type:varchar(100);not null"`
    Email              string                `gorm:"column:email;type:varchar(100);not null"`
    Status             int16                 `gorm:"column:status;type:smallint;not null;default:1"`
    CreatedAt          int64                 `gorm:"column:created_at;autoCreateTime:milli"`
    UpdatedAt          int64                 `gorm:"column:updated_at;autoUpdateTime:milli"`
    DeletedAt          soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli"`
}
```

#### Repository `internal/repository/user_repo.go`

```go
package repository

import (
    "context"
    "admin/internal/dal/model"
    "admin/internal/dal/query"
    "admin/pkg/xcontext"
    "gorm.io/gorm"
)

type UserRepo struct {
    db *gorm.DB
    q  *query.Query
}

func NewUserRepo(db *gorm.DB) *UserRepo {
    return &UserRepo{db: db, q: query.Use(db)}
}

// GetByID — 租户过滤由 Callback 自动注入
func (r *UserRepo) GetByID(ctx context.Context, userID string) (*model.User, error) {
    return r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(userID)).First()
}

// GetByEmail — 跨租户查询，手动跳过
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    ctx = xcontext.SkipTenantCheck(ctx)
    return r.q.User.WithContext(ctx).Where(r.q.User.Email.Eq(email)).First()
}

// ListWithFilters — 带筛选的分页查询
func (r *UserRepo) ListWithFilters(ctx context.Context, offset, limit int,
    nicknameFilter string, statusFilter int) ([]*model.User, int64, error) {
    q := r.q.User.WithContext(ctx)
    if nicknameFilter != "" {
        q = q.Where(r.q.User.Nickname.Like("%" + nicknameFilter + "%"))
    }
    if statusFilter != 0 {
        q = q.Where(r.q.User.Status.Eq(int16(statusFilter)))
    }
    total, err := q.Count()
    if err != nil {
        return nil, 0, err
    }
    users, err := q.Order(r.q.User.CreatedAt.Desc()).Offset(offset).Limit(limit).Find()
    return users, total, err
}

// CountByTenantIDs — 复杂聚合查询，回退原生 GORM
func (r *UserRepo) CountByTenantIDs(ctx context.Context, tenantIDs []string) (map[string]int64, error) {
    ctx = xcontext.SkipTenantCheck(ctx)
    type Result struct {
        TenantID  string
        UserCount int64
    }
    var results []Result
    err := r.db.WithContext(ctx).Table("users").
        Select("tenant_id, count(*) as user_count").
        Where("tenant_id IN ?", tenantIDs).
        Group("tenant_id").Find(&results).Error
    // ... 转为 map
    countMap := make(map[string]int64, len(results))
    for _, r := range results {
        countMap[r.TenantID] = r.UserCount
    }
    return countMap, err
}
```

#### 事务 `pkg/database/tx.go`

```go
func InTransactionWithCtx(ctx context.Context, db *gorm.DB, fn func(ctx context.Context, tx *Tx) error) error {
    return db.WithContext(ctx).Transaction(func(txDB *gorm.DB) error {
        return fn(ctx, &Tx{DB: txDB, Q: query.Use(txDB)})
    })
}
```

### 5.3 多租户实现对比

| 特性 | GORM Callback（当前） | sqlc 显式传参 |
|------|---------------------|-------------|
| 自动注入 | 自动（Callback） | 手动（每个查询传 tenantID） |
| 遗漏风险 | 低（Callback 自动处理） | 中（需人工/AI 确保每个查询都有） |
| 可控性 | 中（有时不清楚哪些查询被过滤） | 高（每个查询显式声明） |
| 调试难度 | 中（需理解 Callback 链） | 低（SQL 直接可见） |
| 跨租户操作 | `SkipTenantCheck(ctx)` | 直接不传 tenantID |
| 代码量 | 少（Callback 一次性写好） | 多（每个 SQL 都写 tenant_id 条件） |

---

## 6. 最终建议

### 6.1 核心结论

**对本项目而言，不建议立即迁移，建议优化当前 GORM + Gen 方案。**

理由：

1. **已有投入沉没成本大**：21 个模型、12 个 Repo、多租户 Callback、Casbin Adapter、Soft Delete 插件 — 迁移工作量大
2. **GORM Callback 多租户方案是核心优势**：这套机制已经成熟稳定，sqlc 无法直接复制这种优雅
3. **Casbin GORM Adapter 无替代品**：切换需要自己实现 Adapter 接口
4. **AI 编程可以通过优化提示词弥补**：为 GORM Gen 提供充分的示例代码和 CLAUDE.md 规范

### 6.2 如果要迁移到 sqlc

**迁移路线图**：

1. **保留 GORM 仅用于 Casbin Adapter** — Casbin 仍用 GORM，业务代码迁移到 sqlc
2. **新建 `queries/` 目录** — 按 Repo 组织 SQL 文件
3. **实现多租户封装层** — 在 Repo 层统一注入 tenantID
4. **逐个模块迁移** — 先迁移简单的 DictType/DictItem，验证可行性
5. **并行运行** — GORM 和 sqlc 共存期间，新功能用 sqlc

**预期工作量**：
- SQL 文件编写：每个 Repo 约 2-3 小时（AI 生成，人工审核）
- Repo 层改写：每个 Repo 约 1-2 小时
- 多租户封装：约 4-8 小时
- 总计（12 个 Repo）：约 40-60 小时

### 6.3 优化当前 GORM + Gen 方案

**立即可行的改进**：

1. **在 CLAUDE.md 中增加 GORM Gen 代码规范**
   - 提供 Gen API 的标准用法模板
   - 列出常见错误和正确写法对比
   - 这能大幅提升 AI 生成 Gen 代码的准确率

2. **减少 Gen 回退到原生 GORM 的场景**
   - 复杂查询直接用 `db.Raw()` + 明确的 SQL
   - 在 CLAUDE.md 中规定：复杂查询用原生 SQL，简单 CRUD 用 Gen

3. **为 AI 提供项目专属的代码模板**
   ```
   # 新增 Repo 模板
   type XxxRepo struct {
       db *gorm.DB
       q  *query.Query
   }
   func NewXxxRepo(db *gorm.DB) *XxxRepo { ... }
   // 自动租户过滤方法：不用 SkipTenantCheck
   // 跨租户方法：使用 xcontext.SkipTenantCheck(ctx)
   ```

4. **新模块考虑 sqlc**
   - 如果未来新增独立模块（如报表、数据分析），可以用 sqlc 实现
   - GORM 和 sqlc 可以共存，共享同一个数据库连接

### 6.4 最终选择矩阵

| 条件 | 选择 |
|------|------|
| 当前项目继续维护 | **保留 GORM + Gen**，优化 AI 提示 |
| 全新项目（无历史包袱） | **sqlc** + golang-migrate |
| 团队不熟悉 GORM | **sqlx**（最简单的中间方案） |
| 需要极强类型 + Graph API | **ent** |
| 纯数据分析 / ETL | **database/sql** 或 **sqlx** |

### 6.5 给 AI 编程为主团队的建议

**核心原则**：让 AI 做它最擅长的事 — 写 SQL。

无论选择哪种方案，都应该：
1. **数据库优先**：先写迁移 SQL → 再生成代码 → 最后写业务逻辑
2. **复杂查询用原生 SQL**：不要用 ORM 的链式 API 拼 JOIN
3. **在 CLAUDE.md 中维护代码模板**：AI 按模板生成，减少幻觉
4. **Repo 层严格封装**：业务层永远不直接操作数据库
5. **每个查询都写测试**：AI 生成的代码必须有测试覆盖

---

> **总结**：对于本项目，保留 GORM + Gen 是最务实的选择。如果要为新项目选型，sqlc 是 AI 时代的最佳方案。关键不在于用哪个工具，而在于建立清晰的开发规范，让 AI 在框架内高效工作。
