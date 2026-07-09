# GORM vs 原生 SQL：AI 全程写代码，还值得用 ORM 吗

> 2026-07-09。本文回答一个具体问题：既然本项目代码全程由 AI 写、没人手写 SQL，
> 那能不能干脆去掉 GORM，直接用 Go 原生 `database/sql` + 手写 model/repo，
> 让技术栈少一个组件、更「简单」？
>
> 复用 [[02-数据库访问层选型调研]] 的 7 条约束与 [[06-GORM与golang-migrate最佳实践]]
> 的「两种心智模型」框架，不重新论证。02 号文比的是 GORM vs Bob vs sqlc（都是代码
> 生成方案），本文补的是它们没正面碰过的选项：**手写 `database/sql`/sqlx**——也就是
> 用户真正在问的那个。日常写代码不必读，想搞明白「为什么不去掉 GORM」时看这里。

## 结论速览

**保持 GORM + gorm/gen，不换原生 SQL。**

一句话理由：**AI 时代选型标准变了**——不是「哪个初写更快」（AI 写哪个都快），
而是「哪个把 AI 的手误挡在编译期、哪个改 schema 时不用 AI 手动同步散落各处的
列顺序」。这两条 GORM+gen 全胜。原生 SQL 省下的是「一个依赖」，换来的是「每类
AI 手误都要等到运行时才炸 + 每次改列都要人肉对齐 SELECT/Scan」的长期维护税。

## 一、"AI 全程写代码"这个前提，把问题问反了

用户的直觉是：反正 AI 写代码，那 GORM 的类型安全查询 API、gen 生成的 model
这些「让人写起来舒服」的东西，AI 又不需要，去掉不就更简单？

这个推理里藏了一个错误前提——**它假设「选型是为了让写代码的人省力」**。在 AI
全程写代码的场景下，选型的真正目的变了，是三个别的问题：

**问题 1：哪个能在编译期拦住 AI 的手误？**

AI 会拼错字段名、会用错类型，频率不比人低（甚至更高，因为它是概率生成）。区别在
错误何时暴露：

```go
// GORM + gen：字段名拼错 → 编译不过，当场红
r.q.User.Staus.Eq(1)        // ❌ compile error: q.User has no field Staus
r.q.User.Status.Eq("1")     // ❌ compile error: Status is int, not string

// 原生 SQL：拼错藏在字符串里 → 编译通过，跑到那行才炸
db.QueryRowContext(ctx, "SELECT staus FROM users WHERE id=$1", id)  // ✅ 编译通过
                                                                     // 💥 运行时 pq: column "staus" does not exist
```

编译期 vs 运行时，是「push 前就被 IDE/CI 拦下」和「上线后某个冷门分支才触发」的
区别。AI 生成的代码里，字符串 SQL 的错误没有任何静态检查网兜着。

**问题 2：改一列时，哪个不用 AI 手动同步散落各处的列顺序？**

这是原生 SQL 最隐蔽的坑。`Scan()` 靠**位置**匹配，列顺序和 struct 字段顺序必须
逐一对齐。加一列、删一列、调顺序，AI 得找出所有相关的 `SELECT` 和 `Scan()` 一起改，
漏一个就是运行时错位——而且可能不报错，只是把 email 塞进了 phone 字段（静默数据错乱）。

GORM 这边 `make gen-db` 从活库反射重新生成 model + query，一条命令收敛，AI 不需要
记住「哪些地方引用了这张表」。

**问题 3：哪个的错误更容易在 code review 里被人看出来？**

AI 写的代码终究要人 review。`r.q.User.Status.Eq(1)` 的错误 IDE 直接标红；一段
20 行的 `SELECT col1, col2, ... FROM ... Scan(&a, &b, ...)`，人眼要逐列核对顺序，
review 者很难发现第 7 个和第 8 个 Scan 参数被 AI 写反了。

