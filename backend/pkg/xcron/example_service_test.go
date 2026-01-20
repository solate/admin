package xcron_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	xcron "admin/pkg/xcron"
)

// ============================================
// 示例：如何在 service 层使用 xcron
// ============================================

// CronService 业务层的定时任务服务
type CronService struct {
	cronMgr *xcron.Manager
	repo    *MockCronJobRepository
}

// MockCronJobRepository 模拟的数据库仓库
type MockCronJobRepository struct {
	jobs    map[int64]*CronJob
	history []JobExecution
	nextID  int64
	mu      sync.Mutex
}

// CronJob 业务层的任务定义（包含业务字段）
type CronJob struct {
	ID          int64
	Name        string
	Spec        string
	Description string
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastRunAt   *time.Time
	// 根据业务需求扩展更多字段
	// Category    string
	// Priority    int
	// Owner       string
}

// JobExecution 任务执行记录
type JobExecution struct {
	JobID     int64
	StartedAt time.Time
	Error     error
}

// CreateJobRequest 创建任务请求
type CreateJobRequest struct {
	Name        string
	Spec        string
	Description string
}

// NewCronService 创建定时任务服务
func NewCronService(cronMgr *xcron.Manager, repo *MockCronJobRepository) *CronService {
	return &CronService{
		cronMgr: cronMgr,
		repo:    repo,
	}
}

// CreateJob 创建定时任务
func (s *CronService) CreateJob(ctx context.Context, req *CreateJobRequest) error {
	// 1. 验证 cron 表达式（这里简化，实际可以用 cron 库验证）
	if req.Name == "" || req.Spec == "" {
		return fmt.Errorf("任务名称和表达式不能为空")
	}

	// 2. 保存到数据库
	job := &CronJob{
		Name:        req.Name,
		Spec:        req.Spec,
		Description: req.Description,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, job); err != nil {
		return fmt.Errorf("保存任务失败: %w", err)
	}

	// 3. 注册到调度器
	// 注意：这里用闭包捕获 jobID，确保每次执行时能获取最新状态
	return s.cronMgr.Add(job.Name, job.Spec, func() {
		s.executeJob(context.Background(), job.ID)
	})
}

// executeJob 执行任务（由调度器调用）
func (s *CronService) executeJob(ctx context.Context, jobID int64) {
	startTime := time.Now()

	// 1. 从数据库获取最新状态
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		fmt.Printf("❌ 任务 %d 不存在: %v\n", jobID, err)
		return
	}

	// 2. 检查是否启用
	if !job.Enabled {
		fmt.Printf("⏸️  任务 %s (%d) 已禁用，跳过执行\n", job.Name, jobID)
		return
	}

	fmt.Printf("▶️  开始执行任务: %s (%d)\n", job.Name, jobID)

	// 3. 执行业务逻辑
	var execErr error
	switch job.Name {
	case "清理临时文件":
		execErr = s.cleanupTempFiles(ctx)
	case "数据备份":
		execErr = s.backupData(ctx)
	case "生成报表":
		execErr = s.generateReport(ctx)
	default:
		execErr = fmt.Errorf("未知任务: %s", job.Name)
	}

	// 4. 记录执行历史
	s.repo.RecordExecution(ctx, jobID, startTime, execErr)

	// 5. 更新最后执行时间
	s.repo.UpdateLastRun(ctx, jobID, startTime)

	if execErr != nil {
		fmt.Printf("❌ 任务 %s 执行失败: %v\n", job.Name, execErr)
	} else {
		fmt.Printf("✅ 任务 %s 执行成功\n", job.Name)
	}
}

// UpdateJob 更新任务
func (s *CronService) UpdateJob(ctx context.Context, jobID int64, req *CreateJobRequest) error {
	// 1. 获取现有任务
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	// 2. 如果修改了 spec，需要重新注册到调度器
	if job.Spec != req.Spec {
		// 移除旧的
		s.cronMgr.Remove(job.Name)

		// 添加新的
		// 注意：捕获 jobID 到局部变量，避免闭包问题
		currentJobID := jobID
		if err := s.cronMgr.Add(req.Name, req.Spec, func() {
			s.executeJob(context.Background(), currentJobID)
		}); err != nil {
			return err
		}
	}

	// 3. 更新数据库
	job.Spec = req.Spec
	job.Description = req.Description
	job.UpdatedAt = time.Now()

	return s.repo.Update(ctx, job)
}

// DisableJob 禁用任务
func (s *CronService) DisableJob(ctx context.Context, jobID int64) error {
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	job.Enabled = false
	job.UpdatedAt = time.Now()

	return s.repo.Update(ctx, job)
}

// EnableJob 启用任务
func (s *CronService) EnableJob(ctx context.Context, jobID int64) error {
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	job.Enabled = true
	job.UpdatedAt = time.Now()

	return s.repo.Update(ctx, job)
}

// DeleteJob 删除任务
func (s *CronService) DeleteJob(ctx context.Context, jobID int64) error {
	// 1. 从数据库获取任务
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	// 2. 从调度器移除
	s.cronMgr.Remove(job.Name)

	// 3. 从数据库删除
	return s.repo.Delete(ctx, jobID)
}

