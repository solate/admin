# Step 04: 数据库模型 — GORM Gen + 迁移

## 目标

设计完整的数据库 schema，用 GORM Gen 生成模型和查询，创建迁移脚本。

## 前置条件

- Step 02 完成，数据库连接正常

## 文件清单

```
internal/dal/model/             # GORM Gen 生成（勿手动编辑）
internal/dal/query/             # GORM Gen 生成（勿手动编辑）
internal/repository/            # 手写仓储层（后续步骤填充）
scripts/migrations/
└── 000001_init_schema.up.sql   # 完整建表 SQL
└── 000001_init_schema.down.sql # 回滚 SQL
scripts/
└── gen.go                     # GORM Gen 生成入口
```

## 实现细节

### 1. 完整数据库 Schema

```sql
-- 全局表（无 tenant_id）
-- permissions：权限定义（API/MENU/BUTTON/DATA）
-- menus：菜单树

-- 租户表（有 tenant_id）
-- tenants：租户
-- users：用户（绑定 tenant_id）
-- roles：角色（绑定 tenant_id，支持 parent_role_id 继承）
-- departments：部门树（绑定 tenant_id）
-- positions：岗位（绑定 tenant_id）

-- 关联表（有 tenant_id）
-- user_roles：用户-角色（user_id + role_id + tenant_id）
-- role_permissions：角色-权限（role_id + permission_id + tenant_id）

-- 审计表（有 tenant_id）
-- login_logs：登录日志
-- operation_logs：操作日志
```

### 2. 核心表设计

#### tenants

```sql
CREATE TABLE tenants (
    tenant_id   VARCHAR(20)  PRIMARY KEY,
    tenant_code VARCHAR(50)  NOT NULL UNIQUE,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500) DEFAULT '',
    contact_name  VARCHAR(50) DEFAULT '',
    contact_phone VARCHAR(20) DEFAULT '',
    status      SMALLINT     NOT NULL DEFAULT 1,  -- 1=启用 2=禁用
    created_at  BIGINT       NOT NULL DEFAULT 0,
    updated_at  BIGINT       NOT NULL DEFAULT 0,
    deleted_at  BIGINT       NOT NULL DEFAULT 0
);
```

#### users

```sql
CREATE TABLE users (
    user_id       VARCHAR(20)  PRIMARY KEY,
    tenant_id     VARCHAR(20)  NOT NULL,
    email         VARCHAR(100) NOT NULL,
    phone         VARCHAR(20)  DEFAULT '',
    password      VARCHAR(200) NOT NULL,
    user_name     VARCHAR(50)  NOT NULL,
    avatar        VARCHAR(500) DEFAULT '',
    department_id VARCHAR(20)  DEFAULT '',
    position_id   VARCHAR(20)  DEFAULT '',
    status        SMALLINT     NOT NULL DEFAULT 1,
    must_change_password SMALLINT NOT NULL DEFAULT 2,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    updated_at    BIGINT       NOT NULL DEFAULT 0,
    deleted_at    BIGINT       NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE deleted_at = 0;
CREATE INDEX idx_users_tenant ON users(tenant_id);
```

#### roles

```sql
CREATE TABLE roles (
    role_id       VARCHAR(20)  PRIMARY KEY,
    tenant_id     VARCHAR(20)  NOT NULL,
    role_code     VARCHAR(50)  NOT NULL,
    name          VARCHAR(100) NOT NULL,
    description   VARCHAR(500) DEFAULT '',
    parent_role_id VARCHAR(20) DEFAULT '',
    data_scope    SMALLINT     NOT NULL DEFAULT 5,  -- 1-5 五级数据权限
    data_scope_dept_ids JSONB   DEFAULT '[]'::jsonb, -- 自定义部门列表
    sort          INT          NOT NULL DEFAULT 0,
    status        SMALLINT     NOT NULL DEFAULT 1,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    updated_at    BIGINT       NOT NULL DEFAULT 0,
    deleted_at    BIGINT       NOT NULL DEFAULT 0,
    UNIQUE(tenant_id, role_code, deleted_at)
);
CREATE INDEX idx_roles_tenant ON roles(tenant_id);
```

> **注意**：`data_scope` 和 `data_scope_dept_ids` 在 Step 12 才使用，但建表时就加上，避免后续迁移。

#### permissions

```sql
CREATE TABLE permissions (
    permission_id VARCHAR(20)  PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    type          VARCHAR(20)  NOT NULL,  -- API/MENU/BUTTON/DATA
    resource      VARCHAR(200) NOT NULL,  -- 路径或标识
    action        VARCHAR(20)  DEFAULT '', -- HTTP 方法或操作类型
    parent_id     VARCHAR(20)  DEFAULT '',
    sort          INT          NOT NULL DEFAULT 0,
    status        SMALLINT     NOT NULL DEFAULT 1,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    updated_at    BIGINT       NOT NULL DEFAULT 0,
    deleted_at    BIGINT       NOT NULL DEFAULT 0
);
-- 全局表，无 tenant_id
```

#### 关联表

```sql
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

CREATE TABLE role_permissions (
    id            BIGSERIAL    PRIMARY KEY,
    role_id       VARCHAR(20)  NOT NULL,
    permission_id VARCHAR(20)  NOT NULL,
    tenant_id     VARCHAR(20)  NOT NULL,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    UNIQUE(role_id, permission_id, tenant_id)
);
CREATE INDEX idx_role_permissions_role_tenant ON role_permissions(role_id, tenant_id);
```

### 3. GORM Gen 使用

```go
// scripts/gen.go
// 参考现有 backend 的 scripts/gen.go
// 生成路径：internal/dal/model/ 和 internal/dal/query/
// 生成命令：make gen-db
```

**Makefile 追加**：

```makefile
gen-db:
	go run ./scripts/gen.go

migrate-up:
	go run ./scripts/migrate.go up

migrate-down:
	go run ./scripts/migrate.go down
```

### 4. 设计决策

| 决策 | 选择 | 原因 |
|------|------|------|
| ID 类型 | VARCHAR(20) UUID | 全局唯一，不暴露数据量 |
| 时间戳 | BIGINT 毫秒 | 跨时区一致，便于前端展示 |
| 软删除 | deleted_at BIGINT | 0=未删除，>0=删除时间戳 |
| permissions 全局 | 无 tenant_id | API 路径和菜单是产品定义的 |
| roles 租户级 | 有 tenant_id | 每个租户角色体系独立 |
| data_scope 预置 | 建表就加 | 避免后续 ALTER TABLE |

## 验收标准

- [ ] `make migrate-up` 成功创建所有表
- [ ] `make migrate-down` 成功回滚
- [ ] `make gen-db` 成功生成 model 和 query 文件
- [ ] 表结构、索引、唯一约束与设计一致
- [ ] `go build ./...` 编译通过（无语法错误）

## AI 协作提示

```
请按 step-04-database-model.md 创建完整的数据库 schema。
先写 SQL 迁移文件，然后配置 GORM Gen 生成入口。
参考现有 backend/scripts/dev_schema.sql 的表设计，但做以下改进：
1. roles 表预置 data_scope 和 data_scope_dept_ids 字段
2. 所有 ID 用 VARCHAR(20) + UUID
3. 所有时间戳用 BIGINT 毫秒
4. 确保每个关联表都有 created_at
```
