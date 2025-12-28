# 文件管理系统设计

## 设计原则

- **租户隔离**：每个租户的文件独立存储，互不可见
- **路径安全**：防止目录遍历攻击
- **存储抽象**：支持本地存储、OSS、S3 等多种存储后端

---

## 数据模型

### 文件表 (files)

```sql
CREATE TABLE files (
    file_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    file_name VARCHAR(255) NOT NULL,         -- 原始文件名
    file_path VARCHAR(500) NOT NULL,         -- 存储路径
    file_size BIGINT NOT NULL,               -- 字节
    file_type VARCHAR(50),                   -- MIME 类型
    file_ext VARCHAR(20),                    -- 扩展名
    storage_type VARCHAR(20) DEFAULT 'local', -- local, oss, s3
    uploader_id VARCHAR(36),                 -- 上传者
    created_at BIGINT,
    deleted_at BIGINT,
    INDEX idx_tenant (tenant_id, deleted_at),
    INDEX idx_uploader (uploader_id, deleted_at)
);
```

### 文件分享表 (file_shares)

```sql
CREATE TABLE file_shares (
    share_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    file_id VARCHAR(36) NOT NULL,
    share_code VARCHAR(20) UNIQUE NOT NULL,   -- 分享码
    password VARCHAR(100),                    -- 访问密码（可选）
    expire_at BIGINT,                         -- 过期时间
    share_count INT DEFAULT 0,                -- 分享次数
    download_count INT DEFAULT 0,             -- 下载次数
    created_by VARCHAR(36),
    created_at BIGINT,
    deleted_at BIGINT,
    INDEX idx_tenant (tenant_id, deleted_at),
    INDEX idx_file (file_id)
);
```

---

## 业务逻辑

### 1. 上传文件

```go
func (s *FileService) Upload(ctx context.Context, file *multipart.FileHeader) (*File, error) {
    tenantID := getTenantID(ctx)
    uploaderID := getUserID(ctx)

    // 验证文件类型
    if !isAllowedType(file.Filename) {
        return nil, errors.New("不支持的文件类型")
    }

    // 验证文件大小
    if file.Size > maxFileSize {
        return nil, errors.New("文件大小超出限制")
    }

    // 生成存储路径
    ext := filepath.Ext(file.Filename)
    fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
    relPath := fmt.Sprintf("%s/%s/%s", tenantID, time.Now().Format("2006/01"), fileName)

    // 打开文件
    src, _ := file.Open()
    defer src.Close()

    // 存储文件
    fullPath := filepath.Join(storagePath, relPath)
    os.MkdirAll(filepath.Dir(fullPath), 0755)
    dst, _ := os.Create(fullPath)
    defer dst.Close()
    io.Copy(dst, src)

    // 保存记录
    fileRecord := &File{
        FileID:      uuid.New().String(),
        TenantID:    tenantID,
        FileName:    file.Filename,
        FilePath:    relPath,
        FileSize:    file.Size,
        FileType:    file.Header.Get("Content-Type"),
        FileExt:     strings.TrimPrefix(ext, "."),
        StorageType: "local",
        UploaderID:  uploaderID,
    }

    s.fileRepo.Create(ctx, fileRecord)
    return fileRecord, nil
}
```

### 2. 下载文件

```go
func (s *FileService) Download(ctx context.Context, fileID string) (string, io.ReadCloser, error) {
    file, err := s.fileRepo.GetByID(ctx, fileID)
    if err != nil || file.TenantID != getTenantID(ctx) {
        return "", nil, errors.New("文件不存在")
    }

    // 安全检查：防止路径遍历
    if strings.Contains(file.FilePath, "..") {
        return "", nil, errors.New("非法路径")
    }

    fullPath := filepath.Join(storagePath, file.FilePath)
    reader, err := os.Open(fullPath)
    if err != nil {
        return "", nil, err
    }

    return file.FileName, reader, nil
}
```

### 3. 创建分享链接

