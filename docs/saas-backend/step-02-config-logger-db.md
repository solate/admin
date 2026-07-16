# Step 02: Config / Logger / Database / Redis

## 文件清单

```
internal/config/
├── config.go          # 类型化 Config + InitConfig(调 xviper.Load)
└── validate.go        # 手写白名单校验

pkg/xviper/
└── xviper.go          # Load[T] + settings/Option/WithPath/WithValidate

pkg/xslog/
├── logger.go          # Config{} + New(cfg) *slog.Logger + parseLevel + shortenSource
├── handler.go         # gormSlogger，实现 gormlogger.Interface
├── context.go         # contextHandler：从 ctx 自动注入 request_id/tenant_id
└── attrs.go           # Err / RedactReplaceAttr

pkg/xgorm/
└── xgorm.go           # Config{} + New(cfg, log) (*gorm.DB, error) + Close

pkg/xredis/
└── xredis.go          # Config{} + New(cfg) (*redis.Client, error)

internal/server/
└── server.go          # Options{DB, RDB, Log} → New(opts) → http.Server

config/
├── config.yaml        # base（含 server/database/redis/jwt/log 完整默认值）
└── config.dev.yaml    # dev overlay（约定式，当前为空）

cmd/server/main.go     # 组合根
```

> Config/Options 类型与实现函数**同文件**（无 config.go 拆分）。

## 架构决策

### Config 分两层
- `pkg/xviper`：通用加载机制（YAML base + overlay 合并 + `APP_*` 环境变量），泛型 `Load[T]`，可跨项目复制
- `internal/config`：项目专属 `Config` 结构体 + 手写校验，`InitConfig()` 调 `xviper.Load[Config]`
- 各 `pkg` 不知道 YAML；`internal/config.Config` → `pkg.Config` 的字段映射在 `main.go` 完成

### Logger
- `xslog.New(cfg) *slog.Logger`，依赖注入不设全局（第三方库需要时才 `slog.SetDefault`）
- context 字段（`request_id`/`tenant_id`）由 `contextHandler` 自动注入，业务代码零负担

### Database / Redis
- 返回原生 `*gorm.DB` / `*redis.Client`，**不套接口、不做单例**
- `xredis`：连接池/超时参数仅在 `>0` 时覆盖，为 0 保留 go-redis 默认值
- GORM SQL 日志通过 `gormSlogger` 适配到 `slog`，`Trace` 透传 ctx（自动带 request_id）

### Server
- `Server` 从 Step 02 起通过 `Options` 持有 `db/rdb/log`，Step 03 只换内部 `mux → gin`，main.go 不动

### main.go 组合根
顺序：`InitConfig` → `xslog.New` → `xgorm.New` → `xredis.New` → `server.New` → 启动 + 优雅关闭

- `run() error` 模式：资源用 `defer` 清理，失败 `return fmt.Errorf`，`main` 统一打日志 + `os.Exit(1)`
- 优雅退出超时取自 `cfg.Server.GracefulTimeout`（`time.Duration`，YAML 写 `30s` 自动解析）
- **敏感字段（密码/密钥）只走环境变量**，不进任何 yaml

### 多环境合并语义（xviper）
- 嵌套 map：深合并；标量/slice：overlay **整体替换** base（slice 不追加）
- `APP_ENV` 设了但 overlay 文件不存在 → **fail-fast**（不静默降回 base）
