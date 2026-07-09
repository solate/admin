# Step 04: 数据模型 — 完整 Schema / GORM Gen / Repository 模板

## 目标

设计完整的多租户数据库 schema，配置 GORM Gen 代码生成，建立 Repository 层模板。

## 前置条件

- Step 02 完成，数据库连接正常
- PostgreSQL 数据库 `app_dev` 已创建

## 文件清单

```
migrations/
├── 000001_init.up.sql         # 完整建表 + 索引
└── 000001_init.down.sql       # 回滚

scripts/
├── gen/main.go                # GORM Gen 生成入口
└── migrate/main.go            # 迁移执行工具

internal/
├── model/                     # Gen 生成（勿手动编辑）
├── query/                     # Gen 生成（勿手动编辑）
└── repository/
    └── user_repo.go           # Repository 模板示例

Makefile                       # 追加 gen / migrate 目标
```

## 实现规范

### 1. 数据库设计原则

| 原则 | 说明 |
|------|------|
| ID 类型 | VARCHAR(20)，雪花算法生成 |
| 时间戳 | BIGINT 毫秒级，0 表示未设置 |
| 软删除 | `deleted_at BIGINT DEFAULT 0`，0=未删除 |
| 租户隔离 | 租户表必有 `tenant_id VARCHAR(20) NOT NULL` |
| 全局表 | permissions / menus 无 tenant_id |
| 状态字段 | SMALLINT：1=启用 2=禁用 |
| 排序字段 | INT：sort，默认 0 |

### 2. 完整建表 SQL

#### 全局表（无 tenant_id）

```sql
-- ============================================================
-- 租户表（特殊：自身无 tenant_id）
-- ============================================================
CREATE TABLE tenants (
    tenant_id     VARCHAR(20)  PRIMARY KEY,
    tenant_code   VARCHAR(50)  NOT NULL UNIQUE,
    name          VARCHAR(100) NOT NULL,
    description   VARCHAR(500) DEFAULT '',
    contact_name  VARCHAR(50)  DEFAULT '',
    contact_phone VARCHAR(20)  DEFAULT '',
    status        SMALLINT     NOT NULL DEFAULT 1,
    expired_at    BIGINT       NOT NULL DEFAULT 0,  -- 过期时间，0=永不过期
    max_users     INT          NOT NULL DEFAULT 100, -- 最大用户数
    created_at    BIGINT       NOT NULL DEFAULT 0,
    updated_at    BIGINT       NOT NULL DEFAULT 0,
    deleted_at    BIGINT       NOT NULL DEFAULT 0
);

-- ============================================================
-- 权限表（全局，产品定义）
-- ============================================================
CREATE TABLE permissions (
    permission_id VARCHAR(20)  PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    type          VARCHAR(20)  NOT NULL,   -- API / MENU / BUTTON
    resource      VARCHAR(200) NOT NULL,   -- API 路径或前端标识
    action        VARCHAR(20)  DEFAULT '', -- GET/POST/PUT/DELETE
    parent_id     VARCHAR(20)  DEFAULT '',
    description   VARCHAR(500) DEFAULT '',
    sort          INT          NOT NULL DEFAULT 0,
    status        SMALLINT     NOT NULL DEFAULT 1,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    updated_at    BIGINT       NOT NULL DEFAULT 0
);
CREATE INDEX idx_permissions_type ON permissions(type);
CREATE INDEX idx_permissions_parent ON permissions(parent_id);

-- ============================================================
-- 菜单表（全局，产品定义）
-- ============================================================
CREATE TABLE menus (
    menu_id      VARCHAR(20)  PRIMARY KEY,
    parent_id    VARCHAR(20)  DEFAULT '',
    name         VARCHAR(100) NOT NULL,
    path         VARCHAR(200) DEFAULT '',     -- 前端路由路径
    component    VARCHAR(200) DEFAULT '',     -- 前端组件路径
    icon         VARCHAR(100) DEFAULT '',
    sort         INT          NOT NULL DEFAULT 0,
    type         SMALLINT     NOT NULL DEFAULT 1, -- 1=目录 2=菜单 3=按钮
    visible      SMALLINT     NOT NULL DEFAULT 1, -- 1=显示 2=隐藏
    status       SMALLINT     NOT NULL DEFAULT 1,
    permission   VARCHAR(100) DEFAULT '',     -- 权限标识（关联 permissions.resource）
    created_at   BIGINT       NOT NULL DEFAULT 0,
    updated_at   BIGINT       NOT NULL DEFAULT 0
);
CREATE INDEX idx_menus_parent ON menus(parent_id);
```

