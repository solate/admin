package logger

// Config 是日志层的全部配置。字段值来自 internal/config.LogConfig。
type Config struct {
	Level     string // debug/info/warn/error
	Format    string // json(生产) / text / console —— 后两者均走 TextHandler
	AddSource bool   // 记录调用位置 file:line
}
