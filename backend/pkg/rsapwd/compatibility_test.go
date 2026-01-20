package rsapwd_test

import (
	"fmt"
	"testing"

	"admin/pkg/constants"
	"admin/pkg/rsapwd"
)

// TestRSACompatibility 测试 RSA 前后端兼容性
//
// 问题诊断：
// 1. JSEncrypt 使用 PKCS#1 v1.5 填充，而 Go 后端原本使用 OAEP 填充
// 2. 前端必须使用与后端私钥对应的公钥
// 3. 需要添加 DecryptPKCS1 方法来兼容 JSEncrypt
func TestRSACompatibility(t *testing.T) {
	// 使用后端的实际私钥
	cipher := rsapwd.MustNew(constants.RSAKey)

	// 导出公钥（提供给前端使用）
	publicKeyPEM := cipher.ExportPublicKey()
	fmt.Println("\n=============================================")
	fmt.Println("前端应该使用的公钥（与后端私钥配对）：")
	fmt.Println("=============================================")
	fmt.Println(publicKeyPEM)
	fmt.Println("=============================================")

	// 测试密码
	testPassword := "myPassword123"
	testPasswordHash := rsapwd.HashPassword(testPassword)

	fmt.Printf("\n原始密码: %s\n", testPassword)
	fmt.Printf("SHA256哈希: %s\n", testPasswordHash)

	// 测试1: PKCS#1 v1.5 加密/解密（与 JSEncrypt 兼容）
	fmt.Println("\n=============================================")
	fmt.Println("测试 PKCS#1 v1.5 加密/解密（JSEncrypt 兼容）")
	fmt.Println("=============================================")

	// 加密明文密码
	encryptedPlain, err := cipher.EncryptPKCS1(testPassword)
	if err != nil {
		t.Fatalf("PKCS#1 加密明文失败: %v", err)
	}
	fmt.Printf("PKCS#1 加密明文: %s\n", encryptedPlain)

	// 解密明文密码
	decryptedPlain, err := cipher.DecryptPKCS1(encryptedPlain)
	if err != nil {
		t.Fatalf("PKCS#1 解密明文失败: %v", err)
	}
	fmt.Printf("PKCS#1 解密结果: %s\n", decryptedPlain)

	if decryptedPlain != testPassword {
		t.Errorf("解密结果不匹配: got %s, want %s", decryptedPlain, testPassword)
	}

	// 加密哈希后的密码
	encryptedHash, err := cipher.EncryptPKCS1(testPasswordHash)
	if err != nil {
		t.Fatalf("PKCS#1 加密哈希失败: %v", err)
	}
	fmt.Printf("PKCS#1 加密哈希: %s\n", encryptedHash)

	// 解密哈希后的密码
	decryptedHash, err := cipher.DecryptPKCS1(encryptedHash)
	if err != nil {
		t.Fatalf("PKCS#1 解密哈希失败: %v", err)
	}
	fmt.Printf("PKCS#1 解密结果: %s\n", decryptedHash)

	if decryptedHash != testPasswordHash {
		t.Errorf("解密结果不匹配: got %s, want %s", decryptedHash, testPasswordHash)
	}

}

// TestPublicKeyExport 测试公钥导出
func TestPublicKeyExport(t *testing.T) {
	cipher := rsapwd.MustNew(constants.RSAKey)

	publicKeyPEM := cipher.ExportPublicKey()
	if publicKeyPEM == "" {
		t.Fatal("导出公钥失败")
	}

	fmt.Println("\n导出的公钥（PEM格式）：")
	fmt.Println(publicKeyPEM)

	publicKeyBase64 := cipher.ExportPublicKeyBase64()
	if publicKeyBase64 == "" {
		t.Fatal("导出公钥Base64失败")
	}

	fmt.Printf("\n导出的公钥（Base64格式）：\n%s\n", publicKeyBase64)
}

// TestJSEncryptWorkflow 模拟 JSEncrypt 的工作流程
func TestJSEncryptWorkflow(t *testing.T) {
	fmt.Println("\n=============================================")
	fmt.Println("JSEncrypt 工作流程模拟")
	fmt.Println("=============================================")

	cipher := rsapwd.MustNew(constants.RSAKey)

	// 前端代码示例：
	// const encrypt = new JSEncrypt();
	// encrypt.setPublicKey(publicKey);  // 使用从后端获取的公钥
	// encrypted = encrypt.encrypt(password);  // 使用 PKCS#1 v1.5

	steps := []string{
		"1. 前端从后端获取公钥（调用 /api/public-key 接口）",
		"2. 前端使用 JSEncrypt 设置公钥：encrypt.setPublicKey(publicKey)",
		"3. 前端加密密码：encrypted = encrypt.encrypt(password)",
		"4. 前端发送加密后的密码到后端",
		"5. 后端使用 DecryptPKCS1() 解密",
	}

	for _, step := range steps {
		fmt.Println(step)
	}

	fmt.Println("\n后端解密示例代码：")
	fmt.Println("```go")
	fmt.Println("decryptedPassword, err := s.rsaCipher.DecryptPKCS1(req.Password)")
	fmt.Println("if err != nil {")
	fmt.Println("    log.Error().Err(err).Msg(\"密码解密失败\")")
	fmt.Println("    return nil, xerr.Wrap(xerr.ErrInvalidCredentials.Code, \"密码解密失败\", err)")
	fmt.Println("}")
	fmt.Println("```")

	// 测试完整流程
	testPassword := "user123456"
	encrypted, err := cipher.EncryptPKCS1(testPassword)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	decrypted, err := cipher.DecryptPKCS1(encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if decrypted != testPassword {
		t.Errorf("密码不匹配: got %s, want %s", decrypted, testPassword)
	}

	fmt.Printf("\n测试成功！密码: %s -> 加密 -> 解密 -> %s\n", testPassword, decrypted)
	fmt.Println("=============================================")
}
