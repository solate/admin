// Package xconfig 是可复用的配置加载库:用 viper 读配置文件并反序列化到
// 调用方提供的类型化结构。本包自己不写反射(不 import reflect),
// 反序列化交给 viper/mapstructure;环境变量覆盖走显式 [Override]。
//
// 设计要点(详见 docs/saas-backend/research/config-loading/):
//   - 绝不使用 viper.AutomaticEnv:它与 Unmarshal 不兼容,nested key 会静默失效。
//   - 零全局状态、零 init;所有错误用 %w 包装。
//   - 长生命周期 Loader:未来 Reload/Watch/Get 都挂在 *Loader 上,不破现有 API。
package xconfig

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Loader 持有加载选项,可重复调用 [Loader.Load]。
// 未来 Reload/Watch/运行时 Get 等扩展都挂在 *Loader 上。
type Loader struct {
	file        string
	searchPaths []string
	fileName    string
	fileType    string
	defaults    map[string]any
}

// New 创建 Loader,用 functional options 配置。
func New(opts ...Option) *Loader {
	l := &Loader{
		fileName: "config",
		fileType: "yaml",
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Load 读配置文件并经 viper 反序列化到 target(类型化结构指针)。
// 文件解析、mapstructure 解码、time.Duration("10s")等均由 viper 处理。
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

// Override 用环境变量 key 覆盖 *dst(类型化、无反射)。
// key 存在(即使值为空串)则覆盖 *dst;不存在则保持原值。
// 仅支持 string(密钥类配置都是 string);其他类型请调用方自行 os.LookupEnv。
func Override(dst *string, key string) {
	if dst == nil {
		return
	}
	if v, ok := os.LookupEnv(key); ok {
		*dst = v
	}
}
