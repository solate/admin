# 系统架构全面重新设计文档

> 基于架构问题分析（`architecture-analysis.md`）和技术选型对比（`go-database-comparison.md`），本项目将进行全面重新设计。项目未上线，无兼容需求。

---

## 目录

1. [系统技术架构](#1-系统技术架构)
2. [技术选型](#2-技术选型)
3. [目录结构与包分层](#3-目录结构与包分层)
4. [pkg 目录重构](#4-pkg-目录重构)
5. [数据库层设计](#5-数据库层设计)
6. [Repository 层设计](#6-repository-层设计)
7. [RBAC 设计（替代 Casbin）](#7-rbac-设计替代-casbin)
8. [中间件链设计](#8-中间件链设计)
9. [错误处理](#9-错误处理)
10. [配置管理](#10-配置管理)
11. [请求验证](#11-请求验证)
12. [其他考虑](#12-其他考虑)
13. [实施步骤](#13-实施步骤)
14. [完整代码示例](#14-完整代码示例)

---

## 1. 系统技术架构

### 1.1 分层架构

```
┌─────────────────────────────────────────────────────────┐
│                      HTTP Client                        │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                   Gin Router                             │
│  /api/v1/auth/*  (public)                               │
│  /api/v1/*       (authenticated: Auth → RBAC → Audit)   │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                   Middleware Chain                       │
│  RequestID → Logger → Recovery → CORS → RateLimit       │
│  → Auth(JWT) → RBAC(PermissionCache) → Audit            │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                   Handler 层                             │
│  绑定请求参数 → 调用 Service → 返回统一响应               │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                   Service 层                             │
│  业务逻辑编排、权限校验、审计日志记录                      │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                  Repository 层                           │
│  数据访问封装、显式 TenantScope、动态查询构建             │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│              GORM + Gen (PostgreSQL)                     │
│  自动时间戳(autoCreateTime/autoUpdateTime)               │
│  自动软删除(softDelete)                                  │
│  显式租户过滤(TenantScope helper)                        │
└─────────────────────────────────────────────────────────┘
```

### 1.2 请求处理流程

```
HTTP Request
  → RequestID 中间件（生成请求 ID）
  → Logger 中间件（记录请求日志）
  → Recovery 中间件（panic 恢复）
  → CORS 中间件（跨域处理）
  → RateLimit 中间件（限流）
  → Auth 中间件（JWT 验证 → 注入 context: user_id, tenant_id, role_ids）
  → RBAC 中间件（从 PermissionCache 检查 API 权限）
  → Audit 中间件（提取请求信息存入 context）
  → Handler（绑定参数 → 调用 Service）
  → Service（业务逻辑 → 调用 Repository）
  → Repository（显式 TenantScope(ctx, db) → GORM 查询）
  → Response（统一 JSON 格式返回）
```

### 1.3 与当前系统的核心区别

| 维度 | 当前系统 | 新系统 |
|------|---------|--------|
| 多租户 | GORM Callback 隐式注入 `WHERE tenant_id = ?` | **显式** `TenantScope(ctx, db)` helper |
| 权限 | Casbin Enforce（字符串策略，双写） | **纯数据库 RBAC**（PermissionCache） |
| 跳过租户 | `SkipTenantCheck(ctx)` + `TenantSkipMiddleware` | **不需要** — 跨租户查询直接不加 TenantScope |
| 角色-用户 | Casbin g 策略 `casbin_rule` 表 | `user_roles` 关联表 |
| 角色-权限 | Casbin p 策略 `casbin_rule` 表 | `role_permissions` 关联表 |
| 角色继承 | Casbin g2 策略 + 应用层 DFS | `roles.parent_role_id` + 递归查询 |
| 菜单查询 | 4 步：Casbin 查角色 → DFS → N 次策略查询 → DB | **1 条 SQL JOIN** |

---

## 2. 技术选型

### 2.1 为什么选择 GORM + Gen 而非 sqlc

| 对比维度 | GORM + Gen | sqlc |
|---------|-----------|------|
| 动态 WHERE | 链式 API 自由组合条件 | 每个筛选组合需写不同 SQL |
| 自动时间戳 | `autoCreateTime:milli` tag 零配置 | 需手动传参或依赖 DB Trigger |
| 自动软删除 | `softDelete:milli` tag 零配置 | 每条 SQL 需写 `WHERE deleted_at = 0` |
| AI 友好度 | 中等（需理解 Gen API） | 最高（AI 生成 SQL 最准确） |
| 代码量 | 少（Gen 生成查询构建器） | 多（每个查询写 SQL） |

**结论**：管理后台需要大量动态筛选（按昵称/状态/部门/岗位组合查询），GORM 的链式 API 更适合。sqlc 的静态 SQL 模式不适合这种场景。

### 2.2 完整技术栈

| 类别 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 语言 | Go | 1.24+ | |
| Web 框架 | Gin | v1.11 | 成熟稳定，性能优秀 |
| ORM | GORM | v1.31 | 自动时间戳、软删除、事务 |
| 代码生成 | GORM Gen | v0.3 | 类型安全查询构建器 |
| 数据库 | PostgreSQL | 16+ | |
| 缓存 | Redis | 7+ | JWT 黑名单、权限缓存 |
| 认证 | JWT (golang-jwt/jwt/v5) | v5.3 | Access + Refresh Token |
| RBAC | 纯数据库 | — | user_roles + role_permissions |
| 日志 | Zerolog | v1.34 | 结构化日志 |
| 配置 | Viper | v1.18 | 多环境配置 |
| 验证 | go-playground/validator | v10 | Gin 内建 |
| 密码 | Argon2id (x/crypto) | — | 安全哈希 |
| ID 生成 | Sonyflake | v1.3 | 18 位数字 ID |
| 定时任务 | robfig/cron | v3 | 秒级精度 |
| 迁移 | golang-migrate | — | SQL 文件版本化 |
| API 文档 | Swag | v1.16 | Swagger/OpenAPI |

---

## 3. 目录结构与包分层

### 3.1 完整目录结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go                  # 应用入口
├── internal/
│   ├── constants/                   # 系统常量
│   │   ├── status.go               # 状态码 (Enabled=1, Disabled=2)
│   │   ├── system.go               # 系统常量 (角色码、权限类型等)
│   │   └── operation_log.go        # 操作日志常量
│   ├── dal/                        # 数据访问层（Gen 生成）
│   │   ├── model/                  # Gen 生成的模型结构体
│   │   └── query/                  # Gen 生成的查询构建器
│   ├── dto/                        # 数据传输对象
│   │   ├── auth_dto.go            # 认证相关 DTO
│   │   ├── user_dto.go            # 用户 DTO (含 binding 验证 tag)
│   │   ├── role_dto.go            # 角色 DTO
│   │   └── ...                    # 其他领域 DTO
│   ├── converter/                  # Model ↔ DTO 转换
│   │   ├── user_converter.go
│   │   ├── role_converter.go
│   │   └── ...
│   ├── repository/                 # 数据访问封装
│   │   ├── user_repo.go           # 用户仓储
│   │   ├── role_repo.go           # 角色仓储
│   │   └── ...
│   ├── service/                    # 业务逻辑
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── role_service.go
│   │   └── ...
│   ├── handler/                    # HTTP 处理器
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   └── ...
│   ├── middleware/                  # Gin 中间件
│   │   ├── request_id.go
│   │   ├── logger.go
│   │   ├── recovery.go
│   │   ├── cors.go
│   │   ├── rate_limit.go
│   │   ├── auth.go                # JWT 认证
│   │   ├── rbac.go                # RBAC 权限检查
│   │   └── audit.go               # 审计日志
│   ├── router/                     # 路由定义 + 依赖注入
│   │   ├── app.go                 # App 结构 + DI
│   │   └── router.go              # 路由注册
│   └── rbac/                       # RBAC 核心
│       ├── cache.go               # 权限缓存
│       ├── middleware.go           # RBAC 中间件（从 cache 移到这）
│       └── service.go             # RBAC 业务逻辑
├── pkg/                            # 可复用包
│   ├── database/                   # 数据库
│   │   ├── postgres.go            # GORM 连接初始化
│   │   ├── tenant.go             # ★ TenantScope / SetTenantID helper
│   │   └── tx.go                 # 事务封装
│   ├── jwt/                       # JWT 管理
│   ├── xredis/                    # Redis 封装
│   ├── response/                  # 统一 JSON 响应
│   ├── xerr/                      # 错误类型 + 错误码
│   ├── xcontext/                  # Context helper (tenant, user, roles)
│   ├── logger/                    # Zerolog 封装
│   ├── audit/                     # 审计日志
│   ├── password/                  # Argon2id 哈希
│   ├── captcha/                   # 验证码
│   ├── xcron/                     # 定时任务管理
│   ├── pagination/                # 分页 helper
│   └── utils/                     # 业务无关工具
│       ├── convert/              # 泛型转换
│       ├── csv/                  # CSV 导出
│       ├── httpclient/           # HTTP 客户端
│       ├── useragent/            # UA 解析
│       └── bodyreader/           # Body 读取
├── config/                        # 配置文件
│   ├── config.yaml               # 基础配置
│   ├── config.dev.yaml           # 开发环境
│   └── config.prod.yaml          # 生产环境
├── migrations/                    # 数据库迁移
├── scripts/                       # 工具脚本
│   └── gen_from_db.go            # Gen 代码生成
├── docs/                          # API 文档 + 设计文档
├── sqlc.yaml                      # (不用 sqlc，仅参考)
├── Makefile
├── go.mod
└── CLAUDE.md
```

### 3.2 分层依赖规则

```
handler → service → repository → dal/query (Gen 生成)
   │         │          │
   └─── dto ─┘          └── model (Gen 生成)
                           │
                      converter (model ↔ dto)

middleware → pkg/xcontext, pkg/jwt, internal/rbac
router → handler, middleware, pkg/* (DI 组装)

pkg/ 不依赖 internal/
internal/ 可依赖 pkg/
```

---

## 4. pkg 目录重构

### 4.1 分类表

| 包 | 操作 | 理由 |
|---|---|---|
| `pkg/casbin/` | **删除** | 被 internal/rbac 替代 |
| `pkg/database/scopes.go` | **删除** | GORM Callback 被 TenantScope helper 替代 |
| `pkg/rsapwd/` | **删除** | 安全风险，密码传输改用 HTTPS |
| `pkg/cache/` | **删除** | 租户缓存合并到 service 层 |
| `pkg/constants/` | **移动** → `internal/constants/` | 与业务相关 |
| `pkg/idgen/` | **删除** | 直接用 `github.com/google/uuid` |
| `pkg/convert/` | **移动** → `pkg/utils/convert/` | 业务无关 |
| `pkg/csv/` | **移动** → `pkg/utils/csv/` | 业务无关 |
| `pkg/httpclient/` | **移动** → `pkg/utils/httpclient/` | 业务无关 |
| `pkg/useragent/` | **移动** → `pkg/utils/useragent/` | 业务无关 |
| `pkg/bodyreader/` | **移动** → `pkg/utils/bodyreader/` | 业务无关 |
| `pkg/database/` | **保留**，重写 tenant.go | 连接管理 + TenantScope helper |
| `pkg/jwt/` | **保留** | JWT 管理，依赖 xredis |
| `pkg/xredis/` | **保留** | Redis 封装 |
| `pkg/response/` | **保留** | 统一 JSON 响应，依赖 xerr |
| `pkg/xerr/` | **保留** | 错误类型 |
| `pkg/xcontext/` | **保留** | Context helper |
| `pkg/logger/` | **保留** | Zerolog |
| `pkg/audit/` | **保留** | 审计日志，依赖 database |
| `pkg/password/` | **保留**（重命名 passwordgen→password） | Argon2id |
| `pkg/captcha/` | **保留** | 验证码 |
| `pkg/xcron/` | **保留** | 定时任务 |
| `pkg/pagination/` | **保留** | 分页 helper |

---

## 5. 数据库层设计

### 5.1 核心设计原则

1. **GORM Gen 生成模型和查询** — 数据库优先，写迁移 → gen-db → 业务代码
2. **移除 GORM Callback** — 不再隐式注入 tenant_id
3. **显式 TenantScope helper** — Repository 层显式调用
4. **自动时间戳** — `autoCreateTime:milli` / `autoUpdateTime:milli` tag
5. **自动软删除** — `softDelete:milli` tag

### 5.2 GORM 连接初始化

```go
// pkg/database/postgres.go
package database

import (
    "fmt"
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

    if err := sqlDB.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    // 注意：不再调用 RegisterCallbacks(db)
    // 多租户过滤改为显式 TenantScope helper

    globalDB = db
    return db, nil
}

func Get() *gorm.DB { return globalDB }

func Close() {
    if globalDB != nil {
        sqlDB, _ := globalDB.DB()
        sqlDB.Close()
    }
}
```

### 5.3 显式多租户 Helper（核心改进）

```go
// pkg/database/tenant.go
package database

import (
    "admin/pkg/xcontext"
    "context"
    "reflect"

    "gorm.io/gorm"
)

// TenantScope 为查询添加租户过滤条件
// 在 Repository 的每个查询方法中显式调用
func TenantScope(ctx context.Context, db *gorm.DB) *gorm.DB {
    tenantID := xcontext.GetTenantID(ctx)
    if tenantID == "" {
        return db
    }
    return db.Where("tenant_id = ?", tenantID)
}

// SetTenantID 为新建记录设置租户 ID
// 在 Repository 的 Create 方法中显式调用
func SetTenantID(ctx context.Context, model interface{}) {
    tenantID := xcontext.GetTenantID(ctx)
    if tenantID == "" {
        return
    }
    v := reflect.ValueOf(model)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }
    field := v.FieldByName("TenantID")
    if field.IsValid() && field.CanSet() && field.String() == "" {
        field.SetString(tenantID)
    }
}
```

**使用方式**：

```go
// 查询：显式加租户过滤
db := database.TenantScope(ctx, r.db.WithContext(ctx))
db.Where("user_id = ?", userID).First(&user)

// 创建：显式设置租户 ID
database.SetTenantID(ctx, user)
r.db.WithContext(ctx).Create(user)

// 跨租户查询：不加 TenantScope
r.db.WithContext(ctx).Where("email = ?", email).First(&user)
```

### 5.4 时间戳自动处理

GORM tag 自动处理，无需任何代码：

```go
type User struct {
    // ...
    CreatedAt int64                 `gorm:"column:created_at;autoCreateTime:milli"` // 创建时自动设为当前毫秒时间戳
    UpdatedAt int64                 `gorm:"column:updated_at;autoUpdateTime:milli"` // 创建和更新时自动设为当前毫秒时间戳
    DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli"`     // Delete 时自动设为毫秒时间戳，查询自动过滤
}
```

- `autoCreateTime:milli` — GORM Create 时自动设置为 `time.Now().UnixMilli()`
- `autoUpdateTime:milli` — GORM Create 和 Update 时自动设置
- `softDelete:milli` — GORM Delete 时设置时间戳，查询自动添加 `WHERE deleted_at = 0`

### 5.5 完整表结构 DDL

```sql
-- =====================================================
-- 租户表（全局表，无 tenant_id）
-- =====================================================
CREATE TABLE tenants (
    tenant_id   VARCHAR(20)  PRIMARY KEY,
    tenant_code VARCHAR(50)  NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    contact_name VARCHAR(100) NOT NULL DEFAULT '',
    contact_phone VARCHAR(20) NOT NULL DEFAULT '',
    status      SMALLINT     NOT NULL DEFAULT 1,
    created_at  BIGINT       NOT NULL DEFAULT 0,
    updated_at  BIGINT       NOT NULL DEFAULT 0,
    deleted_at  BIGINT       DEFAULT 0
);
CREATE UNIQUE INDEX uk_tenants_code ON tenants(tenant_code) WHERE deleted_at = 0;

-- =====================================================
-- 用户表（租户隔离）
-- =====================================================
CREATE TABLE users (
    user_id             VARCHAR(20)  PRIMARY KEY,
    tenant_id           VARCHAR(20)  NOT NULL,
    user_name           VARCHAR(100) NOT NULL DEFAULT '',
    password            VARCHAR(100) NOT NULL,
    nickname            VARCHAR(100) NOT NULL,
    avatar              VARCHAR(255) NOT NULL DEFAULT '',
    phone               VARCHAR(20)  NOT NULL DEFAULT '',
    email               VARCHAR(100) NOT NULL DEFAULT '',
    description         TEXT         NOT NULL DEFAULT '',
    department_id       VARCHAR(20)  NOT NULL DEFAULT '',
    position_id         VARCHAR(20)  NOT NULL DEFAULT '',
    status              SMALLINT     NOT NULL DEFAULT 1,
    remark              TEXT         NOT NULL DEFAULT '',
    last_login_time     BIGINT       NOT NULL DEFAULT 0,
    must_change_password SMALLINT    NOT NULL DEFAULT 1,
    created_at          BIGINT       NOT NULL DEFAULT 0,
    updated_at          BIGINT       NOT NULL DEFAULT 0,
    deleted_at          BIGINT       DEFAULT 0
);
CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE UNIQUE INDEX uk_users_email ON users(email) WHERE deleted_at = 0;
CREATE UNIQUE INDEX uk_users_phone ON users(phone) WHERE deleted_at = 0 AND phone != '';

-- =====================================================
-- 角色表（租户隔离，新增 parent_role_id）
-- =====================================================
CREATE TABLE roles (
    role_id         VARCHAR(20)  PRIMARY KEY,
    tenant_id       VARCHAR(20)  NOT NULL,
    role_code       VARCHAR(50)  NOT NULL,
    name            VARCHAR(100) NOT NULL,
    description     TEXT         NOT NULL DEFAULT '',
    parent_role_id  VARCHAR(20)  DEFAULT '',  -- ★ 新增：替代 Casbin g2 策略
    status          SMALLINT     NOT NULL DEFAULT 1,
    created_at      BIGINT       NOT NULL DEFAULT 0,
    updated_at      BIGINT       NOT NULL DEFAULT 0,
    deleted_at      BIGINT       DEFAULT 0
);
CREATE INDEX idx_roles_tenant ON roles(tenant_id);
CREATE UNIQUE INDEX uk_roles_tenant_code ON roles(tenant_id, role_code) WHERE deleted_at = 0;

-- =====================================================
-- ★ 用户角色关联表（替代 Casbin g 策略）
-- =====================================================
CREATE TABLE user_roles (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    VARCHAR(20)  NOT NULL REFERENCES users(user_id),
    role_id    VARCHAR(20)  NOT NULL REFERENCES roles(role_id),
    tenant_id  VARCHAR(20)  NOT NULL,
    created_at BIGINT       NOT NULL DEFAULT 0,
    UNIQUE(user_id, role_id, tenant_id)
);
CREATE INDEX idx_user_roles_user_tenant ON user_roles(user_id, tenant_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);

-- =====================================================
-- ★ 角色权限关联表（替代 Casbin p 策略）
-- =====================================================
CREATE TABLE role_permissions (
    id            BIGSERIAL    PRIMARY KEY,
    role_id       VARCHAR(20)  NOT NULL REFERENCES roles(role_id),
    permission_id VARCHAR(20)  NOT NULL REFERENCES permissions(permission_id),
    tenant_id     VARCHAR(20)  NOT NULL,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    UNIQUE(role_id, permission_id, tenant_id)
);
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id, tenant_id);

-- =====================================================
-- 菜单表（全局表，无 tenant_id）
-- =====================================================
CREATE TABLE menus (
    menu_id     VARCHAR(20)   PRIMARY KEY,
    parent_id   VARCHAR(20)   NOT NULL DEFAULT '',
    name        VARCHAR(100)  NOT NULL,
    path        VARCHAR(255)  NOT NULL DEFAULT '',
    component   VARCHAR(255)  NOT NULL DEFAULT '',
    redirect    VARCHAR(255)  NOT NULL DEFAULT '',
    icon        VARCHAR(100)  NOT NULL DEFAULT '',
    sort        INT           NOT NULL DEFAULT 0,
    status      SMALLINT      NOT NULL DEFAULT 1,
    api_paths   TEXT          NOT NULL DEFAULT '[]',
    description TEXT          NOT NULL DEFAULT '',
    created_at  BIGINT        NOT NULL DEFAULT 0,
    updated_at  BIGINT        NOT NULL DEFAULT 0,
    deleted_at  BIGINT        DEFAULT 0
);

-- =====================================================
-- 权限表（全局表，无 tenant_id）
-- =====================================================
CREATE TABLE permissions (
    permission_id VARCHAR(20)  PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    type          VARCHAR(20)  NOT NULL,
    resource      VARCHAR(255) NOT NULL DEFAULT '',
    action        VARCHAR(50)  NOT NULL DEFAULT '',
    description   TEXT         NOT NULL DEFAULT '',
    created_at    BIGINT       NOT NULL DEFAULT 0,
    updated_at    BIGINT       NOT NULL DEFAULT 0,
    deleted_at    BIGINT       DEFAULT 0
);

-- =====================================================
-- 部门表（租户隔离）
-- =====================================================
CREATE TABLE departments (
    department_id   VARCHAR(20)  PRIMARY KEY,
    tenant_id       VARCHAR(20)  NOT NULL,
    parent_id       VARCHAR(20)  NOT NULL DEFAULT '',
    department_name VARCHAR(100) NOT NULL,
    description     TEXT         NOT NULL DEFAULT '',
    sort            INT          NOT NULL DEFAULT 0,
    status          SMALLINT     NOT NULL DEFAULT 1,
    created_at      BIGINT       NOT NULL DEFAULT 0,
    updated_at      BIGINT       NOT NULL DEFAULT 0,
    deleted_at      BIGINT       DEFAULT 0
);
CREATE INDEX idx_departments_tenant ON departments(tenant_id);

-- =====================================================
-- 岗位表（租户隔离）
-- =====================================================
CREATE TABLE positions (
    position_id   VARCHAR(20)  PRIMARY KEY,
    tenant_id     VARCHAR(20)  NOT NULL,
    position_code VARCHAR(50)  NOT NULL,
    position_name VARCHAR(100) NOT NULL,
    level         INT          NOT NULL DEFAULT 0,
    description   TEXT         NOT NULL DEFAULT '',
    sort          INT          NOT NULL DEFAULT 0,
    status        SMALLINT     NOT NULL DEFAULT 1,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    updated_at    BIGINT       NOT NULL DEFAULT 0,
    deleted_at    BIGINT       DEFAULT 0
);
CREATE INDEX idx_positions_tenant ON positions(tenant_id);
CREATE UNIQUE INDEX uk_positions_tenant_code ON positions(tenant_id, position_code) WHERE deleted_at = 0;

-- =====================================================
-- 租户菜单关联表
-- =====================================================
CREATE TABLE tenant_menus (
    id         BIGSERIAL   PRIMARY KEY,
    tenant_id  VARCHAR(20) NOT NULL,
    menu_id    VARCHAR(20) NOT NULL,
    created_at BIGINT      NOT NULL DEFAULT 0,
    deleted_at BIGINT      DEFAULT 0,
    UNIQUE(tenant_id, menu_id)
);

-- =====================================================
-- 字典类型表（租户隔离）
-- =====================================================
CREATE TABLE dict_types (
    dict_type_id VARCHAR(20)  PRIMARY KEY,
    tenant_id    VARCHAR(20)  NOT NULL,
    dict_name    VARCHAR(100) NOT NULL,
    dict_type    VARCHAR(100) NOT NULL,
    status       SMALLINT     NOT NULL DEFAULT 1,
    remark       TEXT         NOT NULL DEFAULT '',
    created_at   BIGINT       NOT NULL DEFAULT 0,
    updated_at   BIGINT       NOT NULL DEFAULT 0,
    deleted_at   BIGINT       DEFAULT 0
);
CREATE INDEX idx_dict_types_tenant ON dict_types(tenant_id);

-- =====================================================
-- 字典项表（租户隔离）
-- =====================================================
CREATE TABLE dict_items (
    dict_item_id VARCHAR(20)  PRIMARY KEY,
    tenant_id    VARCHAR(20)  NOT NULL,
    dict_type_id VARCHAR(20)  NOT NULL,
    label        VARCHAR(100) NOT NULL,
    value        VARCHAR(100) NOT NULL,
    sort         INT          NOT NULL DEFAULT 0,
    status       SMALLINT     NOT NULL DEFAULT 1,
    remark       TEXT         NOT NULL DEFAULT '',
    created_at   BIGINT       NOT NULL DEFAULT 0,
    updated_at   BIGINT       NOT NULL DEFAULT 0,
    deleted_at   BIGINT       DEFAULT 0
);
CREATE INDEX idx_dict_items_tenant ON dict_items(tenant_id);
CREATE INDEX idx_dict_items_type ON dict_items(dict_type_id);

-- =====================================================
-- 登录日志表（租户隔离，无软删除）
-- =====================================================
CREATE TABLE login_logs (
    log_id          VARCHAR(20)  PRIMARY KEY,
    tenant_id       VARCHAR(20)  NOT NULL,
    user_id         VARCHAR(20)  NOT NULL DEFAULT '',
    user_name       VARCHAR(100) NOT NULL DEFAULT '',
    operation_type  VARCHAR(20)  NOT NULL DEFAULT '',
    login_type      VARCHAR(20)  NOT NULL DEFAULT '',
    login_ip        VARCHAR(50)  NOT NULL DEFAULT '',
    login_location  VARCHAR(255) NOT NULL DEFAULT '',
    user_agent      TEXT         NOT NULL DEFAULT '',
    status          SMALLINT     NOT NULL DEFAULT 1,
    fail_reason     TEXT         NOT NULL DEFAULT '',
    created_at      BIGINT       NOT NULL DEFAULT 0
);
CREATE INDEX idx_login_logs_tenant ON login_logs(tenant_id);
CREATE INDEX idx_login_logs_user ON login_logs(user_id);

-- =====================================================
-- 操作日志表（租户隔离，无软删除）
-- =====================================================
CREATE TABLE operation_logs (
    log_id          VARCHAR(20)  PRIMARY KEY,
    tenant_id       VARCHAR(20)  NOT NULL,
    user_id         VARCHAR(20)  NOT NULL DEFAULT '',
    user_name       VARCHAR(100) NOT NULL DEFAULT '',
    module          VARCHAR(50)  NOT NULL DEFAULT '',
    operation_type  VARCHAR(50)  NOT NULL DEFAULT '',
    resource_type   VARCHAR(50)  NOT NULL DEFAULT '',
    resource_id     VARCHAR(20)  NOT NULL DEFAULT '',
    resource_name   VARCHAR(255) NOT NULL DEFAULT '',
    request_method  VARCHAR(10)  NOT NULL DEFAULT '',
    request_path    VARCHAR(255) NOT NULL DEFAULT '',
    request_params  TEXT         NOT NULL DEFAULT '',
    old_value       TEXT         NOT NULL DEFAULT '',
    new_value       TEXT         NOT NULL DEFAULT '',
    status          SMALLINT     NOT NULL DEFAULT 1,
    error_message   TEXT         NOT NULL DEFAULT '',
    ip_address      VARCHAR(50)  NOT NULL DEFAULT '',
    user_agent      TEXT         NOT NULL DEFAULT '',
    created_at      BIGINT       NOT NULL DEFAULT 0
);
CREATE INDEX idx_operation_logs_tenant ON operation_logs(tenant_id);
CREATE INDEX idx_operation_logs_user ON operation_logs(user_id);
```

---

## 6. Repository 层设计

### 6.1 标准模式

```go
// internal/repository/user_repo.go
package repository

import (
    "admin/internal/dal/model"
    "admin/internal/dal/query"
    "admin/pkg/database"
    "admin/pkg/xcontext"
    "context"

    "gorm.io/gorm"
)

type UserRepo struct {
    db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
    return &UserRepo{db: db}
}

// Create 创建用户 — 显式设置 tenant_id
func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
    database.SetTenantID(ctx, user)
    return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据ID获取 — 显式租户过滤
func (r *UserRepo) GetByID(ctx context.Context, userID string) (*model.User, error) {
    var user model.User
    err := database.TenantScope(ctx, r.db.WithContext(ctx)).
        Where("user_id = ?", userID).
        First(&user).Error
    return &user, err
}

// GetByEmail 根据邮箱获取 — 跨租户查询（登录场景），不加 TenantScope
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).
        Where("email = ?", email).
        First(&user).Error
    return &user, err
}

// GetByPhone 根据手机号获取 — 跨租户查询
func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).
        Where("phone = ?", phone).
        First(&user).Error
    return &user, err
}

// Update 更新用户 — 显式租户过滤
func (r *UserRepo) Update(ctx context.Context, userID string, updates map[string]interface{}) error {
    return database.TenantScope(ctx, r.db.WithContext(ctx)).
        Where("user_id = ?", userID).
        Updates(updates).Error
}

// Delete 软删除 — 显式租户过滤
func (r *UserRepo) Delete(ctx context.Context, userID string) error {
    return database.TenantScope(ctx, r.db.WithContext(ctx)).
        Where("user_id = ?", userID).
        Delete(&model.User{}).Error
}

// BatchDelete 批量软删除
func (r *UserRepo) BatchDelete(ctx context.Context, userIDs []string) error {
    return database.TenantScope(ctx, r.db.WithContext(ctx)).
        Where("user_id IN ?", userIDs).
        Delete(&model.User{}).Error
}

// ListWithFilters 动态条件查询 — 显式租户过滤 + 动态 WHERE
func (r *UserRepo) ListWithFilters(ctx context.Context, offset, limit int,
    nicknameFilter string, statusFilter int) ([]*model.User, int64, error) {

    q := database.TenantScope(ctx, r.db.WithContext(ctx).Model(&model.User{}))

    // 动态 WHERE 条件（GORM 链式 API 优势）
    if nicknameFilter != "" {
        q = q.Where("nickname LIKE ?", "%"+nicknameFilter+"%")
    }
    if statusFilter != 0 {
        q = q.Where("status = ?", statusFilter)
    }

    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    var users []*model.User
    err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error
    return users, total, err
}

// ListByIDsAndFilters 按用户ID列表 + 动态条件查询
func (r *UserRepo) ListByIDsAndFilters(ctx context.Context, userIDs []string,
    offset, limit int, keywordFilter string, statusFilter int) ([]*model.User, int64, error) {

    q := database.TenantScope(ctx, r.db.WithContext(ctx).Model(&model.User{})).
        Where("user_id IN ?", userIDs)

    if keywordFilter != "" {
        q = q.Where("user_name LIKE ? OR nickname LIKE ?",
            "%"+keywordFilter+"%", "%"+keywordFilter+"%")
    }
    if statusFilter != 0 {
        q = q.Where("status = ?", statusFilter)
    }

    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    var users []*model.User
    err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error
    return users, total, err
}

// CountByTenantID 统计指定租户用户数 — 跨租户查询
func (r *UserRepo) CountByTenantID(ctx context.Context, tenantID string) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Where("tenant_id = ?", tenantID).
        Count(&count).Error
    return count, err
}

// CountByTenantIDs 批量统计多租户用户数 — 跨租户查询
func (r *UserRepo) CountByTenantIDs(ctx context.Context, tenantIDs []string) (map[string]int64, error) {
    type result struct {
        TenantID  string
        UserCount int64
    }
    var results []result
    err := r.db.WithContext(ctx).Model(&model.User{}).
        Select("tenant_id, count(*) as user_count").
        Where("tenant_id IN ?", tenantIDs).
        Group("tenant_id").
        Find(&results).Error
    if err != nil {
        return nil, err
    }

    countMap := make(map[string]int64, len(results))
    for _, r := range results {
        countMap[r.TenantID] = r.UserCount
    }
    return countMap, nil
}

// CheckExists 检查用户名是否已存在
func (r *UserRepo) CheckExists(ctx context.Context, userName string) (bool, error) {
    var count int64
    err := database.TenantScope(ctx, r.db.WithContext(ctx).Model(&model.User{})).
        Where("user_name = ?", userName).
        Count(&count).Error
    return count > 0, err
}
```

### 6.2 关键设计点

| 操作 | 方式 | 说明 |
|------|------|------|
| 租户内查询 | `database.TenantScope(ctx, db)` | 显式添加 `WHERE tenant_id = ?` |
| 跨租户查询 | 直接 `r.db.WithContext(ctx)` | 不加 TenantScope |
| 创建记录 | `database.SetTenantID(ctx, model)` | 显式设置 tenant_id |
| 更新/删除 | `database.TenantScope(ctx, db)` + Where 条件 | 确保只操作本租户数据 |
| 动态筛选 | GORM 链式 API `if` 判断 | 无需为每个组合写不同 SQL |
| 时间戳 | GORM tag 自动处理 | 无需手动设置 |
| 软删除 | GORM tag 自动处理 | 查询自动过滤，删除自动设值 |

### 6.3 事务封装

```go
// pkg/database/tx.go
package database

import (
    "context"
    "gorm.io/gorm"
)

func InTransaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
    return db.WithContext(ctx).Transaction(fn)
}
```

---

## 7. RBAC 设计（替代 Casbin）

### 7.1 架构对比

```
当前（Casbin）:
  用户 → Casbin g 策略(内存) → 角色编码
  角色 → Casbin g2 策略(内存) → DFS 继承角色
  角色 → Casbin p 策略(内存) → 权限(resource, action)
  中间件 → Casbin Enforce(username, "default", path, method)

新系统（纯数据库 RBAC）:
  用户 → user_roles 表 → 角色 ID
  角色 → roles.parent_role_id → 递归 CTE 继承
  角色 → role_permissions 表 → permissions 表
  中间件 → PermissionCache(内存) → 检查 API 权限
```

### 7.2 权限缓存

```go
// internal/rbac/cache.go
package rbac

import (
    "admin/internal/dal/model"
    "context"
    "strings"
    "sync"
    "time"

    "github.com/rs/zerolog/log"
    "gorm.io/gorm"
)

type APIPermission struct {
    Path   string
    Method string
}

type PermissionCache struct {
    mu        sync.RWMutex
    apiPerms  map[string][]APIPermission  // roleID → API 权限列表
    menuPerms map[string][]string         // roleID → menuID 列表
    db        *gorm.DB
    ttl       time.Duration
    lastLoad  time.Time
}

func NewPermissionCache(db *gorm.DB, ttl time.Duration) *PermissionCache {
    return &PermissionCache{
        apiPerms:  make(map[string][]APIPermission),
        menuPerms: make(map[string][]string),
        db:        db,
        ttl:       ttl,
    }
}

// Refresh 刷新权限缓存（定时任务调用，如每 30 秒）
func (c *PermissionCache) Refresh(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    newAPIPerms := make(map[string][]APIPermission)
    newMenuPerms := make(map[string][]string)

    // 查询所有角色的 API 权限（含继承角色的权限）
    // 使用递归 CTE 获取角色及其所有祖先角色
    var apiResults []struct {
        RoleID   string
        Resource string
        Action   string
    }
    err := c.db.WithContext(ctx).Raw(`
        WITH RECURSIVE role_tree AS (
            -- 基础：所有角色
            SELECT role_id, parent_role_id FROM roles WHERE deleted_at = 0
            UNION
            -- 递归：父角色
            SELECT r.role_id, r.parent_role_id
            FROM roles r
            JOIN role_tree rt ON r.role_id = rt.parent_role_id
            WHERE r.deleted_at = 0
        )
        SELECT DISTINCT rt.role_id, p.resource, p.action
        FROM role_permissions rp
        JOIN permissions p ON p.permission_id = rp.permission_id AND p.deleted_at = 0
        JOIN role_tree rt ON rt.role_id = rp.role_id
        WHERE p.type = 'API'
    `).Scan(&apiResults).Error
    if err != nil {
        return err
    }

    for _, r := range apiResults {
        newAPIPerms[r.RoleID] = append(newAPIPerms[r.RoleID], APIPermission{
            Path:   r.Resource,
            Method: r.Action,
        })
    }

    // 查询所有角色的菜单权限
    var menuResults []struct {
        RoleID   string
        Resource string
    }
    err = c.db.WithContext(ctx).Raw(`
        WITH RECURSIVE role_tree AS (
            SELECT role_id, parent_role_id FROM roles WHERE deleted_at = 0
            UNION
            SELECT r.role_id, r.parent_role_id
            FROM roles r
            JOIN role_tree rt ON r.role_id = rt.parent_role_id
            WHERE r.deleted_at = 0
        )
        SELECT DISTINCT rt.role_id, p.resource
        FROM role_permissions rp
        JOIN permissions p ON p.permission_id = rp.permission_id AND p.deleted_at = 0
        JOIN role_tree rt ON rt.role_id = rp.role_id
        WHERE p.type = 'MENU'
    `).Scan(&menuResults).Error
    if err != nil {
        return err
    }

    for _, r := range menuResults {
        // resource 格式: "menu:xxx"
        menuID := strings.TrimPrefix(r.Resource, "menu:")
        newMenuPerms[r.RoleID] = append(newMenuPerms[r.RoleID], menuID)
    }

    c.apiPerms = newAPIPerms
    c.menuPerms = newMenuPerms
    c.lastLoad = time.Now()

    log.Info().Int("api_rules", len(apiResults)).Int("menu_rules", len(menuResults)).
        Msg("权限缓存刷新完成")

    return nil
}

// CheckAPI 检查角色是否有指定 API 权限
func (c *PermissionCache) CheckAPI(roleIDs []string, path, method string) bool {
    c.mu.RLock()
    defer c.mu.RUnlock()

    for _, roleID := range roleIDs {
        for _, perm := range c.apiPerms[roleID] {
            if matchPath(perm.Path, path) && matchMethod(perm.Method, method) {
                return true
            }
        }
    }
    return false
}

// GetMenuIDs 获取角色的菜单 ID 列表
func (c *PermissionCache) GetMenuIDs(roleIDs []string) []string {
    c.mu.RLock()
    defer c.mu.RUnlock()

    seen := make(map[string]bool)
    var result []string
    for _, roleID := range roleIDs {
        for _, menuID := range c.menuPerms[roleID] {
            if !seen[menuID] {
                seen[menuID] = true
                result = append(result, menuID)
            }
        }
    }
    return result
}

// matchPath 路径匹配（支持 /api/v1/users/:id 等）
func matchPath(pattern, path string) bool {
    if pattern == path {
        return true
    }
    // 简单的通配符匹配
    patternParts := strings.Split(pattern, "/")
    pathParts := strings.Split(path, "/")
    if len(patternParts) != len(pathParts) {
        return false
    }
    for i := range patternParts {
        if strings.HasPrefix(patternParts[i], ":") || strings.HasPrefix(patternParts[i], "*") {
            continue
        }
        if patternParts[i] != pathParts[i] {
            return false
        }
    }
    return true
}

// matchMethod HTTP 方法匹配
func matchMethod(pattern, method string) bool {
    return pattern == "*" || pattern == method
}
```

### 7.3 RBAC 中间件

```go
// internal/middleware/rbac.go
package middleware

import (
    "admin/internal/rbac"
    "admin/pkg/response"
    "admin/pkg/xcontext"
    "admin/pkg/xerr"

    "github.com/gin-gonic/gin"
)

func RBACMiddleware(cache *rbac.PermissionCache) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        // 超级管理员跳过权限检查
        if xcontext.HasRole(ctx, "super_admin") {
            c.Next()
            return
        }

        roleIDs := xcontext.GetRoleIDs(ctx)
        if len(roleIDs) == 0 {
            response.Error(c, xerr.ErrForbidden)
            c.Abort()
            return
        }

        path := c.Request.URL.Path
        method := c.Request.Method

        if !cache.CheckAPI(roleIDs, path, method) {
            response.Error(c, xerr.ErrForbidden)
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### 7.4 菜单权限解析（一条 SQL 替代 4 步 Casbin）

```go
// 当前（Casbin，4 步）:
// 1. roles := enforcer.GetRolesForUserInDomain(userName, tenantCode)
// 2. allRoleCodes := getAllRoleCodes(roles)  // DFS
// 3. menuIDs := getMenuPermissionsForRoles(allRoleCodes)  // N 次查询
// 4. menus := menuRepo.GetByIDs(ctx, menuIDs)

// 新系统（一条 SQL）:
func (s *UserMenuService) GetUserMenuIDs(ctx context.Context, roleIDs []string) ([]string, error) {
    return s.cache.GetMenuIDs(roleIDs), nil
}
```

---

## 8. 中间件链设计

### 8.1 中间件顺序

```
RequestID → Logger → Recovery → CORS → RateLimit → Auth → RBAC → Audit
```

### 8.2 变更说明

| 中间件 | 变更 | 说明 |
|--------|------|------|
| CasbinMiddleware | **删除** | 替换为 RBAC 中间件 |
| TenantSkipMiddleware | **删除** | 不再需要，无 Callback 需要跳过 |
| AuthMiddleware | **保留** | JWT 验证，注入 context |
| RBACMiddleware | **新增** | 基于 PermissionCache 检查 |
| AuditMiddleware | **保留** | 提取请求信息 |

### 8.3 Context 注入

Auth 中间件从 JWT 提取并注入 context：

```go
// pkg/xcontext/user.go — 新增 RoleIDs
func SetRoleIDs(ctx context.Context, roleIDs []string) context.Context {
    return context.WithValue(ctx, contextKey("role_ids"), roleIDs)
}

func GetRoleIDs(ctx context.Context) []string {
    val, _ := ctx.Value(contextKey("role_ids")).([]string)
    return val
}
```

JWT Claims 需要扩展，包含 `RoleIDs []string`（替代当前的 `Roles []string` 角色编码列表）。

---

## 9. 错误处理

保留现有 `pkg/xerr/` 和 `pkg/response/` 体系，不变。

### 9.1 错误码范围

| 范围 | 分类 | 示例 |
|------|------|------|
| 200 | 成功 | Success |
| 1000-1999 | 通用错误 | Internal, InvalidParams, NotFound, Unauthorized, Forbidden |
| 2000-2099 | 数据库错误 | RecordNotFound, CreateError, UpdateError |
| 2100-2199 | 认证错误 | InvalidCredentials, TokenExpired, UserDisabled |
| 2200-2299 | 租户错误 | TenantNotFound, TenantDisabled |
| 2300-2399 | 角色错误 | RoleNotFound, RoleInUse |
| 2400-2499 | 菜单错误 | MenuNotFound, HasChildren |
| 2500-2599 | 部门错误 | DeptNotFound |
| 2600-2699 | 岗位错误 | PositionNotFound |

---

## 10. 配置管理

保留 Viper 体系，不变。新增 RBAC 缓存刷新间隔配置：

```yaml
# config.yaml
rbac:
  cache_refresh_interval: 30s  # 权限缓存刷新间隔
```

---

## 11. 请求验证

在 DTO 上使用 `binding` tag 定义验证规则：

```go
type CreateUserRequest struct {
    UserName     string `json:"user_name" binding:"required,alphanum,min=3,max=50"`
    Nickname     string `json:"nickname" binding:"required,min=1,max=100"`
    Email        string `json:"email" binding:"required,email"`
    Phone        string `json:"phone" binding:"required,len=11"`
    DepartmentID string `json:"department_id"`
    PositionID   string `json:"position_id"`
    Remark       string `json:"remark"`
}
```

Handler 中统一使用 `c.ShouldBindJSON(&req)` 或 `c.ShouldBindQuery(&req)`。

---

## 12. 其他考虑

### 12.1 审计日志

保留 `pkg/audit/`，写入时改用显式 `database.SetTenantID(ctx, logModel)`。

### 12.2 CORS

修复当前的 `AllowAllOrigins: true` + `AllowCredentials: true` 不兼容问题，改为配置允许的域名列表。

### 12.3 安全

- 移除硬编码的 RSA 私钥，改用配置文件/环境变量
- 密码传输依赖 HTTPS，不再使用 RSA 加密
- 添加安全响应头（X-Frame-Options, X-Content-Type-Options 等）

### 12.4 限流

保留内存限流，后续可升级为 Redis 限流。

### 12.5 健康检查

保留 `/health` 和 `/ping` 端点。

### 12.6 优雅关闭

保留现有的 signal 处理机制。

### 12.7 Makefile

```makefile
# 新增命令
.PHONY: rbac-refresh
rbac-refresh: ## 手动刷新权限缓存
	@curl -s http://localhost:8080/api/v1/system/rbac/refresh
```

### 12.8 测试策略

- **Repository 层**：集成测试（需要数据库）
- **Service 层**：单元测试（mock Repository）
- **Handler 层**：HTTP 测试（httptest）
- **RBAC 缓存**：单元测试（验证权限匹配逻辑）

---

## 13. 实施步骤

### 阶段一：基础架构改造（~2 天）

1. 创建 `pkg/database/tenant.go`（TenantScope + SetTenantID helper）
2. 移除 `pkg/database/scopes.go`（GORM Callback）
3. 更新 `pkg/database/postgres.go`（移除 RegisterCallbacks 调用）
4. 重写所有 Repository（显式 TenantScope 模式）
5. 移除 `TenantSkipMiddleware` 和 `SkipTenantCheck` 相关代码

### 阶段二：RBAC 替换 Casbin（~3 天）

1. 创建迁移文件（user_roles、role_permissions 表，roles 加 parent_role_id）
2. 编写数据迁移脚本（casbin_rule → 新表）
3. 创建 `internal/rbac/cache.go`（PermissionCache）
4. 创建 `internal/middleware/rbac.go`（RBAC 中间件）
5. 重写 `role_service.go`（权限分配改为操作 role_permissions 表）
6. 重写 `user_role_service.go`（角色绑定改为操作 user_roles 表）
7. 重写 `user_menu_service.go`（菜单查询改为 PermissionCache）
8. 移除 `pkg/casbin/`、`internal/middleware/casbin_middleware.go`

### 阶段三：pkg 重构（~1 天）

1. 创建 `pkg/utils/` 目录，移动业务无关包
2. 删除废弃包（casbin、rsapwd、cache、idgen）
3. 移动 `pkg/constants/` → `internal/constants/`

### 阶段四：清理和验证（~1 天）

1. 更新 `cmd/server/main.go`（移除 Casbin 初始化，新增 PermissionCache 初始化）
2. 更新 `internal/router/app.go`（DI 改造）
3. 更新 `CLAUDE.md`（新的架构说明）
4. 端到端验证

**总计：~7 天**

---

## 14. 完整代码示例

### 14.1 应用入口

```go
// cmd/server/main.go
package main

import (
    "admin/internal/router"
    "os"
    "os/signal"
    "syscall"

    "github.com/rs/zerolog/log"
)

func main() {
    app := router.NewApp()

    go func() {
        if err := app.Run(); err != nil {
            log.Fatal().Err(err).Msg("服务启动失败")
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Info().Msg("正在关闭服务...")
    app.Close()
    log.Info().Msg("服务已关闭")
}
```

### 14.2 依赖注入

```go
// internal/router/app.go（关键变更部分）
type App struct {
    Config   *config.Config
    Router   *gin.Engine
    DB       *gorm.DB
    Redis    *xredis.Client
    JWT      *jwt.Manager
    RBAC     *rbac.PermissionCache  // ★ 新增
    Cron     *xcron.Manager
    Handlers *Handlers
}

func (a *App) initRBAC() {
    a.RBAC = rbac.NewPermissionCache(a.DB, 30*time.Second)
    // 启动时加载
    if err := a.RBAC.Refresh(context.Background()); err != nil {
        log.Warn().Err(err).Msg("权限缓存初始化失败")
    }
    // 定时刷新
    a.Cron.AddFunc("*/30 * * * * *", func() {
        if err := a.RBAC.Refresh(context.Background()); err != nil {
            log.Error().Err(err).Msg("权限缓存刷新失败")
        }
    })
}

func (a *App) initMiddleware() {
    // ...
    authorized := a.Router.Group("/api/v1")
    authorized.Use(
        middleware.AuthMiddleware(a.JWT),
        middleware.RBACMiddleware(a.RBAC),  // ★ 替代 CasbinMiddleware
        // TenantSkipMiddleware 已删除
        audit.AuditMiddleware(),
    )
}
```

### 14.3 JWT Claims 扩展

```go
// pkg/jwt/jwt.go — 新增 RoleIDs 字段
type Claims struct {
    TenantID   string   `json:"tenant_id"`
    TenantCode string   `json:"tenant_code"`
    UserID     string   `json:"user_id"`
    UserName   string   `json:"user_name"`
    Roles      []string `json:"roles"`       // 角色编码列表（保留，用于 HasRole 检查）
    RoleIDs    []string `json:"role_ids"`    // ★ 角色ID列表（新增，用于权限缓存查询）
    TokenID    string   `json:"token_id"`
    jwt.RegisteredClaims
}
```

Auth 中间件注入 context 时同时设置 Roles 和 RoleIDs。

---

> **总结**：新系统通过三个核心改进解决当前架构问题：
> 1. **显式 TenantScope** 替代隐式 Callback — AI 可见、无冲突
> 2. **纯数据库 RBAC** 替代 Casbin — 减少 40% 代码、支持 SQL JOIN
> 3. **保留 GORM + Gen** — 动态 WHERE、自动时间戳、自动软删除的优势不变
