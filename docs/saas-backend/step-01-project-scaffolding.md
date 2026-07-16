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
main.go(run()) → server.New(Options) → server.Run(ctx)
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
| `internal/server/` | 创建 Gin engine；注册中间件；调用 router.Setup()；包装 `*http.Server`；`Run(ctx)` 内 errgroup 编排 + 优雅关闭；持有 db/redis/logger（Step 02 起） |
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
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"admin/internal/config"
	"admin/internal/server"
)

func main() {
	// main 只负责退出码，逻辑与资源清理都在 run 里（defer 保证执行）
	if err := run(); err != nil {
		slog.Error("server exited with error", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.InitConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Step 02 起：这里顺序创建 db/redis 等基础设施，defer 逆序 Close

	srv, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}

	// signal.NotifyContext 把 SIGINT/SIGTERM 变成可传播的 ctx
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run 阻塞直到收到信号或任一组件出错，内部 errgroup 统一优雅关闭
	if err := srv.Run(ctx); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	slog.Info("server exited")
	return nil
}
```

### 3. internal/server/server.go（Step 01 骨架）

```go
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	// Run(ctx) 完整实现见 bootstrap doc 01，会用到 errors / golang.org/x/sync/errgroup
	"admin/internal/config"
)

type Server struct {
	httpSrv         *http.Server
	gracefulTimeout time.Duration
	// Step 02 起追加：db *gorm.DB, rdb *redis.Client, log *slog.Logger
	// Step 05 起追加：scheduler *scheduler.Scheduler
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
	return &Server{httpSrv: httpSrv, gracefulTimeout: 30 * time.Second}, nil
}

// Run 用 errgroup 编排长驻组件（组件 1 HTTP + 其带超时优雅关闭；Step 05 起加 scheduler）。
// 完整实现（两个 g.Go、ErrServerClosed 过滤成 nil、g.Wait 收敛）与 bootstrap doc 01
// 第一节的 Run(ctx) 逐字一致，此处不复制——doc 01 是生命周期编排的唯一权威版本。
func (s *Server) Run(ctx context.Context) error { /* errgroup 编排，见 bootstrap doc 01 */ }
```

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
