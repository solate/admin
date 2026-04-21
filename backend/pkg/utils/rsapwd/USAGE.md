# rsapwd 使用指南

## 完整的集成示例

### 1. 生成 RSA 密钥对

```bash
# 生成 2048 位 RSA 私钥
openssl genrsa -out config/private_key.pem 2048

# 提取公钥（给前端使用）
openssl rsa -in config/private_key.pem -pubout -out config/public_key.pem

# 查看公钥内容（用于前端）
cat config/public_key.pem
```

### 2. 配置文件设置

在 `config/config.yaml` 中添加：

```yaml
rsa:
  private_key: |
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpAIBAAKCAQEAwfMyTEQ05zPp0uB30OOQovZr8sYoRupBqsZUesEwnIV37WbU
    1MRwnrwQRytPtinYXBw+poxYZZ4a5+tQuxfMMyELIYqrxqL6D4ILydSTAQA2dSws
    ...
    -----END RSA PRIVATE KEY-----
```

### 3. 后端实现

#### 3.1 初始化全局加密器

```go
// internal/app/app.go
package app

import (
    "admin/pkg/config"
    "admin/pkg/rsapwd"
)

var RSACipher *rsapwd.RSACipher

func InitRSACipher() error {
    privateKeyPEM := config.Get().RSA.PrivateKey
    cipher, err := rsapwd.New(privateKeyPEM)
    if err != nil {
        return err
    }
    RSACipher = cipher
    return nil
}
```

#### 3.2 在 Handler 中使用

```go
// internal/handler/auth_handler.go
package handler

import (
    "context"
    "admin/internal/app"
    "admin/pkg/xerr"
    "admin/pkg/rsapwd"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"` // 前端加密后的密码（base64编码）
}

func (h *AuthHandler) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    // 1. 解密前端传来的加密密码（使用 DecryptPKCS1 兼容 JSEncrypt）
    decryptedPassword, err := app.RSACipher.DecryptPKCS1(req.Password)
    if err != nil {
        return nil, xerr.Wrap(xerr.ErrInvalidCredentials, "密码解密失败", err)
    }

    // 2. 查询用户
    user, err := h.userRepo.GetByUsername(ctx, req.Username)
    if err != nil {
        return nil, err
    }

    // 3. 验证密码（使用 Argon2）
    if !passwordgen.VerifyPassword(decryptedPassword, user.PasswordHash) {
        return nil, xerr.ErrInvalidCredentials
    }

    // 4. 生成 JWT token
    token, err := h.jwtManager.Generate(user)
    if err != nil {
        return nil, err
    }

    return &LoginResponse{
        Token: token,
        User:  user,
    }, nil
}
```

### 4. 前端实现

#### 4.1 安装依赖

```bash
npm install jsencrypt
# 或
yarn add jsencrypt
```

#### 4.2 工具函数

```javascript
// src/utils/crypto.js
import JSEncrypt from 'jsencrypt';

// RSA 公钥（从后端获取或硬编码）
const RSA_PUBLIC_KEY = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwfMyTEQ05zPp0uB30OOQ
...
-----END PUBLIC KEY-----`;

/**
 * SHA256 哈希
 * @param {string} message - 要哈希的消息
 * @returns {Promise<string>} 十六进制格式的哈希值
 */
export async function sha256(message) {
    const msgBuffer = new TextEncoder().encode(message);
    const hashBuffer = await crypto.subtle.digest('SHA256', msgBuffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}

/**
 * RSA 加密
 * @param {string} plaintext - 要加密的明文
 * @returns {string} Base64 编码的密文
 */
export function encryptWithRSA(plaintext) {
    const encrypt = new JSEncrypt();
    encrypt.setPublicKey(RSA_PUBLIC_KEY);
    const encrypted = encrypt.encrypt(plaintext);
    if (!encrypted) {
        throw new Error('RSA 加密失败');
    }
    return encrypted;
}

/**
 * 加密密码（SHA256 + RSA）
 * @param {string} password - 明文密码
 * @returns {Promise<string>} 加密后的密码
 */
export async function encryptPassword(password) {
    // 1. SHA256 哈希
    const hashedPassword = await sha256(password);

    // 2. RSA 加密
    const encrypted = encryptWithRSA(hashedPassword);

    return encrypted;
}
```

#### 4.3 在登录组件中使用

```vue
<!-- src/views/Login.vue -->
<template>
  <form @submit.prevent="handleLogin">
    <input v-model="username" type="text" placeholder="用户名" />
    <input v-model="password" type="password" placeholder="密码" />
    <button type="submit">登录</button>
  </form>
</template>

<script setup>
import { ref } from 'vue';
import { encryptPassword } from '@/utils/crypto';
import { login } from '@/api/auth';

const username = ref('');
const password = ref('');

const handleLogin = async () => {
  try {
    // 加密密码
    const encryptedPassword = await encryptPassword(password.value);

    // 发送登录请求
    const response = await login({
      username: username.value,
      password: encryptedPassword, // 发送加密后的密码
    });

    console.log('登录成功:', response);
  } catch (error) {
    console.error('登录失败:', error);
  }
};
</script>
```

#### 4.4 API 调用

```javascript
// src/api/auth.js
import axios from 'axios';

export async function login(data) {
  const response = await axios.post('/api/v1/auth/login', data);
  return response.data;
}
```

### 5. 测试

#### 5.1 测试解密功能

```go
// pkg/rsapwd/integration_test.go
package rsapwd_test

import (
    "testing"
    "admin/pkg/rsapwd"
)

func TestPasswordEncryptionFlow(t *testing.T) {
    // 初始化加密器
    cipher := rsapwd.MustNew(testRSAPrivateKey)

    // 模拟前端加密流程（使用 JSEncrypt，即 PKCS#1 v1.5）
    plainPassword := "test123"
    encryptedPassword, _ := cipher.EncryptPKCS1(plainPassword)

    // 模拟后端解密流程
    decryptedPassword, err := cipher.DecryptPKCS1(encryptedPassword)
    if err != nil {
        t.Fatalf("解密失败: %v", err)
    }

    // 验证
    if decryptedPassword != plainPassword {
        t.Errorf("密码不匹配: 期望 %s, 得到 %s", plainPassword, decryptedPassword)
    }

    t.Logf("密码加密传输流程测试通过")
}
```

## 安全建议

1. **密钥管理**：
   - 私钥应该存储在环境变量或密钥管理服务中
   - 不要将私钥提交到版本控制系统
   - 使用 `gitignore` 排除私钥文件

2. **HTTPS 必需**：
   - 始终使用 HTTPS 传输
   - RSA 加密作为额外的安全层

3. **前端公钥管理**：
   - 公钥可以硬编码在前端代码中
   - 或者通过 API 端点动态获取

4. **错误处理**：
   - 不要泄露详细的解密错误信息
   - 使用通用的"用户名或密码错误"消息

## 常见问题

**Q: 为什么需要先 SHA256 哈希再 RSA 加密？**

A: 这样可以确保即使 RSA 被破解，攻击者得到的也只是哈希值。同时，哈希后的密码长度固定，更适合 RSA 加密。

**Q: 如何处理密钥轮换？**

A:
1. 生成新的 RSA 密钥对
2. 更新服务端配置
3. 更新前端公钥
4. 逐步迁移用户到新的密钥

**Q: 支持哪些 RSA 密钥格式？**

A: 支持 PKCS1 和 PKCS8 两种格式，程序会自动识别。
