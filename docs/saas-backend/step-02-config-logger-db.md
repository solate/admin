# Step 02: 基础设施 — Config / Logger / Database / Redis

## 目标

实现配置加载、结构化日志、PostgreSQL 连接、Redis 连接。所有基础设施通过构造函数创建，main.go 做组装，并把已就绪的 db/redis/logger 直接接入 server。

## 前置条件

- Step 01 完成，`go build ./...` 通过
- PostgreSQL 可访问（本地或 Docker）
- Redis 可访问（本地或 Docker）

## 文件清单

```
internal/config/                  # 自包含配置层(内联 viper 加载;新微服务复制整目录、只改 struct)
├── config.go                     # 类型化 Config + Load + 私有 loadFromYAML/override + validate
└── config_test.go                # 加载器行为 + 项目集成测试

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
├── config.yaml                   # base：完整默认配置（即 dev 默认值）
└── config.prod.yaml              # 生产 overlay：只写与 base 不同的字段（APP_ENV=prod 触发）

cmd/server/main.go                # 组合根：配置 → 日志 → DB → Redis → server
```

> **多环境（base + overlay）**：base `config.yaml` 承载共享默认值（即 dev 配置）；每个环境的差异放进 `config.{env}.yaml` overlay，由 `APP_ENV` 触发合并（见 §7.1）。**敏感字段（密码、密钥）始终走环境变量覆盖**，不进任何配置文件——overlay 只放非密钥的环境差异（mode / log 级别与格式 / 域名 / 库名等）。

## 实现规范

### 1. Config 架构（自包含单包：内联 viper 加载 + 项目专属 Config）

**核心思想**：配置全部在一个自包含包 `internal/config` 里。
- **内联的 viper 加载机制**（私有，不随项目变）：`loadFromYAML(target, paths...)` 用 viper 读 base（+ overlay 依次合并）并解码到类型化结构（走 yaml tag）；`override(dst, key)` 显式 env 覆盖。我们写的代码**零反射**（反序列化交给 viper）。
- **项目专属**（新微服务只改这里）：类型化 `Config` 结构 + `Load`（拼路径 + 调 `loadFromYAML` + `override` 4 个密钥）+ `validate`。

为什么是单包：用户坚持**不用反射**（配置结构已知，反射是过度设计），所以 env 覆盖/校验是显式类型化代码、因项目而异；而"读取机制"（viper 读文件 + 反序列化）跨项目通用。早期版本（v4/v5）把通用机制抽成独立可复用库 `pkg/xconfig`，但落地发现复用模式是"**复制**"不是"import"（各微服务独立 module），独立库收益=0，于是 v6 合并回单包——加载机制收为私有函数。详见 `research/config-loading/03-封装设计.md`（v6 演进）。

