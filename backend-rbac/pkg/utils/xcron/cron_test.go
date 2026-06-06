package xcron

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestInit 测试初始化
func TestInit(t *testing.T) {
	// 重置全局变量
	globalManager = nil
	once = sync.Once{}

	m, err := Init(Config{WithSeconds: true})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if m == nil {
		t.Fatal("Init() 返回 nil manager")
	}

	if !m.IsRunning() {
		m.Start()
		defer m.Stop()
	}
}

// TestAddAndRemove 测试添加和删除任务
func TestAddAndRemove(t *testing.T) {
	m, _ := Init(Config{WithSeconds: true})
	m.Start()
	defer m.Stop()

	t.Run("正常添加任务", func(t *testing.T) {
		executed := false
		err := m.Add("test_job", "* * * * * ?", func() {
			executed = true
		})

		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		// 等待任务执行
		time.Sleep(2 * time.Second)

		if !executed {
			t.Error("任务未执行")
		}
	})

	t.Run("重复任务名称", func(t *testing.T) {
		err := m.Add("test_job", "* * * * * ?", func() {})
		if err == nil {
			t.Error("期望返回任务已存在的错误")
		}
	})

	t.Run("空任务名称", func(t *testing.T) {
		err := m.Add("", "* * * * * ?", func() {})
		if err == nil {
			t.Error("期望返回错误")
		}
	})

	t.Run("空cron表达式", func(t *testing.T) {
		err := m.Add("test2", "", func() {})
		if err == nil {
			t.Error("期望返回错误")
		}
	})

	t.Run("空任务函数", func(t *testing.T) {
		err := m.Add("test2", "* * * * * ?", nil)
		if err == nil {
			t.Error("期望返回错误")
		}
	})

	t.Run("删除任务", func(t *testing.T) {
		m.Add("to_remove", "* * * * * ?", func() {})
		err := m.Remove("to_remove")
		if err != nil {
			t.Fatalf("Remove() error = %v", err)
		}

		if m.Exists("to_remove") {
			t.Error("删除后任务仍然存在")
		}
	})

	t.Run("删除不存在的任务", func(t *testing.T) {
		err := m.Remove("not_exist")
		if err == nil {
			t.Error("期望返回错误")
		}
	})
}

// TestJobQuery 测试任务查询
func TestJobQuery(t *testing.T) {
	m, _ := Init(Config{WithSeconds: true})
	m.Start()
	defer m.Stop()

	// 添加测试任务
	m.Add("job1", "*/5 * * * * ?", func() {})
	m.Add("job2", "*/10 * * * * ?", func() {})
	m.Add("job3", "*/15 * * * * ?", func() {})

	t.Run("列出所有任务", func(t *testing.T) {
		jobs := m.List()
		if len(jobs) < 3 {
			t.Errorf("任务数量不足: got %d, want >= 3", len(jobs))
		}
	})

	t.Run("检查任务存在", func(t *testing.T) {
		if !m.Exists("job1") {
			t.Error("job1 应该存在")
		}

		if m.Exists("not_exist") {
			t.Error("not_exist 不应该存在")
		}
	})

	t.Run("获取任务spec", func(t *testing.T) {
		spec, err := m.GetSpec("job1")
		if err != nil {
			t.Fatalf("GetSpec() error = %v", err)
		}

		if spec != "*/5 * * * * ?" {
			t.Errorf("spec 不匹配: got %s, want */5 * * * * ?", spec)
		}
	})

	t.Run("获取不存在任务的spec", func(t *testing.T) {
		_, err := m.GetSpec("not_exist")
		if err == nil {
			t.Error("期望返回错误")
		}
	})

	t.Run("任务数量", func(t *testing.T) {
		count := m.Count()
		if count < 3 {
			t.Errorf("任务数量不足: got %d, want >= 3", count)
		}
	})
}

// TestRun 测试手动执行任务
func TestRun(t *testing.T) {
	m, _ := Init(Config{WithSeconds: true})
	m.Start()
	defer m.Stop()

	t.Run("手动执行任务", func(t *testing.T) {
		executed := false
		m.Add("manual_job", "0 0 3 * * ?", func() {
			executed = true
		})

		err := m.Run("manual_job")
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		// 等待异步执行
		time.Sleep(100 * time.Millisecond)

		if !executed {
			t.Error("手动执行未触发任务")
		}
	})

	t.Run("执行不存在的任务", func(t *testing.T) {
		err := m.Run("not_exist")
		if err == nil {
			t.Error("期望返回错误")
		}
	})
}

// TestPanicRecovery 测试panic恢复
func TestPanicRecovery(t *testing.T) {
	m, _ := Init(Config{WithSeconds: true})
	m.Start()
	defer m.Stop()

	panicCalled := false
	m.Add("panic_job", "* * * * * ?", func() {
		panic("test panic")
	})

	// 等待任务执行并panic
	time.Sleep(2 * time.Second)

	// 管理器应该仍然运行
	if !m.IsRunning() {
		t.Error("panic 后管理器应该仍在运行")
	}

	_ = panicCalled
}

// TestClear 测试清空任务
func TestClear(t *testing.T) {
	m, _ := Init(Config{WithSeconds: true})
	m.Start()
	defer m.Stop()

	m.Add("job1", "* * * * * ?", func() {})
	m.Add("job2", "* * * * * ?", func() {})

	if m.Count() == 0 {
		t.Error("任务数应该大于0")
	}

	m.Clear()

	if m.Count() != 0 {
		t.Errorf("清空后任务数应为0, got %d", m.Count())
	}
}

// TestStartStop 测试启动和停止
func TestStartStop(t *testing.T) {
	m, _ := Init(Config{WithSeconds: true})

	if m.IsRunning() {
		t.Error("初始状态应该未运行")
	}

	m.Start()

	if !m.IsRunning() {
		t.Error("Start() 后应该正在运行")
	}

	// 重复调用 Start 应该安全
	m.Start()

	if !m.IsRunning() {
		t.Error("重复调用 Start 后应该仍在运行")
	}

	m.Stop()

	if m.IsRunning() {
		t.Error("Stop() 后应该停止运行")
	}

	// 重复调用 Stop 应该安全
	m.Stop()
}

// TestConcurrency 测试并发安全
func TestConcurrency(t *testing.T) {
	m, _ := Init(Config{WithSeconds: true})
	m.Start()
	defer m.Stop()

	var wg sync.WaitGroup

	// 并发添加任务
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("concurrent_%d", idx)
			m.Add(name, "* * * * * ?", func() {})
		}(i)
	}

	wg.Wait()

	if m.Count() < 10 {
		t.Errorf("并发添加任务数量不足: got %d, want >= 10", m.Count())
	}

	// 并发查询
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.List()
			m.Count()
		}()
	}

	wg.Wait()
}

// BenchmarkAdd 性能测试
func BenchmarkAdd(b *testing.B) {
	m, _ := Init(Config{WithSeconds: true})
	m.Start()
	defer m.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Add(fmt.Sprintf("bench_%d", i), "* * * * * ?", func() {})
	}
}

// Example_basic 基本使用示例
func Example_basic() {
	m, _ := Init(Config{WithSeconds: true})
	m.Start()
	defer m.Stop()

	m.Add("cleanup", "*/5 * * * * ?", func() {
		fmt.Println("清理临时文件...")
	})
}
