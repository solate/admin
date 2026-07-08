# Step 02: 基础设施 — Config / Logger / Database / Redis

## 目标

实现配置加载、结构化日志、PostgreSQL 连接、Redis 连接。所有基础设施通过构造函数创建，main.go 做组装，并把已就绪的 db/redis/logger 直接接入 server。

> **本步已完成**：下述四层基础设施均已落地。本文档只保留**目标 + 设计决策 + 源码链接 + 验收**，实现细节直接看源码与设计文档，不再内联大段代码。

## 前置条件

- Step 01 完成，`go build ./...` 通过
- PostgreSQL 可访问（本地或 Docker）
- Redis 可访问（本地或 Docker）

## 文件清单

```
internal/config/                  # 自包含配置层(项目专属 Config + 校验)
├── config.go                     # 类型化 Config + InitConfig(调 xviper.Load)
├── validate.go                   # 手写白名单校验(拆分独立文件)
└── config_test.go                # 加载器行为 + 项目集成测试

pkg/xviper/                       # 通用配置加载机制(泛型 Load[T],可复制复用)
├── xviper.go                     # Load[T]:读 base + overlay 合并 + env 覆盖 + validate
├── options.go                    # WithPath / WithValidate 选项
└── xviper_test.go                # 自包含单测(t.TempDir/t.Setenv)

pkg/xslog/                         # 结构化日志封装(slog,可复制复用,含 ctx 字段自动注入)
├── config.go                     # xslog.Config{Level, Format, Output, AddSource, ...}
├── logger.go                     # New(cfg) *slog.Logger + parseLevel + shortenSource
├── handler.go / context.go       # contextHandler:从 ctx 自动注入 request_id/tenant_id
├── attrs.go                      # Err / RedactReplaceAttr 辅助
└── logger_test.go                # 自包含单测(buffer 注入,slogtest 合规)

pkg/database/
├── config.go                     # database.Config
└── database.go                   # New(cfg, log) (*gorm.DB, error) + slog GORM 适配器

pkg/xredis/
├── config.go                     # xredis.Config
└── xredis.go                     # New(cfg Config) (*redis.Client, error)

internal/server/
└── server.go                     # Server 接收 Options，持有 db/rdb/log(*slog.Logger)

config/
├── config.yaml                   # base：完整默认配置（即 dev 默认值）
└── config.dev.yaml               # dev overlay(约定式,当前为空注释)

cmd/server/main.go                # 组合根：配置 → 日志 → DB → Redis → server
```

> **多环境（base + overlay）**：base `config.yaml` 承载共享默认值（即 dev 配置）；每个环境的差异放进 `config.{env}.yaml` overlay，由 `APP_ENV` 触发合并。**敏感字段（密码、密钥）始终走环境变量覆盖**，不进任何配置文件——overlay 只放非密钥的环境差异（mode / log 级别与格式 / 域名 / 库名等）。

## 实现规范

### 1. Config — 自包含配置层 + 通用加载机制

**拆两层**：通用加载机制（读 base + overlay 合并 + 环境变量覆盖）抽成泛型包 `pkg/xviper`（`Load[T]`，可复制到其他项目）；项目专属的 `Config` 结构与校验规则放 `internal/config`（复制后改字段）。

- **通用机制**（不随项目变）：[pkg/xviper/xviper.go](../../backend/pkg/xviper/xviper.go) — `Load[T any](opts...) (*T, error)`，约定 `config/config.yaml` base + `config.{APP_ENV}.yaml` overlay，`APP_*` 环境变量覆盖嵌套字段。选项见 [options.go](../../backend/pkg/xviper/options.go)。
- **项目专属**（新微服务只改这里）：[internal/config/config.go](../../backend/internal/config/config.go) 定义类型化 `Config` + `InitConfig()`（调 `xviper.Load[Config]`）；[internal/config/validate.go](../../backend/internal/config/validate.go) 手写白名单校验。