#### 租户表（有 tenant_id）

```sql
-- ============================================================
-- 用户表
-- ============================================================
CREATE TABLE users (
    user_id       VARCHAR(20)  PRIMARY KEY,
    tenant_id     VARCHAR(20)  NOT NULL,
    email         VARCHAR(100) NOT NULL,
    phone         VARCHAR(20)  DEFAULT '',
    password      VARCHAR(200) NOT NULL,
    user_name     VARCHAR(50)  NOT NULL,
    real_name     VARCHAR(50)  DEFAULT '',
    avatar        VARCHAR(500) DEFAULT '',
    department_id VARCHAR(20)  DEFAULT '',
    position_id   VARCHAR(20)  DEFAULT '',
    status        SMALLINT     NOT NULL DEFAULT 1,
    must_change_password SMALLINT NOT NULL DEFAULT 2,  -- 1=需要 2=不需要
    last_login_at BIGINT       NOT NULL DEFAULT 0,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    updated_at    BIGINT       NOT NULL DEFAULT 0,
    deleted_at    BIGINT       NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX idx_users_email_alive ON users(email) WHERE deleted_at = 0;
CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_department ON users(department_id);

-- ============================================================
-- 角色表（支持继承）
-- ============================================================
CREATE TABLE roles (
    role_id         VARCHAR(20)  PRIMARY KEY,
    tenant_id       VARCHAR(20)  NOT NULL,
    role_code       VARCHAR(50)  NOT NULL,
    name            VARCHAR(100) NOT NULL,
    description     VARCHAR(500) DEFAULT '',
    parent_role_id  VARCHAR(20)  DEFAULT '',       -- 父角色（继承）
    data_scope      SMALLINT     NOT NULL DEFAULT 5, -- 1=全部 2=自定义 3=本部门及以下 4=本部门 5=仅本人
    data_scope_dept_ids TEXT     DEFAULT '',       -- 自定义部门ID列表，逗号分隔
    sort            INT          NOT NULL DEFAULT 0,
    status          SMALLINT     NOT NULL DEFAULT 1,
    created_at      BIGINT       NOT NULL DEFAULT 0,
    updated_at      BIGINT       NOT NULL DEFAULT 0,
    deleted_at      BIGINT       NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX idx_roles_code_tenant_alive ON roles(tenant_id, role_code) WHERE deleted_at = 0;
CREATE INDEX idx_roles_tenant ON roles(tenant_id);
CREATE INDEX idx_roles_parent ON roles(parent_role_id);

-- ============================================================
-- 部门表（树形）
-- ============================================================
CREATE TABLE departments (
    department_id VARCHAR(20)  PRIMARY KEY,
    tenant_id     VARCHAR(20)  NOT NULL,
    parent_id     VARCHAR(20)  DEFAULT '',
    name          VARCHAR(100) NOT NULL,
    leader_id     VARCHAR(20)  DEFAULT '',  -- 部门负责人
    sort          INT          NOT NULL DEFAULT 0,
    status        SMALLINT     NOT NULL DEFAULT 1,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    updated_at    BIGINT       NOT NULL DEFAULT 0,
    deleted_at    BIGINT       NOT NULL DEFAULT 0
);
CREATE INDEX idx_departments_tenant ON departments(tenant_id);
CREATE INDEX idx_departments_parent ON departments(parent_id);

-- ============================================================
-- 岗位表
-- ============================================================
CREATE TABLE positions (
    position_id  VARCHAR(20)  PRIMARY KEY,
    tenant_id    VARCHAR(20)  NOT NULL,
    name         VARCHAR(100) NOT NULL,
    code         VARCHAR(50)  NOT NULL,
    sort         INT          NOT NULL DEFAULT 0,
    status       SMALLINT     NOT NULL DEFAULT 1,
    created_at   BIGINT       NOT NULL DEFAULT 0,
    updated_at   BIGINT       NOT NULL DEFAULT 0,
    deleted_at   BIGINT       NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX idx_positions_code_tenant_alive ON positions(tenant_id, code) WHERE deleted_at = 0;
CREATE INDEX idx_positions_tenant ON positions(tenant_id);
```

#### 关联表

