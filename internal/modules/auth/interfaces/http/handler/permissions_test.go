package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appHandler "github.com/sigif/sigif-go/internal/modules/auth/application/handler"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/service"
	httpHandler "github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/router"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type permissionUsers struct {
	user        *userEntity.AppUser
	permissions map[uuid.UUID][]string
	err         error
	lookups     int
}

func (r *permissionUsers) GetByEmail(context.Context, string) (*userEntity.AppUser, error) {
	return r.user, nil
}

func (r *permissionUsers) GetByID(_ context.Context, id uuid.UUID) (*userEntity.AppUser, error) {
	if r.user != nil && r.user.ID == id {
		return r.user, nil
	}
	return nil, nil
}

func (r *permissionUsers) RecordAccess(context.Context, uuid.UUID) error {
	return nil
}

func (r *permissionUsers) PermissionsByRole(_ context.Context, roleID uuid.UUID) ([]string, error) {
	r.lookups++
	if r.err != nil {
		return nil, r.err
	}
	return r.permissions[roleID], nil
}

func newPermissionsApp(users *permissionUsers, authenticatedAs *uuid.UUID) *fiber.App {
	manager := jwt.NewManager(&config.Config{JWT: config.JWTConfig{Secret: "test-secret", AccessTokenExpiry: 15, Issuer: "sigif-test"}})
	svc := service.NewAuthService(users, nil, nil, manager, clock.NewMockClock(time.Now()))
	h := httpHandler.NewAuthHTTPHandler(appHandler.NewAuthCommandHandler(svc), validator.New())

	app := fiber.New()
	if authenticatedAs != nil {
		app.Use(func(c *fiber.Ctx) error {
			c.SetUserContext(middleware.WithUserID(c.UserContext(), *authenticatedAs))
			return c.Next()
		})
	}
	router.RegisterAuthRoutes(app.Group("/api/v1"), h)
	return app
}

func getPermissions(t *testing.T, app *fiber.App, status int) (map[string]any, string) {
	t.Helper()
	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/auth/me/permissions", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode != status {
		t.Fatalf("status %d, want %d: %s", res.StatusCode, status, raw)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("response is not JSON: %s", raw)
	}
	return body, string(raw)
}

func activeUser() *userEntity.AppUser {
	return &userEntity.AppUser{
		ID:           uuid.New(),
		RoleID:       uuid.New(),
		RoleName:     "business_admin",
		FirstName:    "Ana",
		LastName:     "Pérez",
		Email:        "ana@sigif.com",
		PasswordHash: "secret-hash",
		Status:       userEntity.UserStatusActive,
	}
}

func TestMePermissionsReturnsTheCodesOfTheCurrentUser(t *testing.T) {
	user := activeUser()
	users := &permissionUsers{user: user, permissions: map[uuid.UUID][]string{
		user.RoleID: {"permissions.read", "users.list", "users.read"},
		uuid.New():  {"users.delete"},
	}}

	body, raw := getPermissions(t, newPermissionsApp(users, &user.ID), 200)

	if body["success"] != true {
		t.Fatalf("want success true: %s", raw)
	}
	data := body["data"].(map[string]any)
	if data["role"] != "business_admin" || data["role_id"] != user.RoleID.String() {
		t.Fatalf("unexpected role: %v", data)
	}
	codes := data["permissions"].([]any)
	want := []string{"permissions.read", "users.list", "users.read"}
	if len(codes) != len(want) {
		t.Fatalf("permissions = %v, want %v", codes, want)
	}
	for i, code := range want {
		if codes[i] != code {
			t.Fatalf("permissions = %v, want %v", codes, want)
		}
	}
	if len(data) != 3 {
		t.Fatalf("unexpected fields: %v", data)
	}
	for _, secret := range []string{"secret-hash", "password", "ana@sigif.com"} {
		if strings.Contains(raw, secret) {
			t.Fatalf("response leaks %q: %s", secret, raw)
		}
	}
}

func TestMePermissionsIsAnEmptyArrayWhenTheRoleGrantsNothing(t *testing.T) {
	user := activeUser()
	users := &permissionUsers{user: user}

	_, raw := getPermissions(t, newPermissionsApp(users, &user.ID), 200)

	if !strings.Contains(raw, `"permissions":[]`) {
		t.Fatalf("want an empty array, got %s", raw)
	}
}

func TestMePermissionsIsEmptyForAnInactiveUser(t *testing.T) {
	user := activeUser()
	user.Status = userEntity.UserStatusInactive
	users := &permissionUsers{user: user, permissions: map[uuid.UUID][]string{user.RoleID: {"users.list"}}}

	_, raw := getPermissions(t, newPermissionsApp(users, &user.ID), 200)

	if !strings.Contains(raw, `"permissions":[]`) {
		t.Fatalf("an inactive user has no effective permissions, got %s", raw)
	}
	if users.lookups != 0 {
		t.Fatal("permissions of an inactive user must not be resolved")
	}
}

func TestMePermissionsRequiresAuthentication(t *testing.T) {
	user := activeUser()
	users := &permissionUsers{user: user, permissions: map[uuid.UUID][]string{user.RoleID: {"users.list"}}}

	body, _ := getPermissions(t, newPermissionsApp(users, nil), 401)
	if code := body["error"].(map[string]any)["code"]; code != "UNAUTHORIZED" {
		t.Fatalf("want UNAUTHORIZED, got %v", code)
	}

	unknown := uuid.New()
	body, _ = getPermissions(t, newPermissionsApp(users, &unknown), 401)
	if code := body["error"].(map[string]any)["code"]; code != "UNAUTHORIZED" {
		t.Fatalf("want UNAUTHORIZED for a user that no longer exists, got %v", code)
	}
}

func TestMePermissionsFailsWhenPermissionsCannotBeResolved(t *testing.T) {
	user := activeUser()
	users := &permissionUsers{user: user, err: errors.New("database unavailable")}

	body, raw := getPermissions(t, newPermissionsApp(users, &user.ID), 500)
	if code := body["error"].(map[string]any)["code"]; code != "INTERNAL_ERROR" {
		t.Fatalf("want INTERNAL_ERROR, got %v", code)
	}
	if strings.Contains(raw, "database unavailable") {
		t.Fatalf("internal details must not leak: %s", raw)
	}
}

func TestMeReportsNoPermissionsForAnInactiveUser(t *testing.T) {
	user := activeUser()
	user.Status = userEntity.UserStatusInactive
	users := &permissionUsers{user: user, permissions: map[uuid.UUID][]string{user.RoleID: {"users.list"}}}

	res, err := newPermissionsApp(users, &user.ID).Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/auth/me", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 || !strings.Contains(string(raw), `"permissions":[]`) {
		t.Fatalf("me must report no permissions for an inactive user: %d %s", res.StatusCode, raw)
	}
}
