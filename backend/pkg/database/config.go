package database

// Config 数据库连接参数（纯连接配置，不含日志字段）。
// 由 internal/config.DatabaseConfig 逐字段映射而来（见 cmd/server/main.go）。
// SQL 日志的可见性由 slog 的 cfg.Log.Level 统一控制，不在此单独配置。
type Config struct {
	Host            string // 数据库主机地址
	Port            int    // 端口
	User            string // 用户名
	Password        string // 密码
	DBName          string // 库名
	SSLMode         string // SSL 模式：disable / require / verify-full 等
	MaxIdleConns    int    // 连接池最大空闲连接数
	MaxOpenConns    int    // 连接池最大打开连接数
	ConnMaxLifetime int    // 连接最大存活时长（秒）
}