```sql
-- ============================================================
-- 用户-角色关联
-- ============================================================
CREATE TABLE user_roles (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    VARCHAR(20)  NOT NULL,
    role_id    VARCHAR(20)  NOT NULL,
    tenant_id  VARCHAR(20)  NOT NULL,
    created_at BIGINT       NOT NULL DEFAULT 0,
    UNIQUE(user_id, role_id, tenant_id)
);
CREATE INDEX idx_user_roles_user_tenant ON user_roles(user_id, tenant_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);

-- ============================================================
-- 角色-权限关联
-- ============================================================
CREATE TABLE role_permissions (
    id            BIGSERIAL    PRIMARY KEY,
    role_id       VARCHAR(20)  NOT NULL,
    permission_id VARCHAR(20)  NOT NULL,
    tenant_id     VARCHAR(20)  NOT NULL,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    UNIQUE(role_id, permission_id, tenant_id)
);
CREATE INDEX idx_role_permissions_role_tenant ON role_permissions(role_id, tenant_id);

-- ============================================================
-- 角色-菜单关联
-- ============================================================
CREATE TABLE role_menus (
    id         BIGSERIAL    PRIMARY KEY,
    role_id    VARCHAR(20)  NOT NULL,
    menu_id    VARCHAR(20)  NOT NULL,
    tenant_id  VARCHAR(20)  NOT NULL,
    created_at BIGINT       NOT NULL DEFAULT 0,
    UNIQUE(role_id, menu_id, tenant_id)
);
CREATE INDEX idx_role_menus_role_tenant ON role_menus(role_id, tenant_id);
```

#### 审计表

```sql
-- ============================================================
-- 登录日志
-- ============================================================
CREATE TABLE login_logs (
    log_id     VARCHAR(20)  PRIMARY KEY,
    tenant_id  VARCHAR(20)  NOT NULL,
    user_id    VARCHAR(20)  DEFAULT '',
    user_name  VARCHAR(50)  DEFAULT '',
    ip         VARCHAR(50)  DEFAULT '',
    user_agent VARCHAR(500) DEFAULT '',
    status     SMALLINT     NOT NULL DEFAULT 1, -- 1=成功 2=失败
    message    VARCHAR(200) DEFAULT '',
    login_at   BIGINT       NOT NULL DEFAULT 0
);
CREATE INDEX idx_login_logs_tenant_time ON login_logs(tenant_id, login_at DESC);
CREATE INDEX idx_login_logs_user ON login_logs(user_id);

-- ============================================================
-- 操作日志
-- ============================================================
CREATE TABLE operation_logs (
    log_id       VARCHAR(20)   PRIMARY KEY,
    tenant_id    VARCHAR(20)   NOT NULL,
    user_id      VARCHAR(20)   NOT NULL,
    user_name    VARCHAR(50)   DEFAULT '',
    module       VARCHAR(50)   NOT NULL,   -- 模块名
    action       VARCHAR(50)   NOT NULL,   -- 操作类型
    method       VARCHAR(10)   DEFAULT '', -- HTTP 方法
    path         VARCHAR(200)  DEFAULT '', -- 请求路径
    ip           VARCHAR(50)   DEFAULT '',
    user_agent   VARCHAR(500)  DEFAULT '',
    request_body TEXT          DEFAULT '',
    status       SMALLINT      NOT NULL DEFAULT 1,
    duration     INT           NOT NULL DEFAULT 0, -- 耗时(ms)
    created_at   BIGINT        NOT NULL DEFAULT 0
);
CREATE INDEX idx_operation_logs_tenant_time ON operation_logs(tenant_id, created_at DESC);
CREATE INDEX idx_operation_logs_user ON operation_logs(user_id);
```

### 3. GORM Gen 配置

```go
// scripts/gen/main.go
package main

import (
    "gorm.io/driver/postgres"
    "gorm.io/gen"
    "gorm.io/gorm"
)

func main() {
    dsn := "host=localhost port=5432 user=postgres password=postgres dbname=app_dev sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    g := gen.NewGenerator(gen.Config{
        OutPath:      "./internal/query",
        ModelPkgPath: "./internal/model",
        Mode:         gen.WithDefaultQuery | gen.WithQueryInterface,
    })

    g.UseDB(db)

    // 生成所有表
    g.ApplyBasic(
        g.GenerateModel("tenants"),
        g.GenerateModel("users"),
        g.GenerateModel("roles"),
        g.GenerateModel("departments"),
        g.GenerateModel("positions"),
        g.GenerateModel("permissions"),
        g.GenerateModel("menus"),
        g.GenerateModel("user_roles"),
        g.GenerateModel("role_permissions"),
        g.GenerateModel("role_menus"),
        g.GenerateModel("login_logs"),
        g.GenerateModel("operation_logs"),
    )

    g.Execute()
}
```

