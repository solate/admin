# 配置部署路径与 ConfigMap 挂载设计

> **定位**:补充 L1 静态配置加载的部署侧设计,阐述"配置路径约定 / docker-compose 挂载 / k3s ConfigMap 挂载"的完整方案。
>
> - 配置加载机制(base+overlay/环境变量覆盖)见 [`pkg/xviper` 封装设计](../../../blog/从零设计Go配置加载-viper封装与约定式设计.md)
> - 配置结构定义与校验见 [`step-02-config-logger-db.md`](../../step-02-config-logger-db.md)
> - 本文聚焦:**部署时配置文件怎么进容器、路径怎么约定、flag 是否需要**

---

## 一、核心问题:为什么不需要 `-config` flag

### 1.1 flag 在容器部署中是死代码

**本地开发**时,flag 提供灵活性:

```bash
./server -config=custom.yaml  # 换个配置文件调试
```

但在 **docker-compose / k3s** 部署里,配置文件通过**挂载**进容器,启动命令写死在 `Dockerfile CMD` 或 k8s `command` 里,**没人会去传 `-config` 参数**:

| 部署形态 | 配置来源 | 启动命令在哪 | flag 能否传 |
|---------|---------|-------------|-----------|
| 本地开发 | `backend/config/config.yaml` | 手动 `go run ./cmd/server` | ✅ 能传 |
| docker-compose | volume 挂载 or 烤进镜像 | `Dockerfile CMD` 写死 | ❌ 不会改 CMD |
| k3s/k8s | ConfigMap 挂载到固定路径 | Pod spec `command` 写死 | ❌ 不会改 Pod YAML |

**结论**:容器化后,配置路径由 **挂载约定**决定,flag 成了永远不会被覆盖的默认值,等同于硬编码。

### 1.2 挂载覆盖的是"内容",不是"路径"

关键认知:docker/k3s 从不"换路径",而是把不同内容**挂到同一个固定路径**上。

- **docker-compose**:`volumes` 把宿主机文件挂到 `/app/config/config.yaml`(覆盖内容,路径不变)
- **k3s**:ConfigMap 挂到 `/app/config`(覆盖内容,路径不变)
- **环境切换**:靠 `APP_ENV` 走 overlay 文件合并,不是靠换路径

所以"运行时需要传不同路径"这个需求,在容器部署里根本不存在。既然路径永远是那一个,就没必要用 flag 或环境变量去参数化它——那只是在给一个不存在的问题留接口。

> 唯一需要传不同路径的是**单元测试**(`t.TempDir()`),这个由 `xviper.WithPath` option 覆盖,不影响生产路径的纯约定性。

### 1.3 约定优于配置:路径也是约定的一部分

本项目配置已全面拥抱"约定式设计"(见 `pkg/xviper` 约定):

- 格式固定 YAML
- Struct tag 固定 `mapstructure`
- 环境变量前缀固定 `APP_`
- 环境覆盖文件命名固定 `config.{env}.yaml`
- **基础路径固定 `config/config.yaml`**(`xviper` 的 `defaultSettings` 内置默认值)

**路径约定**是最后一块拼图:`xviper.Load()` 不传 `WithPath` 时就用内置的 `config/config.yaml`。容器内路径固定为 `/app/config/config.yaml`(工作目录下 `config/config.yaml`),三种部署形态统一,连兜底的环境变量都不需要。

---

## 二、设计方案:无参 `InitConfig()`,路径全交给约定

### 2.1 代码结构

`internal/config` 只暴露一个无参入口。基础路径由 `xviper` 内置约定提供,项目层不重复声明:

```go
// internal/config/config.go

// InitConfig 按约定加载并校验配置,返回 *Config。
// 基础路径由 xviper 约定为 config/config.yaml;环境切换靠 APP_ENV,
// 敏感字段靠 APP_ 前缀环境变量覆盖。容器/k3s 部署靠挂载覆盖该路径内容,无需传参。
func InitConfig() (*Config, error) {
    cfg, err := xviper.Load[Config](xviper.WithValidate(validate))
    if err != nil {
        return nil, err
    }
    return cfg, nil
}
```

**设计要点**:

- **无参** — 不传 `WithPath`,`xviper` 用 `defaultSettings` 内置的 `config/config.yaml`;路径是纯约定,不出现在函数签名里
- **只传 `WithValidate`** — 项目层唯一需要注入的是校验规则(业务相关),路径/格式/env 前缀全走 xviper 约定
- **测试传路径** — 单测需要临时路径时直接调 `xviper.Load[Config](xviper.WithPath[Config](tmp), ...)`,不污染生产入口

