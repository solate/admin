# 代码架构重组记录

> 执行时间：2026-04-21
> 状态：已完成

## 变更概览

| 指标 | 改造前 | 改造后 |
|------|--------|--------|
| handler 层 | 12 个平铺文件，1 个包 | 12 个域子包，12 个包 |
| service 层 | 12 个平铺文件，1 个包 | 12 个域子包，12 个包 |
| converter | 独立包 `internal/converter/`，9 个文件 | 合并到各 service 子包的 `converter.go` |
| initHandlers | 50+ 行手动 repo→service→handler 组装 | 15 行，每个 handler 只需 `xxx.NewHandler(db, ...)` 显式传参 |
| Setup 签名 | `Setup(r, app)` 依赖整个 App | `Setup(r, handlers, cfg, jwt, rbac)` 显式依赖 |
| UserMenuHandler | 独立 struct（13 个 handler） | 合并到 user.Handler（12 个 handler） |
| pkg 通用包 | 17 个散落在 `pkg/` | 归集到 `pkg/utils/` |

---

## 一、依赖注入方式

App 直接持有全部基础设施依赖，`initHandlers` 显式传递各 handler 所需参数。每个 handler/service 的构造函数只接收自己实际需要的依赖。

```go
// handler 只接收自己需要的依赖
func NewHandler(db *gorm.DB, recorder *audit.Recorder) *Handler

// service 也只接收自己需要的依赖
func NewService(db *gorm.DB, recorder *audit.Recorder) *Service
```

详细改造记录见 `docs/di-architecture-analysis.md`。

---

## 二、Handler 子包文件映射

每个域子包下的主结构文件以域名命名（如 `user.go`），操作按功能拆文件。

| 域 | 旧文件 | 新目录 `handler/{domain}/` | 文件 |
|---|--------|---------------------------|------|
| health | `health_handler.go` | `handler/health/` | `health.go` |
| captcha | `captcha_handler.go` | `handler/captcha/` | `captcha.go` |
| auth | `auth_handler.go` | `handler/auth/` | `auth.go`, `login.go`, `logout.go`, `refresh.go` |
| user | `user_handler.go` + `user_menu_handler.go` | `handler/user/` | `user.go`, `create.go`, `update.go`, `delete.go`, `query.go`, `password.go`, `role.go`, `menu.go` |
| tenant | `tenant_handler.go` | `handler/tenant/` | `tenant.go`, `create.go`, `update.go`, `delete.go`, `query.go` |
| role | `role_handler.go` | `handler/role/` | `role.go`, `create.go`, `update.go`, `delete.go`, `query.go`, `permission.go` |
| menu | `menu_handler.go` | `handler/menu/` | `menu.go`, `create.go`, `update.go`, `delete.go`, `query.go` |
| department | `department_handler.go` | `handler/department/` | `department.go`, `create.go`, `update.go`, `delete.go`, `query.go` |
| position | `position_handler.go` | `handler/position/` | `position.go`, `create.go`, `update.go`, `delete.go`, `query.go` |
| dict | `dict_handler.go` | `handler/dict/` | `dict.go`, `create.go`, `update.go`, `delete.go`, `query.go` |
| loginlog | `login_log_handler.go` | `handler/loginlog/` | `loginlog.go` |
| operationlog | `operation_log_handler.go` | `handler/operationlog/` | `operationlog.go` |

### Handler struct 模式

```go
// handler/user/user.go
type Handler struct {
    svc     *usersvc.Service      // 主 service
    roleSvc *usersvc.RoleService  // 角色子 service
    menuSvc *usersvc.MenuService  // 菜单子 service
}

func NewHandler(db *gorm.DB, recorder *audit.Recorder, rsaCipher *rsapwd.RSACipher, cache *rbac.PermissionCache) *Handler {
    return &Handler{
        svc:     usersvc.NewService(db, recorder, rsaCipher),
        roleSvc: usersvc.NewRoleService(db, recorder),
        menuSvc: usersvc.NewMenuService(db, cache),
    }
}
```

