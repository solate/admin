# Step 02: 基础设施 — Config / Logger / Database / Redis

## 目标

实现配置加载、结构化日志、PostgreSQL 连接、Redis 连接。所有基础设施通过构造函数创建，main.go 做组装。

## 前置条件

- Step 01 完成，`go build ./cmd/server` 通过
- PostgreSQL 可访问（本地或 Docker）
- Redis 可访问（本地或 Docker）

## 文件清单

```
internal/config/
├── config.go              # Config 总结构体
└── load.go                # Load(path) (*Config, error)

pkg/logger/
├── config.go              # logger.Config
└── logger.go              # New(cfg Config) zerolog.Logger

pkg/database/
├── config.go              # database.Config
└── database.go            # New(cfg Config, log zerolog.Logger) (*gorm.DB, error)

pkg/rdb/
├── config.go              # rdb.Config
└── rdb.go                 # New(cfg Config) (*redis.Client, error)

config/
├── config.yaml            # 完整默认配置
└── config.dev.yaml        # 开发环境覆盖

cmd/server/main.go         # 更新：加载配置 → 初始化基础设施
```

## 实现规范

### 1. Config 架构（B+C 模式）

**核心思想**：`internal/config` 定义总配置结构，各 pkg 定义自己的 Config struct，main.go 做映射。

```go
// internal/config/config.go
package config

import "time"

type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    JWT      JWTConfig      `yaml:"jwt"`
    Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
    Port         int           `yaml:"port"`
    Mode         string        `yaml:"mode"`          // debug / release
    ReadTimeout  time.Duration `yaml:"read_timeout"`  // 10s
    WriteTimeout time.Duration `yaml:"write_timeout"` // 10s
}

type DatabaseConfig struct {
    Host            string `yaml:"host"`
    Port            int    `yaml:"port"`
    User            string `yaml:"user"`
    Password        string `yaml:"password"`
    DBName          string `yaml:"dbname"`
    SSLMode         string `yaml:"sslmode"`
    MaxIdleConns    int    `yaml:"max_idle_conns"`
    MaxOpenConns    int    `yaml:"max_open_conns"`
    ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // 秒
}

type RedisConfig struct {
    Addr     string `yaml:"addr"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
}

type JWTConfig struct {
    AccessSecret  string `yaml:"access_secret"`
    RefreshSecret string `yaml:"refresh_secret"`
    AccessTTL     int    `yaml:"access_ttl"`     // 分钟
    RefreshTTL    int    `yaml:"refresh_ttl"`    // 分钟
    Issuer        string `yaml:"issuer"`
}

type LogConfig struct {
    Level  string `yaml:"level"`  // debug/info/warn/error
    Format string `yaml:"format"` // json/console
}
```

```go
// internal/config/load.go
package config

import (
    "fmt"
    "os"
    "gopkg.in/yaml.v3"
)

// Load 加载配置文件，支持环境变量覆盖敏感字段
func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }

    cfg := &Config{}
    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }

    // 环境变量覆盖（敏感字段）
    if v := os.Getenv("DB_PASSWORD"); v != "" {
        cfg.Database.Password = v
    }
    if v := os.Getenv("REDIS_PASSWORD"); v != "" {
        cfg.Redis.Password = v
    }
    if v := os.Getenv("JWT_ACCESS_SECRET"); v != "" {
        cfg.JWT.AccessSecret = v
    }
    if v := os.Getenv("JWT_REFRESH_SECRET"); v != "" {
        cfg.JWT.RefreshSecret = v
    }

    return cfg, nil
}
```

**设计决策**：
- 不用 Viper：yaml.v3 + os.Getenv 足够，没有隐藏魔法
- 不用 `config.Get()` 全局单例：Load 返回值通过参数传递
- 环境变量只覆盖敏感字段（密码、密钥），其他配置走 YAML

### 2. Logger — zerolog

```go
// pkg/logger/config.go
package logger

type Config struct {
    Level  string // debug/info/warn/error
    Format string // json/console
}
```

```go
// pkg/logger/logger.go
package logger

import (
    "io"
    "os"
    "time"
    "github.com/rs/zerolog"
)

// New 创建 zerolog.Logger 实例
func New(cfg Config) zerolog.Logger {
    var output io.Writer = os.Stdout

    if cfg.Format == "console" {
        output = zerolog.ConsoleWriter{
            Out:        os.Stdout,
            TimeFormat: time.RFC3339,
        }
    }

    level, err := zerolog.ParseLevel(cfg.Level)
    if err != nil {
        level = zerolog.InfoLevel
    }

    return zerolog.New(output).
        Level(level).
        With().
        Timestamp().
        Caller().
        Logger()
}
```

**关键约束**：
- Logger 通过参数传递，不设全局 `log` 变量
- 但可以设置 zerolog 的全局默认（`zerolog.SetGlobalLevel`），方便第三方库
- 日志一行写完：`log.Info().Str("k","v").Msg("msg")`

### 3. Database — PostgreSQL + GORM

```go
// pkg/database/config.go
package database

type Config struct {
    Host            string
    Port            int
    User            string
    Password        string
    DBName          string
    SSLMode         string
    MaxIdleConns    int
    MaxOpenConns    int
    ConnMaxLifetime int // 秒
}
```

```go
// pkg/database/database.go
package database

import (
    "fmt"
    "time"

    "github.com/rs/zerolog"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    gormlogger "gorm.io/gorm/logger"
)

// New 创建 GORM DB 实例
func New(cfg Config, log zerolog.Logger) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: newGormLogger(log), // zerolog 适配 GORM logger
    })
    if err != nil {
        return nil, fmt.Errorf("connect database: %w", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("get sql.DB: %w", err)
    }

    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

    return db, nil
}

