# Swagger 文档编写指南

本文档记录了项目中 Swagger/OpenAPI 文档的编写规范和最佳实践。

## 技术栈

- **工具**: [swaggo/swag](https://github.com/swaggo/swag) v1.16.6
- **框架**: Gin
- **规范**: OpenAPI 2.0 (Swagger 2.0)

## 核心原则

### 为什么这样设计？

1. **单一数据源原则**: DTO 结构定义是"真实来源"，Swagger 文档应该反映代码 reality
2. **维护性**: 改了 DTO 字段，Swagger 会自动更新
3. **类型安全**: 编译时发现问题

## 编写规范

### 1. Handler 层注释

在 handler 函数上方添加 Swagger 注释：

```go
// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新的用户账号，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateUserRequest true "创建用户请求参数"
// @Success 200 {object} response.Response{data=dto.UserResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
    // ...
}
```

**注意事项**:
- 使用 `@Success` 表示成功响应（通常为 200）
- 使用 `@Failure` 表示错误响应（400/401/403/500 等）
- 使用 `{object} response.Response{data=dto.XXXResponse}` 表示嵌套的 data 字段

### 2. DTO 层标签

为 DTO 结构体字段添加 swagger 标签：

```go
// UserInfo 用户基础信息（可复用）
type UserInfo struct {
    UserID        string `json:"user_id" example:"123456789012345678"`         // 用户ID
    UserName      string `json:"username" example:"admin"`                    // 用户名
    Nickname      string `json:"nickname" example:"系统管理员"`                  // 昵称/显示名称
    Phone         string `json:"phone" example:"13800138000"`                  // 手机号
    Email         string `json:"email" example:"admin@example.com"`            // 邮箱
    Status        int    `json:"status" example:"1" enum:"1,2"`                // 状态 1:正常 2:禁用
    TenantID      string `json:"tenant_id" example:"123456789012345678"`       // 租户ID
    LastLoginTime int64  `json:"last_login_time" example:"1735206400"`        // 最后登录时间（Unix时间戳）
    CreatedAt     int64  `json:"created_at" example:"1735200000"`             // 创建时间（Unix时间戳）
    UpdatedAt     int64  `json:"updated_at" example:"1735206400"`             // 更新时间（Unix时间戳）
}
```

**常用标签**:
| 标签 | 用途 | 示例 |
|------|------|------|
| `example` | 示例值 | `example:"123456789012345678"` |
| `enum` | 枚举值 | `enum:"1,2"` |
| `format` | 数据格式 | `format:"uuid"` (本项目不使用，用 18 位数字 ID) |

### 3. ID 规范

**重要**: 本项目使用 `idgen` 生成的 18 位数字 ID，**不是 UUID**。

```go
// ✅ 正确
UserID string `json:"user_id" example:"123456789012345678"`

// ❌ 错误
UserID string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
```

### 4. 嵌套对象处理

对于嵌套对象（如指针、切片），**不要**添加 `example` 标签：

```go
// ✅ 正确
type UserResponse struct {
    User *UserInfo `json:"user"` // 用户基础信息
}

type ProfileResponse struct {
    User   *UserInfo   `json:"user"`   // 当前用户信息
    Tenant *TenantInfo `json:"tenant"` // 当前租户信息
    Roles  []*RoleInfo `json:"roles"`  // 用户角色列表
}

// ❌ 错误 - 会导致 swag 解析失败
type UserResponse struct {
    User *UserInfo `json:"user" example:""`
}
```

## 生成文档

### 安装 swag

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### 生成 Swagger 文档

```bash
# 从项目根目录执行
cd backend
swag init -g cmd/server/main.go -o docs --parseInternal
```

**参数说明**:
- `-g cmd/server/main.go`: 指定 main.go 文件路径
- `-o docs`: 输出目录
- `--parseInternal`: 解析 internal 包的代码

### 验证生成的文档

```bash
# 查看 swagger.json 中的定义
cat docs/swagger.json | jq '.definitions["dto.UserInfo"]'

# 查看某个 API 的响应结构
cat docs/swagger.json | jq '.paths["/api/v1/users"].post.responses["200"]'
```

## 文件组织

```
backend/
├── cmd/
│   └── server/
│       └── main.go          # Swagger UI 中间件注册
├── docs/                     # 生成的 Swagger 文档
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── dto/                  # DTO 定义（包含 swagger 标签）
│   │   ├── user_dto.go
│   │   ├── role_dto.go
│   │   └── ...
│   └── handler/              # Handler 注释
│       ├── user_handler.go
│       └── ...
└── pkg/
    ├── pagination/           # 分页 DTO
    └── response/             # 响应结构
```

## 常见问题

### Q: 为什么必须在 DTO 中添加标签？

A: swag 通过 Go 反射扫描结构体定义，从 struct tags 中提取元数据。Handler 注释中的 `{object}` 只是引用，实际字段说明来自结构体本身。

### Q: 能否只在 Handler 注释中定义响应结构？

A: 可以使用 `map[string]interface{}` 但会失去类型检查和字段说明，不推荐。

### Q: 为什么不用 UUID？

A: 项目使用 `idgen` 生成的 18 位数字 ID，性能更好且更短。

### Q: Apifox 导入后 data 字段显示不完整？

A: 确保：
1. 使用最新的 `docs/swagger.json` 文件导入
2. 检查 Apifox 版本是否最新
3. 选择 "OpenAPI 3.0" 格式导入（或尝试 OpenAPI 2.0）

## 响应结构规范

### 统一响应格式

所有 API 响应都遵循统一的 Response 结构：

```json
{
  "code": 200,           // 业务状态码
  "message": "success",  // 响应消息
  "request_id": "abc123",// 请求ID
  "data": {              // 业务数据
    // 具体的业务数据
  }
}
```

### Handler 注释中的 data 字段

使用 `response.Response{data=dto.XXXResponse}` 语法指定具体的 data 类型：

```go
// @Success 200 {object} response.Response{data=dto.UserResponse} "创建成功"
```

swag 会自动生成 `allOf` 结构，将 `response.Response` 和具体的 data 定义合并。

## 参考资源

- [swaggo/swag GitHub](https://github.com/swaggo/swag)
- [swag 官方文档](https://swaggo.github.io/swag/)
- [Swagger 2.0 规范](https://swagger.io/specification/v2/)
- [Modern API Documentation with Swagger in Go GIN](https://articles.wesionary.team/modern-api-documentation-with-swagger-in-go-gin-2aee10cb7bb9)

## 更新日志

- 2026-01-06: 初始版本，记录项目 Swagger 编写规范
