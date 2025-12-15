package jwt

import (
	"context"
)

// Store 定义 Token 存储接口, 可以没有redis
type Store interface {
	Set(ctx context.Context, tokenID string, refreshToken string, expiration int64) error
	Get(ctx context.Context, tokenID string) (string, error)
	Delete(ctx context.Context, tokenID string) error
	Check(ctx context.Context, tokenID string, refreshToken string) error
}

// nopStore 空实现，用于无状态模式
type nopStore struct{}

func (s *nopStore) Set(ctx context.Context, tokenID string, refreshToken string, expiration int64) error {
	return nil
}
func (s *nopStore) Get(ctx context.Context, tokenID string) (string, error) {
	return "", nil
}
func (s *nopStore) Delete(ctx context.Context, tokenID string) error {
	return nil
}
func (s *nopStore) Check(ctx context.Context, tokenID string, refreshToken string) error {
	return nil
}
