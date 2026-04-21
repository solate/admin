# xcron 定时任务管理包

基于 [github.com/robfig/cron/v3](https://github.com/robfig/cron) 封装的**轻量级定时任务调度引擎**。

## 设计理念

- **职责单一**：仅负责任务调度，不包含业务逻辑
- **简单易用**：API 简洁直观，快速上手
- **并发安全**：内置读写锁，保证并发操作安全
- **轻量高效**：核心代码约 240 行，无冗余功能

## 特性

- ✅ 简单的 API：Add、Remove、Run
- ✅ 秒级 cron 支持
- ✅ 并发安全
- ✅ Panic 自动恢复
- ✅ 任务查询：List、Exists、GetSpec、Count

## 快速开始

### 1. 初始化并启动

```go
import "admin/pkg/xcron"

// 初始化
manager, err := xcron.Init(xcron.Config{
    WithSeconds: true, // 支持秒级cron
})
if err != nil {
    log.Fatal().Err(err).Msg("初始化定时任务失败")
}

// 启动管理器
manager.Start()
defer manager.Stop()
```

### 2. 添加定时任务

```go
// 每5秒执行一次
err := manager.Add("cleanup", "*/5 * * * * ?", func() {
    log.Info().Msg("清理临时文件...")
    // 你的任务逻辑
})

// 每天凌晨2点执行
err := manager.Add("daily_backup", "0 0 2 * * ?", func() {
    log.Info().Msg("开始数据备份...")
    // 备份逻辑
})
```

### 3. 任务管理

```go
// 检查任务是否存在
if manager.Exists("cleanup") {
    // 获取任务spec
    spec, _ := manager.GetSpec("cleanup")
    fmt.Printf("任务规则: %s\n", spec)
}

// 列出所有任务
jobs := manager.List()
fmt.Printf("当前任务: %v\n", jobs)

// 手动执行任务
manager.Run("daily_backup")

// 移除任务
manager.Remove("cleanup")

// 获取任务数量
count := manager.Count()
fmt.Printf("任务总数: %d\n", count)
```

## Cron 表达式格式

支持秒级 cron（默认启用），格式如下：

```
┌───────────── 秒 (0 - 59)
│ ┌─────────── 分钟 (0 - 59)
│ │ ┌───────── 小时 (0 - 23)
│ │ │ ┌─────── 日期 (1 - 31)
│ │ │ │ ┌───── 月份 (1 - 12)
│ │ │ │ │ ┌─── 星期 (0 - 6, 0=周日)
│ │ │ │ │ │
* * * * * *
```

### 常用示例

| 表达式 | 说明 |
|--------|------|
| `* * * * * ?` | 每秒执行 |
| `*/5 * * * * ?` | 每5秒执行 |
| `0 */1 * * * ?` | 每1分钟执行 |
| `0 0 * * * ?` | 每小时执行 |
| `0 0 2 * * ?` | 每天凌晨2点 |
| `0 0 2 * * MON` | 每周一凌晨2点 |
| `0 0 2 1 * ?` | 每月1号凌晨2点 |
| `0 0/30 8-10 * * ?` | 每天8-10点，每半小时 |

## API 文档

### Manager

管理器核心类型，提供所有调度功能。

```go
type Manager struct {
    // 内部字段
}
```

### Config

初始化配置。

```go
type Config struct {
    WithSeconds bool         // 是否支持秒级cron
    Logger      cron.Logger  // 可选的日志记录器
}
```

### 方法

| 方法 | 说明 |
|------|------|
| `Init(cfg Config) (*Manager, error)` | 初始化全局管理器 |
| `Get() *Manager` | 获取全局管理器实例 |
| `Start()` | 启动管理器 |
| `Stop()` | 停止管理器（等待任务完成） |
| `IsRunning() bool` | 检查运行状态 |
| `Add(name, spec string, fn func()) error` | 添加任务 |
| `Remove(name string) error` | 移除任务 |
| `Run(name string) error` | 立即执行任务 |
| `List() []string` | 列出所有任务名称 |
| `GetSpec(name string) (string, error)` | 获取任务cron表达式 |
| `Exists(name string) bool` | 检查任务是否存在 |
| `Count() int` | 获取任务数量 |
| `Clear()` | 清空所有任务 |

## 与业务层集成

`pkg/xcron` 只负责调度，业务逻辑应该在 service 层实现：

```go
// internal/service/cron_service.go
type CronService struct {
    cronMgr *xcron.Manager
    repo    repository.CronJobRepository
}

// 业务层的任务定义
type CronJob struct {
    ID          int64
    Name        string
    Spec        string
    Description string
    Enabled     bool
    CreatedAt   time.Time
    // ... 其他业务字段
}

func (s *CronService) CreateJob(ctx context.Context, req *CreateJobReq) error {
    // 1. 保存到数据库
    job := &CronJob{
        Name: req.Name,
        Spec: req.Spec,
        Enabled: true,
    }
    if err := s.repo.Create(ctx, job); err != nil {
        return err
    }

    // 2. 注册到调度器
    return s.cronMgr.Add(job.Name, job.Spec, func() {
        s.executeJob(ctx, job.ID)
    })
}

func (s *CronService) executeJob(ctx context.Context, jobID int64) {
    // 从数据库获取最新状态
    job, err := s.repo.GetByID(ctx, jobID)
    if err != nil || !job.Enabled {
        return
    }

    // 执行业务逻辑
    // ...

    // 记录执行历史
    s.repo.RecordExecution(ctx, jobID, time.Now(), nil)
}
```

## 最佳实践

### 1. 任务函数设计

```go
// ✅ 推荐：独立的任务函数
func cleanupTempFiles() {
    // 清理逻辑
}

manager.Add("cleanup", "*/5 * * * * ?", cleanupTempFiles)

// ❌ 不推荐：匿名函数难以测试
manager.Add("cleanup", "*/5 * * * * ?", func() {
    // 大量逻辑...
})
```

### 2. 错误处理

```go
manager.Add("task", "*/5 * * * * ?", func() {
    if err := doSomething(); err != nil {
        log.Error().Err(err).Msg("任务执行失败")
        // 不要panic，让调度器继续运行
    }
})
```

### 3. 超时控制

```go
manager.Add("task", "*/5 * * * * ?", func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    if err := doSomethingWithContext(ctx); err != nil {
        log.Error().Err(err).Msg("任务执行失败或超时")
    }
})
```

### 4. 资源清理

```go
manager.Add("task", "*/5 * * * * ?", func() {
    resource, err := acquireResource()
    if err != nil {
        return
    }
    defer resource.Release()

    // 处理逻辑
})
```

## 注意事项

1. **并发执行**：同一任务如果执行时间超过调度间隔，可能会有多个实例并发执行
2. **时间精度**：调度精度约为秒级，不适用于毫秒级场景
3. **系统时间**：避免频繁修改系统时间
4. **优雅关闭**：应用关闭时调用 `Stop()` 等待任务完成

## 与之前版本的区别

如果你使用过旧版本的 `xcron`，主要变化：

| 旧版本 | 新版本 |
|--------|--------|
| `AddJob(spec, name, fn, opts...) (int64, error)` | `Add(name, spec, fn func()) error` |
| 返回自增ID | 使用任务名称作为唯一标识 |
| `JobInfo` 结构体（包含业务字段） | 纯调度引擎，无业务字段 |
| 4种钩子函数 | 移到业务层实现 |
| `UpdateJob`、`GetJob` | 业务层负责 |

迁移示例：

```go
// 旧版本
jobID, _ := manager.AddJob("*/5 * * * * ?", "cleanup", func() error {
    cleanup()
    return nil
}, xcron.WithDescription("清理任务"))

// 新版本
manager.Add("cleanup", "*/5 * * * * ?", func() {
    cleanup()
})
```

## 许可证

本项目采用 MIT 许可证。
