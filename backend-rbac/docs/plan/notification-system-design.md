# 通知中心系统设计

## 设计原则

- **多渠道支持**：站内消息、邮件、短信
- **模板管理**：预定义通知模板，支持变量替换
- **异步发送**：消息异步处理，不阻塞业务
- **发送状态跟踪**：记录发送状态和失败重试

---

## 数据模型

### 通知模板表 (notification_templates)

```sql
CREATE TABLE notification_templates (
    template_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20),                  -- "000000000000000000" 表示系统模板（默认租户）
    template_code VARCHAR(50) NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    type VARCHAR(10) NOT NULL,              -- SYSTEM, CUSTOM
    channel VARCHAR(10) NOT NULL,           -- INBOX, EMAIL, SMS
    title VARCHAR(255),                     -- 标题模板
    content TEXT NOT NULL,                  -- 内容模板（支持变量）
    variables TEXT,                         -- 变量列表（JSON）
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY uk_code (template_code, deleted_at),
    INDEX idx_tenant_type (tenant_id, type)
);
```

### 站内消息表 (notifications)

```sql
CREATE TABLE notifications (
    notification_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    receiver_id VARCHAR(20) NOT NULL,
    template_id VARCHAR(20),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    type VARCHAR(50),                       -- SYSTEM, ALERT, REMINDER
    is_read BOOLEAN DEFAULT FALSE,
    read_at BIGINT,
    created_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    INDEX idx_receiver (tenant_id, receiver_id, is_read, created_at)
);
```

### 发送记录表 (notification_send_logs)

```sql
CREATE TABLE notification_send_logs (
    log_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    notification_id VARCHAR(20),
    receiver_id VARCHAR(20) NOT NULL,
    channel VARCHAR(10) NOT NULL,           -- INBOX, EMAIL, SMS
    receiver VARCHAR(255),                  -- 接收地址（邮箱/手机号）
    title VARCHAR(255),
    content TEXT,
    status SMALLINT,                        -- 0:待发送 1:成功 2:失败
    error VARCHAR(500),
    retry_count INT DEFAULT 0,
    sent_at BIGINT,
    created_at BIGINT NOT NULL DEFAULT 0,
    INDEX idx_tenant_status (tenant_id, status, created_at),
    INDEX idx_notification (notification_id)
);
```

---

## 业务逻辑

### 1. 创建模板

```go
func (s *TemplateService) Create(ctx context.Context, req *CreateTemplateRequest) error {
    tenantID := getTenantID(ctx)

    template := &NotificationTemplate{
        TemplateID:   uuid.New().String(),
        TenantID:     tenantID,
        TemplateCode: req.Code,
        TemplateName: req.Name,
        Type:         constants.TemplateTypeCustom,
        Channel:      req.Channel,
        Title:        req.Title,
        Content:      req.Content,
        Variables:    req.Variables,
    }

    return s.templateRepo.Create(ctx, template)
}
```

### 2. 发送通知

```go
func (s *NotificationService) Send(ctx context.Context, req *SendRequest) error {
    tenantID := getTenantID(ctx)

    // 获取模板
    template, err := s.templateRepo.GetByCode(ctx, req.TemplateCode)
    if err != nil {
        return err
    }

    // 渲染内容
    title := renderTemplate(template.Title, req.Variables)
    content := renderTemplate(template.Content, req.Variables)

    // 根据渠道发送
    switch template.Channel {
    case constants.ChannelInbox:
        return s.sendInbox(ctx, tenantID, template.TemplateID, req.ReceiverIDs, title, content, req.Type)
    case constants.ChannelEmail:
        return s.sendEmail(ctx, tenantID, template.TemplateID, req.ReceiverIDs, title, content)
    case constants.ChannelSMS:
        return s.sendSMS(ctx, tenantID, template.TemplateID, req.ReceiverIDs, content)
    }

    return nil
}

// 站内消息
func (s *NotificationService) sendInbox(ctx context.Context, tenantID, templateID string, receiverIDs []string, title, content, notifType string) error {
    for _, receiverID := range receiverIDs {
        notif := &Notification{
            NotificationID: uuid.New().String(),
            TenantID:       tenantID,
            ReceiverID:     receiverID,
            TemplateID:     templateID,
            Title:          title,
            Content:        content,
            Type:           notifType,
            IsRead:         false,
            CreatedAt:      time.Now().UnixMilli(),
        }
        s.notificationRepo.Create(ctx, notif)

        // 记录发送日志
        s.logSend(ctx, notif.NotificationID, receiverID, constants.ChannelInbox, "", title, content, 1)
    }
    return nil
}

// 邮件（异步）
func (s *NotificationService) sendEmail(ctx context.Context, tenantID, templateID string, receiverIDs []string, title, content string) error {
    // 获取用户邮箱
    users, _ := s.userRepo.GetByIDs(ctx, receiverIDs)

    for _, user := range users {
        if user.Email == "" {
            continue
        }

        // 记录待发送
        log := &SendLog{
            LogID:         uuid.New().String(),
            TenantID:      tenantID,
            TemplateID:    templateID,
            ReceiverID:    user.UserID,
            Channel:       constants.ChannelEmail,
            Receiver:      user.Email,
            Title:         title,
            Content:       content,
            Status:        0, // 待发送
            CreatedAt:     time.Now().UnixMilli(),
        }
        s.sendLogRepo.Create(ctx, log)

        // 放入队列异步发送
        s.emailQueue.Publish(log.LogID)
    }
    return nil
}

// 短信（异步）
func (s *NotificationService) sendSMS(ctx context.Context, tenantID, templateID string, receiverIDs []string, content string) error {
    users, _ := s.userRepo.GetByIDs(ctx, receiverIDs)

    for _, user := range users {
        if user.Phone == "" {
            continue
        }

        log := &SendLog{
            LogID:         uuid.New().String(),
            TenantID:      tenantID,
            TemplateID:    templateID,
            ReceiverID:    user.UserID,
            Channel:       constants.ChannelSMS,
            Receiver:      user.Phone,
            Content:       content,
            Status:        0,
            CreatedAt:     time.Now().UnixMilli(),
        }
        s.sendLogRepo.Create(ctx, log)

        s.smsQueue.Publish(log.LogID)
    }
    return nil
}
```