```go
// internal/config/config.go (自包含:内联 viper 加载 + 项目专属 Config)
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// Config 应用总配置结构(项目专属)。仅 yaml tag(loadFromYAML 让 viper 读 yaml tag)。
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Port            int           `yaml:"port"`
	Mode            string        `yaml:"mode"` // debug/release/test → Step 03 喂 gin.SetMode
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	GracefulTimeout time.Duration `yaml:"graceful_timeout"` // 优雅关闭超时
	Cors            CorsConfig    `yaml:"cors"`
}

type CorsConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
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
	AccessTTL     int    `yaml:"access_ttl"`  // 分钟
	RefreshTTL    int    `yaml:"refresh_ttl"` // 分钟
	Issuer        string `yaml:"issuer"`
}

type LogConfig struct {
	Level  string `yaml:"level"`  // debug/info/warn/error
	Format string `yaml:"format"` // json/console
}

// loadFromYAML 把 paths[0] 作为 base 读入,paths[1:] 作为 overlay 依次 MergeInConfig 合并,
// 最后解码到 target(yaml tag + duration)。私有,不随项目变。
// gotcha(测试钉死):不用 AutomaticEnv;嵌套 map 深合并;标量与 slice 由 overlay 整体替换(slice 不追加)。
func loadFromYAML(target any, paths ...string) error {
	if len(paths) == 0 {
		return fmt.Errorf("at least one config path required")
	}
	if target == nil {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	v := viper.New()
	v.SetConfigFile(paths[0])
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config %s: %w", paths[0], err)
	}
	for _, p := range paths[1:] {
		v.SetConfigFile(p)
		if err := v.MergeInConfig(); err != nil {
			return fmt.Errorf("merge config %s: %w", p, err)
		}
	}
	// TagName="yaml" → 让 viper 读 yaml tag(而非默认 mapstructure tag);不影响 decode hook,duration 解码保留。
	if err := v.Unmarshal(target, func(d *mapstructure.DecoderConfig) { d.TagName = "yaml" }); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}

// override 用环境变量覆盖 *dst(类型化、无反射)。key 存在(即使空串)则覆盖;不存在则保持。仅 string。
func override(dst *string, key string) {
	if dst == nil {
		return
	}
	if v, ok := os.LookupEnv(key); ok {
		*dst = v
	}
}

// Load 从 basePath 加载配置:loadFromYAML 读 base(+可选 APP_ENV overlay)→ 显式环境变量覆盖敏感字段 → validate 校验。
func Load(basePath string) (*Config, error) {
	paths := []string{basePath}
	if env := os.Getenv("APP_ENV"); env != "" {
		paths = append(paths, filepath.Join(filepath.Dir(basePath), "config."+env+".yaml"))
	}
	var cfg Config
	if err := loadFromYAML(&cfg, paths...); err != nil {
		return nil, err
	}
	// 显式环境变量覆盖(类型化、无反射、集中在项目层)
	override(&cfg.Database.Password, "DB_PASSWORD")
	override(&cfg.Redis.Password, "REDIS_PASSWORD")
	override(&cfg.JWT.AccessSecret, "JWT_ACCESS_SECRET")
	override(&cfg.JWT.RefreshSecret, "JWT_REFRESH_SECRET")
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
- **自包含单包 `internal/config`**：viper 加载机制内联为私有 `loadFromYAML(target, paths...)` + `override(dst, key)`，与项目专属 `Config`/`Load`/`validate` 同在一个包。我们写的代码零反射（`viper.Unmarshal` 内部用 mapstructure 是 viper 库的事，本包不 import `reflect`）。
- **环境变量覆盖用显式类型化 `override`**（底层 `os.LookupEnv`），不用反射 / env tag / `viper.AutomaticEnv`。**不用 `AutomaticEnv`** 的原因：它与 `Unmarshal` 不兼容（走 `AllSettings()`，不查 `AutomaticEnv`），嵌套 key 覆盖会静默失效（详见 `research/config-loading/`）。`LookupEnv` 区分"未设置"与"设为空"——设为空串也算覆盖（可显式清空密钥）。
- **校验手写在项目层**（`internal/config.validate`），全覆盖；不用 `go-playground/validator`（那是反射），保持零反射依赖。
- `Config` 结构用 `yaml:` tag（`loadFromYAML` 经 `DecoderConfigOption{TagName="yaml"}` 让 viper 读 yaml tag），值与 YAML key 一致。
- 不用全局单例：`Load` 返回 `*Config` 通过参数传递。
- **复用模型**：新微服务**复制整个 `internal/config/` 目录**，只改 `Config` 字段 + `override` 的 env 名 + `validate` 规则；`loadFromYAML`/`override` 是不随项目变的 viper 样板，原样保留。复用模式是"复制"而非"import"（各微服务独立 module）——这正是 v6 把加载机制从独立 `pkg/xconfig` 合并回 `internal/config` 的原因（详见 `03-封装设计.md` v6 演进）。

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

### 7.1 多环境：base + overlay 合并

`config.yaml` 是 **base**（共享默认值，内容即 dev 配置）。每个环境的差异放进同目录的 `config.{env}.yaml`，由环境变量 **`APP_ENV`** 触发叠加：

- 未设 `APP_ENV` → 仅 base（即 dev，零意外）。
- `APP_ENV=prod` → base + `config/config.prod.yaml` 合并（overlay 只写差异行）。
- `APP_ENV` 设了但 overlay 文件缺失 → **fail-fast**（merge 报错，启动前暴露）。

**合并语义**（viper `MergeInConfig`）：嵌套 map **深合并**；标量与 slice 用 overlay 值**整体替换** base（slice 不追加）。故 overlay 里凡要改的 list（如 `cors.allowed_origins`）必须写全。密钥仍走环境变量，不写进 overlay。

`config/config.prod.yaml`（示例，密钥留 env）：

```yaml
server:
  mode: release
  cors:
    allowed_origins: ["https://admin.example.com"]   # slice 整体替换，写全
database:
  dbname: admin_prod
log:
  level: info
  format: json
  add_source: true
```

运行：

```bash
go run ./cmd/server -config config/config.yaml                          # dev（仅 base）
APP_ENV=prod DB_PASSWORD=... go run ./cmd/server -config config/config.yaml   # prod（base + overlay）
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
go build ./pkg/database && go build ./pkg/logger && go build ./pkg/rdb && go build ./internal/config
# 期望：各自独立编译无错误

# 3. 配置层单测全绿（无反射、无外部服务）
go test ./internal/config/ -v
# 期望：loadFromYAML(加载器行为) / override / 项目 Load(validate、env 覆盖、overlay 合并) 全部 PASS

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

# 5.5 多环境 overlay（base + APP_ENV）
APP_ENV=prod DB_PASSWORD=postgres JWT_ACCESS_SECRET=a JWT_REFRESH_SECRET=b \
  go run ./cmd/server -config config/config.yaml 2>&1 | head -1
# 期望：JSON 日志 + 连 admin_prod 库（prod overlay 的 json 格式与 dbname 合并生效）
APP_ENV=staging go run ./cmd/server -config config/config.yaml 2>&1 | head -1
# 期望：panic "merge config config/config.staging.yaml: ..."（overlay 缺失 fail-fast）

# 6. 日志格式（debug + console 模式）
go run ./cmd/server 2>&1 | head -3
# 期望：带颜色和 RFC3339 时间的 console 格式输出
```

## AI 协作提示

```
请按 step-02-config-logger-db.md 实现基础设施层。

要点：
1. internal/config/config.go 自包含单文件:类型化 Config(yaml tag)+ 私有 loadFromYAML(target,paths...)(viper 读 base+overlay 合并,零反射)+ 私有 override(dst,key)(os.LookupEnv)+ Load(调 loadFromYAML + override + validate)+ 手写 validate 全覆盖
2. internal/config/config_test.go 自包含测试(t.TempDir/t.Setenv):loadFromYAML 加载器行为(含 overlay 合并 gotcha)+ override(set-to-empty)+ 项目 Load(validate、env 覆盖)
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