### 2.2 main.go 调用

```go
// cmd/server/main.go

func main() {
    // 1. 加载配置(路径/环境/env 覆盖全由约定处理)
    cfg, err := config.InitConfig()
    if err != nil {
        panic("load config: " + err.Error())
    }

    // 2. 初始化日志
    log := xslog.New(xslog.Config{...})
    ...
}
```

**对比改动前**:

```diff
- configPath := flag.String("config", "config/config.yaml", "配置文件路径")
- flag.Parse()
- cfg, err := config.Load(*configPath)
+ cfg, err := config.InitConfig()
```

`main.go` 减少 3 行,`flag` import 一并移除。组合根不再关心配置从哪来——这正是约定式设计的收益。

---

## 三、docker-compose 部署:烤进镜像 + volume 可选覆盖

### 3.1 Dockerfile — 配置烤进镜像作为兜底

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
COPY config ./config              # ← 配置目录烤进镜像(默认配置)
COPY migrations ./migrations

EXPOSE 8080
ENTRYPOINT ["./server"]
# 不再需要 CMD ["-config", "..."]
```

**容器内路径**:`/app/config/config.yaml`(与本地开发相对路径一致,工作目录为 `/app`)

### 3.2 docker-compose.yml — 两种配置注入方式

```yaml
services:
  backend:
    build: .
    environment:
      # 方式 1:环境变量覆盖配置字段(推荐,敏感信息)
      - APP_DATABASE_PASSWORD=postgres
      - APP_JWT_ACCESS_SECRET=dev-secret
      - APP_ENV=dev                 # 触发 xviper 读 config.dev.yaml
    volumes:
      # 方式 2:volume 挂载覆盖整个配置文件(可选,本地调试)
      - ./config/config.dev.yaml:/app/config/config.yaml:ro
```

**两种方式适用场景**:

| 方式 | 适用场景 | 优先级 |
|-----|---------|-------|
| `APP_*` 环境变量 | 敏感字段(密码/密钥)、环境切换(`APP_ENV`) | 最高(覆盖 YAML) |
| volume 挂载 | 本地开发改配置不重构建镜像 | 中(覆盖镜像里的文件) |

**推荐组合**:镜像里烤进默认配置 + `APP_ENV`/`APP_DATABASE_PASSWORD` 等环境变量注入 — 既能本地直接 `docker run`,又能生产注入密钥。

---

## 四、k3s/k8s 部署:ConfigMap 挂载配置 + Secret 注入密钥

### 4.1 ConfigMap — 存储配置文件内容

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: backend-config
  namespace: admin
data:
  config.yaml: |
    server:
      port: 8080
      mode: release
      read_timeout: 30
      write_timeout: 30
      graceful_timeout: 30
      cors:
        allowed_origins: ["https://admin.example.com"]
        allowed_methods: ["GET", "POST", "PUT", "DELETE"]
        allowed_headers: ["Content-Type", "Authorization"]
        allow_credentials: true

    database:
      host: postgres-service      # k8s Service DNS
      port: 5432
      user: admin_user
      password: ""                # 留空,走环境变量注入
      dbname: admin_prod
      sslmode: require
      max_idle_conns: 20
      max_open_conns: 200
      conn_max_lifetime: 3600

    redis:
      addr: redis-service:6379
      password: ""                # 留空,走环境变量注入
      db: 0

    jwt:
      access_secret: ""           # 留空,走环境变量注入
      refresh_secret: ""          # 留空,走环境变量注入
      access_ttl: 120
      refresh_ttl: 10080
      issuer: admin-system

    log:
      level: info
      format: json
      add_source: false
```

### 4.2 Secret — 存储敏感信息

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: backend-secret
  namespace: admin
type: Opaque
stringData:
  db-password: "YourStrongPassword123"
  redis-password: "RedisPassword456"
  jwt-access-secret: "AccessSecretXXXXXXXXXXXXXXXXXXXX"
  jwt-refresh-secret: "RefreshSecretYYYYYYYYYYYYYYYYY"
