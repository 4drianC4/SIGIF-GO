package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/authcookie"
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
	app.Use(AuthRequired(newTestJWTManager(t), stubSessionReader{active: true}, authcookie.Default()))
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
	app.Use(AuthRequired(manager, stubSessionReader{active: false}, authcookie.Default()))
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
	app.Use(AuthRequired(manager, stubSessionReader{active: true}, authcookie.Default()))
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
	app.Use(AuthRequired(manager, stubSessionReader{active: true}, authcookie.Default()))
	app.Get("/users", RequirePermission(denyChecker{}, "users", "list"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	token, _, _ := manager.GenerateAccessToken(uuid.NewString(), uuid.NewString(), "", "u@e.com", "business_admin")
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

func newCookieTestApp(manager *jwt.JWTManager, active bool) *fiber.App {
	app := fiber.New()
	app.Use(AuthRequired(manager, stubSessionReader{active: active}, authcookie.Default()))
	app.Get("/protected", func(c *fiber.Ctx) error {
		userID, _ := UserIDFromContext(c.UserContext())
		return c.SendString(userID.String())
	})
	return app
}

func newTestToken(t *testing.T, manager *jwt.JWTManager, userID string) string {
	t.Helper()
	token, _, err := manager.GenerateAccessToken(userID, uuid.NewString(), "", "user@example.com", "superadmin")
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	return token
}

func TestAuthRequiredAcceptsAccessTokenCookie(t *testing.T) {
	manager := newTestJWTManager(t)
	token := newTestToken(t, manager, "11111111-1111-1111-1111-111111111111")

	req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: authcookie.DefaultName, Value: token})
	resp, err := newCookieTestApp(manager, true).Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected %d, got %d", fiber.StatusOK, resp.StatusCode)
	}
}

func TestAuthRequiredPrefersExplicitBearerOverCookie(t *testing.T) {
	manager := newTestJWTManager(t)
	bearerUser := "11111111-1111-1111-1111-111111111111"
	cookieUser := "33333333-3333-3333-3333-333333333333"

	req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+newTestToken(t, manager, bearerUser))
	req.AddCookie(&http.Cookie{Name: authcookie.DefaultName, Value: newTestToken(t, manager, cookieUser)})
	resp, err := newCookieTestApp(manager, true).Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != fiber.StatusOK || string(body) != bearerUser {
		t.Fatalf("expected 200 for bearer user %s, got %d %q", bearerUser, resp.StatusCode, body)
	}
}

func TestAuthRequiredDoesNotFallBackToCookieOnMalformedHeader(t *testing.T) {
	manager := newTestJWTManager(t)

	req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic abc")
	req.AddCookie(&http.Cookie{Name: authcookie.DefaultName, Value: newTestToken(t, manager, uuid.NewString())})
	resp, err := newCookieTestApp(manager, true).Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestAuthRequiredRejectsInvalidCookieAndExpiresIt(t *testing.T) {
	manager := newTestJWTManager(t)
	cases := map[string]struct {
		token  string
		active bool
	}{
		"malformed token": {token: "not-a-jwt", active: true},
		"revoked session": {token: newTestToken(t, manager, uuid.NewString()), active: false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
			req.AddCookie(&http.Cookie{Name: authcookie.DefaultName, Value: tc.token})
			resp, err := newCookieTestApp(manager, tc.active).Test(req)
			if err != nil {
				t.Fatalf("app.Test returned error: %v", err)
			}
			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Fatalf("expected %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
			}
			setCookie := resp.Header.Get(fiber.HeaderSetCookie)
			if !strings.Contains(setCookie, authcookie.DefaultName+"=;") || !strings.Contains(setCookie, "expires=Thu, 01 Jan 1970") {
				t.Fatalf("expected the auth cookie to be expired, got %q", setCookie)
			}
		})
	}
}

func TestAuthRequiredRejectsExpiredCookieToken(t *testing.T) {
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "super-secret-key", AccessTokenExpiry: -1, Issuer: "sigif-test"}}
	expiredManager := jwt.NewManager(cfg)
	token := newTestToken(t, expiredManager, uuid.NewString())

	req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: authcookie.DefaultName, Value: token})
	resp, err := newCookieTestApp(newTestJWTManager(t), true).Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}
