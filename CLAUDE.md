# CLAUDE.md

此文件为 Claude Code (claude.ai/code) 在此代码库中工作时提供指导。

## 开发命令

### 快速开始
```bash
# 首次设置
make init                    # 安装依赖、迁移数据库、生成代码

# 开发工作流
make dev                     # 运行 migrate + gen-db + 启动服务器
make run                     # 仅运行服务器（不执行迁移/生成）
```

### 数据库操作
```bash
make migrate-up              # 应用待执行的迁移
make migrate-down            # 回滚一个迁移
make migrate-reset           # 完全重置数据库（破坏性操作）
make migrate-create NAME=xxx # 创建新的迁移文件
make gen-db                  # 从数据库架构生成 GORM 模型
```

### 代码质量与测试
```bash
make test                    # 运行所有测试并生成覆盖率报告
make lint                    # 运行 golangci-lint 代码检查
make fmt                     # 格式化代码并整理 go.mod
make swagger                 # 生成 Swagger 文档
```

### 数据库配置
设置以下环境变量可覆盖默认配置：
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=root
export DB_PASSWORD=root
export DB_NAME=admin_db      # Makefile 中默认为 content_center
```

## 架构概览

这是一个基于 Go 构建的**多租户管理系统**，采用**数据库优先**的开发方法。

### 核心架构模式
- **清洁架构**：handler → service → repository 分层结构
- **多租户**：通过租户上下文注入实现数据隔离
- **数据库优先**：编写迁移 → 生成模型 → 实现业务逻辑
- **代码生成**：使用 GORM Gen 生成类型安全的查询

### 关键技术栈
- **Web 框架**：Gin v1.11.0
- **ORM**：GORM v1.31.1 配合 GORM Gen 进行代码生成
- **数据库**：PostgreSQL，支持软删除
- **身份认证**：JWT 配合 Redis 黑名单
- **权限授权**：Casbin RBAC，支持多租户（域）
- **缓存**：Redis 用于令牌管理
- **日志**：Zerolog 结构化日志
- **配置**：Viper 支持环境感知的配置管理

### 项目结构
```
backend/
├── cmd/server/              # 应用程序入口点
├── internal/
│   ├── constants/           # 系统常量（状态码等）
│   ├── dal/                # 数据访问层
│   │   ├── model/          # 生成的模型
│   │   └── query/          # 生成的查询（gen.go）
│   ├── dto/                # 数据传输对象
│   ├── handler/            # HTTP 处理器/控制器
│   ├── middleware/         # HTTP 中间件链
│   ├── service/            # 业务逻辑层
│   └── router/             # 路由定义
├── pkg/                    # 可复用的包
├── config/                 # 配置文件
├── migrations/             # 数据库迁移文件
└── scripts/                # 实用脚本
```

## 开发工作流

### 1. 数据库架构变更
1. 创建迁移：`make migrate-create NAME=add_feature_table`
2. 在 `migrations/xxx_add_feature_table.up.sql` 中编写 SQL
3. 应用迁移：`make migrate-up`
4. 生成代码：`make gen-db`

### 2. 使用生成的代码
- 模型生成在 `internal/dal/model/`
- 查询生成在 `internal/dal/query/gen.go`
- 使用生成的查询进行类型安全的数据库操作
- **切勿直接编辑生成的文件** - 它们会被覆盖

### 3. 添加新功能
1. 遵循分层架构：Handler → Service → Repository
2. 使用 `main.go` 中的依赖注入模式
3. 使用自定义错误类型实现适当的错误处理
4. 按正确顺序将中间件添加到链中
5. 更新 Swagger 文档

### 4. 身份认证与授权
- JWT 令牌：访问令牌（1小时）+ 刷新令牌（7天）
- 中间件顺序：Logger → Recovery → CORS → Auth → Casbin
- Casbin 策略格式：`sub, dom, obj, act`（用户、租户、资源、操作）
- 租户上下文由中间件自动注入

## 配置管理

### 环境层次结构
1. `config/config.yaml` - 基础配置
2. `config/config.{env}.yaml` - 环境特定覆盖
3. 环境变量 - 最高优先级

### 环境检测
- `APP_ENV` > `GIN_MODE` > 配置文件值 > "dev"
- `GIN_MODE=release` 映射为 "prod" 环境

### 主要配置部分
- 数据库连接池
- Redis 集群与单节点配置
- JWT 密钥和过期时间
- 日志级别和格式（json/console）
- 限流配置

## 重要实现细节

### 多租户支持
- 所有数据库查询都应限定在租户范围内
- 租户 ID 从 JWT 中提取并注入到上下文
- 使用租户过滤的生成查询
- Casbin 策略是租户隔离的

### 中间件链顺序
对正常运行至关重要：
1. RequestID（追踪）
2. Logger（请求日志）
3. Recovery（panic 处理）
4. CORS（跨域）
5. RateLimit（可选）
6. JWTAuth（身份认证）
7. Casbin（权限授权）

### 错误处理
- 使用 `pkg/errors/` 中的自定义错误类型
- 通过 `pkg/response/` 实现一致的 JSON 响应格式
- 适当的 HTTP 状态码和错误详情
- 使用适当的上下文记录错误

### 数据库模式
- 全局启用软删除
- 使用生成的查询确保类型安全
- 按环境配置连接池
- 迁移文件版本化和顺序化

## 测试

### 测试结构
- 单元测试：源代码文件旁的 `*_test.go` 文件
- 集成测试：测试数据库操作
- 当前覆盖率：基础设施组件（JWT、配置、日志）

### 运行测试
```bash
make test                    # 运行所有测试
go test ./pkg/...           # 测试特定包
go test -v ./internal/...   # 详细输出
```

## 代码生成

### GORM Gen
- 生成类型安全的数据库查询
- 架构变更后运行：`make gen-db`
- 默认排除 `schema_migrations` 表
- 模型和查询都会重新生成

### Swagger
- 从注解生成 API 文档
- 使用 `make swagger` 生成
- 输出目录：`docs/`
- 访问地址：`/swagger/index.html`

## 必需的开发工具

安装这些工具以获得完整的开发体验：
```bash
# 数据库迁移
brew install golang-migrate

# Swagger 文档
go install github.com/swaggo/swag/cmd/swag@latest

# 代码检查
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```