> 核心洞察：**AI 写代码不是"随便选哪个都行"，恰恰相反——因为 AI 会犯概率性手误，
> 更需要一层编译期类型系统当安全网。** 原生 SQL 把这层网撤了，等于让 AI 的每个
> 手误都有机会溜到运行时。GORM+gen 的类型安全在 AI 时代不是"锦上添花"，是"刚需"。

## 二、逐场景对比

用本项目的真实写法（对照 `.claude/rules/repo-transaction-convention.md`）对比五个高频场景。

### 场景 1：model 从数据库映射

**GORM + gen 写法（本项目现状）：**

```bash
# schema 改了，一条命令重新生成
make gen-db
```

```go
// internal/dal/model/user.gen.go（自动生成，不手写）
type User struct {
    UserID       string `gorm:"column:user_id;primaryKey" json:"user_id"`
    Username     string `gorm:"column:username" json:"username"`
    Email        string `gorm:"column:email" json:"email"`
    Status       int    `gorm:"column:status" json:"status"`
    DepartmentID string `gorm:"column:department_id" json:"department_id"`
    TenantID     string `gorm:"column:tenant_id" json:"tenant_id"`
    CreatedAt    int64  `gorm:"column:created_at" json:"created_at"`
    UpdatedAt    int64  `gorm:"column:updated_at" json:"updated_at"`
}
```

- 从活库 schema 自动推导类型（`varchar → string`，`bigint → int64`）
- 字段 tag 自动补全
- schema 加列 → 重跑 `gen-db` → struct 自动多一个字段

**原生 SQL 写法：**

```go
// internal/dal/model/user.go（手写并手工维护）
type User struct {
    UserID       string `json:"user_id"`
    Username     string `json:"username"`
    Email        string `json:"email"`
    Status       int    `json:"status"`
    DepartmentID string `json:"department_id"`
    TenantID     string `json:"tenant_id"`
    CreatedAt    int64  `json:"created_at"`
    UpdatedAt    int64  `json:"updated_at"`
}
```

- AI 要自己写字段定义、猜类型（`bigint` 该用 `int` 还是 `int64`？`numeric` 该用 `float64` 还是 `string`？）
- schema 加列 → AI 要记得来这个文件手动补字段，漏了就是运行时 `Scan` 数量不匹配

| 维度 | GORM + gen | 原生 SQL |
|------|-----------|---------|
| model 来源 | ✅ 自动生成，schema 是唯一真相源 | ❌ 手写，AI 手动同步 |
| 类型推导 | ✅ 从 DB 类型自动映射 | ❌ AI 猜（容易猜错，如 `numeric`） |
| schema 变更传播 | ✅ 一条命令 | ❌ AI 人肉找所有引用 |

### 场景 2：单条查询 GetByID

**GORM + gen 写法（本项目 repo 实际写法）：**

```go
func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
    return r.q.User.WithContext(ctx).Where(r.q.User.UserID.Eq(id)).First()
}
```

- 字段引用 `r.q.User.UserID` 类型安全，拼错编译不过
- 返回 `*model.User`，AI 不需要手写 `Scan()` 每个字段

**原生 SQL 写法：**

```go
func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
    var user model.User
    err := r.db.QueryRowContext(ctx, `
        SELECT user_id, username, email, status, department_id, 
               tenant_id, created_at, updated_at
        FROM users 
        WHERE user_id = $1
    `, id).Scan(
        &user.UserID, &user.Username, &user.Email, &user.Status,
        &user.DepartmentID, &user.TenantID, &user.CreatedAt, &user.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, gorm.ErrRecordNotFound  // 还得手动翻译错误
    }
    return &user, err
}
```

- `SELECT` 列顺序、`Scan()` 参数顺序、struct 字段顺序，三者必须完全一致
- AI 漏一列、改顺序、拼错字段名，编译都能过，运行时才炸
- 加一列 `phone` → 三处都要 AI 手动补

