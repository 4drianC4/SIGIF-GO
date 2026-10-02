package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/jwt"
)

func TestAuthRequiredRejectsMissingToken(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{JWT: config.JWTConfig{
		Secret:             "super-secret-key",
		AccessTokenExpiry:  60,
		RefreshTokenExpiry: 1440,
		Issuer:             "sigif-test",
	}}
	jwtManager := jwt.NewManager(cfg)

	app.Use(AuthRequired(jwtManager))
	app.Get("/protected", func(c *fiber.Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestAuthRequiredAcceptsValidAccessToken(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{JWT: config.JWTConfig{
		Secret:             "super-secret-key",
		AccessTokenExpiry:  60,
		RefreshTokenExpiry: 1440,
		Issuer:             "sigif-test",
	}}
	jwtManager := jwt.NewManager(cfg)

	pair, err := jwtManager.GeneratePair("11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222", "user@example.com", []string{"admin"})
	if err != nil {
		t.Fatalf("GeneratePair returned error: %v", err)
	}

	app.Use(AuthRequired(jwtManager))
	app.Get("/protected", func(c *fiber.Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected %d, got %d", fiber.StatusOK, resp.StatusCode)
	}
}
