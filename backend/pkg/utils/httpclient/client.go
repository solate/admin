package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

// Client HTTP 客户端
type Client struct {
	resty      *resty.Client
	config     *Config
	middleware *LoggingMiddleware
}

// New 创建新的 HTTP 客户端
func New(opts ...Option) *Client {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	client := resty.New().
		SetTimeout(config.Timeout).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetRetryCount(config.MaxRetries).
		SetRetryWaitTime(config.RetryWaitTime).
		SetRetryMaxWaitTime(config.RetryMaxWaitTime).
		SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		})

	// 设置连接池
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableKeepAlives:   config.DisableKeepAlives,
	}
	client.SetTransport(transport)

	if config.BaseURL != "" {
		client.SetBaseURL(config.BaseURL)
	}

	httpClient := &Client{
		resty:      client,
		config:     config,
		middleware: NewLoggingMiddleware(),
	}

	// 添加默认重试条件
	condition := DefaultRetryCondition()
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		resp := &Response{
			StatusCode: r.StatusCode(),
			Header:     r.Header(),
			Body:       r.Body(),
		}
		return condition(resp, err)
	})

	// 添加调试日志
	if config.Debug {
		client.OnBeforeRequest(httpClient.beforeRequestLog)
		client.OnAfterResponse(httpClient.afterResponseLog)
	}

	return httpClient
}

// SetDebug 设置调试模式
func (c *Client) SetDebug(debug bool) *Client {
	c.config.Debug = debug
	if debug {
		c.resty.OnBeforeRequest(c.beforeRequestLog)
		c.resty.OnAfterResponse(c.afterResponseLog)
	}
	return c
}

// EnableBodyLog 启用请求/响应体日志
func (c *Client) EnableBodyLog() *Client {
	c.middleware.WithRequestBodyLog().WithResponseBodyLog()
	return c
}

// beforeRequestLog 请求前日志
func (c *Client) beforeRequestLog(client *resty.Client, req *resty.Request) error {
	ctx := req.Context()
	log.Ctx(ctx).Debug().
		Str("method", req.Method).
		Str("url", req.URL).
		Interface("headers", req.Header).
		Msg("发送 HTTP 请求")
	return nil
}

// afterResponseLog 响应后日志
func (c *Client) afterResponseLog(client *resty.Client, resp *resty.Response) error {
	ctx := resp.Request.Context()
	log.Ctx(ctx).Debug().
		Int("status_code", resp.StatusCode()).
		Dur("duration", resp.Time()).
		Str("status", resp.Status()).
		Msg("收到 HTTP 响应")
	return nil
}

// SetAuthToken 设置认证令牌
func (c *Client) SetAuthToken(token string) *Client {
	c.resty.SetAuthToken(token)
	return c
}

// SetBasicAuth 设置基本认证
func (c *Client) SetBasicAuth(username, password string) *Client {
	c.resty.SetBasicAuth(username, password)
	return c
}

// Request 创建请求
type Request struct {
	client *Client
	r      *resty.Request
	ctx    context.Context
}

//.NewRequest 创建新请求
func (c *Client) NewRequest() *Request {
	r := c.resty.R().
		SetContext(context.Background())

	return &Request{
		client: c,
		r:      r,
		ctx:    context.Background(),
	}
}

// WithContext 设置上下文
func (r *Request) WithContext(ctx context.Context) *Request {
	r.ctx = ctx
	r.r.SetContext(ctx)
	return r
}

// SetHeader 设置请求头
func (r *Request) SetHeader(key, value string) *Request {
	r.r.SetHeader(key, value)
	return r
}

// SetHeaders 设置多个请求头
func (r *Request) SetHeaders(headers map[string]string) *Request {
	r.r.SetHeaders(headers)
	return r
}

// SetQueryParam 设置查询参数
func (r *Request) SetQueryParam(key, value string) *Request {
	r.r.SetQueryParam(key, value)
	return r
}

