package rsapwd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

// 测试用的RSA私钥（2048位）
const testRSAPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQC+1GvoPXuNkVyR
yPmC+Qw3AmlYjf+ol6cyB6jM++sB9U7vlY6owOxzLeK/R2JAiY61nhcO90vJWgmz
CeE9uq6704xrf2ONCCll5DNd3zz3WhpknTHuOXS2/3e2u26eiFR94ZHMctnD34jo
LA6VL+qHi3mTqZ6RlDeES26IvbKZB0NhsbZ2cmaN7oxBVoJzHipHyf0/2YWJQQJS
IukIJbl9RNc4IDhdCPeyfK7SJGUbZIjsiraBs9fn1Sz89jpVjeMHXnHQDGMver7c
XI8StrmdSbd5zVoxc8ZG5r1TpyQPEeyb/uE8zxwsCKusGMMcB37J5fFvUWMSJ4HA
kSBoQO5/AgMBAAECggEABMxyqjRhlv3Axim3nIOGuxtkasWnWCX4Hlny9LShBDuW
8I9iNvwi9gKBYS36WoUbAZYoHkg5r6aD9+yXrWW0XyTCszFQ34sE/3rtj769Wbr6
Tu1lBAiN1sw1xnKQJYxoE4JImEuLDlHgr3XsJ/Q8gYwQUpZBVofTnZAIB4g9pXtu
JKmU4OkyLQaQqZwvIjtWtVjNSeC+VGqh3p5FP13jv2kBlzFksrtYYY/yDB6Uj4c8
Grj8RwOEmz02txsUU3TPbcYSiLjjUwcK9pgHic3uPMSv4wy6Mvm6ZA75XWgBOsrd
Lg6l1tsz8Fz4utnfRbVIdsrQ9z3LHLTTnnOpKf3PAQKBgQDpAu6wtVEgnxzGyFup
Weciwk9zP3qXTzD68rNy4EuUo0z6crbbo++aHJ10l6GKtTjXnvBPfMsoi9zqOCEQ
8dpNMe7e+X2mMCGuJz8e83OOGedroGi8BWoNRyysRZO3BdXzPja48Dx2wRgqQ5Km
pBrDuORctPJv9c98rzxXwv/plwKBgQDRqB/8qw9lRu/0qy2vF5aSmXAwHAQuC+69
2697flrnIozVSK0bdfxMqvKM2/eQITtInFixlRiWsq1k5N1T3o3bBko+hT/Rg/N5
doOhi8WmH4MgZRp7bAyeAMHsVfDXea+NG1K8kgDQW+Kqcg+5Hb3KlrXQDxGJF2Hp
N/ejuoovWQKBgQCi4X3g4J5ZY2BGRIBunX3I+nN3aIRViPIAOe/e+ZNbz9tbpxzT
5ID1BdO7UNOHlq6pa10o819AdKR0xc+3fJjRJXqJO3Xt2e9xQdYJ2LyKNOlkfrk3
1cEQjxRXSDu90MKCSpcOKEDb8pbl1F6LRmO/NVvMwmBGi1oDGqvf3Vvu+QKBgQCL
MozKPOij3U1DrMNQFOErxCPwTSmZSOLhuxHvdBz2iMHoebA1I0i3vmf7ja/4SZgK
xYM9pDgHFep5qloobQLSAIMar22HtYvZgQ40G5DGkvWEdJv4hex6mxYly4l0Bp6/
mPx9ppJTxC3h7Ijz5wMzloxv7xE9bADdzwLj+d31QQKBgQDkTwtclVg90MLMOyx+
9XNbucBnnKMQvMd753n1y/2whhgqdgffMGgRALIhkHLsGkM4irACJytFU6T5mq7n
xqto3Ux5up3yALGP0jfXfcV5IjfbXkkgu+w0xMaWvXx5hgiftY/SBeZz/ujUJ2XG
aSyRaef1MpYPbhjZmFIY0+dI9w==
-----END PRIVATE KEY-----`

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		privateKey  string
		expectError bool
	}{
		{
			name:        "有效的PKCS1格式私钥",
			privateKey:  testRSAPrivateKey,
			expectError: false,
		},
		{
			name:        "空的私钥",
			privateKey:  "",
			expectError: true,
		},
		{
			name:        "无效的私钥",
			privateKey:  "invalid key",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipher, err := New(tt.privateKey)
			if tt.expectError {
				if err == nil {
					t.Error("期望返回错误，但没有")
				}
				if cipher != nil {
					t.Error("期望返回nil，但得到了cipher")
				}
			} else {
				if err != nil {
					t.Errorf("不期望返回错误: %v", err)
				}
				if cipher == nil {
					t.Error("期望返回cipher，但得到了nil")
				}
				if cipher != nil && cipher.privateKey == nil {
					t.Error("cipher.privateKey不应该为nil")
				}
			}
		})
	}
}

func TestMustNew(t *testing.T) {
	t.Run("有效的私钥不应该panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("不应该panic: %v", r)
			}
		}()
		cipher := MustNew(testRSAPrivateKey)
		if cipher == nil {
			t.Error("期望返回cipher，但得到了nil")
		}
	})

	t.Run("无效的私钥应该panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望panic，但没有")
			}
		}()
		MustNew("invalid key")
	})
}

func TestHashPassword(t *testing.T) {
	t.Run("相同的密码应该产生相同的哈希", func(t *testing.T) {
		password := "my_password"
		hash1 := HashPassword(password)
		hash2 := HashPassword(password)

		if hash1 != hash2 {
			t.Error("相同的密码应该产生相同的哈希")
		}

		// SHA256哈希应该是64个字符
		if len(hash1) != 64 {
			t.Errorf("SHA256哈希应该是64个字符，得到: %d", len(hash1))
		}
	})

	t.Run("不同的密码应该产生不同的哈希", func(t *testing.T) {
		hash1 := HashPassword("password1")
		hash2 := HashPassword("password2")

		if hash1 == hash2 {
			t.Error("不同的密码应该产生不同的哈希")
		}
	})

	t.Run("哈希示例验证", func(t *testing.T) {
		// 验证示例中的密码 "toor"
		pwd := "toor"
		hash := HashPassword(pwd)
		t.Logf("密码 '%s' 的SHA256哈希: %s", pwd, hash)
	})
}

// generateTestKey 生成测试用的RSA密钥对
func generateTestKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// TestPKCS8Key 测试PKCS8格式的私钥
func TestPKCS8Key(t *testing.T) {
	// 生成一个新的RSA密钥对
	priv, err := generateTestKey()
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}

	// 转换为PKCS8格式
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("转换PKCS8失败: %v", err)
	}

	// 编码为PEM
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}
	pemKey := pem.EncodeToMemory(block)

	// 测试解析
	cipher, err := New(string(pemKey))
	if err != nil {
		t.Fatalf("解析PKCS8密钥失败: %v", err)
	}

	// 测试加密解密
	plaintext := "test_password"
	ciphertext, err := cipher.EncryptPKCS1(plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	decrypted, err := cipher.DecryptPKCS1(ciphertext)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密结果不匹配: 期望 %s, 得到 %s", plaintext, decrypted)
	}
}