| 维度 | GORM + gen | 原生 SQL |
|------|-----------|---------|
| 字段引用类型安全 | ✅ `r.q.User.UserID`（编译期检查） | ❌ `"user_id"` 字符串（运行时才知道错） |
| Scan 顺序易错 | ✅ 不写 Scan，gen 处理 | ❌ 三处顺序对齐，AI 易出错 |
| 列变更影响 | ✅ `gen-db` 自动 | ❌ 每个查询手动改 |

### 场景 3：动态多条件 WHERE + 分页（核心痛点）

这是 [[02-数据库访问层选型调研]] 直接淘汰 sqlc 的地方——可选过滤 + 分页是高频场景。

**GORM + gen 写法（本项目现状）：**

```go
func (r *UserRepo) List(ctx context.Context, filter *UserFilter, page, pageSize int) ([]*model.User, int64, error) {
    query := r.q.User.WithContext(ctx).Where(r.q.User.TenantID.Eq(filter.TenantID))
    
    if filter.DepartmentID != "" {
        query = query.Where(r.q.User.DepartmentID.Eq(filter.DepartmentID))
    }
    if filter.Status != nil {
        query = query.Where(r.q.User.Status.Eq(*filter.Status))
    }
    if filter.Keyword != "" {
        query = query.Where(r.q.User.Username.Like("%" + filter.Keyword + "%"))
    }
    
    count, err := query.Count()
    if err != nil {
        return nil, 0, err
    }
    
    users, err := query.Order(r.q.User.CreatedAt.Desc()).
        Order(r.q.User.UserID.Desc()).
        Offset((page - 1) * pageSize).
        Limit(pageSize).
        Find()
    return users, count, err
}
```

- 链式条件，每个 `if` 独立
- 字段引用 `r.q.User.DepartmentID` 编译期检查
- 可读、AI 易生成、不易错

**原生 SQL 写法：**

```go
func (r *UserRepo) List(ctx context.Context, filter *UserFilter, page, pageSize int) ([]*model.User, int64, error) {
    var conditions []string
    var args []interface{}
    argPos := 1
    
    conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argPos))
    args = append(args, filter.TenantID)
    argPos++
    
    if filter.DepartmentID != "" {
        conditions = append(conditions, fmt.Sprintf("department_id = $%d", argPos))
        args = append(args, filter.DepartmentID)
        argPos++
    }
    if filter.Status != nil {
        conditions = append(conditions, fmt.Sprintf("status = $%d", argPos))
        args = append(args, *filter.Status)
        argPos++
    }
    if filter.Keyword != "" {
        conditions = append(conditions, fmt.Sprintf("username LIKE $%d", argPos))
        args = append(args, "%"+filter.Keyword+"%")
        argPos++
    }
    
    whereClause := "WHERE " + strings.Join(conditions, " AND ")
    
    // Count 查询
    var count int64
    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
    if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&count); err != nil {
        return nil, 0, err
    }
    
    // List 查询
    listQuery := fmt.Sprintf(`
        SELECT user_id, username, email, status, department_id,
               tenant_id, created_at, updated_at
        FROM users
        %s
        ORDER BY created_at DESC, user_id DESC
        LIMIT $%d OFFSET $%d
    `, whereClause, argPos, argPos+1)
    args = append(args, pageSize, (page-1)*pageSize)
    
    rows, err := r.db.QueryContext(ctx, listQuery, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    var users []*model.User
    for rows.Next() {
        var user model.User
        if err := rows.Scan(
            &user.UserID, &user.Username, &user.Email, &user.Status,
            &user.DepartmentID, &user.TenantID, &user.CreatedAt, &user.UpdatedAt,
        ); err != nil {
            return nil, 0, err
        }
        users = append(users, &user)
    }
    return users, count, rows.Err()
}
```

- 手动管理 `$1 $2 $3` 占位符计数（`argPos++`），AI 容易算错
- 两次查询（count + list）要共享 `conditions/args`，AI 容易漏同步
- `Scan()` 八个字段顺序对齐，AI 错一个就数据错位