// SetQueryParams 设置多个查询参数
func (r *Request) SetQueryParams(params map[string]string) *Request {
	r.r.SetQueryParams(params)
	return r
}

// SetBody 设置请求体
func (r *Request) SetBody(body interface{}) *Request {
	r.r.SetBody(body)
	return r
}

// SetResult 设置成功响应的解析目标
func (r *Request) SetResult(result interface{}) *Request {
	r.r.SetResult(result)
	return r
}

// SetError 设置错误响应的解析目标
func (r *Request) SetError(error interface{}) *Request {
	r.r.SetError(error)
	return r
}

// Get 发送 GET 请求
func (c *Client) Get(url string) (*Response, error) {
	return c.NewRequest().Get(url)
}

// Post 发送 POST 请求
func (c *Client) Post(url string, body interface{}) (*Response, error) {
	return c.NewRequest().SetBody(body).Post(url)
}

// Put 发送 PUT 请求
func (c *Client) Put(url string, body interface{}) (*Response, error) {
	return c.NewRequest().SetBody(body).Put(url)
}

// Patch 发送 PATCH 请求
func (c *Client) Patch(url string, body interface{}) (*Response, error) {
	return c.NewRequest().SetBody(body).Patch(url)
}

// Delete 发送 DELETE 请求
func (c *Client) Delete(url string) (*Response, error) {
	return c.NewRequest().Delete(url)
}

// Get 发送 GET 请求
func (r *Request) Get(url string) (*Response, error) {
	return r.execute(http.MethodGet, url, nil)
}

// Post 发送 POST 请求
func (r *Request) Post(url string) (*Response, error) {
	return r.execute(http.MethodPost, url, nil)
}

// PostJSON 发送 JSON POST 请求
func (r *Request) PostJSON(url string, body interface{}) (*Response, error) {
	return r.execute(http.MethodPost, url, body)
}

// Put 发送 PUT 请求
func (r *Request) Put(url string) (*Response, error) {
	return r.execute(http.MethodPut, url, nil)
}

// PutJSON 发送 JSON PUT 请求
func (r *Request) PutJSON(url string, body interface{}) (*Response, error) {
	return r.execute(http.MethodPut, url, body)
}

// Patch 发送 PATCH 请求
func (r *Request) Patch(url string) (*Response, error) {
	return r.execute(http.MethodPatch, url, nil)
}

// PatchJSON 发送 JSON PATCH 请求
func (r *Request) PatchJSON(url string, body interface{}) (*Response, error) {
	return r.execute(http.MethodPatch, url, body)
}

// Delete 发送 DELETE 请求
func (r *Request) Delete(url string) (*Response, error) {
	return r.execute(http.MethodDelete, url, nil)
}

// execute 执行请求
func (r *Request) execute(method, url string, body interface{}) (*Response, error) {
	startTime := time.Now()

	req := r.r
	if body != nil {
		req.SetBody(body)
	}

	// 转换 header 格式
	headers := make(map[string]string)
	for k, v := range req.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// 获取请求体
	var bodyBytes []byte
	if req.Body != nil {
		if b, ok := req.Body.([]byte); ok {
			bodyBytes = b
		} else if str, ok := req.Body.(string); ok {
			bodyBytes = []byte(str)
		}
	}

	// 记录请求日志
	r.client.middleware.LogRequest(r.ctx, method, url, headers, bodyBytes)

	// 执行请求
	restyResp, err := req.Execute(method, url)

	// 记录响应日志
	duration := time.Since(startTime)
	if restyResp != nil {
		r.client.middleware.LogResponse(r.ctx, restyResp.StatusCode(), duration, restyResp.Body())
	}

	if err != nil {
		return nil, NewError(0, "请求失败", err)
	}

	// 检查状态码
	if restyResp.StatusCode() >= 400 {
		return nil, NewError(restyResp.StatusCode(), getErrorMessage(restyResp.StatusCode()), nil)
	}

	return &Response{
		StatusCode: restyResp.StatusCode(),
		Header:     restyResp.Header(),
		Body:       restyResp.Body(),
	}, nil
}

