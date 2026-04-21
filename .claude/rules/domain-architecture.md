# 域子包架构规则

> 2026-04-21 架构重组后确立

## 规则 1：Handler/Service 按域分子包

每个业务域在 `internal/handler/{domain}/` 和 `internal/service/{domain}/` 下各有独立子包。包名即域名（无下划线）。

当前 12 个域：auth, captcha, department, dict, health, loginlog, menu, operationlog, position, role, tenant, user

## 规则 2：依赖链方向

```
router → handler/{domain} → service/{domain} → repository
```

每层只 import 下一层。handler 和 service 同名包不冲突（从不在同一文件 import）。

## 规则 3：构造函数直接注入

每个 handler/service 的构造函数只接收自己实际需要的依赖参数。

```go
// handler 构造函数 — 只接收所需依赖
func NewHandler(db *gorm.DB, recorder *audit.Recorder) *Handler {
    return &Handler{svc: xxx.NewService(db, recorder)}
}

// service 构造函数 — 只接收所需依赖
func NewService(db *gorm.DB, recorder *audit.Recorder) *Service {
    return &Service{repo: repository.NewXxxRepo(db), recorder: recorder}
}
```

App（`internal/router/app.go`）持有全部基础设施依赖，`initHandlers` 显式传递。

## 规则 4：文件命名

域子包内主结构文件以域名命名（`user.go`、`role.go`），不使用 `handler.go`/`service.go`。操作文件按功能拆分（`create.go`、`update.go`、`delete.go`、`query.go`）。

## 规则 5：Converter 可见性

- 同域使用：unexported（`modelToUserInfo`）
- 跨域引用（role/menu/tenant）：exported（`ModelToRoleInfo`），通过别名 import

```go
import roleconv "admin/internal/service/role"
```

## 规则 6：pkg 分层

- `pkg/`：业务包（会随业务变化修改）— audit, cache, config, constants, database, response, xcontext, xerr
- `pkg/utils/`：通用工具（不因业务变更）— bodyreader, captcha, convert, csv, httpclient, idgen, jwt, logger, pagination, passwordgen, rsapwd, useragent, xcron, xredis

## 规则 7：Router 解耦

`Setup(r *gin.Engine, handlers *Handlers, cfg *config.Config, jwtMgr *jwt.Manager, rbacCache *rbac.PermissionCache)` — 显式依赖，不引用 App struct。
