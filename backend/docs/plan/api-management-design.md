# API管理系统设计

## 设计原则

- **API密钥管理**：为租户生成和管理API密钥
- **限流控制**：基于租户/密钥的请求限流
- **鉴权机制**：API请求通过密钥验证身份
- **使用统计**：记录API调用次数和流量

---

## 数据模型

### API密钥表 (api_keys)

```sql
CREATE TABLE api_keys (
    key_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    key_name VARCHAR(100) NOT NULL,         -- 密钥名称
    access_key VARCHAR(64) NOT NULL UNIQUE,  -- Access Key
    secret_key VARCHAR(128) NOT NULL,        -- Secret Key (加密存储)
    status TINYINT DEFAULT 1,                -- 1:启用 0:禁用
    rate_limit INT DEFAULT 1000,             -- 每分钟请求限制
    expire_at BIGINT,                        -- 过期时间
    last_used_at BIGINT,                     -- 最后使用时间
    created_by VARCHAR(36),
    created_at BIGINT,
    deleted_at BIGINT,
    INDEX idx_tenant (tenant_id, deleted_at),
    INDEX idx_access_key (access_key)
);
```

### API使用统计表 (api_usage_logs)

```sql
CREATE TABLE api_usage_logs (
    log_id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id VARCHAR(20) NOT NULL,
    key_id VARCHAR(36) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,          -- 请求路径
    method VARCHAR(10) NOT NULL,             -- GET, POST, etc.
    status_code SMALLINT,                    -- 响应状态码
    response_time INT,                       -- 响应时间(ms)
    request_size INT,                        -- 请求大小(bytes)
    response_size INT,                       -- 响应大小(bytes)
    created_at BIGINT,
    INDEX idx_tenant_time (tenant_id, created_at),
    INDEX idx_key_time (key_id, created_at)
);
```

---

## 业务逻辑

### 1. 创建API密钥

```go
func (s *APIService) CreateKey(ctx context.Context, req *CreateKeyRequest) (*APIKey, error) {
    tenantID := getTenantID(ctx)

    // 生成密钥对
    accessKey := generateAccessKey()    // ak_xxxxx
    secretKey := generateSecretKey()    // 32字节随机
    secretEncrypted := encryptSecret(secretKey)

    key := &APIKey{
        KeyID:      uuid.New().String(),
        TenantID:   tenantID,
        KeyName:    req.KeyName,
        AccessKey:  accessKey,
        SecretKey:  secretEncrypted,
        RateLimit:  req.RateLimit,
        ExpireAt:   req.ExpireAt,
        CreatedBy:  getUserID(ctx),
    }

    s.keyRepo.Create(ctx, key)

    // 返回明文 secret（仅一次）
    return &APIKey{
        AccessKey:  accessKey,
        SecretKey:  secretKey, // 明文返回
    }, nil
}

func generateAccessKey() string {
    return fmt.Sprintf("ak_%s", randomString(24))
}

func generateSecretKey() string {
    b := make([]byte, 32)
    rand.Read(b)
    return hex.EncodeToString(b)
}
```

### 2. API认证中间件

```go
func APIAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 跳过非API路径
        if !strings.HasPrefix(c.Request.URL.Path, "/api/v1") {
            c.Next()
            return
        }

        // 检查JWT Token（优先）
        if token := c.GetHeader("Authorization"); token != "" {
            c.Next()
            return
        }

        // 检查API密钥
        accessKey := c.GetHeader("X-Access-Key")
        signature := c.GetHeader("X-Signature")
        timestamp := c.GetHeader("X-Timestamp")

        if accessKey == "" || signature == "" {
            c.JSON(401, gin.H{"message": "缺少认证信息"})
            c.Abort()
            return
        }

        // 验证密钥
        key, err := s.keyRepo.GetByAccessKey(c, accessKey)
        if err != nil || key.Status != 1 {
            c.JSON(401, gin.H{"message": "无效的密钥"})
            c.Abort()
            return
        }

        // 验证签名
        if !verifySignature(c.Request, secretKey, signature, timestamp) {
            c.JSON(401, gin.H{"message": "签名验证失败"})
            c.Abort()
            return
        }

        // 检查过期
        if key.ExpireAt > 0 && key.ExpireAt < time.Now().UnixMilli() {
            c.JSON(401, gin.H{"message": "密钥已过期"})
            c.Abort()
            return
        }

        // 注入上下文
        c.Set("tenant_id", key.TenantID)
        c.Set("api_key_id", key.KeyID)
        c.Set("auth_type", "api_key")

        c.Next()
    }
}

// 签名验证
func verifySignature(req *http.Request, secretKey, signature, timestamp string) bool {
    // 重放攻击检查：时间戳必须在5分钟内
    ts, _ := strconv.ParseInt(timestamp, 10, 64)
    if time.Now().Unix()-ts > 300 {
        return false
    }

    // 构造签名字符串
    body, _ := io.ReadAll(req.Body)
    req.Body = io.NopCloser(bytes.NewReader(body))

    signStr := fmt.Sprintf("%s\n%s\n%s", req.URL.Path, timestamp, string(body))
    expected := hmacSha256(signStr, secretKey)

    return hmac.Equal([]byte(signature), []byte(expected))
}
```

