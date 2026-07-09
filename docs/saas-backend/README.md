# SaaS 后端：从零到生产的演进指南

> 技术栈：Go 1.22+ / Gin / GORM Gen / PostgreSQL / Redis / JWT / 纯数据库 RBAC

---

## 为什么重写

| 旧项目问题 | 新方案改进 |
|------------|-----------|
| pkg/config 全局单例 | `internal/config` 自包含单包(内联 viper 加载 + 显式 env 覆盖 + 手写 validate)，组合根映射 |
| 缺少优雅退出 | signal.NotifyContext + http.Server.Shutdown |
| Repository 和 Service 文件过大 | 域子包 + 方法单文件（create.go, update.go...） |
| 依赖关系隐式 | 构造函数显式注入，无 DI 框架 |
| 错误处理不统一 | xerr 集中定义 + middleware 统一捕获 |
| 缺乏验收自动化 | 每步附带可执行检查命令 |

---

## 核心设计原则

| # | 原则 | 说明 |
|---|------|------|
| 1 | 组合根在 main | `cmd/server/main.go` 是唯一了解所有依赖的地方 |
| 2 | 显式优于隐式 | 构造函数接收依赖，不用全局变量、不用 DI 框架 |
| 3 | 域子包 + 单方法文件 | `service/user/create.go` 一个方法一个文件，AI 友好 |
| 4 | 各 pkg 自有 Config | `database.Config`、`xslog.Config`，pkg 不知道 YAML |
| 5 | 多租户三层隔离 | JWT Claims → xcontext → Repository WHERE |
| 6 | 错误码集中管理 | xerr 定义错误 → Service 层 Wrap → middleware 统一响应 |
| 7 | 双排序字段 | 所有列表 `ORDER BY created_at DESC, pk DESC` |
| 8 | ID 全部 string | `idgen.NextID()` 雪花算法，不用 UUID 也不用自增 |
| 9 | 优雅退出 | signal + context 传播 + 资源清理 |
| 10 | 零全局状态 | 没有 init()、没有包级 var，所有状态通过参数传递 |

---

## 项目结构总览

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # 组合根：加载配置 → 创建依赖 → 启动服务 → 优雅退出
├── internal/
│   ├── config/                  # 项目配置层(类型化 Config;加载委托 pkg/xviper)
│   │   ├── config.go            # 类型化 Config 结构 + Load(调 xviper.Load)
│   │   └── validate.go          # 手写 validate,按业务块拆分子校验函数
│   ├── server/
│   │   └── server.go            # HTTP Server 封装（启动 + shutdown）
│   ├── router/
│   │   ├── router.go            # 路由注册入口 Setup()
│   │   └── handlers.go          # Handlers 聚合结构体
│   ├── middleware/
│   │   ├── request_id.go        # X-Request-ID
│   │   ├── logger.go            # 请求日志
│   │   ├── recovery.go          # panic 恢复
│   │   ├── cors.go              # CORS
│   │   ├── auth.go              # JWT 认证
│   │   ├── rbac.go              # 权限检查
│   │   └── tenant.go            # 租户上下文注入
│   ├── handler/{domain}/        # HTTP 处理器（按域分包）
│   │   ├── handler.go           # Handler struct + NewHandler()
│   │   ├── create.go            # Create 方法
│   │   ├── update.go            # Update 方法
│   │   ├── delete.go            # Delete 方法
│   │   ├── query.go             # List / Get 方法
│   │   └── dto.go               # 该域的 Request/Response DTO
│   ├── service/{domain}/        # 业务逻辑（按域分包）
│   │   ├── service.go           # Service struct + NewService()
│   │   ├── create.go
│   │   ├── update.go
│   │   ├── delete.go
│   │   ├── query.go
│   │   └── converter.go         # model ↔ DTO 转换
│   ├── repository/              # 数据仓储（集中式，不按域分包）
│   │   ├── user_repo.go
│   │   ├── role_repo.go
│   │   └── ...
│   ├── model/                   # GORM Gen 生成的 model（勿手动编辑）
│   ├── query/                   # GORM Gen 生成的 query（勿手动编辑）
│   └── rbac/
│       └── cache.go             # PermissionCache 内存缓存
├── pkg/                         # 可复用工具包（各包自有 Config struct）
│   ├── database/
│   │   ├── config.go            # database.Config struct
│   │   └── database.go          # New(cfg Config) → *gorm.DB
│   ├── xredis/
│   │   ├── config.go            # xredis.Config struct
│   │   └── xredis.go            # New(cfg Config) → *redis.Client
│   ├── xslog/
│   │   ├── config.go            # xslog.Config struct
│   │   └── logger.go            # New(cfg Config) → *slog.Logger（标准库结构化日志）
│   ├── jwt/
│   │   ├── config.go            # jwt.Config struct
│   │   ├── claims.go            # Claims 结构
│   │   └── manager.go           # Manager：签发/验证/吊销
│   ├── xcontext/                # 多租户认证上下文
│   │   ├── tenant.go
│   │   ├── user.go
│   │   ├── role.go
│   │   └── copy.go              # CopyContext
│   ├── xerr/                    # 业务错误码
│   │   ├── codes.go             # 错误码常量
│   │   └── error.go             # BizError struct + Wrap
│   ├── response/
│   │   └── response.go          # OK / Fail / Page 响应封装
│   ├── idgen/
│   │   └── idgen.go             # 雪花算法 ID 生成
│   └── password/
│       └── password.go          # bcrypt 加密/验证
├── config/                      # YAML 配置文件
│   └── config.yaml              # 默认配置(敏感字段用环境变量覆盖,不分 dev/prod)
├── migrations/                  # SQL 迁移文件
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
├── scripts/
│   ├── gen/main.go              # GORM Gen 生成入口
│   └── seed/main.go             # 种子数据
├── Makefile
├── Dockerfile
├── .air.toml                    # 热重载配置
└── go.mod
```

### 依赖流向（严格单向）

```
cmd/server/main.go
    ↓ 创建