```

### 4.3 Deployment — 挂载 ConfigMap + 注入 Secret

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: admin
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
        - name: backend
          image: your-registry/admin-backend:v1.0.0
          ports:
            - containerPort: 8080
          env:
            # 环境标识(触发 xviper overlay 合并)
            - name: APP_ENV
              value: "prod"
            # 密钥通过 Secret 注入
            - name: APP_DATABASE_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: backend-secret
                  key: db-password
            - name: APP_REDIS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: backend-secret
                  key: redis-password
            - name: APP_JWT_ACCESS_SECRET
              valueFrom:
                secretKeyRef:
                  name: backend-secret
                  key: jwt-access-secret
            - name: APP_JWT_REFRESH_SECRET
              valueFrom:
                secretKeyRef:
                  name: backend-secret
                  key: jwt-refresh-secret
          volumeMounts:
            # ConfigMap 挂载到容器内固定路径,覆盖镜像里烤进去的 config/
            - name: config
              mountPath: /app/config
              readOnly: true
          livenessProbe:
            httpGet:
              path: /api/v1/health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /api/v1/health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
      volumes:
        - name: config
          configMap:
            name: backend-config
```

**挂载逻辑**:

1. 镜像里 `COPY config ./config` 烤进默认配置(兜底,CI/本地能直接跑)
2. k8s `volumeMount` 把 ConfigMap **覆盖挂载**到 `/app/config`
3. 容器内最终读到的是 ConfigMap 的内容,不是镜像里的
4. `APP_*` 环境变量优先级最高,覆盖 ConfigMap 里的字段

**路径统一**:

| 环境 | 配置来源 | 容器/进程内路径 | 密钥注入 |
|------|---------|---------------|---------|
| 本地开发 | `backend/config/config.yaml` | `./config/config.yaml` | 明文写在 `config.dev.yaml` |
| docker-compose | 镜像烤进去 + volume 可选覆盖 | `/app/config/config.yaml` | `APP_*` 环境变量 |
| k3s/k8s | ConfigMap 挂载覆盖 | `/app/config/config.yaml` | Secret → `APP_*` 环境变量 |

三种环境**容器内路径完全一致**,代码读的都是 `config/config.yaml`(相对于 `WORKDIR /app`),无需任何路径参数。

---

## 五、两种部署形态对比

### 5.1 docker-compose vs k3s 配置管理

| 维度 | docker-compose | k3s/k8s |
|-----|---------------|---------|
| **配置存储** | 宿主机文件 volume 挂载 | ConfigMap(etcd 存储) |
| **密钥管理** | compose 文件明文 or `.env` | Secret(base64,可接入 Vault) |
| **更新流程** | 改 YAML + `docker-compose restart` | `kubectl apply` + 滚动更新 |
| **版本控制** | Git 管理 `config.yaml` | ConfigMap 本身可 Git,但通常走 GitOps |
| **多环境** | 多个 `docker-compose.{env}.yml` | 多个 namespace or 多套 ConfigMap |
| **扩缩容** | 单机,手动改 `replicas` | 自动扩缩容(HPA) |

### 5.2 配置来源优先级(三层统一)

无论哪种部署,`xviper` 的三层优先级始终生效:

```
环境变量(APP_*)
    ↓ 覆盖
环境覆盖文件(config.prod.yaml,由 APP_ENV 触发)
    ↓ 覆盖
基础配置文件(config.yaml)
```

| 字段类型 | 推荐存储位置 | 理由 |
|---------|------------|------|
| 敏感信息(密码/密钥) | Secret → `APP_*` 环境变量 | 不进 Git,不进镜像,不进 ConfigMap |
| 环境差异(数据库地址/日志级别) | ConfigMap `config.yaml` or 镜像内 `config.prod.yaml` | 明文可查,易于审计 |
| 不变配置(超时/端口) | 镜像内 `config.yaml` | 烤进镜像,无需外部挂载 |

---

## 六、迁移与验证

### 6.1 代码改动清单

| 文件 | 改动 | 说明 |
|-----|------|------|
| `internal/config/config.go` | `Load(basePath)` → 无参 `InitConfig()` | 去掉路径参数,完全依赖 xviper 内置约定路径 `config/config.yaml` |
| `cmd/server/main.go` | 删除 `flag.String("config", ...)`<br>改为 `config.InitConfig()` | 简化启动代码,移除 `flag` import |
| `Dockerfile` | 删除 `CMD ["-config", ...]` | 不再需要传配置路径参数 |
| `step-02` 文档(可选) | 更新 `main.go` 示例 | 反映最新 API |

### 6.2 本地开发验证

