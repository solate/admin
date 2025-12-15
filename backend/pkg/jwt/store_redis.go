package jwt

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// 刷新令牌键前缀：refresh_token:{tokenID}
	RefreshTokenKeyPrefix = "refresh_token:"
	// 令牌元数据键前缀（可选扩展使用）：token_meta:{tokenID}
	TokenMetaKeyPrefix = "token_meta:"
	// 用户会话集合键前缀：user_tokens:{userID or tenantID:userID}
	UserTokensKeyPrefix = "user_tokens:"
	// 黑名单键前缀：blacklist:{tokenID}
	BlacklistKeyPrefix = "blacklist:"
)

// 使用redis 存储 refresh token
type redisStore struct {
	client redis.UniversalClient
}

func NewRedisStore(client redis.UniversalClient) Store {
	return &redisStore{client: client}
}

func (s *redisStore) Set(ctx context.Context, tokenID, refreshToken string, expiration int64) error {
	// 写入 refresh_token:{tokenID} = refreshToken，TTL 为 refresh token 生命周期
	key := RefreshTokenKeyPrefix + tokenID
	err := s.client.Set(ctx, key, refreshToken, time.Duration(expiration)*time.Second).Err()
	if err != nil {
		return err
	}

	// // 4. 可选：存储token元数据
	// tokenMetadata := map[string]interface{}{
	// 	"user_id":   userID,
	// 	"username":  username,
	// 	"issued_at": refreshClaims.IssuedAt.Unix(),
	// }
	// err = m.redis.HSet(ctx, fmt.Sprintf("token_meta:%s", refreshToken), tokenMetadata).Err()
	// if err != nil {
	// 	// 记录日志，但不阻止返回
	// 	fmt.Printf("Warning: failed to store token metadata: %v\n", err)
	// }

	return nil

}

func (s *redisStore) Get(ctx context.Context, tokenID string) (string, error) {
	// 读取 refresh_token:{tokenID}
	key := RefreshTokenKeyPrefix + tokenID
	refreshToken, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *redisStore) Delete(ctx context.Context, tokenID string) error {
	// 删除 refresh_token:{tokenID}
	key := RefreshTokenKeyPrefix + tokenID
	return s.client.Del(ctx, key).Err()
}

func (s *redisStore) AddUserToken(ctx context.Context, userID string, tokenID string, expiration int64) error {
	// 将 tokenID 加入 user_tokens:{userID} 集合，并设置集合 TTL（便于自动清理）
	key := UserTokensKeyPrefix + userID
	if err := s.client.SAdd(ctx, key, tokenID).Err(); err != nil {
		return err
	}
	if expiration > 0 {
		if err := s.client.Expire(ctx, key, time.Duration(expiration)*time.Second).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *redisStore) RemoveUserToken(ctx context.Context, userID string, tokenID string) error {
	// 从 user_tokens:{userID} 集合移除 tokenID
	key := UserTokensKeyPrefix + userID
	return s.client.SRem(ctx, key, tokenID).Err()
}

func (s *redisStore) GetUserTokens(ctx context.Context, userID string) ([]string, error) {
	// 获取 user_tokens:{userID} 集合的所有 tokenID
	key := UserTokensKeyPrefix + userID
	members, err := s.client.SMembers(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []string{}, nil
		}
		return nil, err
	}
	return members, nil
}

func (s *redisStore) BlacklistToken(ctx context.Context, tokenID string, expiration int64) error {
	// 设置 blacklist:{tokenID} = 1，并设置 TTL，用于标记该 tokenID 已撤销
	key := BlacklistKeyPrefix + tokenID
	var ttl time.Duration
	if expiration > 0 {
		ttl = time.Duration(expiration) * time.Second
	}
	return s.client.Set(ctx, key, "1", ttl).Err()
}

func (s *redisStore) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	// 判断 blacklist:{tokenID} 是否存在
	key := BlacklistKeyPrefix + tokenID
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