pkg/* (database, xredis, xslog, jwt, idgen, password)
    ↓ 注入
internal/config      ← 加载 YAML
internal/server      ← 包装 http.Server
internal/router      ← 路由注册
internal/middleware  ← 中间件链
internal/handler/*   ← HTTP 层
internal/service/*   ← 业务层
internal/repository  ← 数据层
internal/model       ← GORM Gen
internal/query       ← GORM Gen
```

**禁止**：
- handler 不能直接调 repository
- service 之间不能互相调用（如果需要，提取到共享 service 或在 handler 层编排）
- pkg 之间不能互相 import
- internal 不能 import handler/service（只有上层 import 下层）

---

## 演进步骤总览

### Phase 1：骨架与基础设施（Step 01-03）

| 步骤 | 文档 | 内容 | 产出 |
|------|------|------|------|
| 01 | [项目骨架](step-01-project-scaffolding.md) | go mod、目录、Makefile、main.go | 能 `go build` |
| 02 | [基础设施](step-02-config-logger-db.md) | Config / Logger / Database / Redis | 能连接数据库与 Redis |
| 03 | [HTTP 框架](step-03-http-framework.md) | Gin 封装、中间件、response、xerr | curl health 返回 JSON |

**里程碑 1**：`make run` 启动，`curl /health` 返回 `{"code":0}`

### Phase 2：数据层与认证（Step 04-06）

| 步骤 | 文档 | 内容 | 产出 |
|------|------|------|------|
| 04 | [数据模型](step-04-database-schema.md) | 完整 SQL、GORM Gen、Repository 模板 | `make migrate && make gen` |
| 05 | [多租户与 JWT](step-05-multi-tenant-jwt.md) | xcontext / JWT / Auth 中间件 | 签发和验证 Token |
| 06 | [登录认证](step-06-auth-login.md) | 登录/登出/刷新 Token/租户切换 | 完整登录流程 |

**里程碑 2**：能登录、拿到 Token、访问受保护接口

### Phase 3：RBAC 权限（Step 07-08）

| 步骤 | 文档 | 内容 | 产出 |
|------|------|------|------|
| 07 | [RBAC 核心](step-07-rbac-permission-cache.md) | PermissionCache、角色继承、RBAC 中间件 | 权限检查生效 |
| 08 | [权限管理接口](step-08-permission-management.md) | 权限 CRUD、角色分配权限 | 能动态配置权限 |

**里程碑 3**：完整的认证 + 鉴权体系

### Phase 4：核心业务域（Step 09-13）

| 步骤 | 文档 | 内容 | 产出 |
|------|------|------|------|
| 09 | [租户管理](step-09-domain-tenant.md) | 租户 CRUD | 租户增删改查 |
| 10 | [用户管理](step-10-domain-user.md) | 用户 CRUD + 角色分配 | 用户增删改查 |
| 11 | [角色管理](step-11-domain-role.md) | 角色 CRUD + 权限分配 | 角色增删改查 |
| 12 | [部门管理](step-12-domain-department.md) | 部门树 CRUD | 树形部门 |
| 13 | [菜单管理](step-13-domain-menu.md) | 菜单树 + 按角色返回 | 动态菜单 |

**里程碑 4**：核心 CRUD 完成，前后端可对接

### Phase 5：高级功能（Step 14-17）

| 步骤 | 文档 | 内容 | 产出 |
|------|------|------|------|
| 14 | [数据权限](step-14-data-permission.md) | data_scope 五级 + 部门数据过滤 | 行级数据隔离 |
| 15 | [审计日志](step-15-audit-log.md) | 操作日志 + 登录日志 | 审计追踪 |
| 16 | [Swagger 文档](step-16-swagger.md) | swaggo 生成 API 文档 | 在线文档 |
| 17 | [部署与运维](step-17-deployment.md) | Dockerfile / docker-compose / 健康检查 | 容器化部署 |

> 配置路径约定、flag 去留、docker-compose 与 k3s ConfigMap 挂载的深入设计见 [research/config-loading/05-配置部署路径与ConfigMap挂载设计](research/config-loading/05-配置部署路径与ConfigMap挂载设计.md)
>
> **数据库系列文档**（选型 → 迁移 → 演进 → 落地）：
> - [02 数据库访问层选型调研](research/database/02-数据库访问层选型调研.md) — GORM+gen 方向、7 约束、sqlc 动态查询硬伤
> - [03 迁移工具与数据初始化方案](research/database/03-迁移工具与数据初始化方案.md) — golang-migrate vs Atlas、seed 三层分类
> - [04 schema优先与数据库优先](research/database/04-schema优先与数据库优先.md) — 两种真相源方向对比
> - [05 数据库演进与迁移规范](research/database/05-数据库演进与迁移规范.md) — expand-contract、schema/data 分离、安全 DDL
> - [06 GORM与golang-migrate最佳实践](research/database/06-GORM与golang-migrate最佳实践.md) — 避坑约定、单一 migrations/ 目录、data migration 落地（可执行铁律见 [`.claude/rules/migration-data-convention.md`](../../.claude/rules/migration-data-convention.md)）
>
> **日志系列文档与博客**：
> - [01 日志库选型调研](research/logging/01-日志库选型调研.md) — 为何 2026 选 slog（选型 ADR）
> - [06 OpenTelemetry 集成路径](research/logging/06-OpenTelemetry集成路径.md) — 演进指南：从单体日志到微服务全链路追踪
> - [博客：从零设计 Go 结构化日志](../blog/从零设计Go结构化日志-slog封装与Gin集成.md) — 完整封装设计、契约、踩坑清单（唯一权威实现文档）
>
> **Redis 系列文档**：
> - [01 Redis 客户端封装选型调研](research/redis/01-redis-客户端封装选型调研.md) — 为何 2026 选 go-redis/v9 + 具体 `*redis.Client`（不套接口、不做单例）、UniversalClient/rueidis 对比、集群迁移边界

**里程碑 5**：生产就绪

---

## 每步工作流程

```
1. 阅读当前步骤文档 → 理解目标和验收标准
2. 告诉 AI "请按 step-XX 文档实现" → AI 按文档写代码
3. 运行验收命令 → 通过则进入下一步
4. 如有问题 → 修正后重新验收
```

### 验收标准格式

每个步骤文档都包含：
- **目标**：这一步要达成什么
- **前置条件**：依赖哪些已完成的步骤
- **文件清单**：要创建/修改哪些文件
- **实现规范**：关键代码的设计决策和接口定义
- **验收标准**：可执行的检查命令（`go build`、`curl`、`make test`）
- **AI 协作提示**：复制给 AI 的指令模板

---

## 技术选型决策

| 领域 | 选择 | 原因 | 备选方案（不选） |
|------|------|------|----------------|
| 配置加载 | 泛型封装 `pkg/xviper`(viper + `ExperimentalBindStruct`)+ 项目层 `internal/config`(类型化 Config + 手写校验) | base+overlay 多环境、`APP_` 前缀环境变量覆盖嵌套字段、约定式极简 API | yaml.v3(功能弱)、纯 `AutomaticEnv`(与 `Unmarshal` 不兼容,靠 `ExperimentalBindStruct` 解决) |
| 日志 | slog（标准库）+ pkg/xslog 封装 | 零依赖、结构化、Handler 可换后端、otelslog 直通 OTel | zerolog（多一个依赖，slog 桥接慢 46×） |
| ORM | GORM + Gen | 类型安全查询、代码生成 | sqlx（手写 SQL 太多） |
| 路由 | Gin | 性能好、生态成熟、中间件丰富 | Echo（社区稍小） |
| ID 生成 | 雪花算法（sony/sonyflake） | 有序、紧凑、全局唯一 | UUID（太长、无序） |
| 密码 | bcrypt | 工业标准、自带盐 | argon2（overkill） |
| 缓存 | Redis + 内存 map | 热数据内存、持久化 Redis | 纯 Redis（延迟高） |
| Redis 客户端 | go-redis/v9 + pkg/xredis 封装（返回具体 `*redis.Client`） | 官方维护、最成熟、单节点直用不套接口不做单例 | UniversalClient 接口/单例（单节点用不上，见 [研究](research/redis/01-redis-客户端封装选型调研.md)）、rueidis（新依赖、API 陌生） |
| DI | 手动构造函数注入 | 显式、可追踪、无魔法 | Wire/dig（学习成本） |

---

## 与现有项目的关系

```
admin/
├── backend/              ← 新后端（本文档指导从零实现）
├── 老项目 A/         ← 旧后端（保留参考，不复制代码）
├── frontend-uiux/        ← 前端（共用，新后端保持 API 兼容）
└── docs/
    └── saas-backend/     ← 本文档系列
```

新后端的 API 路径和响应格式与旧后端保持一致，确保前端无需修改。

---

*最后更新：2026-06-18*
