package jwt

import (
	"context"
)

// Store 会话与黑名单的统一存储抽象
// 说明：
// - 刷新令牌持久化：用 tokenID 作为键，值为 refreshToken，本质是维持会话可刷新能力
// - 用户会话索引：用 userID 作为集合键（可传入组合键 tenantID:userID），集合成员为该用户的所有 tokenID
// - 黑名单：对被撤销的 tokenID 建立短期标记（TTL 建议为 access token 剩余有效时间），用于即时失效
type Store interface {
	// 刷新令牌存储：tokenID -> refreshToken
	Set(ctx context.Context, tokenID string, refreshToken string, expiration int64) error
	Get(ctx context.Context, tokenID string) (string, error)
	Delete(ctx context.Context, tokenID string) error

	// 用户会话索引：userID -> Set{tokenID...}
	AddUserToken(ctx context.Context, userID string, tokenID string, expiration int64) error
	RemoveUserToken(ctx context.Context, userID string, tokenID string) error
	GetUserTokens(ctx context.Context, userID string) ([]string, error)

	// 黑名单：撤销某个 tokenID（通常 TTL 设为 access token 剩余时间）
	BlacklistToken(ctx context.Context, tokenID string, expiration int64) error
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
}