```bash
# 默认读 config/config.yaml(xviper 内置约定路径)
cd backend && go run ./cmd/server
# 期望:启动成功,日志输出 "已加载基础配置 config/config.yaml"

# 切换环境(读 config.dev.yaml overlay)
APP_ENV=dev go run ./cmd/server
# 期望:日志输出 "已合并环境配置 config/config.dev.yaml"
```

### 6.3 docker-compose 验证

```bash
# 1. 构建镜像(配置烤进去)
docker build -t admin-backend:latest .

# 2. 启动(环境变量注入密钥)
docker-compose up -d

# 3. 查看配置加载日志
docker-compose logs backend | grep "已加载基础配置"
# 期望:xviper: 已加载基础配置 config/config.yaml
#       xviper: 已合并环境配置 config/config.dev.yaml(若 APP_ENV=dev)

# 4. 验证环境变量覆盖
docker-compose exec backend env | grep APP_
# 期望:看到 APP_DATABASE_PASSWORD=postgres 等

# 5. 健康检查
curl http://localhost:8080/api/v1/health
# 期望:{"code":0,"data":{"status":"ok","database":"healthy","redis":"healthy"}}
```

### 6.4 k3s 验证(部署后)

```bash
# 1. 应用 ConfigMap + Secret + Deployment
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/deployment.yaml

# 2. 查看 Pod 启动日志
kubectl logs -n admin deployment/backend --tail=50 | grep "已加载基础配置"
# 期望:看到配置加载日志

# 3. 进入 Pod 验证挂载路径
kubectl exec -n admin deployment/backend -it -- ls -la /app/config/
# 期望:看到 config.yaml(来自 ConfigMap,非镜像)

# 4. 验证环境变量注入
kubectl exec -n admin deployment/backend -- env | grep APP_
# 期望:APP_DATABASE_PASSWORD 等(来自 Secret)

# 5. 端口转发测试健康检查
kubectl port-forward -n admin deployment/backend 8080:8080
curl http://localhost:8080/api/v1/health
# 期望:返回健康状态
```

---

## 七、FAQ

### Q1:为什么不用 `config.prod.yaml` 烤进镜像,而是用 ConfigMap?

**A**: 两者可以组合,不是二选一:

- **镜像里烤 `config.prod.yaml`** — 作为兜底默认值,CI 能直接跑镜像(无需外部依赖)
- **ConfigMap 覆盖挂载** — 生产环境用,改配置不用重新构建镜像

推荐流程:镜像里烤进 `config.yaml`(通用默认)+ `config.prod.yaml`(生产默认),k8s 部署时用 ConfigMap **覆盖**。

### Q2:`APP_ENV=prod` 时,是读 ConfigMap 还是镜像里的 `config.prod.yaml`?

**A**: 两者都生效,叠加优先级:

1. `xviper` 先读基础文件 `/app/config/config.yaml`(来自 ConfigMap 挂载,覆盖了镜像里的)
2. 发现 `APP_ENV=prod`,尝试合并 `/app/config/config.prod.yaml`:
   - 如果 ConfigMap 里包含 `config.prod.yaml`(多文件 ConfigMap),就读它
   - 如果 ConfigMap 只有 `config.yaml`,但镜像里烤了 `config.prod.yaml`,就读镜像里的(挂载不覆盖其他文件)
3. 最后 `APP_*` 环境变量覆盖前两步的结果

**最佳实践**:生产环境 ConfigMap 只放一个完整的 `config.yaml`(已合并所有生产配置),设 `APP_ENV=prod` 只是为了触发日志/监控的环境标识,不依赖 overlay 文件。

### Q3:本地开发时,每次改配置都要重启服务吗?

**A**: 是的,配置加载在启动时一次性完成,运行时不监听文件变化。如需热重载:

- **方案 1(推荐)**:用 `air` 工具监听文件变化自动重启(项目已配置 `.air.toml`)
- **方案 2**:L2 动态配置(见 `research/04-三层架构...md`),从数据库 `sys_config` 表读,提供 API 刷新缓存

### Q4:能否完全不烤配置进镜像,纯依赖 ConfigMap?

**A**: 可以,但不推荐:

```dockerfile
# 不 COPY config/
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
# ConfigMap 挂载到 /app/config
```

**缺点**:
- 本地 `docker run` 没配置挂载会启动失败(需要额外 `-v` 参数)
- CI 流程需要额外准备配置文件
- 违背"镜像自包含"原则

**推荐**:镜像里保留默认配置,生产用 ConfigMap 覆盖 — 镜像能独立运行,部署有灵活性。

---

*最后更新:2026-07-03*
