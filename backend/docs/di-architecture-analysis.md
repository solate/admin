# 依赖注入（DI）架构改造记录

> 执行时间：2026-04-21
> 状态：已完成

## 变更概览

| 指标 | 改造前 | 改造后 |
|------|--------|--------|
| DI 容器 | `svc.ServiceContext`（God Object，7 字段） | 无，App 直接传显式参数 |
| Handler 构造函数 | `NewHandler(svcCtx)` 统一签名 | `NewHandler(db, recorder, ...)` 按需签名 |
| Service 构造函数 | `NewService(svcCtx)` 统一签名 | `NewService(db, ...)` 按需签名 |
| 依赖透明度 | 签名隐藏真实依赖 | 签名即文档 |
| svc 包 | `internal/svc/service_context.go` | 已删除 |

---

## 一、改造方案：构造函数直接注入

每个 handler/service 只接收自己实际需要的依赖参数。App 直接持有全部基础设施依赖，`initHandlers` 显式传递。

### Handler 依赖分布

| Handler | 依赖 | 参数 |
|---------|------|------|
| health | 无 | `NewHandler()` |
| captcha | Redis | `NewHandler(rdb)` |
| loginlog | DB | `NewHandler(db)` |
| operationlog | DB | `NewHandler(db)` |
| tenant | DB, Audit | `NewHandler(db, recorder)` |
| department | DB, Audit | `NewHandler(db, recorder)` |
| position | DB, Audit | `NewHandler(db, recorder)` |
| dict | DB, Audit | `NewHandler(db, recorder)` |
| role | DB, Audit, RBAC | `NewHandler(db, recorder, cache)` |
| menu | DB, Audit, RBAC | `NewHandler(db, recorder, cache)` |
| user | DB, Audit, RSA, RBAC | `NewHandler(db, recorder, rsa, cache)` |
| auth | DB, JWT, Redis, Audit, RSA, Config | `NewHandler(db, jwt, rdb, recorder, rsa, cfg)` |

### App struct

```go
type App struct {
    Config    *config.Config
    Router    *gin.Engine
    DB        *gorm.DB
    Redis     redis.UniversalClient
    JWT       *jwt.Manager
    RBAC      *rbac.PermissionCache
    RSACipher *rsapwd.RSACipher
    Cron      *xcron.Manager
    Handlers  *Handlers
    Audit     *audit.Recorder
}
```

### initHandlers

```go
func (s *App) initHandlers() error {
    s.Handlers = &Handlers{
        HealthHandler:       health.NewHandler(),
        CaptchaHandler:      captcha.NewHandler(s.Redis),
        AuthHandler:         auth.NewHandler(s.DB, s.JWT, s.Redis, s.Audit, s.RSACipher, s.Config),
        UserHandler:         user.NewHandler(s.DB, s.Audit, s.RSACipher, s.RBAC),
        TenantHandler:       tenant.NewHandler(s.DB, s.Audit),
        RoleHandler:         role.NewHandler(s.DB, s.Audit, s.RBAC),
        MenuHandler:         menu.NewHandler(s.DB, s.Audit, s.RBAC),
        LoginLogHandler:     loginlog.NewHandler(s.DB),
        OperationLogHandler: operationlog.NewHandler(s.DB),
        DepartmentHandler:   department.NewHandler(s.DB, s.Audit),
        PositionHandler:     position.NewHandler(s.DB, s.Audit),
        DictHandler:         dict.NewHandler(s.DB, s.Audit),
    }
    return nil
}
```

### Router Setup

```go
func Setup(r *gin.Engine, handlers *Handlers, cfg *config.Config, jwtMgr *jwt.Manager, rbacCache *rbac.PermissionCache)
```

Router 接收具体依赖（Config、JWT、RBAC），不再依赖 svc 包或 App struct。

---

## 二、删除的文件

- `internal/svc/service_context.go` — God Object 中间层
- `internal/svc/` 目录

---

## 三、变更文件清单

### Service 层（12 个文件）

| 域 | 文件 | 改动 |
|---|------|------|
| loginlog | `service/loginlog/loginlog.go` | `NewService(db *gorm.DB)` |
| operationlog | `service/operationlog/operationlog.go` | `NewService(db *gorm.DB)` |
| tenant | `service/tenant/tenant.go` | `NewService(db, recorder)` |
| department | `service/department/department.go` | `NewService(db, recorder)` |
| position | `service/position/position.go` | `NewService(db, recorder)` |
| dict | `service/dict/dict.go` | `NewService(db, recorder)` |
| user | `service/user/user.go` | `NewService(db, recorder, rsaCipher)` |
| user/role | `service/user/role.go` | `NewRoleService(db, recorder)` |
| user/menu | `service/user/menu.go` | `NewMenuService(db, cache)` |
| role | `service/role/role.go` | `NewService(db, recorder, cache)` |
| menu | `service/menu/menu.go` | `NewService(db, recorder, cache)` |
| auth | `service/auth/auth.go` | `NewService(db, jwt, rdb, recorder, rsa, cfg)` |

### Handler 层（11 个文件）

| 域 | 文件 | 改动 |
|---|------|------|
| loginlog | `handler/loginlog/loginlog.go` | `NewHandler(db)` |
| operationlog | `handler/operationlog/operationlog.go` | `NewHandler(db)` |
| tenant | `handler/tenant/tenant.go` | `NewHandler(db, recorder)` |
| department | `handler/department/department.go` | `NewHandler(db, recorder)` |
| position | `handler/position/position.go` | `NewHandler(db, recorder)` |
| dict | `handler/dict/dict.go` | `NewHandler(db, recorder)` |
| captcha | `handler/captcha/captcha.go` | `NewHandler(rdb)` |
| role | `handler/role/role.go` | `NewHandler(db, recorder, cache)` |
| menu | `handler/menu/menu.go` | `NewHandler(db, recorder, cache)` |
| user | `handler/user/user.go` | `NewHandler(db, recorder, rsa, cache)` |
| auth | `handler/auth/auth.go` | `NewHandler(db, jwt, rdb, recorder, rsa, cfg)` |

### Router 层（2 个文件）

| 文件 | 改动 |
|------|------|
| `router/app.go` | 移除 SvcCtx 字段和 svc import，initHandlers 改为显式传参，initRouter 传 cfg/jwt/rbac |
| `router/router.go` | Setup 签名改为 `(r, handlers, cfg, jwtMgr, rbacCache)`，移除 svc import |

---

## 四、设计决策

### 为什么选构造函数注入而非 DI 框架

1. 项目 12 域、~30 个 service 文件、单二进制，手动注入足够清晰
2. Go 社区主流模式（标准库 net/http、database/sql 均用）
3. 依赖在函数签名中完全透明，无需 IDE 跳转即可理解
4. Wire/Fx 学习曲线高，对本项目规模 ROI 不高

### 接口按需引入

不提前定义接口。只在需要为某个 service 写单元测试时，才为外部依赖定义接口。遵循 Go "accept interfaces, return structs" 哲学。
