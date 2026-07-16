# 05 - 定时任务调度器集成：main 与多组件生命周期编排

> 调研时间：2026-07-16  
> 解答问题：如何在 main 层优雅集成 HTTP Server + Cron Scheduler，动态任务管理放哪里，pkg/xcron 该如何设计。

---

## 一、当前现状与问题

**现状**：

```
backend/cmd/server/main.go
  └─ run()
       ├─ 第一层：基础设施（config/log/db/redis）
       └─ 第二层：server.New(opts) → srv.Run(ctx)
                    └─ errgroup
                         ├─ httpSrv.ListenAndServe()
                         └─ ctx.Done() → httpSrv.Shutdown()
```

server.go 里已有注释 `// Step 05 起追加：scheduler *scheduler.Scheduler`，预留了调度器位置。

**问题**：  
用户打算把 cron 加进来，问的是：

1. Server 内嵌 scheduler，还是和 Server 平级？
2. pkg/xcron 该封装到什么粒度？现在已有一个极简版。
3. 动态 Add/Remove 任务的 API 放哪里？
4. 任务注册（把哪些 job 加进去）应该在 main 还是别的地方？

---

## 二、成熟项目的组织方式

### 模式 A："All-in-Server" —— 调度器作为 Server 字段（etcd、wire 典型模式）

```
Server struct {
    httpSrv   *http.Server
    scheduler *xcron.Scheduler
}

func (s *Server) Run(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)
    g.Go(启动 HTTP)
    g.Go(启动 HTTP 优雅关闭)
    g.Go(启动 Scheduler)
    return g.Wait()
}
```

**适合**：scheduler 与 HTTP 业务强绑定，任务函数需要访问 server 持有的 db/rdb 等依赖。单入口，生命周期统一在一处。

### 模式 B："Multi-component" —— 组件平级，main 编排（Kubernetes、Grafana 模式）

```go
// main.go / run()
srv := server.New(opts)
scheduler := cron.New()
registerJobs(scheduler, deps)   // 集中注册任务

g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return srv.Run(ctx) })
g.Go(func() error { return scheduler.Start(ctx) })
_ = g.Wait()
```

**适合**：scheduler 生命周期和 HTTP 独立，或者后续要做多进程拆分（scheduler 可以单独运行为独立进程，只改 main 即可）。

### 模式 C：App struct 持有全部组件（最重，Spring/Fx 风格）

所有组件塞进一个 App struct，统一 Start/Stop 接口，生命周期由 App 管理。重，通常在 Go 里属于过度设计。

---

## 三、本项目推荐方案

### 结论

**推荐「模式 A + 轻量 jobs 注册包」组合**：scheduler 作为 Server 的字段，与 HTTP 在同一 errgroup 里；任务注册独立到 `cmd/server/jobs` 包。

理由：
- 当前是单体，scheduler 和 HTTP 共享 db/rdb 依赖，放进 Server 字段最自然
- 生命周期天然统一，不需要 main 做额外协调
- "后续动态 Add/Remove" 本质是通过 HTTP API 操作 scheduler 字段 —— 必须让 scheduler 对 handler 可见，放 Server 里最直接
- 模式 B 的最大优势（进程拆分）目前不需要，等真的要拆再重构，成本不高

---

## 四、分层设计

```
backend/
├── cmd/server/
│   ├── main.go             ← 只负责基础设施初始化 + run()
│   └── jobs/
│       └── register.go     ← 集中注册所有 cron 任务（初始静态任务）
│
├── internal/server/
│   └── server.go           ← Server struct 持有 httpSrv + scheduler；Run() 编排
│
└── pkg/xcron/
    └── xcron.go            ← 极简封装：New()/Start(ctx)，暴露原生 Add/Remove API
```

### 4.1 pkg/xcron — 极简封装（现有方案已经足够好）

当前 `pkg/xcron/xcron.go` 的设计已经非常正确：

```go
type Scheduler struct {
    *cron.Cron   // 嵌入，直接暴露 AddFunc / AddJob / Remove / Entries
}

func New() *Scheduler      // 带 slog 适配 + panic 恢复
func (s *Scheduler) Start(ctx context.Context) error  // 阻塞，ctx 取消时优雅停
```

