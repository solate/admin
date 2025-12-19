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
	TenantID   string   `json:"tenant_id"`          // 租户ID
	TenantCode string   `json:"tenant_code"`        // 租户编码
	UserID     string   `json:"user_id"`            // 用户ID
	UserName   string   `json:"user_name"`          // 用户名
	RoleType   int32    `json:"role_type"`          // 角色类型(1:普通用户, 2:租户管理员, 3:平台超级管理员)
	Roles      []string `json:"roles"`              // 角色列表
	TokenID    string   `json:"token_id,omitempty"` // refresh token的唯一标识
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenID      string `json:"token_id,omitempty"` // refresh token的唯一标识
	ExpiresIn    int64  `json:"expires_in"`         // access token 过期时间（秒）
}

type JWTConfig struct {
	AccessSecret  []byte // access token密钥（支持字符串格式）
	AccessExpire  int64  // access token过期时间（秒）
	RefreshSecret []byte // refresh token密钥（支持字符串格式）
	RefreshExpire int64  // refresh token过期时间（秒）
	Issuer        string // 发行者（可选）
}

// 公开错误变量用于中间件与业务层识别
var (
	ErrTokenExpired      = jwt.ErrTokenExpired
	ErrTokenBlacklisted  = errors.New("token blacklisted")
	ErrInvalidToken      = errors.New("invalid token")
	ErrMissingToken      = errors.New("missing token")
	ErrInvalidClaims     = errors.New("invalid claims")
	ErrInvalidSignMethod = errors.New("invalid signing method")
)

// GenerateTokenPair 生成令牌对（access + refresh）
// 注意：
// - 使用随机 TokenID 作为会话标识，便于后续刷新和撤销
// - ExpiresIn 返回 access token 的过期时间（秒）
func GenerateTokenPair(tenantID, tenantCode, userID, userName string, roleType int32, roles []string, config *JWTConfig) (*TokenPair, error) {
	// 生成refresh token的唯一ID
	tokenID := uuid.New().String()

	// 生成 access token
	accessToken, err := generateToken(tenantID, tenantCode, userID, userName, roleType, roles, tokenID, config.AccessExpire, config.AccessSecret, config.Issuer)
	if err != nil {
		return nil, err
	}

	// 生成 refresh token
	refreshToken, err := generateToken(tenantID, tenantCode, userID, userName, roleType, roles, tokenID, config.RefreshExpire, config.RefreshSecret, config.Issuer)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenID:      tokenID,
		ExpiresIn:    config.AccessExpire,
	}, nil
}

// VerifyToken 验证任意 token（去除 Bearer 前缀，校验签名与过期）
func VerifyToken(tokenString string, secret []byte) (*Claims, error) {
	return verifyToken(tokenString, secret)
}

// generateToken 生成单个 token（带 Claims）
func generateToken(tenantID, tenantCode, userID, userName string, roleType int32, roles []string, tokenID string, expire int64, secret []byte, issuer string) (string, error) {
	now := time.Now()
	claims := &Claims{
		TenantID:   tenantID,
		TenantCode: tenantCode,
		UserID:     userID,
		UserName:   userName,
		RoleType:   roleType,
		Roles:      roles,
		TokenID:    tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenStr, nil
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