// JSONToResult 将响应体解析到 result
func (r *Response) JSONToResult(result interface{}) error {
	return json.Unmarshal(r.Body, result)
}

// String 返回响应体字符串
func (r *Response) String() string {
	return string(r.Body)
}

// AddRetryCondition 添加重试条件
func (c *Client) AddRetryCondition(condition RetryCondition) *Client {
	c.resty.AddRetryCondition(func(r *resty.Response, err error) bool {
		resp := &Response{
			StatusCode: r.StatusCode(),
			Header:     r.Header(),
			Body:       r.Body(),
		}
		return condition(resp, err)
	})
	return c
}

// buildURL 构建完整 URL
func (c *Client) buildURL(path string, params map[string]string) (string, error) {
	baseURL := c.config.BaseURL
	if baseURL == "" {
		baseURL = path
	} else {
		baseURL = strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")
	}

	if len(params) > 0 {
		u, err := url.Parse(baseURL)
		if err != nil {
			return "", err
		}
		query := u.Query()
		for k, v := range params {
			query.Set(k, v)
		}
		u.RawQuery = query.Encode()
		baseURL = u.String()
	}

	return baseURL, nil
}

// getErrorMessage 根据状态码获取错误消息
func getErrorMessage(statusCode int) string {
	messages := map[int]string{
		400: "请求参数错误",
		401: "未授权",
		403: "无权限访问",
		404: "资源不存在",
		405: "方法不允许",
		409: "资源冲突",
		429: "请求过于频繁",
		500: "服务器内部错误",
		502: "网关错误",
		503: "服务不可用",
		504: "网关超时",
	}

	if msg, ok := messages[statusCode]; ok {
		return msg
	}
	return fmt.Sprintf("HTTP 错误: %d", statusCode)
}

// DownloadFile 下载文件
func (c *Client) DownloadFile(url, filePath string) error {
	_, err := c.resty.R().
		SetOutput(filePath).
		Get(url)

	return err
}

// UploadFile 上传文件
func (c *Client) UploadFile(url, fieldName, filePath string) (*Response, error) {
	resp, err := c.resty.R().
		SetFile(fieldName, filePath).
		Post(url)

	if err != nil {
		return nil, NewError(0, "上传失败", err)
	}

	if resp.StatusCode() >= 400 {
		return nil, NewError(resp.StatusCode(), "上传失败", nil)
	}

	return &Response{
		StatusCode: resp.StatusCode(),
		Header:     resp.Header(),
		Body:       resp.Body(),
	}, nil
}

// UploadFileFromBytes 上传文件（从字节数组）
func (c *Client) UploadFileFromBytes(url, fieldName, fileName string, fileData []byte) (*Response, error) {
	resp, err := c.resty.R().
		SetFileReader(fieldName, fileName, bytes.NewReader(fileData)).
		Post(url)

	if err != nil {
		return nil, NewError(0, "上传失败", err)
	}

	if resp.StatusCode() >= 400 {
		return nil, NewError(resp.StatusCode(), "上传失败", nil)
	}

	return &Response{
		StatusCode: resp.StatusCode(),
		Header:     resp.Header(),
		Body:       resp.Body(),
	}, nil
}

// GetRestyClient 获取底层的 Resty 客户端（用于高级用法）
func (c *Client) GetRestyClient() *resty.Client {
	return c.resty
}

// Close 关闭客户端并清理资源
func (c *Client) Close() error {
	// Resty 使用 http.DefaultClient 或自定义的 http.Client
	// 如果需要清理连接池，可以在这里实现
	c.resty.GetClient().CloseIdleConnections()
	return nil
}
