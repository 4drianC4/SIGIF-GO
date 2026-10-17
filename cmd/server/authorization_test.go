package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"regexp"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/database"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

var publicRoutes = map[string]bool{
	"GET /health":             true,
	"POST /api/v1/auth/login": true,
}

var sessionOnlyRoutes = map[string]bool{
	"POST /api/v1/auth/logout":        true,
	"GET /api/v1/auth/me":             true,
	"GET /api/v1/auth/me/permissions": true,
	"GET /api/v1/units-of-measure":    true,
	"GET /api/v1/taxes":               true,
}

var routeParam = regexp.MustCompile(`:[A-Za-z_]+`)

type denyAllChecker struct {
	calls atomic.Int64
}

func (c *denyAllChecker) HasPermission(context.Context, uuid.UUID, string, string) (bool, error) {
	c.calls.Add(1)
	return false, nil
}

type activeSessions struct{}

func (activeSessions) IsActive(context.Context, string) (bool, error) {
	return true, nil
}

type apiRoute struct {
	method string
	key    string
	url    string
}

func newAuthorizationApp(t *testing.T, checker middleware.PermissionChecker) (*fiber.App, string) {
	t.Helper()
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "authorization-test-secret", AccessTokenExpiry: 15, Issuer: "sigif-test"}}

	var app *fiber.App
	var manager *jwt.JWTManager
	fxApp := fx.New(
		fx.NopLogger,
		fx.Supply(cfg, &database.Database{}),
		fx.Provide(
			jwt.NewManager,
			validator.New,
			newFiberApp,
			fx.Annotate(clock.NewRealClock, fx.As(new(clock.Clock))),
		),
		modules,
		fx.Decorate(func() middleware.PermissionChecker { return checker }),
		fx.Decorate(func() middleware.SessionReader { return activeSessions{} }),
		fx.Populate(&app, &manager),
	)
	if err := fxApp.Err(); err != nil {
		t.Fatalf("wire application: %v", err)
	}

	token, _, err := manager.GenerateAccessToken(uuid.NewString(), uuid.NewString(), uuid.NewString(), "user@sigif.com", "employee")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return app, token
}

func apiRoutes(app *fiber.App) []apiRoute {
	methods := map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true}
	seen := map[string]bool{}
	routes := []apiRoute{}
	for _, route := range app.GetRoutes(true) {
		key := route.Method + " " + route.Path
		if len(route.Path) > 1 && route.Path[len(route.Path)-1] == '/' {
			key = route.Method + " " + route.Path[:len(route.Path)-1]
		}
		if !methods[route.Method] || seen[key] {
			continue
		}
		seen[key] = true
		routes = append(routes, apiRoute{
			method: route.Method,
			key:    key,
			url:    routeParam.ReplaceAllString(route.Path, uuid.NewString()),
		})
	}
	return routes
}

func call(t *testing.T, app *fiber.App, route apiRoute, token string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(route.method, route.url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("%s: %v", route.key, err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("%s: %v", route.key, err)
	}
	var body map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("%s: response is not JSON (%d): %s", route.key, res.StatusCode, raw)
		}
	}
	return res.StatusCode, body
}

func assertError(t *testing.T, route apiRoute, body map[string]any, code string) {
	t.Helper()
	if body["success"] != false {
		t.Fatalf("%s: want success false, got %v", route.key, body)
	}
	detail, ok := body["error"].(map[string]any)
	if !ok || detail["code"] != code {
		t.Fatalf("%s: want error code %s, got %v", route.key, code, body)
	}
	if message, ok := detail["message"].(string); !ok || message == "" {
		t.Fatalf("%s: missing error message: %v", route.key, body)
	}
}

func TestEveryRouteRequiresAuthentication(t *testing.T) {
	app, _ := newAuthorizationApp(t, &denyAllChecker{})

	checked := 0
	for _, route := range apiRoutes(app) {
		if publicRoutes[route.key] {
			continue
		}
		status, body := call(t, app, route, "")
		if status != fiber.StatusUnauthorized {
			t.Fatalf("%s: status %d without a token, want 401", route.key, status)
		}
		assertError(t, route, body, "UNAUTHORIZED")
		checked++
	}
	if checked == 0 {
		t.Fatal("no protected routes were found")
	}
}

func TestEveryBusinessRouteRequiresAPermission(t *testing.T) {
	checker := &denyAllChecker{}
	app, token := newAuthorizationApp(t, checker)

	guarded := 0
	for _, route := range apiRoutes(app) {
		if publicRoutes[route.key] || sessionOnlyRoutes[route.key] {
			continue
		}
		before := checker.calls.Load()
		status, body := call(t, app, route, token)
		if status != fiber.StatusForbidden {
			t.Fatalf("%s: status %d for a user without permissions, want 403: the route is not guarded by a permission", route.key, status)
		}
		assertError(t, route, body, "FORBIDDEN")
		if checker.calls.Load() == before {
			t.Fatalf("%s: answered 403 without consulting the permission checker", route.key)
		}
		guarded++
	}
	if guarded < 20 {
		t.Fatalf("only %d guarded routes were found; the route table looks incomplete", guarded)
	}
}

func TestPublicAndSessionOnlyRoutesAreNotGuardedByPermissions(t *testing.T) {
	checker := &denyAllChecker{}
	app, _ := newAuthorizationApp(t, checker)

	registered := map[string]bool{}
	for _, route := range apiRoutes(app) {
		registered[route.key] = true
	}
	for _, key := range []string{"GET /health", "POST /api/v1/auth/login", "POST /api/v1/auth/logout", "GET /api/v1/auth/me", "GET /api/v1/auth/me/permissions"} {
		if !registered[key] {
			t.Fatalf("%s is not registered", key)
		}
	}

	status, _ := call(t, app, apiRoute{method: "GET", key: "GET /health", url: "/health"}, "")
	if status != fiber.StatusOK {
		t.Fatalf("health must stay public, got %d", status)
	}
	if checker.calls.Load() != 0 {
		t.Fatal("public routes must not consult the permission checker")
	}
}
