package xconfig

// Option 是配置 [Loader] 的 functional option。
type Option func(*Loader)

// WithFile 指定显式配置文件路径,优先级高于 [WithSearchPaths]。
func WithFile(path string) Option {
	return func(l *Loader) { l.file = path }
}

// WithSearchPaths 追加配置文件搜索目录;未指定 WithFile 时生效。
func WithSearchPaths(paths ...string) Option {
	return func(l *Loader) { l.searchPaths = append(l.searchPaths, paths...) }
}

// WithFileName 设置配置文件名(不含扩展名),默认 "config"。
func WithFileName(name string) Option {
	return func(l *Loader) { l.fileName = name }
}

// WithFileType 设置配置文件类型,默认 "yaml"。
func WithFileType(ft string) Option {
	return func(l *Loader) { l.fileType = ft }
}

// WithDefaults 设置兜底默认值(viper.SetDefault,优先级低于文件)。
// key 用点分嵌套,如 "server.port"。
func WithDefaults(defaults map[string]any) Option {
	return func(l *Loader) {
		if l.defaults == nil {
			l.defaults = make(map[string]any)
		}
		for k, v := range defaults {
			l.defaults[k] = v
		}
	}
}
