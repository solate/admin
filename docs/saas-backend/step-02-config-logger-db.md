# Step 02: 基础设施 — Config / Logger / Database / Redis

## 目标

实现配置加载、结构化日志、PostgreSQL 连接、Redis 连接。所有基础设施通过构造函数创建，main.go 做组装，并把已就绪的 db/redis/logger 直接接入 server。

## 前置条件

- Step 01 完成，`go build ./...` 通过
- PostgreSQL 可访问（本地或 Docker）
- Redis 可访问（本地或 Docker）

## 文件清单

```
pkg/xconfig/                      # 可复用加载库(整目录逐字复用到任意项目)
├── xconfig.go                    # Loader 结构体:New(opts) + (l).Load(target) + Override
├── options.go                    # functional options(WithFile/WithSearchPaths/WithFileName/WithFileType/WithDefaults)
└── xconfig_test.go               # Loader / Override 自包含测试

internal/config/                  # 项目专属配置层(新项目只改这里)
├── config.go                     # 类型化 Config + Load + validate
└── config_test.go                # 项目配置集成测试

internal/server/
└── server.go                     # Server 接收 Options，持有 db/rdb/log

pkg/logger/
├── config.go                     # logger.Config
└── logger.go                     # New(cfg Config) zerolog.Logger

pkg/database/
├── config.go                     # database.Config
└── database.go                   # New(cfg, log) (*gorm.DB, error) + GORM 日志适配器

pkg/rdb/
├── config.go                     # rdb.Config
└── rdb.go                        # New(cfg Config) (*redis.Client, error)

config/
└── config.yaml                   # 完整默认配置

cmd/server/main.go                # 组合根：配置 → 日志 → DB → Redis → server
```

> 敏感字段（密码、密钥）通过环境变量覆盖，不需要单独的 `config.dev.yaml`。

## 实现规范

### 1. Config 架构（可复用 Loader + 项目专属 Config）

**核心思想**：配置分两层。
- `pkg/xconfig` = **可复用加载库**：`Loader` 结构体用 viper 读文件并反序列化到调用方提供的类型化结构。这是"读取机制"，跨项目通用，**整目录逐字复用**。我们写的代码**零反射**（反序列化交给 viper/mapstructure）。
- `internal/config` = **项目专属**：本项目的类型化 `Config` 结构 + 显式环境变量覆盖 + 显式校验。**新项目只改这一层**。

为什么这样拆：`pkg/xconfig` 要做成整套后台框架的地基（跨项目复用、不重写读取逻辑），但用户又坚持**不用反射**（配置结构已知，反射是过度设计）。二者张力靠分层解决——"读取机制"通用、可复用、不反射；"Config 字段/env 名/校验"因项目而异、显式写在项目里。详见 `research/config-loading/03-封装设计.md`。

```go
// pkg/xconfig/xconfig.go (可复用,零反射)
package xconfig

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Loader 持有加载选项,可重复 Load;未来 Reload/Watch/Get 挂在 *Loader 上。
type Loader struct {
	file        string
	searchPaths []string
	fileName    string
	fileType    string
	defaults    map[string]any
}

func New(opts ...Option) *Loader {
	l := &Loader{fileName: "config", fileType: "yaml"}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Load 读配置文件并 viper 反序列化到 target(类型化结构指针)。
func (l *Loader) Load(target any) error {
	if target == nil {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	v := viper.New()
	if l.file != "" {
		v.SetConfigFile(l.file)
	} else {
		for _, p := range l.searchPaths {
			v.AddConfigPath(p)
		}
		v.SetConfigName(l.fileName)
		v.SetConfigType(l.fileType)
	}
	for k, val := range l.defaults {
		v.SetDefault(k, val)
	}
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	if err := v.Unmarshal(target); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}

// Override 用环境变量覆盖 *dst(类型化、无反射)。
// key 存在(即使空串)则覆盖;不存在则保持。仅 string。
func Override(dst *string, key string) {
	if dst == nil {
		return
	}
	if v, ok := os.LookupEnv(key); ok {
		*dst = v
	}
}
```

```go
// pkg/xconfig/options.go
package xconfig

type Option func(*Loader)

func WithFile(path string) Option        { return func(l *Loader) { l.file = path } }
func WithSearchPaths(p ...string) Option { return func(l *Loader) { l.searchPaths = append(l.searchPaths, p...) } }
func WithFileName(name string) Option    { return func(l *Loader) { l.fileName = name } }
func WithFileType(ft string) Option      { return func(l *Loader) { l.fileType = ft } }
func WithDefaults(d map[string]any) Option {
	return func(l *Loader) {
		if l.defaults == nil {
			l.defaults = make(map[string]any)
		}
		for k, v := range d {
			l.defaults[k] = v
		}
	}
}
```

