package passwordgen

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	passwordChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+,.?/:;{}[]~"
	saltLength    = 16 // 推荐128位盐值
)

// GeneratePassword 生成符合复杂要求的随机密码
func GeneratePassword(length int) (string, error) {
	if length < 12 {
		return "", errors.New("密码长度最少需要12个字符")
	}

	var password []byte
	charSets := []string{
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"abcdefghijklmnopqrstuvwxyz",
		"0123456789",
		"!@#$%^&*()-_=+,.?/:;{}[]~",
	}

	// 确保每个字符集至少包含一个字符
	for _, set := range charSets {
		char, err := randomChar(set)
		if err != nil {
			return "", err
		}
		password = append(password, char)
	}

	// 填充剩余字符
	for i := 4; i < length; i++ {
		char, err := randomChar(passwordChars)
		if err != nil {
			return "", err
		}
		password = append(password, char)
	}

	shuffle(password)
	return string(password), nil
}

// GenerateSalt 生成加密安全的随机盐值，返回 base64 编码的字符串
func GenerateSalt() (string, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

// Argon2Hash 使用Argon2id算法进行密码哈希，返回格式化的哈希字符串
func Argon2Hash(password string, saltBase64 string) (string, error) {
    // 解码 base64 盐值
    salt, err := base64.StdEncoding.DecodeString(saltBase64)
    if err != nil {
        return "", fmt.Errorf("解码盐值失败: %v", err)
    }

    // 推荐参数配置（根据硬件性能调整）
    timeCost := uint32(3)           // 迭代次数
    memoryCost := uint32(64 * 1024) // 内存消耗（KB）
    parallelism := uint8(2)         // 并行度
    keyLength := uint32(32)         // 输出密钥长度

    hash := argon2.IDKey(
        []byte(password),
        salt,
        timeCost,
        memoryCost,
        parallelism,
        keyLength,
    )

    // 使用标准编码格式存储参数
    encodedHash := base64.StdEncoding.EncodeToString(hash)

    // 返回格式化的字符串，包含所有参数
    return fmt.Sprintf(
        "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
        argon2.Version,
        memoryCost,
        timeCost,
        parallelism,
        saltBase64,
        encodedHash,
    ), nil
}

// VerifyPassword 验证密码与哈希是否匹配
func VerifyPassword(password, encodedHash string) bool {
    // 解析哈希参数
    parts := strings.Split(encodedHash, "$")
    if len(parts) != 6 || parts[1] != "argon2id" {
        return false
    }

    // 解析盐值（现在是 base64 编码的）
    salt, err := base64.StdEncoding.DecodeString(parts[4])
    if err != nil {
        return false
    }

    // 解析存储的哈希值
    storedHash, err := base64.StdEncoding.DecodeString(parts[5])
    if err != nil {
        return false
    }

    // 解析算法参数
    params := strings.Split(parts[3], ",")
    if len(params) != 3 {
        return false
    }

    timeCost, err := parseUint32(params[1], "t=", 10)
    if err != nil {
        return false
    }

    memoryCost, err := parseUint32(params[0], "m=", 10)
    if err != nil {
        return false
    }

    parallelism, err := parseUint32(params[2], "p=", 10)
    if err != nil {
        return false
    }

    // 重新计算哈希
    newHash := argon2.IDKey(
        []byte(password),
        salt,
        timeCost,
        memoryCost,
        uint8(parallelism),
        uint32(len(storedHash)),
    )

    // 恒定时间比较
    return subtle.ConstantTimeCompare(storedHash, newHash) == 1
}

// 辅助函数
func randomChar(charset string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

func shuffle(s []byte) {
	for i := len(s) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		s[i], s[j.Int64()] = s[j.Int64()], s[i]
	}
}

func parseUint32(s, prefix string, base int) (uint32, error) {
	value := strings.TrimPrefix(s, prefix)
	if value == s {
		return 0, fmt.Errorf("前缀 %s 未找到", prefix)
	}

	n := new(big.Int)
	_, success := n.SetString(value, base)
	if !success {
		return 0, fmt.Errorf("无效的数字格式: %s", value)
	}

	if !n.IsUint64() || n.Uint64() > math.MaxUint32 {
		return 0, fmt.Errorf("数值超出32位范围: %s", value)
	}

	return uint32(n.Uint64()), nil
}
