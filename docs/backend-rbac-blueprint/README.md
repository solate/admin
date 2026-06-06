# backend 蓝图：多租户 SaaS 后端从零到一

> 技术栈：Go + Gin + GORM Gen + PostgreSQL + Redis + JWT + 纯数据库 RBAC

---

## 设计原则

| # | 原则 | 说明 |
|---|------|------|
| 1 | **域子包 + 方法单文件** | `service/user/create.go` 一个方法一个文件，代码量少，AI 友好 |
| 2 | **构造函数直接注入** | 不用 wire/dig，`NewHandler(db, svc)` 显式传参 |
| 3 | **Repository 集中式** | `repository/` 扁平结构，GORM Gen 生成 query，不按域拆子包 |
| 4 | **多租户三层隔离** | JWT Claims → xcontext → Repository `WHERE tenant_id = ?` |
| 5 | **超管绕行一行** | `if HasRole("super_admin") { c.Next(); return }` |
| 6 | **缓存事件驱动** | PermissionCache + `NotifyRefresh()`，权限变更即时生效 |
| 7 | **双排序字段** | 所有列表 `ORDER BY created_at DESC, id DESC`，确保确定性 |
| 8 | **ID 全部 string** | `idgen.GenerateUUID()`，不用 int64 自增 |
| 9 | **错误码集中管理** | `xerr` 包统一定义，Service 层用 `xerr.Wrap` 包装 |
| 10 | **一行日志** | zerolog 链式调用写在一行，禁止多行 |

---

## 架构总览

```
backend/
├── cmd/server/main.go              # 入口
├── internal/
│   ├── handler/{domain}/           # HTTP 处理器（域子包）
│   ├── service/{domain}/           # 业务逻辑（域子包）
│   ├── repository/                 # 数据仓储（集中式，GORM Gen）
│   ├── router/                     # 路由 + App 初始化
│   ├── middleware/                  # 中间件链
│   ├── dto/                        # 数据传输对象
│   ├── dal/model/                  # GORM Gen 生成（勿手动编辑）
│   ├── dal/query/                  # GORM Gen 生成（勿手动编辑）
│   └── rbac/                       # RBAC 权限缓存
├── pkg/
│   ├── config/                     # 配置加载
│   ├── database/                   # 数据库连接
│   ├── jwt/                        # JWT 工具
│   ├── xcontext/                   # 多租户认证上下文
│   ├── xerr/                       # 业务错误码
│   ├── response/                   # HTTP 响应封装
│   ├── idgen/                      # ID 生成器
│   ├── password/                   # 密码工具
│   └── logger/                     # 日志工具
├── migrations/                     # SQL 迁移文件（根目录）
├── scripts/                       # 种子数据、生成脚本
├── docs/                           # 文档
├── Makefile
└── go.mod
```

### 依赖链

```
main.go → router.Setup() → handler/{domain} → service/{domain} → repository
                ↓
         middleware chain:
         RequestID → Logger → Recovery → CORS → JWTAuth → RBAC → Audit
```

### 数据模型关系

```
tenants (1)──(N) users (N)──(N) roles (N)──(N) permissions
                  │              │
                  └──user_roles──┘
                     (tenant_id)
                  └──role_permissions──┘
                     (tenant_id)

全局表（无 tenant_id）：permissions, menus
租户表（有 tenant_id）：users, roles, departments, positions, ...
```

---

## 步骤索引

### 基础设施（Step 01-07）

| 步骤 | 文档 | 内容 | 预估时间 |
|------|------|------|----------|
| 01 | [step-01-project-scaffolding.md](step-01-project-scaffolding.md) | 项目骨架、go mod、Makefile、目录结构 | 30 分钟 |
| 02 | [step-02-infrastructure.md](step-02-infrastructure.md) | Config、Logger、Database、Redis 连接 | 2 小时 |
| 03 | [step-03-gin-router-middleware.md](step-03-gin-router-middleware.md) | Gin 封装、中间件链、错误处理 | 2 小时 |
| 04 | [step-04-database-model.md](step-04-database-model.md) | GORM Gen、完整数据模型、迁移脚本 | 3 小时 |
| 05 | [step-05-multi-tenant-foundation.md](step-05-multi-tenant-foundation.md) | xcontext、JWT Claims、租户隔离 | 2 小时 |
| 06 | [step-06-auth.md](step-06-auth.md) | 登录、JWT、Token 刷新、租户切换 | 3 小时 |
| 07 | [step-07-rbac-core.md](step-07-rbac-core.md) | PermissionCache、RBAC 中间件、角色继承 | 3 小时 |

**里程碑 1**：基础设施完成，能启动服务器、登录、鉴权 ✅

### 域业务（Step 08-11）

| 步骤 | 文档 | 内容 | 预估时间 |
|------|------|------|----------|
| 08 | [step-08-domain-tenant.md](step-08-domain-tenant.md) | 租户管理 CRUD | 2 小时 |
| 09 | [step-09-domain-user.md](step-09-domain-user.md) | 用户管理 CRUD | 3 小时 |
| 10 | [step-10-domain-role.md](step-10-domain-role.md) | 角色 + 权限分配 | 3 小时 |
| 11 | [step-11-domain-department.md](step-11-domain-department.md) | 部门管理（树形） | 2 小时 |

**里程碑 2**：核心 CRUD 完成，前后端能对接 ✅

### 高级功能（Step 12-14）

| 步骤 | 文档 | 内容 | 预估时间 |
|------|------|------|----------|
| 12 | [step-12-data-permission.md](step-12-data-permission.md) | 数据权限（data_scope 五级） | 3 小时 |
| 13 | [step-13-audit.md](step-13-audit.md) | 操作审计、登录日志 | 2 小时 |
| 14 | [step-14-swagger-test.md](step-14-swagger-test.md) | Swagger 文档、集成测试 | 2 小时 |

**里程碑 3**：生产就绪 ✅

---

## 实施流程

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Step 01-04  │────▶│   Step 05-07  │────▶│  Step 08-11   │────▶│  Step 12-14  │
│  基础骨架     │     │  认证+RBAC    │     │  业务 CRUD    │     │  高级功能     │
│              │     │              │     │              │     │              │
│  能编译运行   │     │  能登录鉴权   │     │  前后端对接   │     │  生产就绪     │
└─────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
      ↑                    ↑                    ↑                    ↑
  里程碑 0             里程碑 1             里程碑 2             里程碑 3
```

### 每一步的工作流程

```
1. 阅读当前步骤的文档
2. 理解目标和验收标准
3. 让 AI 按文档实现
4. 运行验收标准中的检查命令
5. 如果有问题，修正后重新验收
6. 通过后进入下一步
```

---

## 与现有 backend 的关系

```
admin/
├── backend/                  ← 新后端（从零实现）
├── backend-rbac/             ← 旧后端（保留参考）
├── frontend/                 ← 前端（共用）
└── docs/
    ├── rbac-analysis/        ← RBAC 分析文档（设计依据）
    └── backend-rbac-blueprint/  ← 本蓝图
```

新项目可以参考旧项目的：
- 数据库 schema（`scripts/dev_schema.sql`）
- 种子数据（`scripts/init_data/`）
- GORM Gen 使用方式
- 前端 API 对接格式

但 **不复制代码**，按蓝图从零写。

---

*最后更新：2026-06-06*
