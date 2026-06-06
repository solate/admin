# CLAUDE.md

此文件为 Claude Code (claude.ai/code) 在此代码库中工作时提供指导。

## 语言规范
- 所有对话和文档都使用中文
- 文档使用 markdown 格式

## 开发命令

```bash
make init                    # 安装依赖、迁移数据库、生成代码
make dev                     # 运行 migrate + gen-db + 启动服务器
make run                     # 仅运行服务器（不执行迁移/生成）
make migrate-up              # 应用待执行的迁移
make migrate-down            # 回滚一个迁移
make migrate-create NAME=xxx # 创建新的迁移文件
make gen-db                  # 从数据库架构生成 GORM 模型
make test                    # 运行所有测试
make lint                    # 运行 golangci-lint
make swagger                 # 生成 Swagger 文档
```

数据库配置环境变量：`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`

## 架构概览

基于 Go 的多租户管理系统，采用数据库优先 + 域子包分层架构。

### 核心技术栈
Gin + GORM Gen + PostgreSQL + Redis + JWT + 纯数据库 RBAC（PermissionCache）

### 项目结构
```
internal/
├── handler/{domain}/    # HTTP 处理器（12 个域子包）
├── service/{domain}/    # 业务逻辑 + converter（12 个域子包）
├── repository/          # 数据仓储层（集中式，GORM Gen）
├── router/              # 路由定义 + App 初始化（Setup 签名解耦）
├── dto/                 # 数据传输对象
├── dal/model/           # 生成的模型（切勿手动编辑）
├── dal/query/           # 生成的查询（切勿手动编辑）
├── middleware/           # HTTP 中间件链
└── rbac/                # RBAC 权限缓存

pkg/
├── utils/               # 通用工具（jwt, logger, idgen, captcha 等 14 个包）
├── audit/               # 审计日志
├── cache/               # 租户缓存
├── config/              # 应用配置（Config 结构定义）
├── constants/           # 业务常量
├── database/            # 数据库连接
├── response/            # HTTP 响应封装
├── xcontext/            # 多租户认证上下文
└── xerr/                # 业务错误码
```

### 依赖链
```
router → handler/{domain} → service/{domain} → repository
```

每个 handler/service 的构造函数只接收自己实际需要的依赖参数（构造函数直接注入）。

### 添加新功能
1. `internal/handler/{domain}/` 创建子包，`{domain}.go` 定义 Handler struct + `NewHandler(db, recorder, ...)`
2. `internal/service/{domain}/` 创建子包，`{domain}.go` 定义 Service struct + `NewService(db, ...)`
3. Converter 写在 `service/{domain}/converter.go`（同域 unexported，跨域 exported）
4. `internal/router/app.go` 的 Handlers struct 添加新 handler，`initHandlers` 中传显式参数
5. `internal/router/router.go` 注册路由
6. Swagger 注解写在 handler 方法上，`make swagger` 生成

### 中间件链顺序
RequestID → Logger → Recovery → CORS → RateLimit（可选）→ JWTAuth → RBAC → Audit（仅认证路由）