### 3. 模板渲染

```go
// 简单变量替换 {{变量名}}
func renderTemplate(template string, variables map[string]string) string {
    result := template
    for k, v := range variables {
        result = strings.ReplaceAll(result, "{{"+k+"}}", v)
    }
    return result
}

// 示例：
// 模板: "您好 {{username}}，您的订单 {{orderNo}} 已支付成功"
// 变量: {"username": "张三", "orderNo": "20231201001"}
// 结果: "您好 张三，您的订单 20231201001 已支付成功"
```

### 4. 邮件发送Worker

```go
func (w *EmailWorker) Process(logID string) error {
    log, err := w.sendLogRepo.GetByID(context.Background(), logID)
    if err != nil {
        return err
    }

    // 调用邮件服务
    err = w.emailService.Send(log.Receiver, log.Title, log.Content)

    // 更新状态
    if err != nil {
        log.Status = 2 // 失败
        log.Error = err.Error()
        log.RetryCount++
    } else {
        log.Status = 1 // 成功
        log.SentAt = time.Now().UnixMilli()
    }
    w.sendLogRepo.Update(context.Background(), log)

    return err
}
```

### 5. 标记已读

```go
func (s *NotificationService) MarkRead(ctx context.Context, notificationID string) error {
    userID := getUserID(ctx)

    notif, err := s.notificationRepo.GetByID(ctx, notificationID)
    if err != nil || notif.ReceiverID != userID {
        return errors.New("消息不存在")
    }

    notif.IsRead = true
    notif.ReadAt = time.Now().UnixMilli()
    return s.notificationRepo.Update(ctx, notif)
}

// 全部标记已读
func (s *NotificationService) MarkAllRead(ctx context.Context) error {
    userID := getUserID(ctx)
    return s.notificationRepo.MarkAllReadByUser(ctx, userID)
}
```

---

## API 设计

### 模板管理接口

```
GET    /api/v1/notification/templates         获取模板列表
POST   /api/v1/notification/templates         创建模板
GET    /api/v1/notification/templates/:id     获取模板详情
PUT    /api/v1/notification/templates/:id     更新模板
DELETE /api/v1/notification/templates/:id     删除模板
```

### 发送接口

```
POST   /api/v1/notification/send              发送通知
POST   /api/v1/notification/send/batch        批量发送
```

### 站内消息接口

```
GET    /api/v1/notifications                  获取消息列表
GET    /api/v1/notifications/:id              获取消息详情
PUT    /api/v1/notifications/:id/read         标记已读
PUT    /api/v1/notifications/read-all         全部标记已读
DELETE /api/v1/notifications/:id              删除消息
GET    /api/v1/notifications/unread-count     未读数量
```

---

## 系统模板示例

```sql
-- 用户注册欢迎
INSERT INTO notification_templates VALUES
('tpl-001', NULL, 'user_welcome', '用户注册欢迎', 'SYSTEM', 'INBOX',
 '欢迎加入', '欢迎 {{username}} 注册系统', '["username"]', ...);

-- 订单支付成功
INSERT INTO notification_templates VALUES
('tpl-002', NULL, 'order_paid', '订单支付成功', 'SYSTEM', 'INBOX',
 '订单支付成功', '您的订单 {{orderNo}} 已支付成功，金额：{{amount}}元', '["orderNo", "amount"]', ...);

-- 密码重置
INSERT INTO notification_templates VALUES
('tpl-003', NULL, 'password_reset', '密码重置', 'SYSTEM', 'EMAIL',
 '密码重置', '点击链接重置密码：{{resetUrl}}', '["resetUrl"]', ...);
```

---

## 常量定义

```go
package constants

const (
    // 通知渠道
    ChannelInbox = "INBOX"
    ChannelEmail = "EMAIL"
    ChannelSMS   = "SMS"

    // 模板类型
    TemplateTypeSystem = "SYSTEM"
    TemplateTypeCustom = "CUSTOM"

    // 消息类型
    NotifTypeSystem   = "SYSTEM"
    NotifTypeAlert    = "ALERT"
    NotifTypeReminder = "REMINDER"

    // 发送状态
    SendStatusPending = 0 // 待发送
    SendStatusSuccess = 1 // 成功
    SendStatusFailed  = 2 // 失败
)
```
