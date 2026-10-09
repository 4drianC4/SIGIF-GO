package handler_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
	"github.com/sigif/sigif-go/internal/modules/user/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func (s *testServer) changeRoleStatus(t *testing.T, id, body, token string) (int, apiResponse, roleItem) {
	t.Helper()
	status, resp := s.do(t, fiber.MethodPatch, "/api/v1/roles/"+id+"/status", body, token)
	var item roleItem
	if resp.Success {
		if err := json.Unmarshal(resp.Data, &item); err != nil {
			t.Fatalf("data is not a role: %s", resp.Data)
		}
	}
	return status, resp, item
}

func TestChangeRoleStatusDeactivatesActiveRole(t *testing.T) {
	s := newTestServer(t)
	role := s.seedRole(t, companyA, "Vendedor A", false, entity.RoleStatusActive)
	s.grantPermissions(t, role, "customers.list", "customers.read")
	token := s.tokenWithCompany(t, companyA, "roles.status")

	status, resp, item := s.changeRoleStatus(t, role.ID.String(), `{"status":"inactive"}`, token)

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if item.Status != "inactive" {
		t.Fatalf("status = %q, want inactive", item.Status)
	}
	if item.Name != "Vendedor A" || item.Type != "custom" {
		t.Fatalf("configuration changed: %+v", item)
	}
	if item.PermissionsCount != 2 {
		t.Fatalf("permissions_count = %d, want 2 (assignments must be preserved)", item.PermissionsCount)
	}
	if stored := s.roles.Roles[role.ID]; stored == nil || stored.Status != entity.RoleStatusInactive {
		t.Fatalf("role not persisted as inactive: %+v", stored)
	}
}

func TestChangeRoleStatusReactivatesInactiveRole(t *testing.T) {
	s := newTestServer(t)
	role := s.seedRole(t, companyA, "Vendedor A", false, entity.RoleStatusInactive)
	s.grantPermissions(t, role, "customers.list")
	token := s.tokenWithCompany(t, companyA, "roles.status")

	status, resp, item := s.changeRoleStatus(t, role.ID.String(), `{"status":"active"}`, token)

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if item.Status != "active" {
		t.Fatalf("status = %q, want active", item.Status)
	}
	if item.PermissionsCount != 1 {
		t.Fatalf("permissions_count = %d, want 1 (assignments must be preserved)", item.PermissionsCount)
	}
	if stored := s.roles.Roles[role.ID]; stored == nil || stored.Status != entity.RoleStatusActive {
		t.Fatalf("role not persisted as active: %+v", stored)
	}
}

func TestChangeRoleStatusIsIdempotent(t *testing.T) {
	s := newTestServer(t)
	role := s.seedRole(t, companyA, "Vendedor A", false, entity.RoleStatusActive)
	token := s.tokenWithCompany(t, companyA, "roles.status")
	writes := s.roles.Writes

	status, resp, item := s.changeRoleStatus(t, role.ID.String(), `{"status":"active"}`, token)

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if item.Status != "active" {
		t.Fatalf("status = %q, want active", item.Status)
	}
	if s.roles.Writes != writes {
		t.Fatalf("writes = %d, want no write for a repeated status", s.roles.Writes-writes)
	}
}

func TestChangeRoleStatusProtectedSuperadmin(t *testing.T) {
	s := newTestServer(t)
	role := s.seedRole(t, "", entity.RoleSuperadmin, true, entity.RoleStatusActive)
	token := s.tokenWithCompany(t, "", "roles.status")

	status, resp, _ := s.changeRoleStatus(t, role.ID.String(), `{"status":"inactive"}`, token)

	if status != fiber.StatusConflict || resp.Error == nil || resp.Error.Code != "CONFLICT" {
		t.Fatalf("got %d %+v, want 409 CONFLICT", status, resp.Error)
	}
	if stored := s.roles.Roles[role.ID]; stored.Status != entity.RoleStatusActive {
		t.Fatalf("protected role was deactivated: %+v", stored)
	}
}

func TestChangeRoleStatusNotFound(t *testing.T) {
	s := newTestServer(t)
	token := s.tokenWithCompany(t, companyA, "roles.status")

	status, resp, _ := s.changeRoleStatus(t, uuid.NewString(), `{"status":"inactive"}`, token)

	if status != fiber.StatusNotFound || resp.Error == nil || resp.Error.Code != "NOT_FOUND" {
		t.Fatalf("got %d %+v, want 404 NOT_FOUND", status, resp.Error)
	}
}

func TestChangeRoleStatusInvalidID(t *testing.T) {
	s := newTestServer(t)
	token := s.tokenWithCompany(t, companyA, "roles.status")

	status, resp, _ := s.changeRoleStatus(t, "not-a-uuid", `{"status":"inactive"}`, token)

	if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != "BAD_REQUEST" {
		t.Fatalf("got %d %+v, want 400 BAD_REQUEST", status, resp.Error)
	}
}

func TestChangeRoleStatusInvalidStatus(t *testing.T) {
	s := newTestServer(t)
	role := s.seedRole(t, companyA, "Vendedor A", false, entity.RoleStatusActive)
	token := s.tokenWithCompany(t, companyA, "roles.status")

	tests := []struct {
		name string
		body string
	}{
		{name: "unknown value", body: `{"status":"blocked"}`},
		{name: "missing value", body: `{}`},
		{name: "empty value", body: `{"status":""}`},
		{name: "malformed json", body: `{"status":`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.changeRoleStatus(t, role.ID.String(), tt.body, token)
			if status != fiber.StatusBadRequest || resp.Error == nil {
				t.Fatalf("got %d %+v, want 400", status, resp.Error)
			}
		})
	}
}

