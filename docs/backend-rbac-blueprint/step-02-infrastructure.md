# Step 02: 基础设施 — Config / Logger / Database / Redis

## 目标

引入配置管理、结构化日志、PostgreSQL 连接、Redis 连接，为后续业务提供基础设施。

## 前置条件

- Step 01 完成，项目能编译运行

## 文件清单

```
pkg/
├── config/
│   └── config.go              # Config 结构体 + 加载逻辑
├── database/
│   └── database.go            # PostgreSQL 连接 + GORM 初始化
└── logger/
    └── logger.go              # zerolog 初始化

internal/
└── router/
    └── app.go                 # App 结构体，持有基础设施依赖
```

## 实现细节

### 1. Config — YAML 配置

```go
// pkg/config/config.go
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    JWT      JWTConfig      `yaml:"jwt"`
}

type ServerConfig struct {
    Port int    `yaml:"port"`
    Mode string `yaml:"mode"` // debug/release
}

type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
    DBName   string `yaml:"dbname"`
    SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
    Addr     string `yaml:"addr"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
}

type JWTConfig struct {
    AccessSecret   string `yaml:"access_secret"`
    RefreshSecret  string `yaml:"refresh_secret"`
    AccessTTL      int    `yaml:"access_ttl"`    // 分钟
    RefreshTTL     int    `yaml:"refresh_ttl"`   // 分钟
    Issuer         string `yaml:"issuer"`
}
```

**设计决策**：
- 用 YAML 文件配置（`config.yaml`），环境变量覆盖（`CONFIG_PATH`）
- 不用 Viper，直接 `gopkg.in/yaml.v3` + `os.Getenv`，简单够用
- 敏感信息（密码、密钥）支持环境变量覆盖：`DB_PASSWORD` 覆盖 `database.password`

### 2. Logger — zerolog

```go
// pkg/logger/logger.go
// 初始化 zerolog：JSON 格式、带时间戳、日志级别从 config 读取
// 提供 Init(cfg *config.Config) 函数
// 日志一律一行写完：log.Info().Str("key","val").Msg("msg")
```

**关键约束**：日志必须写在一行，禁止多行链式调用（见 logging-style.md 规则）。

### 3. Database — PostgreSQL + GORM

```go
// pkg/database/database.go
// 初始化 GORM：
// 1. 连接 PostgreSQL
// 2. 配置连接池（MaxIdleConns=10, MaxOpenConns=100, ConnMaxLifetime=1h）
// 3. 启用 Logger（用 zerolog 适配 GORM logger 接口）
// 4. 返回 *gorm.DB
```

**设计决策**：
- GORM Gen 在 Step 04 引入，这里只建立 `*gorm.DB` 连接
- 连接池参数可配置

### 4. App 结构体

```go
// internal/router/app.go
type App struct {
    Config *config.Config
    DB     *gorm.DB
    RDB    *redis.Client  // 或这里先不留，Redis 在 auth 步骤引入
}

func NewApp() *App {
    // 1. 加载配置
    // 2. 初始化日志
    // 3. 连接数据库
    // 4. 返回 App
}
```

**设计决策**：
- App 是全局唯一的，持有所有基础设施依赖
- 不用依赖注入框架，手动构造
- main.go 调用 `app := NewApp()`，然后把 `app.DB` 传给需要的组件

### 5. main.go 更新

```go
func main() {
    app := router.NewApp()
    defer app.Close()

    log.Info().Msg("server starting")
    // 后续步骤在这里启动 Gin
}
```

## 依赖引入

```
go get gorm.io/gorm gorm.io/driver/postgres
go get github.com/rs/zerolog
go get github.com/redis/go-redis/v9
go get gopkg.in/yaml.v3
```

## 验收标准

- [ ] `config.yaml` 能正确加载（数据库连接、端口等）
- [ ] 环境变量 `DB_PASSWORD` 能覆盖配置文件中的密码
- [ ] `go build ./cmd/server` 编译通过
- [ ] 启动后日志输出 `server starting`（JSON 格式）
- [ ] 数据库连接成功（日志无错误）
- [ ] Redis 连接成功（如果配置了）
- [ ] 关闭程序时 `defer app.Close()` 正确释放连接

## AI 协作提示

```
请按 step-02-infrastructure.md 实现基础设施层。
参考现有 backend 的 pkg/config/、pkg/database/、pkg/utils/logger/ 的实现模式，
但不要直接复制代码。重新实现，保持简洁。
Config 用 YAML，Logger 用 zerolog，Database 用 GORM。
确保日志一行写完的规则。
```