```go
// internal/config/config.go (项目专属)
package config

import (
	"fmt"
	"time"

	"admin/pkg/xconfig"
)

// Config 应用总配置结构(项目专属)。仅 mapstructure tag(viper 用)。
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	Mode            string        `mapstructure:"mode"` // debug/release/test → Step 03 喂 gin.SetMode
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	GracefulTimeout time.Duration `mapstructure:"graceful_timeout"` // 优雅关闭超时
	Cors            CorsConfig    `mapstructure:"cors"`
}

type CorsConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 秒
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	AccessSecret  string `mapstructure:"access_secret"`
	RefreshSecret string `mapstructure:"refresh_secret"`
	AccessTTL     int    `mapstructure:"access_ttl"`  // 分钟
	RefreshTTL    int    `mapstructure:"refresh_ttl"` // 分钟
	Issuer        string `mapstructure:"issuer"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug/info/warn/error
	Format string `mapstructure:"format"` // json/console
}

// Load 从 path 加载配置:xconfig 读文件 → 显式环境变量覆盖敏感字段 → validate 校验。
func Load(path string) (*Config, error) {
	var cfg Config
	loader := xconfig.New(xconfig.WithFile(path))
	if err := loader.Load(&cfg); err != nil {
		return nil, err
	}
	// 显式环境变量覆盖(类型化、无反射、集中在项目层)
	xconfig.Override(&cfg.Database.Password, "DB_PASSWORD")
	xconfig.Override(&cfg.Redis.Password, "REDIS_PASSWORD")
	xconfig.Override(&cfg.JWT.AccessSecret, "JWT_ACCESS_SECRET")
	xconfig.Override(&cfg.JWT.RefreshSecret, "JWT_REFRESH_SECRET")
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return &cfg, nil
}

// validate 显式校验(无反射、无第三方校验库),覆盖全部关键字段。
func validate(c *Config) error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be in [1,65535], got %d", c.Server.Port)
	}
	switch c.Server.Mode {
	case "", "debug", "release", "test":
	default:
		return fmt.Errorf("server.mode invalid: %q", c.Server.Mode)
	}
	if c.Server.ReadTimeout <= 0 || c.Server.WriteTimeout <= 0 || c.Server.GracefulTimeout <= 0 {
		return fmt.Errorf("server timeouts must be positive")
	}
	if len(c.Server.Cors.AllowedOrigins) == 0 {
		return fmt.Errorf("server.cors.allowed_origins required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database.host required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("database.dbname required")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("database.port must be in [1,65535]")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("database.password required (set DB_PASSWORD or config)")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis.addr required")
	}
	if c.JWT.AccessSecret == "" || c.JWT.RefreshSecret == "" {
		return fmt.Errorf("jwt secrets required (set JWT_ACCESS_SECRET/JWT_REFRESH_SECRET)")
	}
	if c.JWT.AccessTTL < 1 || c.JWT.RefreshTTL < 1 {
		return fmt.Errorf("jwt ttl must be positive (minutes)")
	}
	switch c.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log.level invalid: %q", c.Log.Level)
	}
	switch c.Log.Format {
	case "json", "console":
	default:
		return fmt.Errorf("log.format invalid: %q", c.Log.Format)
	}
	return nil
}
```

**设计决策**：
- **可复用加载库 `pkg/xconfig`**（package `xconfig`，x 前缀）：`Loader` 结构体 + functional options，整目录逐字复用到任意项目。我们写的代码零反射（`viper.Unmarshal` 内部用 mapstructure 是 viper 库的事，本包不 import `reflect`）。
- **环境变量覆盖用显式类型化 `xconfig.Override`**（底层 `os.LookupEnv`），不用反射 / env tag / `viper.AutomaticEnv`。**不用 `AutomaticEnv`** 的原因：它与 `Unmarshal` 不兼容（走 `AllSettings()`，不查 `AutomaticEnv`），嵌套 key 覆盖会静默失效（详见 `research/config-loading/`）。`LookupEnv` 区分"未设置"与"设为空"——设为空串也算覆盖（可显式清空密钥）。
- **校验手写在项目层**（`internal/config.validate`），全覆盖；不用 `go-playground/validator`（那是反射），保持零反射依赖。
- `Config` 结构用 `mapstructure:` tag（viper.Unmarshal 用），值与 YAML key 一致。
- 不用全局单例：`Load` 返回 `*Config` 通过参数传递。
- **复用模型**：`pkg/xconfig/` 整目录逐字复用；新项目只改 `internal/config/`（Config 字段 + `Override` 的 env 名 + `validate` 规则）。`pkg/xconfig` 适用 `.claude/rules/reusable-package.md`（它是真可复用库），但仍是"轻量类型化 Loader"，不套用该规则的"泛型/反射驱动"模板。

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

	zerolog.SetGlobalLevel(level)

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
- `zerolog.SetGlobalLevel(level)` 设全局默认级别，方便第三方库（如 GORM 适配器）按级别过滤
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
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func New(cfg Config, log zerolog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newGormLogger(log),
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

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// gormZerologger 将 GORM SQL 日志桥接到 zerolog
type gormZerologger struct {
	log      zerolog.Logger
	logLevel gormlogger.LogLevel
}

func newGormLogger(log zerolog.Logger) gormlogger.Interface {
	return &gormZerologger{
		log:      log,
		logLevel: gormlogger.Info,
	}
}

func (l *gormZerologger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &gormZerologger{log: l.log, logLevel: level}
}

func (l *gormZerologger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.log.Info().Msgf(msg, args...)
}

func (l *gormZerologger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.log.Warn().Msgf(msg, args...)
}

func (l *gormZerologger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.log.Error().Msgf(msg, args...)
}

func (l *gormZerologger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	ev := l.log.Debug().Dur("elapsed", elapsed).Int64("rows", rows).Str("sql", sql)
	if err != nil {
		ev = l.log.Error().Err(err).Dur("elapsed", elapsed).Int64("rows", rows).Str("sql", sql)
	}
	ev.Msg("sql")
}
```