// Close 关闭数据库连接
func Close(db *gorm.DB) error {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}
```

**GORM Logger 适配**：实现 `gormlogger.Interface`，将 SQL 日志输出到 zerolog。

### 4. Redis

```go
// pkg/rdb/config.go
package rdb

type Config struct {
    Addr     string
    Password string
    DB       int
}
```

```go
// pkg/rdb/rdb.go
package rdb

import (
    "context"
    "fmt"
    "github.com/redis/go-redis/v9"
)

// New 创建 Redis 客户端
func New(cfg Config) (*redis.Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     cfg.Addr,
        Password: cfg.Password,
        DB:       cfg.DB,
    })

    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("ping redis: %w", err)
    }

    return client, nil
}
```

### 5. main.go 更新 — 组合根

```go
package main

import (
    "context"
    "flag"
    "os"
    "os/signal"
    "syscall"
    "time"

    "admin/internal/config"
    "admin/pkg/database"
    "admin/pkg/logger"
    "admin/pkg/rdb"
)

func main() {
    // 解析命令行参数
    configPath := flag.String("config", "config/config.yaml", "配置文件路径")
    flag.Parse()

    // 1. 加载配置
    cfg, err := config.Load(*configPath)
    if err != nil {
        panic("load config: " + err.Error())
    }

    // 2. 初始化日志
    log := logger.New(logger.Config{
        Level:  cfg.Log.Level,
        Format: cfg.Log.Format,
    })

    // 3. 连接数据库
    db, err := database.New(database.Config{
        Host:            cfg.Database.Host,
        Port:            cfg.Database.Port,
        User:            cfg.Database.User,
        Password:        cfg.Database.Password,
        DBName:          cfg.Database.DBName,
        SSLMode:         cfg.Database.SSLMode,
        MaxIdleConns:    cfg.Database.MaxIdleConns,
        MaxOpenConns:    cfg.Database.MaxOpenConns,
        ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
    }, log)
    if err != nil {
        log.Fatal().Err(err).Msg("connect database failed")
    }
    defer database.Close(db)

    // 4. 连接 Redis
    rdbClient, err := rdb.New(rdb.Config{
        Addr:     cfg.Redis.Addr,
        Password: cfg.Redis.Password,
        DB:       cfg.Redis.DB,
    })
    if err != nil {
        log.Fatal().Err(err).Msg("connect redis failed")
    }
    defer rdbClient.Close()

    log.Info().Int("port", cfg.Server.Port).Msg("all infrastructure initialized")

    // 5. TODO: 启动 HTTP 服务器（Step 03）

    // 优雅退出
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Info().Msg("shutting down...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = ctx // Step 03 用于 server.Shutdown(ctx)

    log.Info().Msg("server exited")
}
```

**关键设计**：
- `internal/config.Config` → 各 `pkg.Config` 的映射在 main 中完成
- 各 pkg 不知道 YAML 的存在，只接收自己的 Config struct
- 这种方式让 pkg 完全可复用（换一个项目只需换 YAML 和 main 的映射）

### 6. config/config.yaml

```yaml
server:
  port: 8080
  mode: debug
  read_timeout: 10s
  write_timeout: 10s

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: admin_dev
  sslmode: disable
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600

redis:
  addr: localhost:6379
  password: ""
  db: 0

jwt:
  access_secret: "your-access-secret-change-in-production"
  refresh_secret: "your-refresh-secret-change-in-production"
  access_ttl: 30
  refresh_ttl: 10080
  issuer: "admin"

log:
  level: debug
  format: console
```

## 依赖引入

```bash
go get gopkg.in/yaml.v3
go get github.com/rs/zerolog
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/redis/go-redis/v9
```

## 验收标准

```bash
# 1. 编译通过
go build ./cmd/server
# 期望：无错误

# 2. 配置加载（有 config.yaml）
go run ./cmd/server -config config/config.yaml
# 期望：日志输出 "all infrastructure initialized"
# 如果数据库/Redis 没启动，期望看到明确错误信息

# 3. 环境变量覆盖
DB_PASSWORD=wrong go run ./cmd/server 2>&1 | grep "connect database"
# 期望：连接失败（密码被覆盖了）

# 4. 日志格式
go run ./cmd/server 2>&1 | head -3
# 期望：debug 模式下有 console 格式输出（带颜色和时间）

# 5. 各 pkg 独立可编译
go build ./pkg/database
go build ./pkg/logger
go build ./pkg/rdb
# 期望：各自独立编译无错误
```

## AI 协作提示

```
请按 step-02-config-logger-db.md 实现基础设施层。

要点：
1. internal/config/config.go 定义总 Config 结构体
2. internal/config/load.go 用 yaml.v3 加载 + 环境变量覆盖敏感字段
3. pkg/logger/ 有自己的 Config struct，New() 返回 zerolog.Logger
4. pkg/database/ 有自己的 Config struct，New() 返回 *gorm.DB
5. pkg/rdb/ 有自己的 Config struct，New() 返回 *redis.Client
6. cmd/server/main.go 做配置映射（internal/config → 各 pkg Config）
7. 不要用全局变量，所有依赖通过参数传递
8. GORM Logger 适配 zerolog（实现 gormlogger.Interface）
9. config/config.yaml 写完整默认配置
```

---

*上一步：[Step 01 - 项目骨架](step-01-project-scaffolding.md) | 下一步：[Step 03 - HTTP 框架](step-03-http-framework.md)*
