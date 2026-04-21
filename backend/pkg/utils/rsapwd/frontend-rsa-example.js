/**
 * 前端 RSA 加密示例
 *
 * 此文件展示如何正确使用 JSEncrypt 库与后端 RSA 解密配合
 *
 * 关键点：
 * 1. 前端必须使用从后端获取的公钥（与后端私钥配对）
 * 2. JSEncrypt 默认使用 PKCS#1 v1.5 填充，后端使用 DecryptPKCS1() 解密
 * 3. 公钥应该从后端接口获取，或配置在前端环境变量中
 */

import JSEncrypt from 'jsencrypt';

// ============================================
// 方式一：公钥硬编码（不推荐，仅用于测试）
// ============================================
const PUBLIC_KEY = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwfMyTEQ05zPp0uB30OOQ
ovZr8sYoRupBqsZUesEwnIV37WbU1MRwnrwQRytPtinYXBw+poxYZZ4a5+tQuxfM
MyELIYqrxqL6D4ILydSTAQA2dSws5nQZkspQup0DLVGkAU+HUDCYyOUPAgom7wJ6
EaarhNDRlAitJGM4sha27Xh55Nbhw7/Fj2BLaOa8EWRvW0giYf1AFRU0BzMejGdH
FVxlgsSM2/E/62U6VTIh3Uw2GqSJWRri0bQSw79j01Q8Q5gF1ahaCa56fT6hOVUG
qt6Cc0Zc6ei1u0z8jcBdS6+B9T0oU6Aalry1wARcu0AKnuwPb72C+19AM4U8Ki/T
hQIDAQAB
-----END PUBLIC KEY-----`;

// ============================================
// 方式二：从后端获取公钥（推荐）
// ============================================
async function getPublicKey() {
  try {
    // 调用后端公钥接口（需要后端实现）
    const response = await fetch('/api/public-key');
    const data = await response.json();
    return data.publicKey;
  } catch (error) {
    console.error('获取公钥失败:', error);
    // 降级使用硬编码的公钥
    return PUBLIC_KEY;
  }
}

// ============================================
// 加密函数
// ============================================
/**
 * 使用 RSA 加密密码
 * @param {string} password - 明文密码
 * @param {string} publicKey - RSA 公钥（可选，默认使用内置公钥）
 * @returns {string} - Base64 编码的加密密文
 */
export function encryptPassword(password, publicKey = PUBLIC_KEY) {
  const encrypt = new JSEncrypt();
  encrypt.setPublicKey(publicKey);

  const encrypted = encrypt.encrypt(password);
  if (!encrypted) {
    throw new Error('密码加密失败');
  }

  return encrypted;
}

// ============================================
// 完整的登录流程示例
// ============================================
/**
 * 手机号登录示例
 */
async function loginByPhone(phone, password) {
  try {
    // 1. 获取公钥（可选，如果后端公钥不变可以跳过）
    const publicKey = await getPublicKey();

    // 2. 加密密码
    const encryptedPassword = encryptPassword(password, publicKey);

    // 3. 发送登录请求
    const response = await fetch('/api/auth/login/phone', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        phone: phone,
        password: encryptedPassword, // 发送加密后的密码
        captcha_id: captchaId,
        captcha: captchaCode,
      }),
    });

    const data = await response.json();

    if (data.code === 200) {
      console.log('登录成功');
      return data.data;
    } else {
      console.error('登录失败:', data.message);
      return null;
    }
  } catch (error) {
    console.error('登录错误:', error);
    return null;
  }
}

/**
 * 邮箱登录示例
 */
async function loginByEmail(email, password) {
  try {
    // 1. 获取公钥（可选）
    const publicKey = await getPublicKey();

    // 2. 加密密码
    const encryptedPassword = encryptPassword(password, publicKey);

    // 3. 发送登录请求
    const response = await fetch('/api/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email: email,
        password: encryptedPassword, // 发送加密后的密码
        captcha_id: captchaId,
        captcha: captchaCode,
      }),
    });

    const data = await response.json();

    if (data.code === 200) {
      console.log('登录成功');
      return data.data;
    } else {
      console.error('登录失败:', data.message);
      return null;
    }
  } catch (error) {
    console.error('登录错误:', error);
    return null;
  }
}

// ============================================
// 修改密码示例
// ============================================
/**
 * 修改密码
 */
async function changePassword(oldPassword, newPassword) {
  try {
    // 1. 获取公钥
    const publicKey = await getPublicKey();

    // 2. 加密旧密码和新密码
    const encryptedOldPassword = encryptPassword(oldPassword, publicKey);
    const encryptedNewPassword = encryptPassword(newPassword, publicKey);

    // 3. 发送修改密码请求
    const response = await fetch('/api/user/password', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${accessToken}`, // 需要携带认证令牌
      },
      body: JSON.stringify({
        old_password: encryptedOldPassword,
        new_password: encryptedNewPassword,
      }),
    });

    const data = await response.json();

    if (data.code === 200) {
      console.log('密码修改成功');
      return true;
    } else {
      console.error('密码修改失败:', data.message);
      return false;
    }
  } catch (error) {
    console.error('修改密码错误:', error);
    return false;
  }
}

