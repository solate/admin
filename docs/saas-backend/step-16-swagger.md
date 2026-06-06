# Step 16: Swagger API 文档

## 目标

使用 swaggo 自动生成 Swagger/OpenAPI 文档，提供在线 API 测试界面。

## 前置条件

- 所有业务接口已实现（Step 09-15）

## 文件清单

```
docs/swagger/                  # swag init 生成（勿手动编辑）
├── docs.go
├── swagger.json
└── swagger.yaml

cmd/server/main.go             # 注册 swagger 路由
internal/router/router.go      # 添加 swagger 路由
Makefile                       # 追加 swag 目标
```

## 实现规范

### 1. 安装 swaggo

```bash
go install github.com/swaggo/swag/cmd/swag@latest
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
```

### 2. main.go 注解

```go
// @title           Admin SaaS API
// @version         1.0
// @description     多租户 SaaS 后端管理系统 API
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer Token 认证
func main() { ... }
```

### 3. Handler 注解规范

```go
// Login 用户登录
// @Summary      用户登录
// @Description  使用邮箱和密码登录，返回 Token
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "登录请求"
// @Success      200 {object} response.Response{data=LoginResponse}
// @Failure      200 {object} response.Response
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) { ... }
```

### 4. 注解规则

| 规则 | 说明 |
|------|------|
| ID 字段用 string | `@Param role_id query string true "角色ID"` |
| @Success 必须指定 data | `{object} response.Response{data=dto.RoleInfo}` |
| DTO 字段必须有 example | `UserID string json:"user_id" example:"123456"` |
| Enums 逗号后空格 | `Enums(A, B, C)` |
| 时间戳注明单位 | `"开始时间(毫秒级时间戳)"` |

### 5. 路由注册

```go
import (
    ginSwagger "github.com/swaggo/gin-swagger"
    swaggerFiles "github.com/swaggo/files"
    _ "admin/docs/swagger" // 导入生成的 docs
)

// 只在 debug 模式注册
if cfg.Server.Mode == "debug" {
    engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
```

### 6. Makefile

```makefile
# Swagger 文档生成
swagger:
	swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal

# 格式化 swagger 注解
swagger-fmt:
	swag fmt
```

## 验收标准

```bash
# 1. 生成文档
make swagger
# 期望：docs/swagger/ 下生成 3 个文件

# 2. 编译通过
go build ./...

# 3. 访问文档
# 启动后访问 http://localhost:8080/swagger/index.html
# 期望：看到完整 API 列表

# 4. 在线测试
# Swagger UI 中尝试 Login 接口 → 获取 Token
# 点击 Authorize 填入 Bearer Token
# 访问需认证接口 → 成功返回

# 5. 生产模式不暴露
# mode=release 时访问 /swagger → 404
```

## AI 协作提示

```
请按 step-16-swagger.md 为所有接口添加 Swagger 注解。

要点：
1. main.go 顶部添加 API 基本信息注解
2. 每个 handler 方法添加完整注解（Summary/Tags/Param/Success）
3. 遵守 swagger-rules.md 中的规则（ID 用 string、Enums 有空格等）
4. DTO struct 添加 example tag
5. 路由注册 swagger handler（仅 debug 模式）
6. make swagger 生成文档
```

---

*上一步：[Step 15 - 审计日志](step-15-audit-log.md) | 下一步：[Step 17 - 部署与运维](step-17-deployment.md)*
