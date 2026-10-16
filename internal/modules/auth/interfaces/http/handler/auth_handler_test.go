package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appHandler "github.com/sigif/sigif-go/internal/modules/auth/application/handler"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
	authService "github.com/sigif/sigif-go/internal/modules/auth/domain/service"
	httpHandler "github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/router"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/authcookie"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

const (
	testEmail    = "admin.empresa@sigif.com"
	testPassword = "admin123"
	testOrigin   = "http://localhost:4200"
)

type memoryUsers struct{ user *userEntity.AppUser }

func (m memoryUsers) GetByEmail(_ context.Context, email string) (*userEntity.AppUser, error) {
	if m.user.Email == email {
		return m.user, nil
	}
	return nil, nil
}

func (m memoryUsers) GetByID(_ context.Context, id uuid.UUID) (*userEntity.AppUser, error) {
	if m.user.ID == id {
		return m.user, nil
	}
	return nil, nil
}

func (memoryUsers) RecordAccess(context.Context, uuid.UUID) error { return nil }

func (memoryUsers) PermissionsByRole(context.Context, uuid.UUID) ([]string, error) {
	return []string{"users.list"}, nil
}

type memorySessions struct {
	mu       sync.Mutex
	clock    clock.Clock
	sessions map[string]*entity.UserSession
}

func (m *memorySessions) Create(_ context.Context, s *entity.UserSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.TokenHash] = s
	return nil
}

func (m *memorySessions) GetByTokenHash(_ context.Context, hash string) (*entity.UserSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[hash], nil
}

func (m *memorySessions) IsActive(_ context.Context, hash string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[hash]
	return ok && s.IsActive(m.clock), nil
}

func (m *memorySessions) End(_ context.Context, id uuid.UUID, endedAt time.Time, reason entity.SessionCloseReason) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		if s.ID == id {
			s.EndedAt = &endedAt
			s.CloseReason = &reason
		}
	}
	return nil
}

func (m *memorySessions) EndAllByUserID(context.Context, uuid.UUID, time.Time, entity.SessionCloseReason) error {
	return nil
}

type noopAttempts struct{}

func (noopAttempts) Create(context.Context, *entity.LoginAttempt) error { return nil }

// newAuthApp wires the auth routes behind the same middleware chain the
// server uses (OriginGuard + AuthRequired), backed by in-memory repositories.
func newAuthApp(t *testing.T) *fiber.App {
	t.Helper()

	clk := clock.NewRealClock()
	user, err := userEntity.NewUser(clk, userEntity.RegisterUserParams{
		RoleID:    uuid.New(),
		FirstName: "Admin",
		LastName:  "Empresa",
		Email:     testEmail,
		Password:  testPassword,
	})
	if err != nil {
		t.Fatalf("NewUser returned error: %v", err)
	}

	jwtManager := jwt.NewManager(&config.Config{JWT: config.JWTConfig{
		Secret:            "super-secret-key",
		AccessTokenExpiry: 15,
		Issuer:            "sigif-test",
	}})
	sessions := &memorySessions{clock: clk, sessions: map[string]*entity.UserSession{}}
	service := authService.NewAuthService(memoryUsers{user: user}, sessions, noopAttempts{}, jwtManager, clk)
	cookie := authcookie.Default()

	app := fiber.New()
	app.Use(middleware.OriginGuard([]string{testOrigin}))
	app.Use(middleware.AuthRequired(jwtManager, sessions, cookie))
	router.RegisterAuthRoutes(app.Group("/api/v1"), httpHandler.NewAuthHTTPHandler(
		appHandler.NewAuthCommandHandler(service), validator.New(), cookie,
	))
	return app
}

func login(t *testing.T, app *fiber.App) *http.Response {
	t.Helper()
	body := `{"email":"` + testEmail + `","password":"` + testPassword + `"}`
	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", testOrigin)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	return resp
}

func authCookie(t *testing.T, resp *http.Response) *http.Cookie {
	t.Helper()
	for _, c := range resp.Cookies() {
		if c.Name == authcookie.DefaultName {
			return c
		}
	}
	t.Fatalf("response has no %s cookie (Set-Cookie: %q)", authcookie.DefaultName, resp.Header.Values("Set-Cookie"))
	return nil
}

func send(t *testing.T, app *fiber.App, method, path string, cookie *http.Cookie) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Origin", testOrigin)
	if cookie != nil {
		req.AddCookie(&http.Cookie{Name: cookie.Name, Value: cookie.Value})
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s failed: %v", method, path, err)
	}
	return resp
}

func TestLoginSetsHttpOnlyCookieAndHidesTokenFromBody(t *testing.T) {
	resp := login(t, newAuthApp(t))
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	c := authCookie(t, resp)
	if c.Value == "" || !c.HttpOnly || c.Path != authcookie.DefaultPath || c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected cookie attributes: %+v", c)
	}
	if c.MaxAge != 900 {
		t.Fatalf("expected Max-Age aligned with expires_in (900), got %d", c.MaxAge)
	}

	raw, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(raw), c.Value) || strings.Contains(string(raw), "access_token") {
		t.Fatalf("login body must not expose the access token: %s", raw)
	}

	var body struct {
		Data struct {
			User struct {
				Email string `json:"email"`
			} `json:"user"`
			ExpiresIn int `json:"expires_in"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if body.Data.User.Email != testEmail || body.Data.ExpiresIn != 900 {
		t.Fatalf("unexpected login body: %s", raw)
	}
}

func TestLoginWithInvalidCredentialsSetsNoCookie(t *testing.T) {
	app := newAuthApp(t)
	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"email":"`+testEmail+`","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if len(resp.Cookies()) != 0 {
		t.Fatalf("failed login must not set cookies, got %v", resp.Header.Values("Set-Cookie"))
	}
}

func TestLoginFromForeignOriginIsRejected(t *testing.T) {
	app := newAuthApp(t)
	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"email":"`+testEmail+`","password":"`+testPassword+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://evil.example")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestMeAuthenticatesWithCookie(t *testing.T) {
	app := newAuthApp(t)
	c := authCookie(t, login(t, app))

	resp := send(t, app, fiber.MethodGet, "/api/v1/auth/me", c)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(raw), testEmail) {
		t.Fatalf("expected the authenticated user in the body, got %s", raw)
	}
}

func TestMeWithoutCredentialsIsUnauthorized(t *testing.T) {
	resp := send(t, newAuthApp(t), fiber.MethodGet, "/api/v1/auth/me", nil)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogoutRevokesSessionAndExpiresCookie(t *testing.T) {
	app := newAuthApp(t)
	c := authCookie(t, login(t, app))

	resp := send(t, app, fiber.MethodPost, "/api/v1/auth/logout", c)
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	cleared := authCookie(t, resp)
	if cleared.Value != "" || cleared.Path != authcookie.DefaultPath || !cleared.HttpOnly ||
		!cleared.Expires.Before(time.Now()) {
		t.Fatalf("expected an expired auth cookie, got %+v", cleared)
	}

	// The old cookie value no longer authenticates: the session was revoked.
	resp = send(t, app, fiber.MethodGet, "/api/v1/auth/me", c)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", resp.StatusCode)
	}
}

func TestLogoutFromForeignOriginIsRejected(t *testing.T) {
	app := newAuthApp(t)
	c := authCookie(t, login(t, app))

	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}

	// The session is still valid for the legitimate origin.
	if resp := send(t, app, fiber.MethodGet, "/api/v1/auth/me", c); resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected the session to survive a rejected cross-origin logout, got %d", resp.StatusCode)
	}
}
