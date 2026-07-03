package xviper

// settings 是解析后的内部配置(小写不导出)。
type settings[T any] struct {
	path     string
	validate func(cfg *T) error
}

// defaultSettings 返回内建约定默认值:基础配置文件 config/config.yaml。
func defaultSettings[T any]() settings[T] {
	return settings[T]{
		path: "config/config.yaml",
	}
}

// Option 修改单个配置项;不传则全部走 defaultSettings 约定。
type Option[T any] func(*settings[T])

// WithPath 指定基础配置文件路径。默认 config/config.yaml。
func WithPath[T any](p string) Option[T] {
	return func(s *settings[T]) { s.path = p }
}

// WithValidate 设置反序列化后的校验钩子。返回 error 则 Load 失败。
func WithValidate[T any](fn func(*T) error) Option[T] {
	return func(s *settings[T]) { s.validate = fn }
}