设计细节（为何 xviper 泛型、overlay 合并语义、为何不用 `AutomaticEnv`）见配置设计文档：
- 设计蓝图：[research/config-loading/04-三层架构与动态多租户配置蓝图.md](research/config-loading/04-三层架构与动态多租户配置蓝图.md)
- 部署路径：[research/config-loading/05-配置部署路径与ConfigMap挂载设计.md](research/config-loading/05-配置部署路径与ConfigMap挂载设计.md)
- 通俗讲解：[../blog/从零设计Go配置加载-viper封装与约定式设计.md](../blog/从零设计Go配置加载-viper封装与约定式设计.md)

**关键设计决策**：
- **通用加载与项目 Config 分层**：`xviper`（机制）复制即用、`internal/config`（结构）复制后改字段，关注点分离。
- **校验手写全覆盖**：反序列化交给 viper，字段校验在 `validate.go` 手写，规则集中一处、可跳转可调试。
- **不用全局单例**：`InitConfig` 返回 `*Config`，通过参数向下传递。
- **敏感字段走 env**：密码/密钥不进任何 yaml，只经环境变量覆盖。

### 2. Logger — slog + pkg/xslog

日志用 Go 标准库 `log/slog`，封装在 [pkg/xslog/](../../backend/pkg/xslog/)，通过自定义 `slog.Handler` 实现请求级字段（`request_id`/`tenant_id`）自动注入。选型理由（为何 2026 新项目用 slog 而非 zerolog）见 [research/logging/01-日志库选型调研.md](research/logging/01-日志库选型调研.md)；完整的封装设计、逐行原理、踩坑与决策推演见博客 [从零设计Go结构化日志-slog封装与context注入实战](../blog/从零设计Go结构化日志-slog封装与Gin集成.md)（唯一权威实现文档）。

源码：
- [pkg/xslog/logger.go](../../backend/pkg/xslog/logger.go) — `New(cfg) *slog.Logger`；`parseLevel`（复用 `slog.Level.UnmarshalText`，大小写不敏感）；`shortenSource` 把调用位置裁成 `dir/file:line`。
- [pkg/xslog/handler.go](../../backend/pkg/xslog/handler.go)、[context.go](../../backend/pkg/xslog/context.go) — `contextHandler` 从 ctx 自动注入字段；`WithField`/`WithFields`/`ContextExtractor` 扩展点。

**关键设计决策**：
- **ctx 字段自动注入**：自定义 `contextHandler` 包裹底层 handler，在每条日志写入前从 ctx 提取 `request_id`/`tenant_id` 注入，业务代码零负担（用 `*Context` 变体即可）。这是本包的核心价值。
- **依赖倒置解耦**：日志包只认 `ContextExtractor` 函数签名，不 import 业务包（xcontext）；由 main 组装。接 OTel 只需再注册一个 extractor。
- **不提供 Fatal**：对齐 slog 官方哲学（日志只记录、不控制流程）；启动失败用 `run() error` 模式，`main` 统一 `os.Exit`，保证 defer 执行。
- **依赖注入，不设全局**：`New` 返回 `*slog.Logger` 经参数传递（第三方库需要时才 `slog.SetDefault`）。
- **格式**：`json`（生产，也是空值默认）走 `JSONHandler`，`text`（开发）走 `TextHandler`；输出默认 stdout（`Output` 可注入 buffer 供测试），容器/systemd 负责轮转。
- **命名 `xslog`**：可复用封装包用 `x` 前缀，避免与标准库 `log`/`slog` 及项目内包冲突（对齐 `xviper`/`xcontext`/`xerr`）。

### 3. Database — PostgreSQL + GORM

源码：[pkg/database/database.go](../../backend/pkg/database/database.go)、[config.go](../../backend/pkg/database/config.go)。

- `New(cfg Config, log *slog.Logger) (*gorm.DB, error)`：拼 DSN、`gorm.Open(postgres)`、设连接池（MaxIdle/MaxOpen/ConnMaxLifetime）。`Close(db)` 优雅关闭。
- **GORM → slog 适配器**：`gormSlogger` 实现 `gormlogger.Interface`（LogMode/Info/Warn/Error/Trace），把 SQL 执行日志统一接到 slog。`Trace` 用 `*Context` 变体（`LogAttrs`/`InfoContext` 等）把 ctx 透传给 slog，`contextHandler` 会自动带上 ctx 里的 request_id；正常走 Debug、出错走 Error，带 `elapsed`/`rows`/`sql` 字段。

