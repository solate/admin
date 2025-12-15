package jwt

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func testConfig() *JWTConfig {
	return &JWTConfig{
		AccessSecret:  []byte("access-secret"),
		AccessExpire:  3600,
		RefreshSecret: []byte("refresh-secret"),
		RefreshExpire: 7200,
	}
}

func TestGenerateTokenPairAndParse(t *testing.T) {
	cfg := testConfig()

	pair, err := GenerateTokenPair("tenant-1", "user-1", "role-1", cfg)
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("GenerateTokenPair returned empty tokens")
	}
	if pair.TokenID == "" {
		t.Fatalf("GenerateTokenPair returned empty token id")
	}

	accessClaims, err := VerifyToken(pair.AccessToken, []byte(cfg.AccessSecret))
	if err != nil {
		t.Fatalf("VerifyToken(access) returned error: %v", err)
	}
	if accessClaims.TenantID != "tenant-1" || accessClaims.UserID != "user-1" || accessClaims.RoleID != "role-1" {
		t.Fatalf("unexpected access claims: %+v", accessClaims)
	}
	if accessClaims.TokenID != pair.TokenID {
		t.Fatalf("access token token_id = %s, want %s", accessClaims.TokenID, pair.TokenID)
	}

	refreshClaims, err := VerifyToken(pair.RefreshToken, []byte(cfg.RefreshSecret))
	if err != nil {
		t.Fatalf("VerifyToken(refresh) returned error: %v", err)
	}
	if refreshClaims.TokenID != pair.TokenID {
		t.Fatalf("refresh token token_id = %s, want %s", refreshClaims.TokenID, pair.TokenID)
	}
}

func TestVerifyTokenInvalidSignature(t *testing.T) {
	cfg := testConfig()

	pair, err := GenerateTokenPair("tenant-1", "user-1", "role-1", cfg)
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	_, err = VerifyToken(pair.AccessToken, []byte("wrong-secret"))
	if err == nil {
		t.Fatalf("VerifyToken should have failed with wrong secret")
	}
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Fatalf("expected ErrTokenSignatureInvalid, got: %v", err)
	}
}

func TestParseTokenExpired(t *testing.T) {
	secret := []byte("secret")
	expiredCfg := &JWTConfig{
		AccessSecret:  secret,
		AccessExpire:  -1,
		RefreshSecret: secret,
		RefreshExpire: -1,
		Issuer:        "admin",
	}

	token, err := generateToken("tenant-1", "user-1", "role-1", "token-1", expiredCfg.AccessExpire, secret, expiredCfg.Issuer)
	if err != nil {
		t.Fatalf("generateToken returned error: %v", err)
	}

	time.Sleep(2 * time.Second)

	_, err = VerifyToken(token, secret)
	if err == nil {
		t.Fatalf("VerifyToken should have returned expiration error")
	}
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got: %v", err)
	}
}

func TestRemoveBearerPrefix(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no_prefix", "token", "token"},
		{"single_space", "Bearer token", "token"},
		{"multiple_space", "Bearer   token", "token"},
		{"leading_space", "   Bearer token", "token"},
		{"trailing_space", "Bearer token   ", "token"},
		{"lowercase_prefix", "bearer token", "bearer token"},
	}

	for _, tt := range tests {
		got := RemoveBearerPrefix(tt.input)
		if got != tt.want {
			t.Fatalf("RemoveBearerPrefix(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