func TestChangeRoleStatusIsolatesCompanies(t *testing.T) {
	s := newTestServer(t)
	roleB := s.seedRole(t, companyB, "Vendedor B", false, entity.RoleStatusActive)
	token := s.tokenWithCompany(t, companyA, "roles.status")

	status, resp, _ := s.changeRoleStatus(t, roleB.ID.String(), `{"status":"inactive"}`, token)

	if status != fiber.StatusNotFound || resp.Error == nil || resp.Error.Code != "NOT_FOUND" {
		t.Fatalf("got %d %+v, want 404 NOT_FOUND", status, resp.Error)
	}
	if stored := s.roles.Roles[roleB.ID]; stored.Status != entity.RoleStatusActive {
		t.Fatalf("another company's role was modified: %+v", stored)
	}
}

func TestChangeRoleStatusAuthorization(t *testing.T) {
	s := newTestServer(t)
	role := s.seedRole(t, companyA, "Vendedor A", false, entity.RoleStatusActive)

	tests := []struct {
		name   string
		token  string
		status int
		code   string
	}{
		{name: "without token", token: "", status: 401, code: "UNAUTHORIZED"},
		{name: "invalid token", token: "not-a-jwt", status: 401, code: "UNAUTHORIZED"},
		{name: "without any permission", token: s.tokenWithCompany(t, companyA), status: 403, code: "FORBIDDEN"},
		{name: "with only roles.list", token: s.tokenWithCompany(t, companyA, "roles.list"), status: 403, code: "FORBIDDEN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.changeRoleStatus(t, role.ID.String(), `{"status":"inactive"}`, tt.token)
			if status != tt.status || resp.Error == nil || resp.Error.Code != tt.code {
				t.Fatalf("got %d %+v, want %d %s", status, resp.Error, tt.status, tt.code)
			}
		})
	}
}

func TestChangeRoleStatusDoesNotGrantPermissionsToInactiveRole(t *testing.T) {
	s := newTestServer(t)
	role := s.seedRole(t, companyA, "Vendedor A", false, entity.RoleStatusActive)
	token := s.tokenWithCompany(t, companyA, "roles.status")

	if _, resp, _ := s.changeRoleStatus(t, role.ID.String(), `{"status":"inactive"}`, token); !resp.Success {
		t.Fatalf("deactivation failed: %+v", resp.Error)
	}

	// The listing still exposes the inactive role so an administrator can
	// reactivate it.
	status, resp, items := s.listRoles(t, "status=inactive", s.tokenWithCompany(t, companyA, "roles.list"))
	if status != fiber.StatusOK || resp.Meta.Total != 1 {
		t.Fatalf("status = %d total = %v, want the inactive role listed", status, resp.Meta)
	}
	if len(items) != 1 || items[0].Name != "Vendedor A" {
		t.Fatalf("items = %+v, want the inactive role", items)
	}
}

// TestInactiveRoleRevokesPermissionsForExistingSession exercises the RBAC
// middleware with the real UserService as the permission checker: an already
// issued session stops holding a permission as soon as its role is deactivated,
// without invalidating the session (HU-082-04, CA-02/CA-17).
func TestInactiveRoleRevokesPermissionsForExistingSession(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", AccessTokenExpiry: 60, Issuer: "sigif-test"},
	}
	jwtManager := jwt.NewManager(cfg)
	users := testutil.NewMemoryUserRepository()
	roles := testutil.NewMemoryRoleRepository()
	permRepo := testutil.NewMemoryPermissionRepository()
	svc := service.NewUserService(users, roles, permRepo, clock.NewMockClock(time.Now()), testutil.NewMemoryCompanyRepository())

	permission := &entity.Permission{ID: uuid.New(), Module: "roles", Operation: "status"}
	if err := permRepo.Seed(permission); err != nil {
		t.Fatalf("seed permission: %v", err)
	}
	role := &entity.Role{ID: uuid.New(), Name: "manager", Status: entity.RoleStatusActive}
	roles.Seed(role)
	roles.Grant(role.ID, permission.ID)
	permRepo.AssignToRole(role.ID, permission.ID)

	user := &entity.AppUser{ID: uuid.New(), RoleID: role.ID, Status: entity.UserStatusActive}
	if err := users.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	app := fiber.New()
	app.Use(middleware.AuthRequired(jwtManager, activeSessions{}))
	app.Patch("/api/v1/roles/:id/status",
		middleware.RequirePermission(svc, "roles", "status"),
		func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	token, _, err := jwtManager.GenerateAccessToken(user.ID.String(), uuid.NewString(), uuid.NewString(), "manager@sigif.com", "manager")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	call := func() int {
		req := httptest.NewRequest(fiber.MethodPatch, "/api/v1/roles/"+role.ID.String()+"/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}

	if got := call(); got != fiber.StatusOK {
		t.Fatalf("active role: status = %d, want 200", got)
	}

	if _, err := svc.ChangeRoleStatus(context.Background(), nil, role.ID, entity.RoleStatusInactive); err != nil {
		t.Fatalf("ChangeRoleStatus: %v", err)
	}

	if got := call(); got != fiber.StatusForbidden {
		t.Fatalf("inactive role: status = %d, want 403 for the same session", got)
	}
}
