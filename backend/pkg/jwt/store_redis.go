package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	RefreshTokenKeyPrefix = "refresh_token:" // refresh_token:user_id
	TokenMetaKeyPrefix    = "token_meta:"    // token_meta:user_id
)

// 使用redis 存储 refresh token
type redisStore struct {
	client redis.UniversalClient
}

func NewRedisStore(client redis.UniversalClient) Store {
	return &redisStore{client: client}
}

func (s *redisStore) Set(ctx context.Context, tokenID, refreshToken string, expiration int64) error {

	// 将Refresh Token存储到Redis
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
	// 从Redis获取Refresh Token
	key := RefreshTokenKeyPrefix + tokenID
	refreshToken, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *redisStore) Check(ctx context.Context, tokenID string, refreshToken string) error {
	storeToken, err := s.Get(ctx, tokenID)
	if err != nil {
		return err
	}
	if storeToken != refreshToken {
		return errors.New("token mismatch")
	}

	return nil
}

func (s *redisStore) Delete(ctx context.Context, tokenID string) error {
	// 从Redis删除Refresh Token
	key := RefreshTokenKeyPrefix + tokenID
	return s.client.Del(ctx, key).Err()
}
