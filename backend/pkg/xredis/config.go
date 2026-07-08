package xredis

// Config Redis 连接参数。无 mapstructure tag，由 main 从 internal/config 逐字段映射。
type Config struct {
	Addr         string // 地址 host:port
	Password     string // 密码（建议 APP_REDIS_PASSWORD 注入）
	DB           int    // 库编号
	PoolSize     int    // 连接池大小，0 用 go-redis 默认
	MinIdleConns int    // 最小空闲连接，0 用默认
	MaxRetries   int    // 命令重试次数，0 用默认
	DialTimeout  int    // 建连超时(秒)，0 用默认
	ReadTimeout  int    // 读超时(秒)，0 用默认
	WriteTimeout int    // 写超时(秒)，0 用默认
}