### 3. 限流中间件

```go
func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        keyID, exists := c.Get("api_key_id")
        if !exists {
            c.Next()
            return
        }

        key, _ := s.keyRepo.GetByID(c, keyID.(string))

        // Redis限流：滑动窗口
        key := fmt.Sprintf("ratelimit:api:%s", keyID)
        window := time.Minute
        limit := key.RateLimit

        count, _ := redis.Incr(c, key)
        if count == 1 {
            redis.Expire(c, key, int(window.Seconds()))
        }

        if count > limit {
            c.JSON(429, gin.H{
                "message": "请求过于频繁",
                "limit":   limit,
                "reset":   time.Now().Add(window).Unix(),
            })
            c.Abort()
            return
        }

        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-count))

        c.Next()
    }
}
```

### 4. 记录使用日志

```go
func APILogMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // 读取请求大小
        requestSize := c.Request.ContentLength

        c.Next()

        // 异步记录
        go func() {
            keyID, _ := c.Get("api_key_id")
            if keyID == nil {
                return
            }

            log := &APIUsageLog{
                TenantID:     c.GetString("tenant_id"),
                KeyID:        keyID.(string),
                Endpoint:     c.Request.URL.Path,
                Method:       c.Request.Method,
                StatusCode:   c.Writer.Status(),
                ResponseTime: int(time.Since(start).Milliseconds()),
                RequestSize:  int(requestSize),
                ResponseSize: c.Writer.Size(),
                CreatedAt:    time.Now().UnixMilli(),
            }
            s.usageLogRepo.Create(context.Background(), log)
        }()
    }
}
```

### 5. 使用统计

```go
func (s *APIService) GetUsageStats(ctx context.Context, keyID string, days int) (*UsageStats, error) {
    since := time.Now().AddDate(0, 0, -days).UnixMilli()

    logs, _ := s.usageLogRepo.ListByKeySince(ctx, keyID, since)

    stats := &UsageStats{}
    for _, log := range logs {
        stats.TotalRequests++
        stats.TotalTraffic += log.RequestSize + log.ResponseSize

        if log.StatusCode >= 200 && log.StatusCode < 300 {
            stats.SuccessCount++
        } else {
            stats.ErrorCount++
        }

        stats.AvgResponseTime += log.ResponseTime
    }

    if len(logs) > 0 {
        stats.AvgResponseTime /= len(logs)
    }

    return stats, nil
}
```

---

## API设计

```
# 密钥管理
POST   /api/v1/api-keys                 创建密钥
GET    /api/v1/api-keys                 获取密钥列表
GET    /api/v1/api-keys/:id             获取密钥详情
PUT    /api/v1/api-keys/:id             更新密钥
DELETE /api/v1/api-keys/:id             删除密钥
POST   /api/v1/api-keys/:id/disable     禁用密钥
POST   /api/v1/api-keys/:id/enable      启用密钥

# 使用统计
GET    /api/v1/api-keys/:id/stats       使用统计
GET    /api/v1/api-keys/:id/logs        使用日志
```

---

## 签名算法

### 客户端签名

```go
func SignRequest(req *http.Request, accessKey, secretKey string) {
    timestamp := fmt.Sprintf("%d", time.Now().Unix())

    // 读取Body
    body, _ := io.ReadAll(req.Body)

    // 构造签名字符串
    signStr := fmt.Sprintf("%s\n%s\n%s", req.URL.Path, timestamp, string(body))

    // HMAC-SHA256
    h := hmac.New(sha256.New, []byte(secretKey))
    h.Write([]byte(signStr))
    signature := hex.EncodeToString(h.Sum(nil))

    // 设置Headers
    req.Header.Set("X-Access-Key", accessKey)
    req.Header.Set("X-Signature", signature)
    req.Header.Set("X-Timestamp", timestamp)

    // 恢复Body
    req.Body = io.NopCloser(bytes.NewReader(body))
}
```

### 服务端验证（已在中间件实现）

---

## 常量定义

```go
package constants

const (
    // 签名算法
    SignAlgorithmHMACSHA256 = "hmac-sha256"

    // 时间戳有效期（秒）
    TimestampTTL = 300

    // 默认限流
    DefaultRateLimit = 1000 // 每分钟

    // 密钥状态
    APIKeyStatusEnabled  = 1
    APIKeyStatusDisabled = 0
)
```

---

## Swagger配置

```go
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Access-Key
func (h *Handler) SomeHandler(c *gin.Context) {
    // ...
}

// @Security ApiKeyAuth
// @Param X-Access-Key header string true "Access Key"
// @Param X-Signature header string true "Request Signature"
// @Param X-Timestamp header string true "Unix Timestamp"
```
