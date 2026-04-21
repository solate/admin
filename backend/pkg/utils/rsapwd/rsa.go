package rsapwd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
)

var (
	// ErrInvalidPrivateKey 私钥无效
	ErrInvalidPrivateKey = errors.New("私钥无效")
	// ErrInvalidBase64 base64编码无效
	ErrInvalidBase64 = errors.New("base64编码无效")
	// ErrDecryptFailed 解密失败
	ErrDecryptFailed = errors.New("解密失败")
)

// RSACipher RSA加密解密器
type RSACipher struct {
	privateKey *rsa.PrivateKey
}

// New 从PEM格式的私钥字符串创建RSACipher
func New(privateKeyPEM string) (*RSACipher, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, ErrInvalidPrivateKey
	}

	// 尝试解析PKCS1格式的私钥
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return &RSACipher{privateKey: priv}, nil
	}

	// 尝试解析PKCS8格式的私钥
	priv8, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPrivateKey, err)
	}

	rsaKey, ok := priv8.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidPrivateKey
	}

	return &RSACipher{privateKey: rsaKey}, nil
}

// MustNew 从PEM格式的私钥字符串创建RSACipher，如果失败则panic
// 用于初始化阶段，确保私钥格式正确
func MustNew(privateKeyPEM string) *RSACipher {
	cipher, err := New(privateKeyPEM)
	if err != nil {
		panic(fmt.Sprintf("初始化RSA加密器失败: %v", err))
	}
	return cipher
}

// Encrypt OAEP 加密（已弃用，与 JSEncrypt 不兼容）
// 使用 EncryptPKCS1 代替
/*
func (r *RSACipher) Encrypt(plaintext string) (string, error) {
	if r.privateKey == nil {
		return "", ErrInvalidPrivateKey
	}

	// 使用公钥加密
	ciphertext, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		&r.privateKey.PublicKey,
		[]byte(plaintext),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
*/

// Decrypt OAEP 解密（已弃用，与 JSEncrypt 不兼容）
// 使用 DecryptPKCS1 代替
/*
func (r *RSACipher) Decrypt(ciphertextBase64 string) (string, error) {
	if r.privateKey == nil {
		return "", ErrInvalidPrivateKey
	}

	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidBase64, err)
	}

	// RSA解密
	plaintext, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		r.privateKey,
		ciphertext,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	return string(plaintext), nil
}
*/

// HashPassword 对密码进行SHA256哈希（前端预处理）
// 将明文密码转换为SHA256哈希值后再传输
func HashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

// DecryptAndHash 解密密码并验证SHA256哈希（已弃用）
// 此方法与 JSEncrypt 不兼容
/*
func (r *RSACipher) DecryptAndHash(ciphertextBase64 string) (string, error) {
	// 解密
	plaintext, err := r.Decrypt(ciphertextBase64)
	if err != nil {
		return "", err
	}

	// 验证是否已经是SHA256格式（64个字符的十六进制字符串）
	if len(plaintext) == 64 {
		// 尝试验证是否为有效的十六进制字符串
		if _, err := hex.DecodeString(plaintext); err == nil {
			// 已经是SHA256哈希，直接返回
			return plaintext, nil
		}
	}

	// 不是SHA256格式，进行哈希
	return HashPassword(plaintext), nil
}
*/

// DecryptPKCS1 解密使用 PKCS#1 v1.5 填充加密的密文
// 与 JSEncrypt 库兼容
// 返回解密后的明文
func (r *RSACipher) DecryptPKCS1(ciphertextBase64 string) (string, error) {
	if r.privateKey == nil {
		return "", ErrInvalidPrivateKey
	}

	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidBase64, err)
	}

	// 使用 PKCS#1 v1.5 解密（与 JSEncrypt 兼容）
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, r.privateKey, ciphertext)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	return string(plaintext), nil
}

// EncryptPKCS1 使用 PKCS#1 v1.5 填充加密明文
// 返回base64编码的密文
func (r *RSACipher) EncryptPKCS1(plaintext string) (string, error) {
	if r.privateKey == nil {
		return "", ErrInvalidPrivateKey
	}

	// 使用公钥加密
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, &r.privateKey.PublicKey, []byte(plaintext))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// ExportPublicKey 导出PEM格式的公钥
// 用于提供给前端进行加密
func (r *RSACipher) ExportPublicKey() string {
	if r.privateKey == nil {
		return ""
	}

	// 将公钥转换为 PKIX 格式
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&r.privateKey.PublicKey)
	if err != nil {
		return ""
	}

	// PEM 编码
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return string(pubKeyPEM)
}

// ExportPublicKeyBase64 导出Base64编码的公钥（不含PEM头尾）
// 用于某些需要裸公钥的场景
func (r *RSACipher) ExportPublicKeyBase64() string {
	if r.privateKey == nil {
		return ""
	}

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&r.privateKey.PublicKey)
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(pubKeyBytes)
}
