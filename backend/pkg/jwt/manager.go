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

func (m *JWTManager) generateUserKey(tenantID, userID string) string {
	return tenantID + ":" + userID
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

	// 将会话索引到用户集合（键建议采用 tenantID:userID）
	userKey := m.generateUserKey(tenantID, userID)
	if err := m.store.AddUserToken(ctx, userKey, tokenPair.TokenID, m.config.RefreshExpire); err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// VerifyAccessToken 验证 access token（签名/过期/黑名单）
// 说明：
// - 先校验签名与过期；若过期将返回 ErrTokenExpired（来自第三方库）
// - 再检查是否命中黑名单，命中则返回 ErrTokenBlacklisted
func (m *JWTManager) VerifyAccessToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims, err := VerifyToken(tokenString, m.config.AccessSecret)
	if err != nil {
		return nil, err
	}
	// 黑名单校验：命中即视为撤销
	blacklisted, err := m.store.IsBlacklisted(ctx, claims.TokenID)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, ErrTokenBlacklisted
	}
	return claims, nil
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

	// 撤销旧会话：将旧 tokenID 置入黑名单，TTL 建议为 access token 生命周期
	// 说明：刷新后旧 access token 需要立即失效，避免并发窗口
	if err := m.store.BlacklistToken(ctx, claims.TokenID, m.config.AccessExpire); err != nil {
		return nil, err
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

	// 维护用户会话索引：移除旧 tokenID，加入新 tokenID
	userKey := m.generateUserKey(claims.TenantID, claims.UserID)
	if err := m.store.RemoveUserToken(ctx, userKey, claims.TokenID); err != nil {
		return nil, err
	}
	if err := m.store.AddUserToken(ctx, userKey, tokenPair.TokenID, m.config.RefreshExpire); err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// RevokeToken 撤销指定 tokenID 的会话
// 说明：
// - 将 tokenID 放入黑名单（access 生命周期）
// - 删除对应的 refresh token
func (m *JWTManager) RevokeToken(ctx context.Context, tokenID string) error {
	// 黑名单标记
	if err := m.store.BlacklistToken(ctx, tokenID, m.config.AccessExpire); err != nil {
		return err
	}
	// 删除对应 refresh token
	if err := m.store.Delete(ctx, tokenID); err != nil {
		return err
	}
	return nil
}

// RevokeAllUserTokens 撤销某个用户的所有会话（跨设备登出）
// 说明：
// - 读取用户会话集合，批量黑名单标记并删除 refresh token
// - 最后逐个从集合移除 tokenID
func (m *JWTManager) RevokeAllUserTokens(ctx context.Context, tenantID, userID string) error {
	userKey := m.generateUserKey(tenantID, userID)
	tokenIDs, err := m.store.GetUserTokens(ctx, userKey)
	if err != nil {
		return err
	}
	for _, tid := range tokenIDs {
		if err := m.store.BlacklistToken(ctx, tid, m.config.AccessExpire); err != nil {
			return err
		}
		if err := m.store.Delete(ctx, tid); err != nil {
			return err
		}
		if err := m.store.RemoveUserToken(ctx, userKey, tid); err != nil {
			return err
		}
	}
	return nil
}
