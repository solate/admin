package jwt

import (
	"context"
	"errors"
	"fmt"
)

type Manager struct {
	config *JWTConfig
	store  Store
}

// NewJWTManager 创建 JWT 管理器
// 参数：
// - config: JWT 配置信息
// - store: token 存储后端（Redis 或其他）
func NewManager(config *JWTConfig, store Store) *Manager {
	return &Manager{
		config: config,
		store:  store,
	}
}

// generateUserKey 生成用户会话索引 key
// 格式：tenantID:userID
func (m *Manager) generateUserKey(tenantID, userID string) string {
	return fmt.Sprintf("%s:%s", tenantID, userID)
}

// GenerateTokenPair 生成令牌对（access + refresh）
// 说明：
// - 生成新的 access token 和 refresh token
// - 将 refresh token 存储到 store
// - 维护用户会话索引（便于后续跨设备登出）
func (m *Manager) GenerateTokenPair(ctx context.Context, tenantID, tenantCode, userID, userName string, roles []string) (*TokenPair, error) {
	tokenPair, err := GenerateTokenPair(tenantID, tenantCode, userID, userName, roles, m.config)
	if err != nil {
		return nil, fmt.Errorf("generate token pair failed: %w", err)
	}

	// 将 refresh token 存储
	if err := m.store.Set(ctx, tokenPair.TokenID, tokenPair.RefreshToken, m.config.RefreshExpire); err != nil {
		return nil, fmt.Errorf("store refresh token failed: %w", err)
	}

	// 将会话索引到用户集合（键建议采用 tenantCode:userID）
	userKey := m.generateUserKey(tenantCode, userID)
	if err := m.store.AddUserToken(ctx, userKey, tokenPair.TokenID, m.config.RefreshExpire); err != nil {
		return nil, fmt.Errorf("add user token index failed: %w", err)
	}

	return tokenPair, nil
}

// VerifyAccessToken 验证 access token（签名/过期/黑名单）
// 说明：
// - 先校验签名与过期；若过期将返回 ErrTokenExpired（来自第三方库）
// - 再检查是否命中黑名单，命中则返回 ErrTokenBlacklisted
// 返回值：
// - Claims: token 中的声明信息
// - error: 验证失败的错误原因
func (m *Manager) VerifyAccessToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims, err := VerifyToken(tokenString, []byte(m.config.AccessSecret))
	if err != nil {
		return nil, err
	}

	// 黑名单校验：命中即视为撤销
	blacklisted, err := m.store.IsBlacklisted(ctx, claims.TokenID)
	if err != nil {
		return nil, fmt.Errorf("check blacklist failed: %w", err)
	}
	if blacklisted {
		return nil, ErrTokenBlacklisted
	}

	return claims, nil
}

// VerifyRefreshToken 验证 refresh token（签名/过期/黑名单/匹配）
// 说明：
// - 先校验签名与过期；若过期将返回 ErrTokenExpired（来自第三方库）
// - 再检查是否命中黑名单，命中则返回 ErrTokenBlacklisted
// - 检查 refresh token 是否匹配存储值
// 返回值：
// - TokenPair: 刷新后的令牌对（access + refresh）
// - error: 验证失败的错误原因
func (m *Manager) VerifyRefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// 验证 refresh token
	claims, err := VerifyToken(refreshToken, []byte(m.config.RefreshSecret))
	if err != nil {
		return nil, err
	}

	// 从 store 中获取存储的 refresh token
	storedToken, err := m.store.Get(ctx, claims.TokenID)
	if err != nil {
		return nil, fmt.Errorf("get stored refresh token failed: %w", err)
	}

	// 验证 refresh token 是否匹配
	if storedToken != refreshToken {
		return nil, errors.New("refresh token not match")
	}

	// 撤销旧会话：将旧 tokenID 置入黑名单，TTL 为 access token 生命周期
	// 说明：刷新后旧 access token 需要立即失效，避免并发窗口
	if err := m.store.BlacklistToken(ctx, claims.TokenID, m.config.AccessExpire); err != nil {
		return nil, fmt.Errorf("blacklist old token failed: %w", err)
	}

	// 生成新的 token 对
	tokenPair, err := GenerateTokenPair(claims.TenantID, claims.TenantCode, claims.UserID, claims.UserName, claims.Roles, m.config)
	if err != nil {
		return nil, fmt.Errorf("generate new token pair failed: %w", err)
	}

	// 存储新的 refresh token
	if err := m.store.Set(ctx, tokenPair.TokenID, tokenPair.RefreshToken, m.config.RefreshExpire); err != nil {
		return nil, fmt.Errorf("store new refresh token failed: %w", err)
	}

	// 删除旧的 refresh token
	if err := m.store.Delete(ctx, claims.TokenID); err != nil {
		return nil, fmt.Errorf("delete old refresh token failed: %w", err)
	}

	// 维护用户会话索引：移除旧 tokenID，加入新 tokenID
	userKey := m.generateUserKey(claims.TenantCode, claims.UserID)
	if err := m.store.RemoveUserToken(ctx, userKey, claims.TokenID); err != nil {
		return nil, fmt.Errorf("remove old user token index failed: %w", err)
	}
	if err := m.store.AddUserToken(ctx, userKey, tokenPair.TokenID, m.config.RefreshExpire); err != nil {
		return nil, fmt.Errorf("add new user token index failed: %w", err)
	}

	return tokenPair, nil
}

// RevokeToken 撤销指定 tokenID 的会话
// 说明：
// - 将 tokenID 放入黑名单（TTL 为 access token 生命周期）
// - 删除对应的 refresh token
func (m *Manager) RevokeToken(ctx context.Context, tokenID string) error {
	// 黑名单标记
	if err := m.store.BlacklistToken(ctx, tokenID, m.config.AccessExpire); err != nil {
		return fmt.Errorf("blacklist token failed: %w", err)
	}
	// 删除对应 refresh token
	if err := m.store.Delete(ctx, tokenID); err != nil {
		return fmt.Errorf("delete refresh token failed: %w", err)
	}
	return nil
}

// RevokeAllUserTokens 撤销某个用户的所有会话（跨设备登出）
// 说明：
// - 读取用户会话集合，批量黑名单标记并删除 refresh token
// - 最后逐个从集合移除 tokenID
func (m *Manager) RevokeAllUserTokens(ctx context.Context, tenantID, userID string) error {
	userKey := m.generateUserKey(tenantID, userID)
	tokenIDs, err := m.store.GetUserTokens(ctx, userKey)
	if err != nil {
		return fmt.Errorf("get user tokens failed: %w", err)
	}

	for _, tid := range tokenIDs {
		if err := m.store.BlacklistToken(ctx, tid, m.config.AccessExpire); err != nil {
			return fmt.Errorf("blacklist user token failed: %w", err)
		}
		if err := m.store.Delete(ctx, tid); err != nil {
			return fmt.Errorf("delete user token failed: %w", err)
		}
		if err := m.store.RemoveUserToken(ctx, userKey, tid); err != nil {
			return fmt.Errorf("remove user token index failed: %w", err)
		}
	}

	return nil
}