**GORM Logger 适配**：`gormZerologger` 实现 `gormlogger.Interface`（LogMode/Info/Warn/Error/Trace），把 SQL 执行日志统一接到 zerolog。`Trace` 里正常走 Debug、出错走 Error；err 分支重新构造 Event（未调 `.Msg()` 的 Debug Event 会被丢弃，行为正确）。

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

func New(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
```

### 5. Server 接入基础设施

`internal/server/server.go` 在 Step 02 起持有 db/redis/logger，通过 `Options` 传入：

```go
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"admin/internal/config"
)

type Server struct {
	httpSrv *http.Server
	db      *gorm.DB
	rdb     *redis.Client
	log     zerolog.Logger
	// Step 05 起追加：cronRunner *cron.Runner
}

type Options struct {
	Config *config.Config
	DB     *gorm.DB
	RDB    *redis.Client
	Log    zerolog.Logger
}

func New(opts Options) (*Server, error) {
	// Step 03 起替换为：engine := gin.New(); router.Setup(engine, ...)
	// gin.SetMode(opts.Config.Server.Mode) // Step 03 接入:把 Server.Mode 喂给 gin
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", opts.Config.Server.Port),
		Handler:      mux,
		ReadTimeout:  opts.Config.Server.ReadTimeout,
		WriteTimeout: opts.Config.Server.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}
	return &Server{
		httpSrv: httpSrv,
		db:      opts.DB,
		rdb:     opts.RDB,
		log:     opts.Log,
	}, nil
}

func (s *Server) Start() error {
	// Step 05 起追加：go s.cronRunner.Start(ctx)
	s.log.Info().Str("addr", s.httpSrv.Addr).Msg("server starting")
	if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) {
	// Step 05 起逆序追加：s.cronRunner.Stop()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		s.log.Error().Err(err).Msg("server shutdown error")
	}
}
```

**设计决策：为什么 Step 02 就把 db/redis/logger 接入 server**

延续 Step 01 的核心保证——**main.go 是最终结构，永不改动**。基础设施在 Step 02 就绪后，直接通过 `Options` 传入 server，于是：

- Step 03 只需把 `New()` 内部的 `mux` 换成 `gin.New()` + `router.Setup()`，server 签名和 main.go 都不动
- Step 05 cron 同理，只在 `Start()`/`Stop()` 内部追加，main.go 无感知

如果把 server 的接入推迟到 Step 03，main.go 就要在 Step 03 再改一次（从「不启动 server」变成「启动 server」），破坏稳定性保证。

### 6. main.go — 组合根

```go
package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"admin/internal/config"
	"admin/internal/server"
	"admin/pkg/database"
	"admin/pkg/logger"
	"admin/pkg/rdb"
)

