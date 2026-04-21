# HTTP Client

基于 [Resty](https://github.com/go-resty/resty) 封装的 HTTP 客户端包，提供简洁易用的 API 和丰富的功能。

## 特性

- ✅ 简洁的 API 设计
- ✅ 自动 JSON 编解码
- ✅ 可配置的重试机制
- ✅ 支持多种认证方式 (Bearer Token, Basic Auth)
- ✅ 请求/响应日志记录
- ✅ 超时控制
- ✅ 连接池管理
- ✅ 文件上传/下载
- ✅ 上下文支持
- ✅ 类型安全的错误处理

## 安装

```bash
go get github.com/go-resty/resty/v2
```

## 快速开始

### 基本使用

```go
package main

import (
    "admin/pkg/httpclient"
    "time"
)

func main() {
    // 创建客户端
    client := httpclient.New(
        httpclient.WithBaseURL("https://api.example.com"),
        httpclient.WithTimeout(10*time.Second),
    )

    // GET 请求
    resp, err := client.Get("/users/1")
    if err != nil {
        panic(err)
    }
    println(resp.String())

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
        panic(err)
    }
    println(resp.String())
}
```

### 使用 Request 构建器

```go
resp, err := client.NewRequest().
    SetHeader("X-Custom-Header", "value").
    SetQueryParam("page", "1").
    SetQueryParam("size", "10").
    Get("/users")
```

### JSON 响应解析

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

resp, err := client.Get("/users/1")
if err != nil {
    panic(err)
}

var user User
if err := resp.JSONToResult(&user); err != nil {
    panic(err)
}

fmt.Printf("用户: %+v\n", user)
```

## 高级用法

### 认证

```go
// Bearer Token
client.SetAuthToken("your-token-here")

// Basic Auth
client.SetBasicAuth("username", "password")
```

### 重试机制

```go
client := httpclient.New(
    httpclient.WithRetry(
        5,                       // 最大重试次数
        200*time.Millisecond,    // 重试等待时间
        2*time.Second,          // 最大等待时间
    ),
)

// 添加自定义重试条件
client.AddRetryCondition(httpclient.RetryIf(429, 500, 502, 503))
```

### 上下文支持

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.NewRequest().
    WithContext(ctx).
    Get("/users")
```

### 调试模式

```go
client := httpclient.New(
    httpclient.WithDebug(true),
)

// 或动态启用
client.SetDebug(true)

// 启用请求/响应体日志
client.EnableBodyLog()
```

### 文件上传

```go
// 从文件上传
resp, err := client.UploadFile("/upload", "file", "/path/to/file.pdf")

// 从字节数组上传
fileData := []byte("file content")
resp, err := client.UploadFileFromBytes("/upload", "file", "document.pdf", fileData)
```

### 文件下载

```go
err := client.DownloadFile("/files/document.pdf", "/path/to/save/document.pdf")
```

## 错误处理

```go
resp, err := client.Get("/users/999")
if err != nil {
    // 检查错误类型
    if httpclient.IsTimeout(err) {
        fmt.Println("请求超时")
    } else if httpclient.IsNetworkError(err) {
        fmt.Println("网络错误")
    } else if httpclient.IsStatusCodeError(err, 404) {
        fmt.Println("用户不存在")
    } else {
        fmt.Printf("其他错误: %v\n", err)
    }
    return
}
```

## 配置选项

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `WithBaseURL` | string | "" | 基础 URL |
| `WithTimeout` | time.Duration | 30s | 总超时时间 |
| `WithConnectionTimeout` | time.Duration | 10s | 连接超时时间 |
| `WithRetry` | (int, time.Duration, time.Duration) | (3, 100ms, 1s) | 重试配置 |
| `WithMaxIdleConns` | int | 100 | 最大空闲连接数 |
| `WithMaxIdleConnsPerHost` | int | 10 | 每个主机的最大空闲连接数 |
| `WithDebug` | bool | false | 调试模式 |
| `WithInsecureSkipVerify` | bool | false | 跳过 SSL 验证 |

## API 参考

### Client 方法

- `New(opts ...Option) *Client` - 创建新客户端
- `Get(url string) (*Response, error)` - GET 请求
- `Post(url string, body interface{}) (*Response, error)` - POST 请求
- `Put(url string, body interface{}) (*Response, error)` - PUT 请求
- `Patch(url string, body interface{}) (*Response, error)` - PATCH 请求
- `Delete(url string) (*Response, error)` - DELETE 请求
- `SetAuthToken(token string) *Client` - 设置认证令牌
- `SetBasicAuth(username, password string) *Client` - 设置基本认证
- `NewRequest() *Request` - 创建新请求
- `AddRetryCondition(condition RetryCondition) *Client` - 添加重试条件
- `UploadFile(url, fieldName, filePath string) (*Response, error)` - 上传文件
- `UploadFileFromBytes(url, fieldName, fileName string, fileData []byte) (*Response, error)` - 从字节数组上传文件
- `DownloadFile(url, filePath string) error` - 下载文件
- `Close() error` - 关闭客户端

### Request 方法

- `WithContext(ctx context.Context) *Request` - 设置上下文
- `SetHeader(key, value string) *Request` - 设置请求头
- `SetHeaders(headers map[string]string) *Request` - 设置多个请求头
- `SetQueryParam(key, value string) *Request` - 设置查询参数
- `SetQueryParams(params map[string]string) *Request` - 设置多个查询参数
- `SetBody(body interface{}) *Request` - 设置请求体
- `Get(url string) (*Response, error)` - GET 请求
- `Post(url string) (*Response, error)` - POST 请求
- `PostJSON(url string, body interface{}) (*Response, error)` - JSON POST 请求
- `Put(url string) (*Response, error)` - PUT 请求
- `PutJSON(url string, body interface{}) (*Response, error)` - JSON PUT 请求
- `Patch(url string) (*Response, error)` - PATCH 请求
- `PatchJSON(url string, body interface{}) (*Response, error)` - JSON PATCH 请求
- `Delete(url string) (*Response, error)` - DELETE 请求

### Response 方法

- `JSONToResult(result interface{}) error` - 解析 JSON 响应
- `String() string` - 获取响应体字符串

## 测试

```bash
# 运行所有测试
go test ./pkg/httpclient/...

# 运行特定测试
go test ./pkg/httpclient/... -run TestClientBasic

# 运行性能测试
go test ./pkg/httpclient/... -bench=.

# 查看测试覆盖率
go test ./pkg/httpclient/... -cover
```

## 最佳实践

1. **复用客户端实例**: 客户端是线程安全的，应该在应用中复用
2. **设置合理的超时**: 根据业务需求设置合适的超时时间
3. **使用上下文**: 对于需要取消的操作，使用 context.Context
4. **处理错误**: 始终检查并处理错误
5. **日志记录**: 在生产环境启用适当的日志级别
6. **连接池配置**: 根据并发量调整连接池大小

## 为什么选择 Resty?

相比原生 `net/http`，Resty 提供了:

- 更简洁的 API
- 自动 JSON 处理
- 内置重试机制
- 更好的可扩展性
- 更少的样板代码

参考: [HTTP Requests in Go: A Comparison of Client Libraries](https://medium.com/@praveendayanithi/http-requests-in-go-a-comparison-of-client-libraries-97ab5a8cc51c)

## 许可证

本项目采用与主项目相同的许可证。