### 4. Redis

源码：[pkg/xredis/xredis.go](../../backend/pkg/xredis/xredis.go)、[config.go](../../backend/pkg/xredis/config.go)。

- `New(cfg Config) (*redis.Client, error)`：`redis.NewClient` + 启动时 `Ping` 探活，失败即 `Close` 并返回错误（fail-fast）。
- 连接池/超时参数（PoolSize/MinIdleConns/MaxRetries/Dial·Read·WriteTimeout）**仅在 >0 时覆盖**，为 0 保留 go-redis 默认值——yaml 不填即用官方默认，无需背默认数字。

**关键设计决策**（选型调研见 [research/redis/01-redis-客户端封装选型调研.md](research/redis/01-redis-客户端封装选型调研.md)）：
- **返回具体 `*redis.Client`，不套接口、不做单例**：与 `pkg/database` 返回原生 `*gorm.DB` 风格一致。参照的原项目 `content-center-backend` 单节点却返回 `redis.UniversalClient` 接口 + `sync.Once` 全局单例，接口没隐藏底层库（消费方仍要 import go-redis 用 `redis.Nil`）、单例又挡住并行测试与多实例——这是被刻意否掉的两点。
- **单节点优先，集群是后话**：admin 后台 95% 是单 key 操作。后期若上集群，只需把 `New` 内部 `redis.NewClient` 换成 `redis.NewUniversalClient` 并调整返回类型，改动收敛在这一处封装（业务调用点因命令 API 相同而基本不动）。
- **命名 `xredis`**：可复用封装包用 `x` 前缀（对齐 `xviper`/`xslog`）。

### 5. Server 接入基础设施

源码：[internal/server/server.go](../../backend/internal/server/server.go)。

`Server` 从 Step 02 起通过 `Options` 持有 `db/rdb/log`，`New(opts Options) (*Server, error)` 组装 `http.Server`（当前是 `/health` stub，Step 03 起替换为 `gin.New()` + `router.Setup()`）。

**设计决策：为什么 Step 02 就把 db/redis/logger 接入 server**

延续 Step 01 的核心保证——**main.go 是最终结构，永不改动**。基础设施在 Step 02 就绪后直接通过 `Options` 传入 server，于是：

- Step 03 只需把 `New()` 内部的 `mux` 换成 `gin.New()` + `router.Setup()`，server 签名和 main.go 都不动。
- Step 05 cron 同理，只在 `Start()`/`Stop()` 内部追加，main.go 无感知。

若把 server 的接入推迟到 Step 03，main.go 就要在 Step 03 再改一次（从「不启动 server」变成「启动 server」），破坏稳定性保证。

### 6. main.go — 组合根

源码：[cmd/server/main.go](../../backend/cmd/server/main.go)。

顺序：`config.InitConfig()` → `xslog.New()` → `database.New()` → `xredis.New()` → `server.New()` → 启动 + `signal.NotifyContext` 优雅关闭。

**关键设计**：
- `config.Config` → 各 `pkg.Config` 的映射在 main 中完成；各 pkg 不知道 YAML 的存在，只接收自己的 Config struct —— 换项目只需换 YAML 和 main 的映射。
- `main` 采用 `run() error` 模式：所有资源用 `defer` 清理，任何一步失败 `return fmt.Errorf`，由 `main` 统一打日志 + `os.Exit(1)`。不用 `xslog.Fatal`（`os.Exit` 会跳过 defer），日志包只记录不控制流程。
- 优雅退出超时取自 `cfg.Server.GracefulTimeout`（不硬编码）。

### 7. config/config.yaml

源码：[config/config.yaml](../../backend/config/config.yaml)（base，含 server/database/redis/jwt/log 完整默认值）、[config/config.dev.yaml](../../backend/config/config.dev.yaml)（dev overlay，约定式占位）。

日志段（base）：

```yaml
log:
  level: debug
  format: text
  add_source: true
```