大多数域只有一个 service，更简单：

```go
// handler/role/role.go
type Handler struct{ svc *rolesvc.Service }
func NewHandler(db *gorm.DB, recorder *audit.Recorder, cache *rbac.PermissionCache) *Handler {
    return &Handler{svc: rolesvc.NewService(db, recorder, cache)}
}
```

---

## 三、Service 子包文件映射

| 域 | 旧文件 | 新目录 `service/{domain}/` | 文件 |
|---|--------|---------------------------|------|
| auth | `auth_service.go` | `service/auth/` | `auth.go`, `login.go`, `logout.go`, `refresh.go` |
| user | `user_service.go` + `user_role_service.go` + `user_menu_service.go` | `service/user/` | `user.go`, `converter.go`, `create.go`, `update.go`, `delete.go`, `query.go`, `password.go`, `role.go`, `menu.go` |
| tenant | `tenant_service.go` | `service/tenant/` | `tenant.go`, `converter.go`, `create.go`, `update.go`, `delete.go`, `query.go` |
| role | `role_service.go` | `service/role/` | `role.go`, `converter.go`, `create.go`, `update.go`, `delete.go`, `query.go`, `permission.go` |
| menu | `menu_service.go` | `service/menu/` | `menu.go`, `converter.go`, `create.go`, `update.go`, `delete.go`, `query.go` |
| department | `department_service.go` | `service/department/` | `department.go`, `converter.go`, `create.go`, `update.go`, `delete.go`, `query.go` |
| position | `position_service.go` | `service/position/` | `position.go`, `converter.go`, `create.go`, `update.go`, `delete.go`, `query.go` |
| dict | `dict_service.go` | `service/dict/` | `dict.go`, `converter.go`, `create.go`, `update.go`, `delete.go`, `query.go` |
| loginlog | `login_log_service.go` | `service/loginlog/` | `loginlog.go`（含 converter） |
| operationlog | `operation_log_service.go` | `service/operationlog/` | `operationlog.go`（含 converter） |

---

## 四、Converter 迁移

`internal/converter/` 整个目录删除，converter 函数合并到对应的 service 子包。

### 可见性规则

| Converter | 所在包 | 可见性 | 调用方 |
|-----------|--------|--------|--------|
| `modelToUserInfo` | `service/user` | unexported | 包内 |
| `modelToRoleInfo` | `service/role` | **exported** | `service/user`（RoleService） |
| `modelListToRoleInfoList` | `service/role` | **exported** | `service/user`（UserService） |
| `modelToMenuInfo` | `service/menu` | **exported** | `service/user`（MenuService） |
| `modelToTenantInfo` | `service/tenant` | **exported** | `service/user`、`service/auth` |
| 其余域 converter | 各 service 子包 | unexported | 包内 |

跨域 converter 通过别名 import 避免包名冲突：

```go
import (
    roleconv   "admin/internal/service/role"
    menuconv   "admin/internal/service/menu"
    tenantconv "admin/internal/service/tenant"
)
```

无循环依赖：被依赖的包（role/menu/tenant）不 import 调用方的包（user/auth）。

---

## 五、User 域特殊结构

user 域是最复杂的，包含 3 个 service struct（都在 `package user` 内）：

| Struct | 文件 | 职责 |
|--------|------|------|
| `Service` | `user.go` | 用户 CRUD、密码管理 |
| `RoleService` | `role.go` | 用户角色分配、查询 |
| `MenuService` | `menu.go` | 用户菜单树、按钮权限 |

`Service` 内部依赖 `RoleService`（构造函数中创建）。

原 `UserMenuHandler` 合并到 `handler/user/`，路由中引用从 `handlers.UserMenuHandler` 改为 `handlers.UserHandler`。

---

## 六、Router 解耦

### 改造前

```go
// app.go
Setup(r, s)  // 传入整个 App

// router.go
func Setup(r *gin.Engine, app *App)
// 内部通过 app.Handlers、app.Config、app.JWT 等访问
```