```go
func (s *FileService) CreateShare(ctx context.Context, fileID string, req *CreateShareRequest) (*FileShare, error) {
    file, err := s.fileRepo.GetByID(ctx, fileID)
    if err != nil || file.TenantID != getTenantID(ctx) {
        return nil, errors.New("文件不存在")
    }

    // 生成分享码
    shareCode := generateShareCode()

    share := &FileShare{
        ShareID:    uuid.New().String(),
        TenantID:   getTenantID(ctx),
        FileID:     fileID,
        ShareCode:  shareCode,
        Password:   hashPassword(req.Password),
        ExpireAt:   req.ExpireAt,
        CreatedBy:  getUserID(ctx),
    }

    s.shareRepo.Create(ctx, share)
    return share, nil
}
```

### 4. 访问分享文件

```go
func (s *FileService) AccessShare(ctx context.Context, shareCode, password string) (string, io.ReadCloser, error) {
    share, err := s.shareRepo.GetByCode(ctx, shareCode)
    if err != nil {
        return "", nil, errors.New("分享不存在")
    }

    // 检查过期
    if share.ExpireAt > 0 && share.ExpireAt < time.Now().UnixMilli() {
        return "", nil, errors.New("分享已过期")
    }

    // 检查密码
    if share.Password != nil && !verifyPassword(password, *share.Password) {
        return "", nil, errors.New("密码错误")
    }

    // 获取文件
    file, _ := s.fileRepo.GetByID(ctx, share.FileID)

    // 更新下载次数
    s.shareRepo.IncrementDownload(ctx, share.ShareID)

    fullPath := filepath.Join(storagePath, file.FilePath)
    reader, _ := os.Open(fullPath)

    return file.FileName, reader, nil
}
```

### 5. 删除文件

```go
func (s *FileService) Delete(ctx context.Context, fileID string) error {
    file, err := s.fileRepo.GetByID(ctx, fileID)
    if err != nil || file.TenantID != getTenantID(ctx) {
        return errors.New("文件不存在")
    }

    // 软删除数据库记录
    s.fileRepo.Delete(ctx, fileID)

    // 异步删除物理文件
    go func() {
        fullPath := filepath.Join(storagePath, file.FilePath)
        os.Remove(fullPath)
    }()

    return nil
}
```

---

## API 设计

```
POST   /api/v1/files/upload              上传文件
GET    /api/v1/files/:id/download        下载文件
GET    /api/v1/files                     获取文件列表
DELETE /api/v1/files/:id                 删除文件

POST   /api/v1/files/:id/shares          创建分享
GET    /api/v1/shares/:code              访问分享文件
GET    /api/v1/shares                    我的分享列表
DELETE /api/v1/shares/:id                取消分享
```

---

## 存储接口抽象

```go
type Storage interface {
    Upload(ctx context.Context, path string, reader io.Reader) error
    Download(ctx context.Context, path string) (io.ReadCloser, error)
    Delete(ctx context.Context, path string) error
    GetURL(ctx context.Context, path string) string
}

// LocalStorage 本地存储
type LocalStorage struct {
    basePath string
}

// OSSStorage 阿里云 OSS
type OSSStorage struct {
    client *oss.Client
    bucket string
}

// S3Storage AWS S3
type S3Storage struct {
    client *s3.Client
    bucket string
}
```

---

## 安全措施

```go
// 允许的文件类型
var allowedExts = map[string]bool{
    "jpg":  true,
    "jpeg": true,
    "png":  true,
    "gif":  true,
    "pdf":  true,
    "doc":  true,
    "docx": true,
    "xls":  true,
    "xlsx": true,
    "zip":  true,
}

// 最大文件大小 10MB
const maxFileSize = 10 * 1024 * 1024

// 防止路径遍历
func sanitizePath(path string) string {
    return strings.ReplaceAll(path, "..", "")
}
```

---

## 常量定义

```go
package constants

const (
    // 存储类型
    StorageTypeLocal = "local"
    StorageTypeOSS   = "oss"
    StorageTypeS3    = "s3"

    // 文件大小限制
    MaxFileSizeImage = 5 * 1024 * 1024      // 5MB
    MaxFileSizeFile  = 10 * 1024 * 1024     // 10MB
    MaxFileSizeVideo = 100 * 1024 * 1024    // 100MB

    // 分享码长度
    ShareCodeLength = 8
)
```
