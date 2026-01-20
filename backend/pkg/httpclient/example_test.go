package httpclient_test

import (
	"admin/pkg/httpclient"
	"context"
	"fmt"
	"testing"
	"time"
)

// DemoClient 基本使用演示
func DemoClient(t *testing.T) {
	// 创建客户端
	client := httpclient.New(
		httpclient.WithBaseURL("https://api.example.com"),
		httpclient.WithTimeout(10*time.Second),
		httpclient.WithDebug(true),
	)

	// GET 请求
	resp, err := client.Get("/users/1")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	fmt.Printf("响应: %s\n", resp.String())

	// POST 请求
	type CreateUserRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	req := &CreateUserRequest{
		Name:  "张三",
		Email: "zhangsan@example.com",
	}

	resp, err = client.Post("/users", req)
	if err != nil {
		fmt.Printf("创建用户失败: %v\n", err)
		return
	}
	fmt.Printf("创建成功: %s\n", resp.String())
}

// DemoClientWithAuth 带认证的演示
func DemoClientWithAuth(t *testing.T) {
	// 方式1: 使用 SetAuthToken
	client := httpclient.New(
		httpclient.WithBaseURL("https://api.example.com"),
	)
	client.SetAuthToken("your-token-here")

	resp, err := client.Get("/protected/resource")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	fmt.Printf("响应: %s\n", resp.String())

	// 方式2: 使用 Basic Auth
	client.SetBasicAuth("username", "password")

	resp, err = client.Get("/protected/resource")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	fmt.Printf("响应: %s\n", resp.String())
}

// DemoRequestWithBuilder 使用 Request 构建器的演示
func DemoRequestWithBuilder(t *testing.T) {
	client := httpclient.New(
		httpclient.WithBaseURL("https://api.example.com"),
	)

	// 使用构建器模式
	resp, err := client.NewRequest().
		SetHeader("X-Custom-Header", "value").
		SetQueryParam("page", "1").
		SetQueryParam("size", "10").
		Get("/users")

	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	fmt.Printf("响应: %s\n", resp.String())
}

// DemoClientWithContext 使用上下文的演示
func DemoClientWithContext(t *testing.T) {
	client := httpclient.New(
		httpclient.WithBaseURL("https://api.example.com"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.NewRequest().
		WithContext(ctx).
		Get("/users")

	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	fmt.Printf("响应: %s\n", resp.String())
}

// DemoClientWithRetry 自定义重试策略演示
func DemoClientWithRetry(t *testing.T) {
	client := httpclient.New(
		httpclient.WithBaseURL("https://api.example.com"),
		httpclient.WithRetry(5, 200*time.Millisecond, 2*time.Second),
	)

	// 添加自定义重试条件
	client.AddRetryCondition(httpclient.RetryIf(429, 500, 502, 503))

	resp, err := client.Get("/unstable-endpoint")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	fmt.Printf("响应: %s\n", resp.String())
}

// DemoClientJSONHandling JSON 处理演示
func DemoClientJSONHandling(t *testing.T) {
	client := httpclient.New(
		httpclient.WithBaseURL("https://api.example.com"),
	)

	// 发送请求
	resp, err := client.Get("/users/1")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	// 解析响应
	type User struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	var user User
	if err := resp.JSONToResult(&user); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		return
	}

	fmt.Printf("用户: %+v\n", user)
}

// DemoClientFileUpload 文件上传演示
func DemoClientFileUpload(t *testing.T) {
	client := httpclient.New(
		httpclient.WithBaseURL("https://api.example.com"),
	)

	// 上传文件
	resp, err := client.UploadFile("/upload", "file", "/path/to/file.pdf")
	if err != nil {
		fmt.Printf("上传失败: %v\n", err)
		return
	}
	fmt.Printf("上传成功: %s\n", resp.String())

	// 从字节数组上传
	fileData := []byte("file content")
	resp, err = client.UploadFileFromBytes("/upload", "file", "document.pdf", fileData)
	if err != nil {
		fmt.Printf("上传失败: %v\n", err)
		return
	}
	fmt.Printf("上传成功: %s\n", resp.String())
}

// DemoClientDownload 文件下载演示
func DemoClientDownload(t *testing.T) {
	client := httpclient.New(
		httpclient.WithBaseURL("https://api.example.com"),
	)

	err := client.DownloadFile("/files/document.pdf", "/path/to/save/document.pdf")
	if err != nil {
		fmt.Printf("下载失败: %v\n", err)
		return
	}
	fmt.Printf("下载成功\n")
}

// DemoClientErrorHandling 错误处理演示
func DemoClientErrorHandling(t *testing.T) {
	client := httpclient.New(
		httpclient.WithBaseURL("https://api.example.com"),
	)

	resp, err := client.Get("/users/999")
	if err != nil {
		// 检查错误类型
		if httpclient.IsTimeout(err) {
			fmt.Println("请求超时")
			return
		}

		if httpclient.IsNetworkError(err) {
			fmt.Println("网络错误")
			return
		}

		if httpclient.IsStatusCodeError(err, 404) {
			fmt.Println("用户不存在")
			return
		}

		fmt.Printf("其他错误: %v\n", err)
		return
	}

	fmt.Printf("响应: %s\n", resp.String())
}

// TestClientBasic 基本测试
func TestClientBasic(t *testing.T) {
	// 创建测试客户端
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
		httpclient.WithTimeout(5*time.Second),
		httpclient.WithDebug(true),
	)

	// 测试 GET 请求
	t.Run("GET", func(t *testing.T) {
		resp, err := client.Get("/posts/1")
		if err != nil {
			t.Fatalf("GET 请求失败: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("期望状态码 200, 实际: %d", resp.StatusCode)
		}

		t.Logf("响应: %s", resp.String())
	})

	// 测试 POST 请求
	t.Run("POST", func(t *testing.T) {
		type Post struct {
			Title  string `json:"title"`
			Body   string `json:"body"`
			UserID int    `json:"userId"`
		}

		post := &Post{
			Title:  "测试标题",
			Body:   "测试内容",
			UserID: 1,
		}

		resp, err := client.Post("/posts", post)
		if err != nil {
			t.Fatalf("POST 请求失败: %v", err)
		}

		if resp.StatusCode != 201 {
			t.Errorf("期望状态码 201, 实际: %d", resp.StatusCode)
		}

		t.Logf("响应: %s", resp.String())
	})
}

// TestClientWithAuth 测试认证
func TestClientWithAuth(t *testing.T) {
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
	)

	// 设置 Bearer Token
	client.SetAuthToken("test-token")

	resp, err := client.Get("/posts/1")
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	t.Logf("响应状态码: %d", resp.StatusCode)
}

// TestClientRetry 测试重试机制
func TestClientRetry(t *testing.T) {
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
		httpclient.WithRetry(3, 100*time.Millisecond, 500*time.Millisecond),
	)

	resp, err := client.Get("/posts/1")
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	t.Logf("响应状态码: %d", resp.StatusCode)
}

// TestClientErrorHandling 测试错误处理
func TestClientErrorHandling(t *testing.T) {
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
	)

	// 测试 404 错误
	_, err := client.Get("/posts/999999")
	if err == nil {
		t.Error("期望返回错误,但得到了成功响应")
	}

	t.Logf("错误信息: %v", err)
}

// BenchmarkClient 性能测试
func BenchmarkClient(b *testing.B) {
	client := httpclient.New(
		httpclient.WithBaseURL("https://jsonplaceholder.typicode.com"),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Get("/posts/1")
		if err != nil {
			b.Fatalf("请求失败: %v", err)
		}
	}
}