### 4. Repository 模板

```go
// internal/repository/user_repo.go
package repository

import (
    "context"
    "admin/internal/model"
    "admin/internal/query"
    "admin/pkg/xcontext"
    "gorm.io/gorm"
)

type UserRepo struct {
    db *gorm.DB
    q  *query.Query
}

func NewUserRepo(db *gorm.DB) *UserRepo {
    return &UserRepo{
        db: db,
        q:  query.Use(db),
    }
}

// GetByID 根据 ID 获取用户（自动带租户隔离）
func (r *UserRepo) GetByID(ctx context.Context, userID string) (*model.User, error) {
    u := r.q.User
    return u.WithContext(ctx).
        Where(u.UserID.Eq(userID)).
        Where(u.TenantID.Eq(xcontext.GetTenantID(ctx))).
        Where(u.DeletedAt.Eq(0)).
        First()
}

// List 分页查询用户列表（双排序字段）
func (r *UserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
    u := r.q.User
    q := u.WithContext(ctx).
        Where(u.TenantID.Eq(xcontext.GetTenantID(ctx))).
        Where(u.DeletedAt.Eq(0))

    total, err := q.Count()
    if err != nil {
        return nil, 0, err
    }

    users, err := q.
        Order(u.CreatedAt.Desc()).
        Order(u.UserID.Desc()). // 双排序字段
        Offset(offset).
        Limit(limit).
        Find()

    return users, total, err
}
```

**Repository 规范**：
- 构造函数接收 `*gorm.DB`，内部 `query.Use(db)` 创建类型安全查询
- 所有租户表查询带 `WHERE tenant_id = xcontext.GetTenantID(ctx)`
- 所有列表查询使用双排序字段：`ORDER BY created_at DESC, pk DESC`
- 软删除条件：`WHERE deleted_at = 0`
- 不在 Repository 内管理事务（由 Service 层控制）

### 5. 迁移工具

```go
// scripts/migrate/main.go
// 简单的迁移执行器：
// - up：执行 migrations/*.up.sql
// - down：执行 migrations/*.down.sql
// 使用 database/sql 直接执行 SQL
// 生产环境可替换为 golang-migrate/migrate
```

### 6. Makefile 追加

```makefile
# 数据库迁移
migrate-up:
	go run ./scripts/migrate/main.go up

migrate-down:
	go run ./scripts/migrate/main.go down

# GORM Gen 代码生成
gen:
	go run ./scripts/gen/main.go
```

## 验收标准

```bash
# 1. 迁移执行成功
make migrate-up
# 期望：所有表创建成功

# 2. 表结构验证（PostgreSQL）
psql app_dev -c "\dt"
# 期望：列出所有 12 张表

# 3. 索引验证
psql app_dev -c "\di"
# 期望：包含所有预定义索引

# 4. Gen 代码生成
make gen
# 期望：internal/model/ 和 internal/query/ 生成文件

# 5. 编译通过
go build ./...
# 期望：无错误

# 6. 回滚测试
make migrate-down
make migrate-up
# 期望：先删后建，无残留
```

## AI 协作提示

```
请按 step-04-database-schema.md 实现数据模型层。

要点：
1. migrations/000001_init.up.sql — 完整建表 SQL（按文档中的表设计）
2. migrations/000001_init.down.sql — DROP TABLE 回滚
3. scripts/gen/main.go — GORM Gen 配置（生成到 internal/model 和 internal/query）
4. scripts/migrate/main.go — 简单迁移工具
5. internal/repository/user_repo.go — Repository 模板（演示租户隔离和双排序）
6. Makefile 追加 gen / migrate-up / migrate-down
7. 所有 ID 用 VARCHAR(20)，时间用 BIGINT，软删除用 deleted_at
8. 关联表：独立主键 BIGSERIAL + UNIQUE 约束 + 双向索引
```

---

*上一步：[Step 03 - HTTP 框架](step-03-http-framework.md) | 下一步：[Step 05 - 多租户与 JWT](step-05-multi-tenant-jwt.md)*
