# 多租户SaaS后端管理系统设计

## 1. 系统概述

基于Go语言的企业级多租户SaaS管理后端系统，采用清洁架构设计模式。

### 技术栈

| 组件 | 技术 |
|------|------|
| Web框架 | Gin v1.11.0 |
| ORM | GORM v1.31.1 + GORM Gen |
| 数据库 | PostgreSQL |
| 认证 | JWT 双令牌 + Redis 黑名单 |
| 权限 | Casbin RBAC（支持多租户域） |
| 缓存 | Redis |
| 日志 | Zerolog |
| 配置 | Viper |

### 架构层次

```
┌─────────────────────────────────────────────────────────────┐
│  外部接口层    HTTP API / CLI                               │
├─────────────────────────────────────────────────────────────┤
│  处理器层      Handlers / Middleware                        │
├─────────────────────────────────────────────────────────────┤
│  业务逻辑层    Services / Domain Models                     │
├─────────────────────────────────────────────────────────────┤
│  数据访问层    Repositories / Models / Queries              │
├─────────────────────────────────────────────────────────────┤
│  基础设施层    Database / Redis / Queue / Storage           │
└─────────────────────────────────────────────────────────────┘
```

## 2. 功能模块

| 模块 | 功能 |
|------|------|
| 租户管理 | 租户注册、信息管理、状态管理、套餐管理、数据隔离 |
| 用户管理 | 用户注册注销、信息管理、状态管理、批量导入导出 |
| 角色权限 | 角色定义、权限点管理、角色权限分配、用户角色绑定、数据权限 |
| 菜单管理 | 多租户菜单配置、菜单权限关联、动态菜单渲染 |
| 认证授权 | 多租户登录、JWT双令牌、SSO、第三方登录、密码策略 |
| 系统配置 | 系统参数、租户级配置、字典管理 |
| 日志审计 | 登录日志、操作日志、数据变更日志 |
| 通知中心 | 站内消息、邮件/短信通知、通知模板 |
| 定时任务 | 任务调度、执行监控、任务日志 |
| 文件管理 | 文件上传下载、租户隔离存储、文件分享 |
| 数据字典 | 字典类型管理、字典数据维护 |
| 监控告警 | 性能监控、异常告警、资源监控、租户用量统计 |
| API管理 | 接口管理、密钥管理、限流鉴权、API文档 |
| 备份恢复 | 数据备份恢复、租户数据迁移 |

## 3. 核心数据模型

### ER关系

```
tenants (租户)
  ├── users (用户) ──┬── user_tenant_role ──┬── roles (角色)
  │                  │                       │
  │                  └── user_departments ───└── departments (部门)
  │
  ├── permissions (权限)
  │
  └── roles ─── role_permissions ─── permissions

多对多关系:
- users ↔ roles (通过 user_tenant_role)
- roles ↔ permissions (通过 role_permissions)
- users ↔ departments (通过 user_departments)
```

### 关键设计

- **多租户隔离**: 所有业务表包含 `tenant_id` 字段
- **复合唯一约束**: `(tenant_id, business_key, deleted_at=0)`
- **软删除**: `deleted_at BIGINT DEFAULT 0`
- **时间戳**: `created_at`, `updated_at` 使用毫秒时间戳

## 4. 多租户隔离

### 数据隔离策略

```go
// 租户上下文
type TenantContext struct {
    TenantID   string
    TenantCode string
    UserID     string
    Roles      []string
}

// GORM Callbacks 自动注入租户过滤
db.Callback().Query().Before("gorm:query").Register("tenant_filter", func(db *gorm.DB) {
    tenantID := getTenantIDFromContext(db)
    if tenantID != "" {
        db.Statement.AddClause(clause.Where{
            Exprs: []clause.Expression{
                clause.Eq{Column: "tenant_id", Value: tenantID},
                clause.Eq{Column: "deleted_at", Value: 0},
            },
        })
    }
})
```

### 权限模型 (Casbin)

```conf
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _     # 用户-角色-租户 (user, role, domain)
g2 = _, _       # 角色继承 (子角色, 父角色)

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```

**说明**:
- `g(r.sub, p.sub, r.dom)`: 验证用户是否拥有某个角色（在同一租户下）
- `r.dom == p.dom`: 确保租户隔离
- `keyMatch2`: 支持路径通配符（如 `/api/v1/users/*`）
- `regexMatch`: 支持HTTP方法通配符（如 `GET|POST`）

## 5. 认证授权

### JWT双令牌

```
登录 → 生成 Access Token (1小时) + Refresh Token (7天)
     ↓
请求 → Bearer Access Token
     ↓
验证失败 → 使用 Refresh Token 刷新
     ↓
登出 → Access Token 加入 Redis 黑名单
```

### 中间件链

```
RequestID → Logger → Recovery → CORS → RateLimit
    → Tenant → Auth → Casbin → Handler
```

## 6. API设计

### 统一响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {...},
  "timestamp": 1640995200000,
  "request_id": "req_123"
}
```

### 分页响应

```json
{
  "code": 200,
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

## 7. 缓存策略

### 缓存键定义

```
user:{tenant_id}:{user_id}           # 用户信息
user:roles:{tenant_id}:{user_id}     # 用户角色
perm:{tenant_id}:{user_id}           # 用户权限
policy:{tenant_id}                   # 租户策略
```

### 多级缓存

```
L1: 本地内存 (5分钟 TTL)
  ↓ 未命中
L2: Redis (1小时 TTL)
  ↓ 未命中
DB: PostgreSQL
```

## 8. 安全设计

| 安全措施 | 实现方式 |
|---------|---------|
| 密码存储 | bcrypt + 随机salt |
| 敏感数据 | AES-256-GCM 加密 |
| API限流 | IP + 用户双重限流 |
| 输入验证 | SQL注入/XSS检测 |
| 审计日志 | 操作日志 + 数据变更记录 |

## 9. 项目目录结构

```
backend/
├── cmd/                 # 应用入口
│   ├── server/          # HTTP服务
│   ├── worker/          # 后台任务
│   └── cli/             # 命令行工具
├── internal/
│   ├── constants/       # 常量
│   ├── dal/             # 数据访问层
│   │   ├── model/       # GORM模型
│   │   └── query/       # GORM Gen查询
│   ├── dto/             # 数据传输对象
│   ├── handler/         # HTTP处理器
│   ├── middleware/      # 中间件
│   ├── repository/      # 仓库实现
│   ├── service/         # 业务逻辑
│   └── router/          # 路由
├── pkg/                 # 可复用包
│   ├── auth/            # 认证
│   ├── cache/           # 缓存
│   ├── config/          # 配置
│   ├── database/        # 数据库
│   ├── errors/          # 错误处理
│   └── logger/          # 日志
├── config/              # 配置文件
├── migrations/          # 数据库迁移
├── docs/                # 文档
└── scripts/             # 脚本
```

## 10. 部署架构

```
┌─────────────────────────────────────────────────────────────┐
│                       LoadBalancer                           │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────┴────────────────────────────────────────┐
│                    Kubernetes Cluster                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                   │
│  │  Pod 1   │  │  Pod 2   │  │  Pod 3   │                   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘                   │
│       │             │             │                          │
└───────┼─────────────┼─────────────┼──────────────────────────┘
        │             │             │
┌───────┴─────────────┴─────────────┴──────────────────────────┐
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │  PostgreSQL  │  │    Redis     │  │ Prometheus   │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```