**多环境合并语义**（viper `MergeInConfig`）：嵌套 map **深合并**；标量与 slice 用 overlay 值**整体替换** base（slice 不追加）。`APP_ENV` 设了但 overlay 缺失 → **fail-fast**。密钥仍走环境变量，不写进任何 overlay。详见配置设计文档（§1 链接）。

## 依赖引入

```bash
go get github.com/spf13/viper        # v1.21.0（YAML 解析由其间接依赖 go.yaml.in/yaml/v3 提供）
go get gorm.io/gorm                  # v1.31.1
go get gorm.io/driver/postgres       # v1.6.0
go get github.com/redis/go-redis/v9  # v9.21.0（当前最新，drop-in、无 breaking change）
# 日志用标准库 log/slog，无需第三方依赖
```

## 验收标准

```bash
cd backend

# 1. 全包编译通过
go build ./...
# 期望：无错误

# 2. 各 pkg 独立可编译
go build ./pkg/database && go build ./pkg/xslog && go build ./pkg/xredis && go build ./pkg/xviper && go build ./internal/config
# 期望：各自独立编译无错误

# 3. 配置层 + 日志层单测全绿（无外部服务依赖）
go test ./internal/config/ ./pkg/xviper/ ./pkg/xslog/ -v
# 期望：加载器行为 / env 覆盖 / overlay 合并 / 日志级别过滤 / JSON / source 裁剪 / ctx 注入 / 脱敏 / slogtest 合规 全部 PASS

# 4. 配置加载 + 基础设施初始化（需 PG / Redis 在跑）
go run ./cmd/server &
sleep 1
curl -s http://localhost:8080/health
# 期望：{"status":"ok"}
#       且日志输出 "all infrastructure initialized"
kill %1
# 若 PG / Redis 未启动，期望看到明确的 "connect database failed" / "connect redis failed"

# 5. 环境变量覆盖敏感字段
DB_PASSWORD=wrong go run ./cmd/server 2>&1 | grep "connect database"
# 期望：连接失败（密码被覆盖）

# 6. 日志格式（debug + text 模式）
go run ./cmd/server 2>&1 | head -3
# 期望：TextHandler 的 key=value 输出，含 source=dir/file:line
```

## AI 协作提示

```
请按 step-02-config-logger-db.md 实现基础设施层。

要点：
1. pkg/xviper 泛型 Load[T]：读 base + config.{APP_ENV}.yaml overlay 合并 + APP_* 环境变量覆盖；WithPath/WithValidate 选项；绝不用 viper.AutomaticEnv
2. internal/config：类型化 Config(mapstructure tag) + InitConfig(调 xviper.Load[Config]) + validate.go 手写白名单校验
3. 自包含单测(t.TempDir/t.Setenv)：xviper 加载器行为 + internal/config 校验/env 覆盖
4. pkg/xslog：Config{Level,Format,AddSource,Output,ContextExtractors,ReplaceAttr}，New() 返回 *slog.Logger；parseLevel 用 slog.Level.UnmarshalText(大小写不敏感)；shortenSource 裁 dir/file:line；自定义 contextHandler 从 ctx 注入 request_id/tenant_id；不提供 Fatal(启动失败走 run() error + os.Exit)；日志用标准库 slog，无第三方依赖
5. pkg/database：New() 返回 *gorm.DB，含 slog 适配的 GORM logger(gormSlogger 实现 gormlogger.Interface，用 *Context 变体携带 request_id)
6. pkg/xredis：New() 返回具体 *redis.Client（不套接口、不做单例），启动 Ping 探活；连接池/超时参数仅在 >0 时覆盖，否则用 go-redis 默认
7. cmd/server/main.go 用 run() error 模式：config.InitConfig()，做 config.Config → 各 pkg.Config 映射；关闭超时用 cfg.Server.GracefulTimeout；任何一步失败 return error，由 main 统一打日志 + os.Exit(1)
8. 不要用全局变量，所有依赖通过参数传递
9. server 通过 Options 接收 db/rdb/log(*slog.Logger)，main.go 启动 server
```

---

*上一步：[Step 01 - 项目骨架](step-01-project-scaffolding.md) | 下一步：[Step 03 - HTTP 框架](step-03-http-framework.md)*
