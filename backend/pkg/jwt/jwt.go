package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims 自定义声明
// 说明：
// - TokenID 为会话唯一标识（access/refresh 均携带），用于黑名单与会话管理
type Claims struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	RoleID   string `json:"role_id"`
	TokenID  string `json:"token_id,omitempty"` // refresh token的唯一标识
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenID      string `json:"token_id,omitempty"` // refresh token的唯一标识
}

type JWTConfig struct {
	AccessSecret  []byte // access token密钥
	AccessExpire  int64  // access token过期时间（秒）
	RefreshSecret []byte // refresh token密钥
	RefreshExpire int64  // refresh token过期时间（秒）
	Issuer        string // 发行者（可选）
}

// 公开错误变量用于中间件与业务层识别
// 说明：
// - ErrTokenExpired 直接引用第三方库的过期错误，便于 errors.Is 判断
// - ErrTokenBlacklisted 表示令牌已被撤销（命中黑名单）
var (
	ErrTokenExpired     = jwt.ErrTokenExpired
	ErrTokenBlacklisted = errors.New("token blacklisted")
)

// GenerateTokenPair 生成令牌对（access + refresh）
// 注意：
// - 使用随机 TokenID 作为会话标识，便于后续刷新和撤销
func GenerateTokenPair(tenantID, userID, roleID string, config *JWTConfig) (*TokenPair, error) {

	// 生成refresh token的唯一ID
	tokenID := uuid.New().String()

	// 生成 access token
	accessToken, err := generateToken(tenantID, userID, roleID, tokenID, config.AccessExpire, config.AccessSecret)
	if err != nil {
		return nil, err
	}

	// 生成 refresh token
	refreshToken, err := generateToken(tenantID, userID, roleID, tokenID, config.RefreshExpire, config.RefreshSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenID:      tokenID,
	}, nil
}

// VerifyToken 验证任意 token（去除 Bearer 前缀，校验签名与过期）
func VerifyToken(tokenString string, secret []byte) (*Claims, error) {
	return verifyToken(tokenString, secret)
}

// refreshTokenPair 根据 refresh token 生成新的令牌对（无状态工具）
// 说明：
// - 仅适用于不依赖存储的场景；在有状态场景下应使用 JWTManager.RefreshToken
func refreshTokenPair(tokenString string, config *JWTConfig) (*TokenPair, error) {
	claims, err := verifyToken(tokenString, config.RefreshSecret)
	if err != nil {
		return nil, err
	}
	tokenPair, err := GenerateTokenPair(claims.TenantID, claims.UserID, claims.RoleID, config)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// generateToken 生成单个 token（带 Claims）
func generateToken(tenantID, userID, roleID, tokenID string, expire int64, secret []byte) (string, error) {
	now := time.Now()
	claims := &Claims{
		TenantID: tenantID,
		UserID:   userID,
		RoleID:   roleID,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "", // 由调用方在构造 JWTConfig 时填充
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// verifyToken 验证 token（签名算法 & 过期）
func verifyToken(tokenString string, secret []byte) (*Claims, error) {
	tokenString = RemoveBearerPrefix(tokenString)
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if jwt.GetSigningMethod(jwt.SigningMethodHS256.Alg()).Alg() != token.Method.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RemoveBearerPrefix 去除 Bearer 前缀的通用函数（兼容多空格）
func RemoveBearerPrefix(tokenString string) string {
	tokenString = strings.TrimSpace(tokenString)
	tokenString = strings.TrimPrefix(tokenString, "Bearer")
	return strings.TrimSpace(tokenString)
}
