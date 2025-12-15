package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// 自定义Claims结构体
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
	AccessExpire  int64  // access token过期时间
	RefreshSecret []byte // refresh token密钥
	RefreshExpire int64  // refresh token过期时间
}

// 生成token对 (access + refresh)
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

func VerifyToken(tokenString string, secret []byte) (*Claims, error) {
	return verifyToken(tokenString, secret)
}

func RefreshToken(tokenString string, config *JWTConfig) (*TokenPair, error) {
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

// 生成单个Token
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
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// 验证token
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

// RemoveBearerPrefix 去除Bearer前缀的通用函数
// 处理各种情况：Bearer token、Bearer  token、Bearer   token等
func RemoveBearerPrefix(tokenString string) string {
	tokenString = strings.TrimSpace(tokenString)
	tokenString = strings.TrimPrefix(tokenString, "Bearer")
	return strings.TrimSpace(tokenString)
}
