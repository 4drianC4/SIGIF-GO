package authcookie

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/shared/config"
)

func TestNewAppliesDefaults(t *testing.T) {
	s, err := New(&config.Config{})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if s.Name != DefaultName || s.Path != DefaultPath || s.SameSite != fiber.CookieSameSiteLaxMode || s.Secure {
		t.Fatalf("unexpected defaults: %+v", s)
	}
}

func TestNewNormalizesSameSite(t *testing.T) {
	s, err := New(&config.Config{AuthCookie: config.AuthCookieConfig{SameSite: "strict"}})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if s.SameSite != fiber.CookieSameSiteStrictMode {
		t.Fatalf("expected Strict, got %q", s.SameSite)
	}
}

func TestNewRejectsInvalidSameSite(t *testing.T) {
	if _, err := New(&config.Config{AuthCookie: config.AuthCookieConfig{SameSite: "sometimes"}}); err == nil {
		t.Fatal("expected error for invalid same_site")
	}
}

func TestNewRejectsSameSiteNoneWithoutSecure(t *testing.T) {
	if _, err := New(&config.Config{AuthCookie: config.AuthCookieConfig{SameSite: "None"}}); err == nil {
		t.Fatal("expected error for SameSite=None without Secure")
	}
	if _, err := New(&config.Config{AuthCookie: config.AuthCookieConfig{SameSite: "None", Secure: true}}); err != nil {
		t.Fatalf("SameSite=None with Secure should be accepted: %v", err)
	}
}

func setCookieHeader(t *testing.T, handler fiber.Handler) string {
	t.Helper()
	app := fiber.New()
	app.Get("/", handler)
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	return resp.Header.Get(fiber.HeaderSetCookie)
}

func TestSetWritesHttpOnlyCookieAlignedWithExpiry(t *testing.T) {
	s := Settings{Name: "sigif_access_token", Path: "/api/v1", Secure: true, SameSite: fiber.CookieSameSiteLaxMode}
	header := setCookieHeader(t, func(c *fiber.Ctx) error {
		s.Set(c, "jwt-value", 900)
		return nil
	})

	for _, want := range []string{"sigif_access_token=jwt-value", "max-age=900", "path=/api/v1", "HttpOnly", "secure", "SameSite=Lax"} {
		if !strings.Contains(header, want) {
			t.Fatalf("Set-Cookie %q does not contain %q", header, want)
		}
	}
}

func TestClearExpiresCookieWithSameAttributes(t *testing.T) {
	s := Default()
	header := setCookieHeader(t, func(c *fiber.Ctx) error {
		s.Clear(c)
		return nil
	})

	for _, want := range []string{"sigif_access_token=;", "expires=Thu, 01 Jan 1970", "path=/api/v1", "HttpOnly", "SameSite=Lax"} {
		if !strings.Contains(header, want) {
			t.Fatalf("Set-Cookie %q does not contain %q", header, want)
		}
	}
}