// ============================================
// Vue 3 Composition API 示例
// ============================================
import { ref } from 'vue';

export function useAuth() {
  const loading = ref(false);
  const error = ref(null);

  const login = async (email, password) => {
    loading.value = true;
    error.value = null;

    try {
      const publicKey = await getPublicKey();
      const encryptedPassword = encryptPassword(password, publicKey);

      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email,
          password: encryptedPassword,
          captcha_id: captchaId,
          captcha: captchaCode,
        }),
      });

      const data = await response.json();

      if (data.code === 200) {
        return { success: true, data: data.data };
      } else {
        error.value = data.message;
        return { success: false, error: data.message };
      }
    } catch (err) {
      error.value = err.message;
      return { success: false, error: err.message };
    } finally {
      loading.value = false;
    }
  };

  return {
    loading,
    error,
    login,
  };
}

// ============================================
// React Hook 示例
// ============================================
import { useState, useCallback } from 'react';

export function useAuth() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const login = useCallback(async (email, password) => {
    setLoading(true);
    setError(null);

    try {
      const publicKey = await getPublicKey();
      const encryptedPassword = encryptPassword(password, publicKey);

      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email,
          password: encryptedPassword,
          captcha_id: captchaId,
          captcha: captchaCode,
        }),
      });

      const data = await response.json();

      if (data.code === 200) {
        return { success: true, data: data.data };
      } else {
        setError(data.message);
        return { success: false, error: data.message };
      }
    } catch (err) {
      setError(err.message);
      return { success: false, error: err.message };
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    loading,
    error,
    login,
  };
}

// ============================================
// 重要说明
// ============================================
/**
 * 后端解密方式说明：
 *
 * 后端使用 DecryptPKCS1() 方法解密，因为：
 * 1. JSEncrypt 库默认使用 PKCS#1 v1.5 填充
 * 2. Go 的 OAEP 解密无法解密 PKCS#1 v1.5 加密的数据
 * 3. 必须使用 DecryptPKCS1() 才能正确解密
 *
 * 后端代码示例：
 *
 * ```go
 * decryptedPassword, err := s.rsaCipher.DecryptPKCS1(req.Password)
 * if err != nil {
 *     log.Error().Err(err).Msg("密码解密失败")
 *     return nil, xerr.Wrap(xerr.ErrInvalidCredentials.Code, "密码解密失败", err)
 * }
 * ```
 *
 * 前端可选的额外安全措施：
 * 在发送前对密码进行 SHA256 哈希：
 *
 * ```javascript
 * import { HashPassword } from '@/utils/crypto';
 * const hashedPassword = HashPassword(password);
 * const encryptedPassword = encryptPassword(hashedPassword, publicKey);
 * ```
 *
 * 测试验证：
 * 运行测试验证前后端兼容性：
 * ```bash
 * go test -v ./pkg/rsapwd/... -run TestRSACompatibility
 * ```
 */
