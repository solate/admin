# 操作日志使用指南

## 架构概述

操作日志系统采用 **中间件 + 上下文 + 异步写入** 的设计，实现无感知的操作日志记录：

```
请求 → AuthMiddleware → Handler → OperationLogMiddleware → 写入日志
       (注入用户信息)     (记录操作)    (收集信息+异步写入)
```

### 职责分离

| 中间件 | 职责 |
|--------|------|
| `AuthMiddleware` | 解析 JWT，将 `tenant_id`、`user_id`、`user_name` 注入 `request.Context` |
| `OperationLogMiddleware` | 检查是否有 `LogContext`，收集请求信息，异步写入日志 |
| 业务代码 (Service) | 设置 `LogContext`，指定 `module`、`operation_type`、`resource` |

## 核心组件

| 文件 | 作用 |
|------|------|
| `pkg/operationlog/context.go` | 日志上下文和构建器 |
| `pkg/operationlog/helper.go` | 便捷辅助函数 |
| `pkg/operationlog/logger.go` | 日志写入器（DB 写入） |
| `internal/middleware/operation_log_middleware.go` | 中间件（收集信息、异步写入） |
| `internal/dal/model/operation_logs.gen.go` | 数据模型 |

## 使用方式

### 方式一：便捷函数（推荐）

```go
import "admin/pkg/operationlog"
import "admin/pkg/constants"

// 在 Service 层记录操作
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
    // ... 业务逻辑 ...

    // 记录创建操作
    ctx = operationlog.RecordCreate(ctx, constants.ModuleUser, "user", user.ID, user.Name, user)

    return resp, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) error {
    // 获取旧数据
    oldUser, _ := s.userRepo.GetByID(ctx, req.UserID)

    // ... 更新逻辑 ...

    // 记录更新操作（包含变更前后数据）
    ctx = operationlog.RecordUpdate(ctx, constants.ModuleUser, "user", req.UserID, oldUser.Name, oldUser, newUser)
    return nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
    // 获取删除前的数据
    oldUser, _ := s.userRepo.GetByID(ctx, userID)

    // ... 删除逻辑 ...

    // 记录删除操作
    ctx = operationlog.RecordDelete(ctx, constants.ModuleUser, "user", userID, oldUser.Name, oldUser)
    return nil
}
```

### 方式二：链式构建器

```go
ctx = operationlog.Record(ctx, constants.ModuleUser, constants.OperationCreate).
    Resource("user", "123", "张三").
    Data(nil, user).
    BuildToContext(ctx)
```

### 方式三：错误记录

```go
func (s *UserService) SomeOperation(ctx context.Context) error {
    err := doSomething()
    if err != nil {
        operationlog.RecordError(ctx, err)
        return err
    }
    return nil
}
```

### 方式四：登录/登出

```go
// 登录
ctx = operationlog.RecordLogin(ctx, userID, userName)

// 登出
ctx = operationlog.RecordLogout(ctx, userID, userName)
```

## 自动收集的信息

| 字段 | 来源 |
|------|------|
| `tenant_id` | `AuthMiddleware` 注入到 `request.Context` |
| `user_id` | `AuthMiddleware` 注入到 `request.Context` |
| `user_name` | `AuthMiddleware` 注入到 `request.Context` |
| `request_method` | HTTP Method |
| `request_path` | URL Path |
| `request_params` | Request Body/Query（OperationLogMiddleware 收集并脱敏） |
| `ip_address` | `useragent.GetClientIP()` |
| `user_agent` | HTTP User-Agent Header |
| `status` | 根据响应状态码自动设置 |

业务代码只需设置：
- `module` - 模块名称
- `operation_type` - 操作类型
- `resource_type/id/name` - 资源信息（可选）
- `old_value/new_value` - 数据变更（可选）

## 脱敏处理

中间件会自动脱敏敏感字段：

- `password`, `passwd`, `pwd`
- `secret`, `token`, `access_token`, `refresh_token`
- `api_key`, `apikey`, `api-key`
- `phone`, `mobile`, `telephone`
- `id_card`, `idcard`

示例：
```json
{"password": "123456"} → {"password": "12******56"}
```

## 中间件执行顺序

```go
// router/router.go
authenticated.Use(middleware.AuthMiddleware(app.JWT))           // 1. 认证，注入用户信息
authenticated.Use(middleware.CasbinMiddleware(app.Enforcer))     // 2. 权限检查
authenticated.Use(middleware.OperationLogMiddleware(logService)) // 3. 操作日志
```

## 数据库表结构

参考 `backend/scripts/dev_schema.sql` 中的 `operation_logs` 表定义。

## 常量定义

模块和操作类型常量在 `pkg/constants/operation_log.go` 中定义：

```go
const (
    ModuleUser       = "user"
    ModuleRole       = "role"
    ModuleTenant     = "tenant"
    // ...
)

const (
    OperationCreate = "CREATE"
    OperationUpdate = "UPDATE"
    OperationDelete = "DELETE"
    // ...
)
```
