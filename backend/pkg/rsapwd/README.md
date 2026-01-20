# rsapwd - RSA 密码传输加密

用于传输过程中密码加密的封装包，提供 RSA 加密解密功能，保护密码在传输过程中的安全性。

## 功能特性

- **RSA 加密解密**：支持 PKCS#1 v1.5 填充的 RSA 加密解密（与 JSEncrypt 兼容）
- **多格式私钥支持**：自动识别 PKCS1 和 PKCS8 格式的私钥
- **SHA256 哈希**：提供密码的 SHA256 哈希功能（可选）
- **Base64 编码**：自动处理 Base64 编码解码
- **错误处理**：完善的错误处理和自定义错误类型

## 快速开始

### 1. 生成 RSA 密钥对

```bash
# 生成私钥
openssl genrsa -out private_key.pem 2048

# 提取公钥（提供给前端）
openssl rsa -in private_key.pem -pubout -out public_key.pem
```

### 2. 初始化加密器

```go
import "admin/pkg/rsapwd"

// 从配置或环境变量中读取私钥
cipher := rsapwd.MustNew(privateKeyPEM)
```

### 3. 前端加密（JavaScript）

```javascript
import JSEncrypt from 'jsencrypt';

const encrypt = new JSEncrypt();
encrypt.setPublicKey(publicKey);
const encryptedPassword = encrypt.encrypt(password);
```

### 4. 后端解密（Go）

```go
// 解密前端传来的加密密码
decryptedPassword, err := cipher.DecryptPKCS1(req.Password)
if err != nil {
    return nil, xerr.Wrap(xerr.ErrInvalidCredentials, "密码解密失败", err)
}
```

## API 参考

### 构造函数

```go
func New(privateKeyPEM string) (*RSACipher, error)
func MustNew(privateKeyPEM string) *RSACipher
```

### 加密解密方法

```go
// 使用 PKCS#1 v1.5 加密（与 JSEncrypt 兼容）
func (r *RSACipher) EncryptPKCS1(plaintext string) (string, error)

// 解密 PKCS#1 v1.5 加密的密文
func (r *RSACipher) DecryptPKCS1(ciphertextBase64 string) (string, error)

// SHA256 哈希（可选的前端预处理）
func HashPassword(password string) string
```

### 公钥导出

```go
// 导出 PEM 格式的公钥
func (r *RSACipher) ExportPublicKey() string

// 导出 Base64 编码的公钥（不含 PEM 头尾）
func (r *RSACipher) ExportPublicKeyBase64() string
```

## 错误类型

- `ErrInvalidPrivateKey`: 私钥无效
- `ErrInvalidBase64`: Base64 编码无效
- `ErrDecryptFailed`: 解密失败

## 常见问题

**Q: 为什么使用 PKCS#1 v1.5 而不是 OAEP？**

A: JSEncrypt 等主流前端 RSA 加密库默认使用 PKCS#1 v1.5 填充。为了保持前后端兼容性，后端必须使用对应的 `DecryptPKCS1()` 方法进行解密。

**Q: 需要先 SHA256 哈希再 RSA 加密吗？**

A: 不是必需的，但作为额外的安全措施可以这样做。前端可以在加密前对密码进行 SHA256 哈希。

**Q: 支持哪些 RSA 私钥格式？**

A: 支持 PKCS1 和 PKCS8 两种格式的私钥，程序会自动识别。

## 测试

```bash
# 运行所有测试
go test ./pkg/rsapwd/... -v

# 运行兼容性测试
go test ./pkg/rsapwd/... -run TestRSACompatibility -v
```

## 示例

- [rsa_test.go](rsa_test.go) - 单元测试
- [compatibility_test.go](compatibility_test.go) - 前后端兼容性测试
- [example_test.go](example_test.go) - 使用示例
- [frontend-rsa-example.js](frontend-rsa-example.js) - 前端完整示例
- [USAGE.md](USAGE.md) - 完整集成指南

## 安全建议

1. **密钥管理**：私钥应存储在环境变量或密钥管理服务中，不要硬编码在代码中
2. **传输安全**：始终使用 HTTPS（TLS/SSL）
3. **密码存储**：使用 Argon2 或 bcrypt 等安全哈希算法存储密码
4. **前端库**：使用 JSEncrypt 等成熟的加密库，确保使用 PKCS#1 v1.5 填充（默认）
