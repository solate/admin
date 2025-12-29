# 操作日志使用指南

## 架构

```
请求 → AuthMiddleware → Handler → OperationLogMiddleware → 写入日志
       (注入用户信息)     (设置LogContext)    (收集信息+异步写入)
```

## 核心组件

| 文件 | 作用 |
|------|------|
| `pkg/operationlog/types.go` | 数据结构 (LogContext, LogEntry) |
| `pkg/operationlog/context.go` | context 存取 (WithLogContext, GetLogContext) |
| `pkg/operationlog/helper.go` | 便捷函数 (Record*) |
| `pkg/operationlog/writer.go` | Writer 写入器 |
| `internal/middleware/operation_log_middleware.go` | 中间件 |

## 使用方式

```go
import "admin/pkg/operationlog"
import "admin/pkg/constants"

// 创建
ctx = operationlog.RecordCreate(ctx, constants.ModuleUser, constants.ResourceTypeUser, user.UserID, user.Nickname, user)

// 更新
ctx = operationlog.RecordUpdate(ctx, constants.ModuleUser, constants.ResourceTypeUser, user.UserID, user.Nickname, oldUser, newUser)

// 删除
ctx = operationlog.RecordDelete(ctx, constants.ModuleUser, constants.ResourceTypeUser, user.UserID, user.Nickname, user)

// 查询
ctx = operationlog.RecordQuery(ctx, constants.ModuleUser, constants.ResourceTypeUser)

// 登录/登出
ctx = operationlog.RecordLogin(ctx, userID, userName)
ctx = operationlog.RecordLogout(ctx, userID, userName)

// 错误
operationlog.RecordError(ctx, err)
```

## 初始化

```go
// app.go
app.OperationLogWriter = operationlog.NewWriter(app.DB)

// router.go
authenticated.Use(middleware.OperationLogMiddleware(app.OperationLogWriter))
```

## 自动收集

中间件自动收集：`tenant_id`、`user_id`、`user_name`、`request_method`、`request_path`、`request_params`（脱敏）、`ip_address`、`user_agent`

业务代码只需设置：`module`、`operation_type`、`resource`、`old_value`、`new_value`
