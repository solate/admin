# Step 17: 部署与运维

## 目标

容器化部署、健康检查、配置管理、日志收集等生产环境准备。

## 前置条件

- 所有功能开发完成

> 配置路径约定、flag 去留决策、k3s ConfigMap + Secret 挂载的深入设计见 [research/config-loading/05-配置部署路径与ConfigMap挂载设计](research/config-loading/05-配置部署路径与ConfigMap挂载设计.md)。本步的 Dockerfile / docker-compose 是其落地实例。

## 文件清单

```
Dockerfile                     # 多阶段构建
docker-compose.yml             # 本地开发完整环境
docker-compose.prod.yml        # 生产配置覆盖
deploy/
├── nginx.conf                 # 反向代理配置
└── init.sql                   # 数据库初始化（Docker 用）
config/
└── config.prod.yaml           # 生产配置模板
```

## 实现规范

### 1. Dockerfile — 多阶段构建

```dockerfile
# ============ Build Stage ============
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 依赖缓存
COPY go.mod go.sum ./
RUN go mod download

# 编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# ============ Runtime Stage ============
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/server .
COPY config/ ./config/
COPY migrations/ ./migrations/

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["./server"]
```

> 无 `-config` flag：路径固定为约定的 `config/config.yaml`（WORKDIR 下相对路径，由 `pkg/xviper` 内置约定），环境切换靠 `APP_ENV`、覆盖配置靠 volume/ConfigMap 挂载或 `APP_*` 环境变量。详见 [research/config-loading/05](research/config-loading/05-配置部署路径与ConfigMap挂载设计.md)。

**构建优化**：
- 多阶段构建：最终镜像 ~20MB（Alpine + 静态二进制）
- 依赖缓存：`go.mod` 变化才重新 `go mod download`
- `-ldflags="-s -w"`：去掉符号表和调试信息，减小二进制体积
- HEALTHCHECK：Docker 原生健康检查

### 2. docker-compose.yml（开发环境）

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: app_dev
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./deploy/init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  backend:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DB_PASSWORD=postgres
      - JWT_ACCESS_SECRET=dev-access-secret
      - JWT_REFRESH_SECRET=dev-refresh-secret
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./config/config.dev.yaml:/app/config/config.yaml

volumes:
  postgres_data:
```

### 3. 生产配置要点

```yaml
# config/config.prod.yaml
server:
  port: 8080
  mode: release
  read_timeout: 30s
  write_timeout: 30s

database:
  host: ${DB_HOST}       # 环境变量注入
  port: 5432
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  dbname: admin_prod
  sslmode: require       # 生产必须 SSL
  max_idle_conns: 20
  max_open_conns: 200
  conn_max_lifetime: 3600

redis:
  addr: ${REDIS_ADDR}
  password: ${REDIS_PASSWORD}
  db: 0

log:
  level: info
  format: json           # 生产用 JSON 方便日志收集