### 改造后

```go
// app.go
auditDB := audit.NewDB(app.DB)
app.Audit = audit.NewRecorder(auditDB)
Setup(r, s.Handlers, s.Config, s.JWT, s.RBAC)

// router.go
func Setup(r *gin.Engine, handlers *Handlers, cfg *config.Config, jwtMgr *jwt.Manager, rbacCache *rbac.PermissionCache)
// handlers: 12 个域 handler
// cfg: 配置（RateLimit 等）
// jwtMgr: JWT 认证中间件
// rbacCache: RBAC 权限中间件
```

router.go 不再 import App struct 或 svc 包，只依赖 Handlers 和具体基础设施依赖。

---

## 七、pkg 分层调整

### 第一轮：通用工具归入 `pkg/utils/`

17 个通用包从 `pkg/` 移至 `pkg/utils/`：

bodyreader, config, convert, csv, httpclient, idgen, jwt, logger, pagination, passwordgen, response, rsapwd, useragent, xcontext, xcron, xerr, xredis

import 路径变更：`"admin/pkg/jwt"` → `"admin/pkg/utils/jwt"`

### 第二轮：业务基础包提升到 `pkg/`

`pkg/utils/` 中混入了业务基础包（会随业务修改），提升到 `pkg/` 层级：

| 包 | 移动 | 原因 |
|---|------|------|
| config | `pkg/utils/config` → `pkg/config` | 定义本应用的 Config 结构（数据库/Redis/JWT 配置） |
| xerr | `pkg/utils/xerr` → `pkg/xerr` | 错误码按业务域划分（租户/角色/菜单等 ~40 个码） |
| xcontext | `pkg/utils/xcontext` → `pkg/xcontext` | 定义多租户认证上下文（TenantID/UserID/Roles） |
| response | `pkg/utils/response` → `pkg/response` | 依赖 xerr 业务错误码，属于业务基础设施 |

同时将通用验证码包降入 `pkg/utils/`：

| 包 | 移动 | 原因 |
|---|------|------|
| captcha | `pkg/captcha` → `pkg/utils/captcha` | 纯通用工具，只依赖 base64Captcha + redis，零业务耦合 |

### 调整后分层

**`pkg/`（业务包，会随业务变化）：** audit, cache, config, constants, database, response, xcontext, xerr

**`pkg/utils/`（通用工具，不因业务变更）：** bodyreader, captcha, convert, csv, httpclient, idgen, jwt, logger, pagination, passwordgen, rsapwd, useragent, xcron, xredis

---

## 八、删除的文件

### handler 平铺文件（12 个）
`internal/handler/` 下的：auth_handler.go, captcha_handler.go, department_handler.go, dict_handler.go, health_handler.go, login_log_handler.go, menu_handler.go, operation_log_handler.go, position_handler.go, role_handler.go, tenant_handler.go, user_handler.go, user_menu_handler.go

### service 平铺文件（12 个）
`internal/service/` 下的：auth_service.go, department_service.go, dict_service.go, login_log_service.go, menu_service.go, operation_log_service.go, position_service.go, role_service.go, tenant_service.go, user_menu_service.go, user_role_service.go, user_service.go

### converter 目录
`internal/converter/` 整个目录（9 个文件）：login_log_converter.go, menu_converter.go, operation_log_converter.go, position_converter.go, role_converter.go, tenant_converter.go, user_converter.go, department_converter.go, dict_converter.go

---

## 九、设计决策

### Handler 模式：Struct vs 闭包

**选择 Struct 模式**。理由：
1. 依赖关系在 struct 定义中一目了然
2. 可直接注入 mock 进行单元测试
3. Go 生态主流（K8s, Docker, HashiCorp 均使用）
4. 闭包在多方法共享 service 时会退化为 struct

### 文件命名

域子包内主结构文件以域名命名（`user.go`、`role.go`），不使用 `handler.go`/`service.go`。操作文件按功能拆分（`create.go`、`update.go`、`delete.go`、`query.go`）。