// GetExecutionHistory 获取执行历史
func (s *CronService) GetExecutionHistory(jobID int64) []JobExecution {
	return s.repo.GetHistory(jobID)
}

// ============ 业务逻辑实现 ============

func (s *CronService) cleanupTempFiles(ctx context.Context) error {
	// 模拟清理逻辑
	time.Sleep(100 * time.Millisecond)
	fmt.Println("  → 清理了 50 个临时文件")
	return nil
}

func (s *CronService) backupData(ctx context.Context) error {
	// 模拟备份逻辑
	time.Sleep(200 * time.Millisecond)
	fmt.Println("  → 数据备份完成，大小: 1.2GB")
	return nil
}

func (s *CronService) generateReport(ctx context.Context) error {
	// 模拟报表生成
	time.Sleep(150 * time.Millisecond)
	fmt.Println("  → 报表生成完成，记录数: 10000")
	return nil
}

// ============ Mock Repository 实现 ============

func (r *MockCronJobRepository) Create(ctx context.Context, job *CronJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	job.ID = r.nextID
	if r.jobs == nil {
		r.jobs = make(map[int64]*CronJob)
	}
	r.jobs[job.ID] = job
	return nil
}

func (r *MockCronJobRepository) GetByID(ctx context.Context, id int64) (*CronJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, exists := r.jobs[id]
	if !exists {
		return nil, fmt.Errorf("任务不存在: %d", id)
	}
	return job, nil
}

func (r *MockCronJobRepository) Update(ctx context.Context, job *CronJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.jobs[job.ID]; !exists {
		return fmt.Errorf("任务不存在: %d", job.ID)
	}
	r.jobs[job.ID] = job
	return nil
}

func (r *MockCronJobRepository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.jobs, id)
	return nil
}

func (r *MockCronJobRepository) RecordExecution(ctx context.Context, jobID int64, startTime time.Time, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.history = append(r.history, JobExecution{
		JobID:     jobID,
		StartedAt: startTime,
		Error:     err,
	})
}

func (r *MockCronJobRepository) UpdateLastRun(ctx context.Context, jobID int64, lastRun time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if job, exists := r.jobs[jobID]; exists {
		job.LastRunAt = &lastRun
	}
	return nil
}

func (r *MockCronJobRepository) GetHistory(jobID int64) []JobExecution {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []JobExecution
	for _, h := range r.history {
		if h.JobID == jobID {
			result = append(result, h)
		}
	}
	return result
}

// ============================================
// 使用示例
// ============================================

// Demo_serviceLayerUsage 展示如何在 service 层使用
func Demo_serviceLayerUsage() {
	// 1. 初始化调度器
	cronMgr, _ := xcron.Init(xcron.Config{WithSeconds: true})
	cronMgr.Start()
	defer cronMgr.Stop()

	// 2. 创建 service
	repo := &MockCronJobRepository{}
	service := NewCronService(cronMgr, repo)

	// 3. 创建任务
	service.CreateJob(context.Background(), &CreateJobRequest{
		Name:        "清理临时文件",
		Spec:        "*/5 * * * * ?", // 每5秒
		Description: "定期清理临时文件",
	})

	service.CreateJob(context.Background(), &CreateJobRequest{
		Name:        "数据备份",
		Spec:        "*/10 * * * * ?", // 每10秒
		Description: "每日数据备份",
	})

	service.CreateJob(context.Background(), &CreateJobRequest{
		Name:        "生成报表",
		Spec:        "0 */1 * * * ?", // 每分钟
		Description: "生成每日报表",
	})

	// 4. 让任务运行一段时间
	time.Sleep(15 * time.Second)

	// 5. 查看执行历史
	fmt.Println("\n执行历史:")
	for _, job := range repo.jobs {
		history := service.GetExecutionHistory(job.ID)
		fmt.Printf("- %s: 执行 %d 次\n", job.Name, len(history))
	}
}

// Demo_jobManagement 展示任务管理操作
func Demo_jobManagement() {
	cronMgr, _ := xcron.Init(xcron.Config{WithSeconds: true})
	cronMgr.Start()
	defer cronMgr.Stop()

	repo := &MockCronJobRepository{}
	service := NewCronService(cronMgr, repo)

	// 创建任务
	ctx := context.Background()
	err := service.CreateJob(ctx, &CreateJobRequest{
		Name: "测试任务",
		Spec: "*/5 * * * * ?",
	})
	if err != nil {
		fmt.Printf("创建任务失败: %v\n", err)
		return
	}

	// 获取创建的任务ID
	jobID := int64(1)

	// 禁用任务
	err = service.DisableJob(ctx, jobID)
	if err != nil {
		fmt.Printf("禁用任务失败: %v\n", err)
		return
	}

	// 重新启用
	err = service.EnableJob(ctx, jobID)
	if err != nil {
		fmt.Printf("启用任务失败: %v\n", err)
		return
	}

	// 删除任务
	err = service.DeleteJob(ctx, jobID)
	if err != nil {
		fmt.Printf("删除任务失败: %v\n", err)
		return
	}
}