| 维度 | GORM + gen | 原生 SQL |
|------|-----------|---------|
| 动态条件复杂度 | ✅ 链式 if，简单直接 | ❌ 手管占位符、拼 SQL、args 数组 |
| AI 易错点 | 几乎没有 | 占位符序号、Scan 顺序、条件同步 |
| 代码行数（本例） | 18 行 | 60+ 行 |

### 场景 4：schema 加一列的改动扩散

假设 users 表加一列 `phone VARCHAR(20)`。

**GORM + gen 工作流：**

1. 写 migration：`ALTER TABLE users ADD COLUMN phone VARCHAR(20)`
2. 跑 migration
3. 跑 `make gen-db` → `internal/dal/model/user.gen.go` 自动多一行 `Phone string`
4. Done — 所有 repo 方法自动能用 `r.q.User.Phone`，不用改任何查询

**原生 SQL 工作流：**

1. 写 migration：`ALTER TABLE users ADD COLUMN phone VARCHAR(20)`
2. 跑 migration
3. 手动在 `model.User` struct 加 `Phone string` 字段
4. **找出所有引用 users 表的查询**（可能散落 10+ 个 repo 方法）
5. 每个 `SELECT user_id, username, ...` 都要加 `phone`
6. 每个对应的 `Scan(&user.UserID, &user.Username, ...)` 都要加 `&user.Phone`（位置必须和 SELECT 列顺序一致）
7. 漏一个 → 编译通过，运行时 `Scan` 参数数量不匹配报错；或者更糟——数量对了但顺序错了，数据静默错位（email 写进 phone）

| 维度 | GORM + gen | 原生 SQL |
|------|-----------|---------|
| 改动点数 | 2 处（migration + `make gen-db`） | N+2 处（migration + model + N 个查询的 SELECT/Scan） |
| AI 易漏 | ✅ 不会漏，自动生成 | ❌ 容易漏某个冷门查询 |
| 错误发现时机 | ✅ 编译期（用了新字段立刻红） | ❌ 运行时（该查询被执行时才炸） |

### 场景 5：事务

**GORM 写法（本项目规范）：**

```go
func (s *Service) Transfer(ctx context.Context, from, to string, amount int) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        userRepo := repository.NewUserRepo(tx)
        orderRepo := repository.NewOrderRepo(tx)
        
        if err := userRepo.Deduct(ctx, from, amount); err != nil {
            return err  // 自动回滚
        }
        if err := userRepo.Add(ctx, to, amount); err != nil {
            return err  // 自动回滚
        }
        return orderRepo.Create(ctx, &model.Order{...})
    })
}
```

- 闭包内任何 `return err` 自动回滚
- 不需要手写 `defer Rollback()` / `Commit()`

**原生 SQL 写法：**

```go
func (s *Service) Transfer(ctx context.Context, from, to string, amount int) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()  // 安全兜底（Commit 后 Rollback 无害）
    
    userRepo := repository.NewUserRepo(tx)
    orderRepo := repository.NewOrderRepo(tx)
    
    if err := userRepo.Deduct(ctx, from, amount); err != nil {
        return err  // 依赖 defer Rollback
    }
    if err := userRepo.Add(ctx, to, amount); err != nil {
        return err
    }
    if err := orderRepo.Create(ctx, &model.Order{...}); err != nil {
        return err
    }
    
    return tx.Commit()
}
```

- AI 容易忘记 `defer tx.Rollback()`（忘了就是事务未回滚）
- 最后必须显式 `Commit()`，AI 容易忘

| 维度 | GORM | 原生 SQL |
|------|------|---------|
| 自动回滚 | ✅ 闭包 return err 自动 | ❌ 需 defer Rollback（AI 易忘） |
| 代码安全性 | ✅ 不会漏 Commit | ❌ AI 易忘最后 Commit |

## 三、迁移成本估算：现在拆 GORM 要付什么