```

### 4. 健康检查增强

```go
// handler/health/health.go
func (h *Handler) Check(c *gin.Context) {
    checks := gin.H{"status": "ok"}

    // 数据库健康
    if err := h.db.Exec("SELECT 1").Error; err != nil {
        checks["database"] = "unhealthy"
        checks["status"] = "degraded"
    } else {
        checks["database"] = "healthy"
    }

    // Redis 健康
    if err := h.rdb.Ping(c.Request.Context()).Err(); err != nil {
        checks["redis"] = "unhealthy"
        checks["status"] = "degraded"
    } else {
        checks["redis"] = "healthy"
    }

    response.OK(c, checks)
}
```

### 5. 优雅退出完整版

生命周期骨架（`run()` 组合根 + `signal.NotifyContext` + `srv.Run(ctx)` 的 errgroup 编排）是**既有实现**，完整代码见 [bootstrap doc 01](./research/bootstrap/01-应用启动与组件生命周期-main组合根与gin-cron编排.md) 第一节，此处不复制。本步只关心一件 step-17 特有的事——**审计日志缓冲的刷新时机**：

- db、redis 是 `run()` 的局部变量，用 `defer` 逆序 Close（redis 先、db 后），HTTP 的带超时优雅关闭由 `srv.Run(ctx)` 内部的 errgroup 负责——这两类都无需在这里手写。
- 审计 recorder 有内存缓冲，进程退出前必须 `Flush()` 落盘，否则丢最后一批审计记录。把它挂成 `defer recorder.Flush()`，且**注册在 db 的 `defer` 之后**（先声明后执行，Flush 先于 db.Close 跑），保证刷缓冲时数据库连接还在：

```go
func run() error {
    // ... cfg/log/db/redis 初始化（详见 bootstrap doc 01）...
    defer xgorm.Close(db)          // 最后关（先声明）
    recorder := audit.NewRecorder(db)
    defer recorder.Flush()         // 在 db.Close 之前刷缓冲（后声明先执行）

    srv, err := server.New(server.Options{Config: cfg, DB: db, RDB: rdb, Log: log, Audit: recorder})
    if err != nil {
        return fmt.Errorf("init server: %w", err)
    }
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()
    return srv.Run(ctx)            // errgroup 编排 HTTP 优雅关闭，见 bootstrap doc 01
}
```

### 6. Makefile 追加

```makefile
# Docker
docker-build:
	docker build -t admin-backend:latest .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f backend

# 生产部署
deploy:
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### 7. 运维清单

| 项目 | 说明 |
|------|------|
| 日志收集 | JSON 格式输出到 stdout，Docker 日志驱动收集 |
| 监控 | /health 端点 + Prometheus metrics（后续扩展） |
| 备份 | PostgreSQL pg_dump 定时备份 |
| 密钥管理 | 生产环境敏感信息走环境变量或 Secret Manager |
| HTTPS | Nginx 反代终结 SSL，后端只监听 HTTP |
| 限流 | Nginx 层面做 rate limiting |

## 验收标准

```bash
# 1. Docker 构建
docker build -t admin-backend:latest .
# 期望：构建成功，镜像 < 30MB

# 2. docker-compose 启动
docker-compose up -d
# 期望：所有服务 healthy

# 3. 健康检查
curl http://localhost:8080/api/v1/health
# 期望：{"code":0,"data":{"status":"ok","database":"healthy","redis":"healthy"}}

# 4. 优雅退出
docker-compose stop backend
# 期望：日志输出 "shutdown signal received" → "server exited cleanly"
# 无报错，残余请求处理完毕

# 5. 生产模式
# mode=release 时无 Swagger、无 debug 日志

# 6. 镜像扫描
docker scout cves admin-backend:latest
# 期望：无 CRITICAL/HIGH 漏洞
```

## AI 协作提示

```
请按 step-17-deployment.md 实现部署配置。

要点：
1. Dockerfile 多阶段构建（Alpine 基础镜像）
2. docker-compose.yml 完整开发环境（postgres + redis + backend）
3. 健康检查增强（检测 DB + Redis 状态）
4. 优雅退出完整流程（停服务 → 刷缓冲 → 关连接）
5. 生产配置：SSL、JSON 日志、环境变量注入
6. Makefile 追加 docker 相关目标
```

---

*上一步：[Step 16 - Swagger 文档](step-16-swagger.md)*

---

## 🎉 恭喜完成！

至此，你拥有了一个完整的多租户 SaaS 后端：

- ✅ 项目骨架与构建工具
- ✅ 配置/日志/数据库/Redis 基础设施
- ✅ Gin HTTP 框架 + 统一响应 + 错误码
- ✅ 完整数据模型 + GORM Gen
- ✅ 多租户 JWT 认证
- ✅ 登录/登出/刷新 Token
- ✅ RBAC 权限缓存 + 中间件
- ✅ 权限/菜单管理
- ✅ 租户/用户/角色/部门 CRUD
- ✅ 五级数据权限
- ✅ 审计日志
- ✅ Swagger 文档
- ✅ 容器化部署