func main() {
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

	// 5. 启动 HTTP 服务器
	srv, err := server.New(server.Options{
		Config: cfg,
		DB:     db,
		RDB:    rdbClient,
		Log:    log,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("init server failed")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("start server failed")
		}
	}()

	<-ctx.Done()
	stop()
	log.Info().Msg("shutdown signal received")

	// 优雅关闭超时取自配置(原为硬编码 30s)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulTimeout)
	defer cancel()
	srv.Stop(shutdownCtx)
	log.Info().Msg("server exited")
}
```

**关键设计**：
- `internal/config.Load` 返回 `*config.Config`,main 直接用;`config.Config` → 各 `pkg.Config` 的映射在 main 中完成
- 各 pkg 不知道 YAML 的存在，只接收自己的 Config struct —— 换项目只需换 YAML 和 main 的映射
- 配置加载失败用 `panic`（此时 logger 尚未初始化）；其余基础设施失败用 `log.Fatal`
- `signal.NotifyContext` + `cfg.Server.GracefulTimeout` 优雅退出超时（取自配置，不再硬编码）

### 7. config/config.yaml

```yaml
server:
  port: 8080
  mode: debug
  read_timeout: 10s
  write_timeout: 10s
  graceful_timeout: 30s
  cors:
    allowed_origins: ["http://localhost:5173"]
    allowed_methods: [GET, POST, PUT, DELETE, OPTIONS]
    allowed_headers: [Content-Type, Authorization]
    allow_credentials: true

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
go get github.com/spf13/viper        # v1.21.0（YAML 解析由其间接依赖 go.yaml.in/yaml/v3 提供）
go get github.com/rs/zerolog         # v1.35.0
go get gorm.io/gorm                  # v1.31.1
go get gorm.io/driver/postgres       # v1.6.0
go get github.com/redis/go-redis/v9  # v9.18.0
```

## 验收标准

```bash
cd backend

# 1. 全包编译通过
go build ./...
# 期望：无错误

# 2. 各 pkg 独立可编译
go build ./pkg/xconfig && go build ./pkg/database && go build ./pkg/logger && go build ./pkg/rdb
# 期望：各自独立编译无错误

# 3. 配置层单测全绿（无反射、无外部服务）
go test ./pkg/xconfig/ ./internal/config/ -v
# 期望：Loader / Override / 项目 Load(validate、env 覆盖) 全部 PASS

# 4. 配置加载 + 基础设施初始化（需 PG / Redis 在跑）
go run ./cmd/server -config config/config.yaml &
sleep 1
curl -s http://localhost:8080/health
# 期望：{"status":"ok"}
#       且日志输出 "all infrastructure initialized"
kill %1
# 若 PG / Redis 未启动，期望看到明确的 "connect database failed" / "connect redis failed"

# 5. 环境变量覆盖敏感字段
DB_PASSWORD=wrong go run ./cmd/server 2>&1 | grep "connect database"
# 期望：连接失败（密码被覆盖）

# 6. 日志格式（debug + console 模式）
go run ./cmd/server 2>&1 | head -3
# 期望：带颜色和 RFC3339 时间的 console 格式输出
```

## AI 协作提示

```
请按 step-02-config-logger-db.md 实现基础设施层。

要点：
1. pkg/xconfig/xconfig.go 可复用 Loader 结构体:New(opts) + (l).Load(target any) + Override(dst,key);零反射(不 import reflect)
2. pkg/xconfig/options.go functional options:WithFile/WithSearchPaths/WithFileName/WithFileType/WithDefaults
3. pkg/xconfig/xconfig_test.go Loader/Override 自包含测试(t.TempDir/t.Setenv)
4. internal/config/config.go 项目专属:类型化 Config(mapstructure tag)+ Load(调 xconfig + Override + validate)+ 手写 validate 全覆盖
5. internal/config/config_test.go 项目集成测试(env 覆盖、validate 失败)
6. pkg/logger/ 有自己的 Config struct,New() 返回 zerolog.Logger
7. pkg/database/ 有自己的 Config struct,New() 返回 *gorm.DB,含 zerolog 适配的 GORM logger
8. pkg/rdb/ 有自己的 Config struct,New() 返回 *redis.Client
9. cmd/server/main.go 调 config.Load,做 config.Config → 各 pkg.Config 映射;关闭超时用 cfg.Server.GracefulTimeout
10. 不要用全局变量,所有依赖通过参数传递;绝不用 viper.AutomaticEnv
11. GORM Logger 适配 zerolog(实现 gormlogger.Interface)
12. config/config.yaml 写完整默认配置(含 graceful_timeout + cors)
13. server 通过 Options 接收 db/rdb/log,main.go 启动 server
```

---

*上一步：[Step 01 - 项目骨架](step-01-project-scaffolding.md) | 下一步：[Step 03 - HTTP 框架](step-03-http-framework.md)*