本项目当前有 12 个业务域（auth / department / dict / menu / role / tenant / user 等），
每域 repo 约 5-10 个方法。拆掉 GORM 换原生 SQL 意味着：

- **一次性重写**：12 域 × 5-10 方法 ≈ **60-120 个 repo 方法**，每个都要手写 SQL、对齐 Scan、处理 `sql.ErrNoRows`、写集成测试验证列顺序没错
- **model 层**：`internal/dal/model` 全部改手写，放弃 `make gen-db`
- **长期税**：此后每次 schema 变更，成本从「跑一条命令」变成「人肉找所有查询改 SELECT/Scan」（见场景 4）

**一句话**：省掉 GORM 这一个依赖，换来的是 60-120 个方法的重写成本 + 每次 schema 变更翻倍的长期维护税。这笔账不划算。

## 四、承认的 tradeoff（不回避）

原生 SQL 不是没有好处，诚实列出来：

- **完全透明**：写的 SQL 就是执行的 SQL，没有 ORM 生成的「猜不透的查询」。（但 GORM 的 SQL 也能在日志里看到，透明度差距没想象中大。）
- **零抽象层**：不经过 ORM 的反射与 builder，`database/sql` 直连驱动。
- **PG 原生特性直用**：jsonb / array / CTE / RETURNING 直接写进 SQL，不受 ORM 抽象限制。
- **无反射开销**：GORM 靠运行时反射做字段映射，原生 SQL 的 `Scan` 是直接赋值，理论上更快。

**但这些好处对本项目不够大到值得换：**

- 透明度：GORM SQL 日志已经够看。
- PG 原生特性：真需要时，GORM 的 `Raw().Scan()` 就能写裸 SQL，不必全盘放弃 ORM。
- 反射开销：本项目不是高频交易系统，ORM 反射远不是瓶颈——**先测量再优化**，不要为了没测过的性能假设付上面那笔迁移账。

**边界（混合模式才是正解）**：90% 的标准 CRUD 走 gen 的类型安全 API，10% 的复杂查询（递归 CTE、复杂报表）用 GORM 的 `Raw().Scan()` 落到裸 SQL：

```go
func (r *DepartmentRepo) TreeReport(ctx context.Context) ([]*ReportRow, error) {
    var results []*ReportRow
    err := r.q.WithContext(ctx).UnderlyingDB().Raw(`
        WITH RECURSIVE hierarchy AS (
            SELECT dept_id, parent_id, name, 1 as level
            FROM departments WHERE parent_id = ''
            UNION ALL
            SELECT d.dept_id, d.parent_id, d.name, h.level + 1
            FROM departments d JOIN hierarchy h ON d.parent_id = h.dept_id
        )
        SELECT * FROM hierarchy ORDER BY level, name
    `).Scan(&results).Error
    return results, err
}
```

要 SQL 的表达力时随时能落到裸 SQL，不必为这 10% 把 90% 的类型安全全丢掉。

## 五、结论

**保持 GORM + gen，不换原生 SQL。**

「AI 全程写代码」不是换原生 SQL 的理由，反而是**留住 GORM 的最强理由**——AI 写代码时手误概率更高（字段名、类型、列顺序），而 GORM 的类型安全把这些错误挡在编译期，原生 SQL 把它们推到运行时。省一个依赖换来的是编译期安全网的消失，不划算。

**什么场景才该反过来选原生 SQL：**

- 项目极小（< 5 张表、纯 CRUD、无动态查询）——ORM 的收益覆盖不了它的重量
- schema 100% 冻结、永不变更——场景 4 的维护税消失
- GORM 被**实测**为性能瓶颈——注意是「实测」，不是「觉得反射慢」；先 profile 再决定

本项目 20+ 张表、SaaS 后台活跃迭代、schema 频繁变更、AI 主力生产代码——**四条全部指向保留 GORM**。真需要 SQL 表达力的地方，用 `Raw().Scan()` 混合模式即可，见第四节。
