# Step 01: 项目骨架搭建

## 目标

创建 `backend/` 项目的完整目录结构、Go Module、Makefile，确保 `go build` 通过。

**这一步不引入任何第三方依赖**，只使用标准库。

## 前置条件

- Go 1.22+ 已安装（`go version`）
- 工作目录：`/path/to/admin/`

## 文件清单

```
backend/
├── cmd/server/main.go
├── internal/
│   ├── config/
│   ├── server/
│   ├── router/
│   ├── middleware/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── model/
│   ├── query/
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
├── .air.toml
├── Makefile
└── go.mod
```

## 实现规范

### 1. Go Module

```bash
mkdir backend && cd backend
go mod init admin
```

模块名 `admin`，与旧项目一致，方便后续 import 路径简短。

### 2. main.go — 最小可运行

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 最小 HTTP 服务器，验证项目骨架
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 优雅退出（从第一天就养成习惯）
	go func() {
		log.Printf("server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
	log.Println("server exited")
}
```

**设计要点**：
- 从第一天就使用优雅退出模式，后续只需替换 handler 不需改结构
- `ReadTimeout` / `WriteTimeout` 防止慢连接占用资源
- 不使用 `http.ListenAndServe(":8080", nil)`（全局默认 mux 是反模式）

### 3. Makefile

```makefile
.PHONY: build run dev clean test lint

# 构建
build:
	go build -o bin/server ./cmd/server

# 运行
run:
	go run ./cmd/server

# 热重载开发（需安装 air）
dev:
	air

# 清理
clean:
	rm -rf bin/ tmp/

# 测试
test:
	go test ./... -v -count=1

# 代码检查
lint:
	golangci-lint run ./...

# 格式化
fmt:
	gofmt -w .
```

### 4. .air.toml（热重载）

```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/server ./cmd/server"
bin = "./tmp/server"
include_ext = ["go", "yaml"]
exclude_dir = ["tmp", "bin", "vendor", "node_modules"]
delay = 1000

[log]
time = false

[misc]
clean_on_exit = true
```

### 5. .gitignore

```
# 构建产物
bin/
tmp/

# IDE
.idea/
.vscode/
*.swp

# 环境
.env
.env.local

# OS
.DS_Store
Thumbs.db

# 依赖（如果 vendor 模式）
# vendor/
```

### 6. config/config.yaml（占位）

```yaml
# 服务器配置
server:
  port: 8080
  mode: debug  # debug / release

# 后续步骤填充
# database:
# redis:
# jwt:
# log:
```

### 7. 空目录占位文件

每个空目录放一个 `.gitkeep` 文件，确保 Git 能跟踪。

或者：每个包目录放一个 `doc.go`：

```go
// Package database provides PostgreSQL connection management.
package database
```

**推荐用 doc.go**：既能让 Git 跟踪目录，又能作为包文档。

## 验收标准

```bash
# 1. Go Module 初始化
cd backend && cat go.mod | grep "module admin"
# 期望：module admin

# 2. 编译通过
go build ./cmd/server
# 期望：无错误，生成 bin/server

# 3. Makefile 工作
make build
ls bin/server
# 期望：文件存在

# 4. 启动并响应
make run &
sleep 1
curl -s http://localhost:8080/health
# 期望：{"status":"ok"}
kill %1

# 5. 目录结构完整
find . -type d | sort
# 期望：包含所有预定义目录
```

## AI 协作提示

```
请按 step-01-project-scaffolding.md 在 admin/backend/ 下创建项目骨架。

要求：
1. go mod init admin
2. main.go 使用标准库 net/http + 优雅退出模式
3. 不引入任何第三方依赖
4. 每个空包目录放 doc.go（包注释）
5. Makefile 包含 build/run/dev/clean/test/lint/fmt 目标
6. .air.toml 热重载配置
7. .gitignore 排除 bin/tmp/.env
```

---

*下一步：[Step 02 - 基础设施](step-02-config-logger-db.md)*
