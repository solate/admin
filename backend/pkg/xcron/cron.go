package xcron

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"github.com/robfig/cron/v3"
)

var (
	globalManager *Manager
	once          sync.Once
)

// Manager 定时任务管理器 - 纯调度引擎
type Manager struct {
	cron    *cron.Cron
	jobs    map[string]*job  // name -> job
	mu      sync.RWMutex
	running atomic.Bool
}

// job 内部任务结构
type job struct {
	entryID cron.EntryID
	name    string
	spec    string
	fn      func()
}

// Config 配置
type Config struct {
	WithSeconds bool  // 是否支持秒级cron（默认支持）
	Logger      cron.Logger // 可选的日志记录器
}

// Init 初始化定时任务管理器
func Init(cfg Config) (*Manager, error) {
	var initErr error
	once.Do(func() {
		opts := []cron.Option{}

		if cfg.Logger != nil {
			opts = append(opts, cron.WithLogger(cfg.Logger))
		}

		if cfg.WithSeconds {
			opts = append(opts, cron.WithSeconds())
		}

		globalManager = &Manager{
			cron: cron.New(opts...),
			jobs: make(map[string]*job),
		}

		log.Info().Msg("定时任务管理器初始化成功")
	})

	if globalManager == nil {
		initErr = fmt.Errorf("定时任务管理器初始化失败")
	}

	return globalManager, initErr
}

// Get 获取全局管理器
func Get() *Manager {
	return globalManager
}

// Start 启动管理器
func (m *Manager) Start() {
	if m.running.CompareAndSwap(false, true) {
		m.cron.Start()
		log.Info().Msg("定时任务管理器已启动")
	}
}

// Stop 停止管理器
func (m *Manager) Stop() {
	if !m.running.Load() {
		return
	}

	ctx := m.cron.Stop()
	log.Info().Msg("定时任务管理器正在停止...")
	<-ctx.Done()
	m.running.Store(false)
	log.Info().Msg("定时任务管理器已停止")
}

// IsRunning 检查是否运行中
func (m *Manager) IsRunning() bool {
	return m.running.Load()
}

// Add 添加任务
func (m *Manager) Add(name, spec string, fn func()) error {
	if name == "" {
		return fmt.Errorf("任务名称不能为空")
	}
	if spec == "" {
		return fmt.Errorf("cron表达式不能为空")
	}
	if fn == nil {
		return fmt.Errorf("任务函数不能为空")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在
	if _, exists := m.jobs[name]; exists {
		return fmt.Errorf("任务已存在: %s", name)
	}

	// 包装函数，添加 panic 恢复
	jobName := name // 捕获变量
	wrappedFn := func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error().Str("job", jobName).Interface("panic", r).Msg("定时任务发生panic")
			}
		}()
		fn()
	}

	entryID, err := m.cron.AddFunc(spec, wrappedFn)
	if err != nil {
		return fmt.Errorf("添加任务失败: %w", err)
	}

	m.jobs[name] = &job{
		entryID: entryID,
		name:    name,
		spec:    spec,
		fn:      fn,
	}

	log.Info().
		Str("name", name).
		Str("spec", spec).
		Int("entry_id", int(entryID)).
		Msg("定时任务添加成功")

	return nil
}

// Remove 移除任务
func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	j, exists := m.jobs[name]
	if !exists {
		return fmt.Errorf("任务不存在: %s", name)
	}

	m.cron.Remove(j.entryID)
	delete(m.jobs, name)

	log.Info().
		Str("name", name).
		Msg("定时任务移除成功")

	return nil
}

// Run 立即执行任务（不等待调度）
func (m *Manager) Run(name string) error {
	m.mu.RLock()
	j, exists := m.jobs[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("任务不存在: %s", name)
	}

	// 在新协程中执行
	go j.fn()

	log.Info().
		Str("name", name).
		Msg("定时任务手动执行成功")

	return nil
}

// List 列出所有任务名称
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.jobs))
	for name := range m.jobs {
		names = append(names, name)
	}
	return names
}

// GetSpec 获取任务的 cron 表达式
func (m *Manager) GetSpec(name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	j, exists := m.jobs[name]
	if !exists {
		return "", fmt.Errorf("任务不存在: %s", name)
	}
	return j.spec, nil
}

// Exists 检查任务是否存在
func (m *Manager) Exists(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.jobs[name]
	return exists
}

// Count 获取任务数量
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.jobs)
}

// Clear 清空所有任务
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, j := range m.jobs {
		m.cron.Remove(j.entryID)
	}

	m.jobs = make(map[string]*job)
	log.Info().Msg("已清空所有定时任务")
}
