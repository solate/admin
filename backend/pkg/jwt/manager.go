package jwt

import (
	"context"
	"errors"
)

type JWTManager struct {
	config *JWTConfig
	store  Store
}

func NewJWTManager(config *JWTConfig, store Store) *JWTManager {
	return &JWTManager{
		config: config,
		store:  store,
	}
}

func (m *JWTManager) GenerateTokenPair(ctx context.Context, tenantID, userID, roleID string) (*TokenPair, error) {
	tokenPair, err := GenerateTokenPair(tenantID, userID, roleID, m.config)
	if err != nil {
		return nil, err
	}

	// 将refresh Token 存储
	err = m.store.Set(ctx, tokenPair.TokenID, tokenPair.RefreshToken, m.config.RefreshExpire)
	if err != nil {
		return nil, err
	}

	// 可选：限制用户设备数量

	return tokenPair, nil
}

// RefreshToken 刷新token
func (m *JWTManager) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// 验证refresh token
	claims, err := VerifyToken(refreshToken, m.config.RefreshSecret)
	if err != nil {
		return nil, err
	}

	// 从store中获取refresh token
	storedToken, err := m.store.Get(ctx, claims.TokenID)
	if err != nil {
		return nil, err
	}

	// 验证refresh token是否匹配
	if storedToken != refreshToken {
		return nil, errors.New("refresh token not match")
	}

	// 生成新的token pair
	tokenPair, err := GenerateTokenPair(claims.TenantID, claims.UserID, claims.RoleID, m.config)
	if err != nil {
		return nil, err
	}

	// 更新store中的refresh token
	err = m.store.Set(ctx, tokenPair.TokenID, tokenPair.RefreshToken, m.config.RefreshExpire)
	if err != nil {
		return nil, err
	}

	// 删除旧的refresh token
	err = m.store.Delete(ctx, claims.TokenID)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}
