// Package authcookie issues, reads and expires the HttpOnly cookie that
// carries the access token for browser clients.
package authcookie

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/shared/config"
)

const (
	DefaultName = "sigif_access_token"
	DefaultPath = "/api/v1"
)

// Settings are the attributes of the access token cookie. The same values are
// used to set and to expire it, so the browser matches and removes the cookie.
type Settings struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	SameSite string
}

// Default returns the development settings: Lax, not Secure.
func Default() Settings {
	return Settings{
		Name:     DefaultName,
		Path:     DefaultPath,
		SameSite: fiber.CookieSameSiteLaxMode,
	}
}

// New builds the settings from the configuration, applying defaults and
// rejecting combinations browsers would ignore (SameSite=None without Secure).
func New(cfg *config.Config) (Settings, error) {
	s := Default()
	c := cfg.AuthCookie
	if name := strings.TrimSpace(c.Name); name != "" {
		s.Name = name
	}
	if path := strings.TrimSpace(c.Path); path != "" {
		s.Path = path
	}
	s.Domain = strings.TrimSpace(c.Domain)
	s.Secure = c.Secure

	if sameSite := strings.TrimSpace(c.SameSite); sameSite != "" {
		switch strings.ToLower(sameSite) {
		case "lax":
			s.SameSite = fiber.CookieSameSiteLaxMode
		case "strict":
			s.SameSite = fiber.CookieSameSiteStrictMode
		case "none":
			s.SameSite = fiber.CookieSameSiteNoneMode
		default:
			return Settings{}, fmt.Errorf("auth_cookie.same_site: invalid value %q (use Lax, Strict or None)", sameSite)
		}
	}
	if s.SameSite == fiber.CookieSameSiteNoneMode && !s.Secure {
		return Settings{}, fmt.Errorf("auth_cookie.same_site=None requires auth_cookie.secure=true")
	}

	return s, nil
}

// Set writes the access token cookie. Its lifetime matches the token's
// expires_in so the cookie never outlives the JWT.
func (s Settings) Set(c *fiber.Ctx, token string, expiresIn int) {
	c.Cookie(&fiber.Cookie{
		Name:     s.Name,
		Value:    token,
		Path:     s.Path,
		Domain:   s.Domain,
		MaxAge:   expiresIn,
		Expires:  time.Now().Add(time.Duration(expiresIn) * time.Second),
		Secure:   s.Secure,
		HTTPOnly: true,
		SameSite: s.SameSite,
	})
}

// Clear expires the access token cookie using the same name, path and
// domain it was created with.
func (s Settings) Clear(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     s.Name,
		Value:    "",
		Path:     s.Path,
		Domain:   s.Domain,
		Expires:  time.Unix(0, 0).UTC(),
		Secure:   s.Secure,
		HTTPOnly: true,
		SameSite: s.SameSite,
	})
}

// Read returns the access token carried by the cookie, or "" if absent.
func (s Settings) Read(c *fiber.Ctx) string {
	return strings.TrimSpace(c.Cookies(s.Name))
}
