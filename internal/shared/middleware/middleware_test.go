package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/jwt"
)

type stubSessionReader struct {
	active bool
}

func (s stubSessionReader) IsActive(_ context.Context, _ string) (bool, error) {
	return s.active, nil
}

func newTestJWTManager(t *testing.T) *jwt.JWTManager {
	t.Helper()
	cfg := &config.Config{JWT: config.JWTConfig{
		Secret:            "super-secret-key",
		AccessTokenExpiry: 60,
		Issuer:            "sigif-test",
	}}
	return jwt.NewManager(cfg)
}

func TestAuthRequiredRejectsMissingToken(t *testing.T) {
	app := fiber.New()
	app.Use(AuthRequired(newTestJWTManager(t), stubSessionReader{active: true}))
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

func TestAuthRequiredRejectsInactiveSession(t *testing.T) {
	manager := newTestJWTManager(t)
	app := fiber.New()
	app.Use(AuthRequired(manager, stubSessionReader{active: false}))
	app.Get("/protected", func(c *fiber.Ctx) error { return c.SendString("ok") })

	token, _, err := manager.GenerateAccessToken(
		uuid.NewString(), uuid.NewString(), "", "user@example.com", "superadmin",
	)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestAuthRequiredAcceptsValidAccessToken(t *testing.T) {
	manager := newTestJWTManager(t)
	app := fiber.New()
	app.Use(AuthRequired(manager, stubSessionReader{active: true}))
	app.Get("/protected", func(c *fiber.Ctx) error { return c.SendString("ok") })

	token, _, err := manager.GenerateAccessToken(
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
		"",
		"user@example.com",
		"superadmin",
	)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected %d, got %d", fiber.StatusOK, resp.StatusCode)
	}
}

func TestRequirePermissionDeniesWithoutPermission(t *testing.T) {
	manager := newTestJWTManager(t)
	app := fiber.New()
	app.Use(AuthRequired(manager, stubSessionReader{active: true}))
	app.Get("/users", RequirePermission(denyChecker{}, "users", "list"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	token, _, _ := manager.GenerateAccessToken(uuid.NewString(), uuid.NewString(), "", "u@e.com", "soporte")
	req := httptest.NewRequest(fiber.MethodGet, "/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected %d, got %d", fiber.StatusForbidden, resp.StatusCode)
	}
}

type denyChecker struct{}

func (denyChecker) HasPermission(_ context.Context, _ uuid.UUID, _, _ string) (bool, error) {
	return false, nil
}
