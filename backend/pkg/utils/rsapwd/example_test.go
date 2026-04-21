package rsapwd_test

import (
	"context"
	"fmt"

	"admin/pkg/constants"
	"admin/pkg/rsapwd"
)

// Example_inHandler 展示了在 HTTP Handler 中使用 rsapwd 的完整示例
func Example_inHandler() {
	// 初始化加密器（全局单例）
	// 注意：实际使用时应该从配置或环境变量中读取私钥
	// privateKeyPEM 应该从 config.Get().RSA.PrivateKey 或 os.Getenv("RSA_PRIVATE_KEY") 获取
	// 这里使用测试密钥仅用于演示
	privateKeyPEM := constants.RSAKey
	globalRSACipher := rsapwd.MustNew(privateKeyPEM)

	// 模拟登录处理函数
	DecryptPassword := func(ctx context.Context, reqPassword string) (string, error) {
		// 解密前端传来的加密密码（使用 DecryptPKCS1 兼容 JSEncrypt）
		decrypted, err := globalRSACipher.DecryptPKCS1(reqPassword)
		if err != nil {
			return "", fmt.Errorf("密码解密失败: %w", err)
		}

		// decrypted 是前端用 JSEncrypt 加密的密码（PKCS#1 v1.5）
		// 可选：前端可能在加密前先做了 SHA256 哈希
		// 直接用于数据库验证
		return decrypted, nil
	}

	// 使用示例
	encryptedPassword := "base64_encoded_rsa_encrypted_password"
	password, err := DecryptPassword(context.Background(), encryptedPassword)
	if err != nil {
		// 返回错误响应
		fmt.Println("密码解密失败:", err)
		return
	}

	// 使用 password 与数据库中的哈希值进行验证
	fmt.Println("解密后的密码:", password)

	// Output:
	// 密码解密失败: 密码解密失败: base64编码无效: illegal base64 data at input byte 6
}
