package jwt

import (
	"testing"
	"time"
)

func TestGenerateIDProducesUniqueValues(t *testing.T) {
	ids := make(map[string]struct{}, 20)
	for i := 0; i < 20; i++ {
		ids[generateID()] = struct{}{}
	}

	if len(ids) != 20 {
		t.Fatalf("expected 20 unique IDs, got %d", len(ids))
	}
}

func TestGeneratePairAssignsTokenType(t *testing.T) {
	manager := &JWTManager{
		secret:             []byte("super-secret-key"),
		accessTokenExpiry:  time.Hour,
		refreshTokenExpiry: 24 * time.Hour,
		issuer:             "sigif-test",
	}

	pair, err := manager.GeneratePair("user-1", "tenant-1", "user@example.com", []string{"admin"})
	if err != nil {
		t.Fatalf("GeneratePair returned error: %v", err)
	}

	accessClaims, err := manager.Validate(pair.AccessToken)
	if err != nil {
		t.Fatalf("Validate access token returned error: %v", err)
	}
	if accessClaims.TokenType != AccessTokenType {
		t.Fatalf("expected access token type %q, got %q", AccessTokenType, accessClaims.TokenType)
	}

	refreshClaims, err := manager.Validate(pair.RefreshToken)
	if err != nil {
		t.Fatalf("Validate refresh token returned error: %v", err)
	}
	if refreshClaims.TokenType != RefreshTokenType {
		t.Fatalf("expected refresh token type %q, got %q", RefreshTokenType, refreshClaims.TokenType)
	}
}
