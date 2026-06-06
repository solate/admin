# Step 01: 项目骨架搭建

## 目标

创建 `backend` 项目的基础目录结构和构建配置，确保能 `go build` 通过。

## 前置条件

- Go 1.22+ 已安装
- 工作目录：`/Users/solate/workspace/go/src/admin/`

## 文件清单

```
backend/
├── cmd/server/main.go          # 入口，最小化：只启动 HTTP server
├── internal/                   # 空目录，后续步骤填充
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── router/
│   ├── middleware/
│   ├── dto/
│   ├── dal/
│   └── rbac/
├── pkg/                        # 空目录，后续步骤填充
├── migrations/                 # SQL 迁移文件（根目录）
├── scripts/                    # 种子数据、生成脚本
├── docs/
├── .gitignore
├── Makefile
└── go.mod
```

## 实现细节

### 1. 初始化 Go Module

```bash
cd /Users/solate/workspace/go/src/admin
mkdir backend && cd backend
go mod init admin
```

### 2. main.go — 最小可运行

```go
package main

import (
    "log"
    "net/http"
)

func main() {
    log.Println("server starting on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}
```

> 不引入 Gin，先确保骨架能编译。Step 03 再引入 Gin。

### 3. Makefile

```makefile
.PHONY: build run dev clean

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

clean:
	rm -rf bin/
```

### 4. .gitignore

```
bin/
*.exe
.env
```

## 验收标准

- [ ] `go mod init admin` 成功
- [ ] `go build ./cmd/server` 编译通过
- [ ] `make build` 生成 `bin/server`
- [ ] `make run` 启动后，`curl localhost:8080` 不报错（可能 404，但不断连）
- [ ] 目录结构与文件清单一致

## AI 协作提示

```
请按 step-01-project-scaffolding.md 在 admin/backend/ 下创建项目骨架。
只创建空目录和最小文件，不引入任何第三方依赖。
main.go 只用标准库 net/http，端口 8080。
```
