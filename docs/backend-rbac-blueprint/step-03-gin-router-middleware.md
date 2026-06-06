# Step 03: Gin 路由与中间件链

## 目标

引入 Gin 框架，设计中间件链和路由注册机制，实现统一的请求响应格式和错误处理。

## 前置条件

- Step 02 完成，App 能初始化配置、数据库、日志

## 文件清单

```
internal/
├── router/
│   ├── app.go                 # App 增加 Gin Engine
│   └── router.go              # Setup() 路由注册 + 中间件链
├── middleware/
│   ├── request_id.go          # 请求 ID
│   ├── logger.go              # 请求日志
│   ├── recovery.go            # panic 恢复（可用 Gin 内置）
│   ├── cors.go                # CORS 跨域
│   └── auth.go                # JWT 认证（占位，Step 05 实现）
├── handler/
│   └── health/
│       └── health.go          # 健康检查（验证框架集成）
└── dto/
    └── common.go              # 通用 DTO（分页请求等）

pkg/
└── response/
    └── response.go            # 统一响应格式
```

## 实现细节

### 1. 统一响应格式

```go
// pkg/response/response.go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// 成功
func OK(c *gin.Context, data interface{})
func OKWithMessage(c *gin.Context, message string, data interface{})

// 失败
func Fail(c *gin.Context, err error)  // 从 xerr 提取 code 和 message
func FailWithCode(c *gin.Context, code int, message string)

// 分页
func OKWithPage(c *gin.Context, list interface{}, total int64, page, pageSize int)
```

**响应格式**：

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

```json
{
  "code": 40100,
  "message": "用户名或密码错误"
}
```

### 2. 中间件链顺序

```go
// router.go
func Setup(r *gin.Engine, ...) {
    // 公共路由（不需要认证）
    public := r.Group("/api/v1")
    public.GET("/health", healthHandler.Check)

    // 认证路由（需要 JWT + RBAC）
    auth := r.Group("/api/v1")
    auth.Use(middleware.JWTAuth(jwtMgr))
    // auth.Use(middleware.RBAC(rbacCache))  // Step 07 添加
    // auth.Use(middleware.Audit(recorder))  // Step 13 添加
    {
        // 后续步骤在这里注册业务路由
    }
}
```

**中间件顺序**：

```
Gin Engine 级别：
  RequestID → Logger → Recovery → CORS

路由组级别：
  JWTAuth → RBAC → Audit
```

### 3. Gin 封装策略

```go
// router/app.go
type App struct {
    Config *config.Config
    DB     *gorm.DB
    Engine *gin.Engine    // 新增
}

func (a *App) initRouter() {
    // 创建中间件和 handler
    // 调用 router.Setup(a.Engine, ...)
}

func (a *App) Run() error {
    return a.Engine.Run(fmt.Sprintf(":%d", a.Config.Server.Port))
}
```

**设计决策**：
- Gin Engine 在 App 中创建，不全局暴露
- 路由注册通过 `Setup()` 函数，接收显式参数，不引用 App
- Handler 构造函数只接收自己需要的依赖

### 4. 健康检查

```go
// handler/health/health.go
type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Check(c *gin.Context) {
    response.OK(c, gin.H{"status": "ok"})
}
```

### 5. 通用 DTO

```go
// dto/common.go
type PageRequest struct {
    Page     int `form:"page" binding:"omitempty,min=1"`
    PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

func (p *PageRequest) GetOffset() int {
    if p.Page <= 0 { p.Page = 1 }
    if p.PageSize <= 0 { p.PageSize = 10 }
    return (p.Page - 1) * p.PageSize
}

func (p *PageRequest) GetLimit() int {
    if p.PageSize <= 0 { p.PageSize = 10 }
    return p.PageSize
}

type IDRequest struct {
    ID string `uri:"id" binding:"required"`
}

type IDsRequest struct {
    IDs []string `json:"ids" binding:"required,min=1"`
}

type StatusRequest struct {
    IDs    []string `json:"ids" binding:"required,min=1"`
    Status int      `json:"status" binding:"required,oneof=1 2"`
}
```

## 依赖引入

```
go get github.com/gin-gonic/gin
```

## 验收标准

- [ ] `go build ./cmd/server` 编译通过
- [ ] 启动后 `curl localhost:8080/api/v1/health` 返回 `{"code":0,"message":"success","data":{"status":"ok"}}`
- [ ] 请求日志输出 Request ID（JSON 格式）
- [ ] 访问不存在的路由返回 `{"code":40400,"message":"路由不存在"}`
- [ ] CORS 头正确（`Access-Control-Allow-Origin`）
- [ ] panic 时 Recovery 中间件捕获并返回 500，不暴露堆栈

## AI 协作提示

```
请按 step-03-gin-router-middleware.md 引入 Gin 框架。
参考现有 backend/internal/router/ 和 internal/middleware/ 的模式，
但重新实现。重点：
1. response 包的统一响应格式
2. 中间件链的正确顺序
3. health handler 作为第一个业务 handler
4. 通用 DTO 定义（PageRequest, IDRequest 等）
```
