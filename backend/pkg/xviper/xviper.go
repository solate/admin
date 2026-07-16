// Package xviper 是对 spf13/viper 的约定式封装。
//
// 约定(不可配置):
//   - 配置格式: YAML
//   - Struct tag: mapstructure (viper 默认)
//   - 环境覆盖: config.{env}.yaml (APP_ENV=dev → config.dev.yaml)
//   - 环境变量: 点/横线转下划线大写 (server.port → SERVER_PORT,加前缀为 APP_SERVER_PORT)
package xviper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	envPrefix  = "APP"     // 环境变量前缀,如 APP_SERVER_PORT
	envVarName = "APP_ENV" // 环境切换变量,读后覆盖默认 env
	envDefault = "dev"     // 默认环境
)

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

// Load 按约定加载 YAML 配置并反序列化为 *T。
//
// 合并优先级(低 → 高): 基础文件 → 环境文件 → 环境变量。
func Load[T any](opts ...Option[T]) (*T, error) {
	s := defaultSettings[T]()
	for _, opt := range opts {
		opt(&s)
	}

	// 用 viper.ExperimentalBindStruct()(v1.20+)让 AutomaticEnv 能覆盖嵌套字段,
	// 否则密钥字段必须在 yaml 里留空占位才能被环境变量注入。
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())

	// 格式固定为 YAML(包约定),显式设定而非依赖路径推断,
	v.SetConfigType("yaml")

	// 1. 基础配置文件(必须存在)
	// 用 SetConfigFile 显式指定完整路径,而非 AddConfigPath+SetConfigName 多路径搜索:
	// 服务端/容器部署下配置路径是确定的,多路径搜索反而有加载到残留同名文件的风险。
	basePath := s.path
	v.SetConfigFile(basePath)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("xviper: 读取配置文件 %q 失败: %w", basePath, err)
	}
	fmt.Printf("xviper: 已加载基础配置 %s\n", v.ConfigFileUsed())

	// 2. 环境覆盖文件(config.{env}.yaml)
	overlay := buildOverlayPath(basePath, getEnvironment(v))
	v.SetConfigFile(overlay)
	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("xviper: 合并环境配置 %q 失败: %w", overlay, err)
	}
	fmt.Printf("xviper: 已合并环境配置 %s\n", overlay)

	// 3. 环境变量覆盖(最高优先级)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	var cfg T
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("xviper: 反序列化配置失败: %w", err)
	}
	if s.validate != nil {
		if err := s.validate(&cfg); err != nil {
			return nil, fmt.Errorf("xviper: 配置校验失败: %w", err)
		}
	}
	return &cfg, nil
}

// buildOverlayPath 按约定把基础路径转成环境覆盖路径。
// config.yaml + dev → config.dev.yaml
func buildOverlayPath(basePath, env string) string {
	ext := filepath.Ext(basePath)
	return strings.TrimSuffix(basePath, ext) + "." + env + ext
}

// getEnvironment 获取当前环境标识。
// 优先级: APP_ENV 环境变量 > 配置文件 app.env 字段 > 默认值 dev。
func getEnvironment(v *viper.Viper) string {
	if env := os.Getenv(envVarName); env != "" {
		return env
	}
	if configEnv := v.GetString("app.env"); configEnv != "" {
		return configEnv
	}
	return envDefault
}
