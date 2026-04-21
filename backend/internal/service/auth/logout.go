package auth

import (
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Logout 用户登出
func (s *Service) Logout(ctx context.Context, r *http.Request) error {
	tokenID := xcontext.GetTokenID(ctx)
	if tokenID == "" {
		return xerr.ErrUnauthorized
	}

	if err := s.jwt.RevokeToken(ctx, tokenID); err != nil {
		log.Error().Err(err).Str("token_id", tokenID).Msg("撤销token失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "撤销token失败", err)
	}

	s.recorder.Logout(ctx)

	return nil
}