**不需要再加 Manager 层（名字 map、running flag）**，原因：
- robfig/cron 内部的 `entries []Entry` 已经是任务列表，通过 `cron.EntryID` 管理
- 加一层 map 既增加锁竞争，又要同步两份状态（cron 内部的 + 外面的 map）
- 对外暴露 `EntryID`，让调用方（service 层）自己维护"名字 → EntryID"的映射

**xcron 需要补充的唯一方法**：运行时动态 Add 要保证调度器未停止时才能 Add（可选，现有实现已够用）。

### 4.2 internal/server — Server struct 加入 scheduler 字段

```go
type Server struct {
    httpSrv         *http.Server
    scheduler       *xcron.Scheduler   // ← 新增
    db              *gorm.DB
    rdb             *redis.Client
    gracefulTimeout time.Duration
}

type Options struct {
    Config    *config.Config
    DB        *gorm.DB
    RDB       *redis.Client
    Scheduler *xcron.Scheduler         // ← 新增（由外部创建后注入）
}
```

**为什么 scheduler 由外部创建后注入，而不是在 server.New 里创建？**
- 任务注册（`scheduler.AddFunc(...)`）需要在 scheduler 创建后、Server.Run 之前完成
- 如果在 server.New 里创建，外部没法在启动前注册任务
- 注入模式让 main 的顺序更清晰：`New Scheduler → Register Jobs → New Server → Run`

Run 方法新增调度器 goroutine：

```go
func (s *Server) Run(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)

    // 组件 1：HTTP server
    g.Go(func() error { ... })
    g.Go(func() error { /* 优雅关闭 */ })

    // 组件 2：Cron scheduler
    g.Go(func() error {
        return s.scheduler.Start(ctx)  // 阻塞，ctx 取消时自动停
    })

    return g.Wait()
}
```

### 4.3 cmd/server/jobs — 集中注册静态任务

```go
// cmd/server/jobs/register.go
package jobs

import (
    "admin/internal/service/xxx"
    "admin/pkg/xcron"
)

// Register 注册所有静态 cron 任务（启动时固定的）。
// 动态任务（运行时 Add/Remove）通过 HTTP API 操作 scheduler，不在这里。
func Register(s *xcron.Scheduler, deps Deps) {
    // 每日凌晨 2 点清理操作日志
    s.AddFunc("0 0 2 * * *", deps.OperationLogSvc.CleanOldLogs)

    // 每 5 分钟刷新权限缓存
    s.AddFunc("0 */5 * * * *", deps.RBACCache.Reload)

    // ... 其他固定任务
}

type Deps struct {
    OperationLogSvc *operationlog.Service
    RBACCache       *rbac.PermissionCache
    // 按需加依赖
}
```

**为什么放在 `cmd/server/jobs` 而不是 `internal/`？**
- 任务注册是"装配逻辑"，不是业务逻辑，属于 cmd 层职责
- 避免 internal 包之间循环依赖（jobs 需要引用多个 service）
- 后续拆微服务时，这个 cmd 目录可以独立成一个 cmd/scheduler 进程，只改 cmd，不动 internal

### 4.4 main.go 的最终形态

```go
func run() error {
    // 第一层：基础设施
    cfg := ...
    log := ...
    db  := ...
    rdb := ...

    // 第二层：长驻组件
    // 2a. 创建调度器（先于 Server，因为注册任务需要它）
    scheduler := xcron.New()

    // 2b. 注册静态任务（在 Start 之前）
    jobs.Register(scheduler, jobs.Deps{
        OperationLogSvc: operationlog.NewService(db),
        RBACCache:       rbac.NewPermissionCache(db),
    })

    // 2c. 组装 Server（注入 scheduler）
    srv, err := server.New(server.Options{
        Config:    cfg,
        DB:        db,
        RDB:       rdbClient,
        Scheduler: scheduler,
    })

    ctx, stop := signal.NotifyContext(...)
    defer stop()

    return srv.Run(ctx)  // 内部 errgroup 同时跑 HTTP + Scheduler
}
```

---

## 五、动态 Add/Remove 任务

用户提到后续要通过 HTTP API 动态添加/删除任务。

### 数据流

```
POST /api/cron/jobs  →  cron handler  →  cron service  →  xcron.Scheduler.AddFunc(...)
DELETE /api/cron/jobs/:name  →  ...  →  xcron.Scheduler.Remove(entryID)
```

