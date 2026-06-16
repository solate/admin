# Step 01: 项目骨架搭建

## 目标

创建 `backend/` 项目的完整目录结构、Go Module、Makefile，确保 `go build` 通过。

**这一步只引入一个第三方依赖**：`gopkg.in/yaml.v3`（读取配置文件）。

## 前置条件

- Go 1.22+ 已安装（`go version`）
- 工作目录：`/path/to/admin/`

## 目录结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   ├── server/       ← Server struct：lifecycle + HTTP 全在此
│   ├── router/       ← 仅路由注册
│   ├── middleware/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── model/        ← gorm-gen 输出，禁止手写
│   ├── query/        ← gorm-gen 输出，禁止手写
│   ├── dto/
│   └── rbac/
├── pkg/
│   ├── database/
│   ├── rdb/
│   ├── logger/
│   ├── jwt/
│   ├── xcontext/
│   ├── xerr/
│   ├── response/
│   ├── idgen/
│   └── password/
├── config/
│   └── config.yaml
├── migrations/
├── scripts/
├── .gitignore
├── Makefile
└── go.mod
```

## 架构选型：为什么只有 `internal/server/`

### Gin 社区单层方案（本项目采用）

主流 Gin boilerplate（vsouza、Massad 等）的标准结构：`internal/server/` 既封装 Gin engine，也管理 HTTP server lifecycle（Start/Stop），同时持有 db/redis/logger 等基础设施依赖。

```
main.go → server.New(cfg) → server.Start() / server.Stop()
```

`internal/server/` 就是 App，没有额外的编排层。

### go-kratos 两层方案（本项目未采用）

go-kratos 为了同时支持 HTTP + gRPC 两个传输层，拆出了独立的编排层：

```
internal/app/    ← kratos.New()，编排多个 server 组件
internal/server/ ← HTTP 组件 / gRPC 组件
```

对纯 Gin 项目（单 HTTP 传输），这一层是多余的。

### 两层职责对应表

| 包 | 职责 |
|---|---|
| `internal/server/` | 创建 Gin engine；注册中间件；调用 router.Setup()；包装 `*http.Server`；Start/Stop lifecycle；持有 db/redis/logger（Step 02 起） |
| `internal/router/` | `Setup(r *gin.Engine, ...)` 注册全部路由，无 lifecycle |

## 实现

### 1. Go Module

```bash
mkdir backend && cd backend
go mod init admin
go get gopkg.in/yaml.v3
```

模块名 `admin`，与旧项目一致，import 路径简短。

### 2. cmd/server/main.go

```go
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"admin/internal/config"
	"admin/internal/server"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("init server: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("start server: %v", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Stop(shutdownCtx)
	log.Println("server exited")
}
```

**设计要点**：
- `signal.NotifyContext`（Go 1.16+）：比手写 channel 更简洁
- 30 秒优雅退出超时（生产级标准，Step 01 就定好，后续不再改）
- main.go 是**最终结构**——后续步骤只填充 `server.New()` 内部，main.go 不再变动

### 3. internal/server/server.go（Step 01 骨架）

```go
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"admin/internal/config"
)

type Server struct {
	httpSrv *http.Server
	// Step 02 起追加：db *gorm.DB, rdb *redis.Client, logger zerolog.Logger
	// Step 05 起追加：cronRunner *cron.Runner
}

func New(cfg *config.Config) (*Server, error) {
	// Step 02 起：初始化 db、redis、logger，传入 Server
	// Step 03 起替换为：engine := gin.New(); router.Setup(engine, ...)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return &Server{httpSrv: httpSrv}, nil
}

func (s *Server) Start() error {
	// Step 05 起追加：go s.cronRunner.Start(ctx)
	log.Printf("server starting on %s", s.httpSrv.Addr)
	if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) {
	// Step 05 起逆序追加：s.cronRunner.Stop()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
```

**设计要点**：
- Step 01 先用标准库 `http.NewServeMux()`，Step 03 替换为 `gin.New()`，接口不变
- 注释标注了各 Step 会追加的字段和逻辑，避免未来忘记位置
- cron 加入后（Step 05），在 `Start()`/`Stop()` 内追加，main.go 无感知

### 4. internal/config/config.go（Step 01 极简版）

```go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	// Step 02 起追加：Database, Redis, JWT, Log
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	return &cfg, nil
}
```

### 5. Makefile

```makefile
.PHONY: build run dev clean test lint fmt

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

dev:
	go run ./cmd/server

clean:
	rm -rf bin/ tmp/

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
```

**注意**：不使用 air 热重载，`dev` 目标直接 `go run`。

### 6. config/config.yaml

```yaml
server:
  port: 8080
  mode: debug  # debug / release

# Step 02 起填充：
# database:
# redis:
# jwt:
# log:
```

### 7. 包文档占位（doc.go）

每个空包目录放 `doc.go`，让 Git 跟踪目录，同时作为包文档：

```go
// Package server manages HTTP server lifecycle and Gin engine setup.
package server
```

## 验收标准

```bash
# 1. 模块名正确
cat go.mod | grep "module admin"

# 2. 全包编译通过
go build ./...

# 3. Makefile 工作
make build
ls bin/server

# 4. 启动并响应
make run &
sleep 1
curl -s http://localhost:8080/health
# 期望：{"status":"ok"}
kill %1
```

---

*下一步：[Step 02 - 基础设施](step-02-config-logger-db.md)*
