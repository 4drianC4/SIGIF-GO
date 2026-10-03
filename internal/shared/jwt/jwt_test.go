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

func TestAccessTokenRoundTrip(t *testing.T) {
	manager := &JWTManager{
		secret:            []byte("super-secret-key"),
		accessTokenExpiry: time.Hour,
		issuer:            "sigif-test",
	}

	token, _, err := manager.GenerateAccessToken(
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
		"33333333-3333-3333-3333-333333333333",
		"user@example.com",
		"superadmin",
	)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	claims, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken returned error: %v", err)
	}

	if claims.TokenType != TokenTypeAccess {
		t.Fatalf("expected token type %q, got %q", TokenTypeAccess, claims.TokenType)
	}
	if claims.UserID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected user id %q", claims.UserID)
	}
	if claims.SessionID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("unexpected session id %q", claims.SessionID)
	}
	if claims.Role != "superadmin" {
		t.Fatalf("unexpected role %q", claims.Role)
	}
}

func TestValidateRejectsInvalidToken(t *testing.T) {
	manager := &JWTManager{
		secret:            []byte("super-secret-key"),
		accessTokenExpiry: time.Hour,
		issuer:            "sigif-test",
	}

	if _, err := manager.ValidateAccessToken("not-a-token"); err == nil {
		t.Fatal("expected error validating an invalid token")
	}
}