### 关键：名字 → EntryID 的映射在哪里管？

robfig/cron 只认 `cron.EntryID`（int），对外无名字概念。需要维护一个 `name → EntryID` 映射。

**方案一（推荐）：在 cron service 里维护**

```go
// internal/service/cronjob/cronjob.go
type Service struct {
    scheduler *xcron.Scheduler
    entries   map[string]cron.EntryID  // 内存，进程重启后静态任务重新注册就行
    mu        sync.RWMutex
}

func (s *Service) AddJob(ctx context.Context, name, spec string, fn func()) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if _, exists := s.entries[name]; exists {
        return xerr.New(xerr.ErrConflict.Code, "任务已存在: "+name)
    }
    id, err := s.scheduler.AddFunc(spec, fn)
    if err != nil { return err }
    s.entries[name] = id
    return nil
}

func (s *Service) RemoveJob(ctx context.Context, name string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    id, exists := s.entries[name]
    if !exists {
        return xerr.New(xerr.ErrNotFound.Code, "任务不存在: "+name)
    }
    s.scheduler.Remove(id)
    delete(s.entries, name)
    return nil
}
```

**方案二（当前 backend-rbac 的 Manager 模式）**：把映射封装进 pkg/xcron 的 Manager。

两者区别：
| | 方案一（service 层管映射） | 方案二（pkg 层 Manager） |
|---|---|---|
| xcron 职责 | 纯调度引擎（薄） | 任务注册表 + 调度引擎 |
| 复用性 | xcron 可无修改跨项目 | Manager 含项目语义（中文 log/error） |
| 适合场景 | **本项目** | 需要多处独立任务表的场景 |

**推荐方案一**：xcron 保持极简（现有代码不改），业务语义的"名字管理"放 cron service 里。

---

## 六、整体目录结构

```
backend/
├── cmd/server/
│   ├── main.go
│   └── jobs/
│       └── register.go        ← 静态任务注册
│
├── internal/
│   ├── server/
│   │   └── server.go          ← Server struct 加 scheduler 字段 + Run 里注册 goroutine
│   ├── handler/cronjob/       ← HTTP 动态管理 API（后续加）
│   │   └── cronjob.go
│   └── service/cronjob/       ← 名字→EntryID 映射 + Add/Remove 业务逻辑（后续加）
│       └── cronjob.go
│
└── pkg/xcron/
    └── xcron.go               ← 极简封装，现有代码无需改动
```

---

## 七、用户原方案的评估

> 用户方案：pkg 下封装 cron，main 启动和 server 同级的 cron，动态添加/删除直接调用 pkg

**评估**：方向正确，但有一处值得调整：

| 问题 | 用户方案 | 建议方案 |
|---|---|---|
| scheduler 位置 | 和 server 平级（main 里各自 g.Go） | Server 的字段，由 server.Run 编排 |
| 动态 Add/Remove | 直接调 pkg/xcron | 通过 service/cronjob 间接调（维护名字映射） |
| 任务注册 | 未明确 | cmd/server/jobs/register.go 集中注册 |

**"server 同级"和"server 字段"的核心差别**：

```
# 同级方案（main 编排）
g.Go(srv.Run(ctx))
g.Go(scheduler.Start(ctx))

# 字段方案（server 编排）
srv.scheduler = scheduler
g.Go(srv.Run(ctx))  // 内部自己跑 scheduler
```

如果未来 scheduler 需要读/写 server 持有的内存状态（比如 PermissionCache 的刷新），字段方案天然可访问，同级方案需要额外传引用。本项目更推荐字段方案。

若用户更偏好同级方案（更像 Kubernetes 风格，独立进程感），也完全可行，只是主调代码在 main 里，可以接受。

---

## 八、结论与下一步

1. **pkg/xcron 不改**：现有极简封装已足够，不需要加 Manager。
2. **server.go**：加 `scheduler *xcron.Scheduler` 字段，Run 里新增 `g.Go(s.scheduler.Start(ctx))`。
3. **cmd/server/jobs/register.go**：新建，集中注册静态任务。
4. **main.go**：顺序改为 `New Scheduler → Register Jobs → New Server(注入) → Run`。
5. **动态管理（后续）**：新建 `internal/service/cronjob` + `internal/handler/cronjob`，service 层维护名字→EntryID 映射。

