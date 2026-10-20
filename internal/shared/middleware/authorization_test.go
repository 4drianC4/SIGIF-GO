package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type recordingChecker struct {
	granted map[string]bool
	err     error
	checked []string
	userIDs []uuid.UUID
}

func (c *recordingChecker) HasPermission(_ context.Context, userID uuid.UUID, module, operation string) (bool, error) {
	code := module + "." + operation
	c.checked = append(c.checked, code)
	c.userIDs = append(c.userIDs, userID)
	if c.err != nil {
		return false, c.err
	}
	return c.granted[code], nil
}

type guardedApp struct {
	app      *fiber.App
	executed int
}

func newGuardedApp(userID *uuid.UUID, guard fiber.Handler) *guardedApp {
	g := &guardedApp{app: fiber.New()}
	if userID != nil {
		g.app.Use(func(c *fiber.Ctx) error {
			c.SetUserContext(WithUserID(c.UserContext(), *userID))
			return c.Next()
		})
	}
	g.app.Post("/operation", guard, func(c *fiber.Ctx) error {
		g.executed++
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true})
	})
	return g
}

func (g *guardedApp) call(t *testing.T) (int, map[string]any) {
	t.Helper()
	res, err := g.app.Test(httptest.NewRequest(fiber.MethodPost, "/operation", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	var body map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("response is not JSON: %s", raw)
		}
	}
	return res.StatusCode, body
}

func TestGuardLetsTheOperationContinueWhenTheRoleHasThePermission(t *testing.T) {
	userID := uuid.New()
	checker := &recordingChecker{granted: map[string]bool{"users.create": true}}
	g := newGuardedApp(&userID, RequirePermission(checker, "users", "create"))

	status, _ := g.call(t)

	if status != fiber.StatusCreated || g.executed != 1 {
		t.Fatalf("status %d, executed %d: the operation must run", status, g.executed)
	}
	if len(checker.checked) != 1 || checker.checked[0] != "users.create" || checker.userIDs[0] != userID {
		t.Fatalf("unexpected permission lookup: %v for %v", checker.checked, checker.userIDs)
	}
}

func TestGuardRejectsWithForbiddenAndDoesNotRunTheOperation(t *testing.T) {
	userID := uuid.New()
	checker := &recordingChecker{granted: map[string]bool{"users.list": true, "customers.create": true}}
	g := newGuardedApp(&userID, RequirePermission(checker, "users", "create"))

	status, body := g.call(t)

	if status != fiber.StatusForbidden {
		t.Fatalf("status %d, want 403", status)
	}
	if g.executed != 0 {
		t.Fatal("the business operation must not run when the permission is missing")
	}
	if body["success"] != false {
		t.Fatalf("want success false: %v", body)
	}
	if _, ok := body["data"]; ok {
		t.Fatalf("a rejected request must not return data: %v", body)
	}
	detail := body["error"].(map[string]any)
	if detail["code"] != "FORBIDDEN" || detail["message"] != PermissionDeniedMessage {
		t.Fatalf("unexpected error: %v", detail)
	}
	if len(detail) != 2 {
		t.Fatalf("403 must only carry code and message: %v", detail)
	}
}

func TestGuardDoesNotReplaceAuthentication(t *testing.T) {
	checker := &recordingChecker{granted: map[string]bool{"users.create": true}}
	g := newGuardedApp(nil, RequirePermission(checker, "users", "create"))

	status, body := g.call(t)

	if status != fiber.StatusUnauthorized {
		t.Fatalf("status %d, want 401 for a request without identity", status)
	}
	if code := body["error"].(map[string]any)["code"]; code != "UNAUTHORIZED" {
		t.Fatalf("want UNAUTHORIZED, got %v", code)
	}
	if g.executed != 0 || len(checker.checked) != 0 {
		t.Fatalf("an unauthenticated request must not reach the checker (%v) or the operation (%d)", checker.checked, g.executed)
	}
}

func TestGuardAnswersUnauthorizedBeforeForbiddenThroughTheWholeChain(t *testing.T) {
	manager := newTestJWTManager(t)
	checker := &recordingChecker{}
	executed := 0
	app := fiber.New()
	app.Use(AuthRequired(manager, stubSessionReader{active: true}))
	app.Delete("/users/:id", RequirePermission(checker, "users", "delete"), func(c *fiber.Ctx) error {
		executed++
		return c.SendStatus(fiber.StatusNoContent)
	})
	token, _, err := manager.GenerateAccessToken(uuid.NewString(), uuid.NewString(), "", "user@sigif.com", "employee")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	cases := []struct {
		name          string
		authorization string
		status        int
	}{
		{"without token", "", fiber.StatusUnauthorized},
		{"with an invalid token", "Bearer not-a-token", fiber.StatusUnauthorized},
		{"with a valid token and no permission", "Bearer " + token, fiber.StatusForbidden},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(fiber.MethodDelete, "/users/"+uuid.NewString(), nil)
		if tc.authorization != "" {
			req.Header.Set("Authorization", tc.authorization)
		}
		res, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		res.Body.Close()
		if res.StatusCode != tc.status {
			t.Fatalf("%s: status %d, want %d", tc.name, res.StatusCode, tc.status)
		}
	}
	if executed != 0 {
		t.Fatal("the operation must not run in any rejected case")
	}
	if len(checker.checked) != 1 {
		t.Fatalf("only the authenticated request may consult permissions, got %v", checker.checked)
	}
}

func TestGuardFailsClosedWhenPermissionsCannotBeResolved(t *testing.T) {
	userID := uuid.New()
	checker := &recordingChecker{granted: map[string]bool{"users.create": true}, err: errors.New("database unavailable")}
	g := newGuardedApp(&userID, RequirePermission(checker, "users", "create"))

	status, body := g.call(t)

	if status != fiber.StatusInternalServerError || g.executed != 0 {
		t.Fatalf("status %d, executed %d: a failed lookup must not let the operation run", status, g.executed)
	}
	if code := body["error"].(map[string]any)["code"]; code != "INTERNAL_ERROR" {
		t.Fatalf("want INTERNAL_ERROR, got %v", code)
	}
}

func TestRequireAnyPermissionAcceptsAnyOfTheCodes(t *testing.T) {
	userID := uuid.New()

	checker := &recordingChecker{granted: map[string]bool{"users.update": true}}
	g := newGuardedApp(&userID, RequireAnyPermission(checker, "users.create", "users.update"))
	if status, _ := g.call(t); status != fiber.StatusCreated || g.executed != 1 {
		t.Fatalf("status %d: the second code must be enough", status)
	}
	if len(checker.checked) != 2 || checker.checked[0] != "users.create" || checker.checked[1] != "users.update" {
		t.Fatalf("unexpected lookups: %v", checker.checked)
	}

	checker = &recordingChecker{granted: map[string]bool{"users.create": true}}
	g = newGuardedApp(&userID, RequireAnyPermission(checker, "users.create", "users.update"))
	if status, _ := g.call(t); status != fiber.StatusCreated {
		t.Fatalf("status %d: the first code must be enough", status)
	}
	if len(checker.checked) != 1 {
		t.Fatalf("lookups must stop at the first granted code: %v", checker.checked)
	}

	checker = &recordingChecker{granted: map[string]bool{"users.read": true}}
	g = newGuardedApp(&userID, RequireAnyPermission(checker, "users.create", "users.update"))
	if status, _ := g.call(t); status != fiber.StatusForbidden || g.executed != 0 {
		t.Fatalf("status %d, executed %d: no matching code must be rejected", status, g.executed)
	}
}

func TestPermissionCodeIsSplitAtTheFirstDot(t *testing.T) {
	userID := uuid.New()
	checker := &recordingChecker{granted: map[string]bool{"users.change_password": true}}
	g := newGuardedApp(&userID, RequireAnyPermission(checker, "users.change_password"))

	if status, _ := g.call(t); status != fiber.StatusCreated {
		t.Fatalf("status %d, want the operation to run", status)
	}
}

func TestInvalidPermissionCodesAreRejectedWhenRegisteringTheRoute(t *testing.T) {
	invalid := [][]string{{}, {"users"}, {".create"}, {"users."}, {"users.create", ""}}
	for _, codes := range invalid {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("codes %v must panic", codes)
				}
			}()
			RequireAnyPermission(&recordingChecker{}, codes...)
		}()
	}
}

func observeDeniedLog(t *testing.T) *observer.ObservedLogs {
	t.Helper()
	core, logs := observer.New(zap.WarnLevel)
	previous := deniedLogger
	deniedLogger = func() *zap.Logger { return zap.New(core) }
	t.Cleanup(func() { deniedLogger = previous })
	return logs
}

func TestRejectedRequestsAreLogged(t *testing.T) {
	logs := observeDeniedLog(t)
	userID, companyID := uuid.New(), uuid.New()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Request-ID", "request-123")
		c.SetUserContext(WithCompanyID(WithUserID(c.UserContext(), userID), companyID))
		return c.Next()
	})
	app.Delete("/users/:id", RequireAnyPermission(&recordingChecker{}, "users.delete", "users.deactivate"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodDelete, "/users/abc", nil))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status %d, want 403", res.StatusCode)
	}

	entries := logs.All()
	if len(entries) != 1 || entries[0].Message != "permission denied" {
		t.Fatalf("want one permission denied entry, got %v", entries)
	}
	fields := entries[0].ContextMap()
	expected := map[string]string{
		"user_id":    userID.String(),
		"company_id": companyID.String(),
		"method":     "DELETE",
		"path":       "/users/abc",
		"request_id": "request-123",
	}
	for key, want := range expected {
		if fields[key] != want {
			t.Fatalf("%s = %v, want %v", key, fields[key], want)
		}
	}
	codes, ok := fields["required_permissions"].([]any)
	if !ok || len(codes) != 2 || codes[0] != "users.delete" || codes[1] != "users.deactivate" {
		t.Fatalf("unexpected required_permissions: %v", fields["required_permissions"])
	}
}

func TestAllowedAndUnauthenticatedRequestsAreNotLoggedAsDenied(t *testing.T) {
	logs := observeDeniedLog(t)
	userID := uuid.New()

	allowed := newGuardedApp(&userID, RequirePermission(&recordingChecker{granted: map[string]bool{"users.create": true}}, "users", "create"))
	allowed.call(t)
	anonymous := newGuardedApp(nil, RequirePermission(&recordingChecker{}, "users", "create"))
	anonymous.call(t)

	if logs.Len() != 0 {
		t.Fatalf("unexpected denied entries: %v", logs.All())
	}
}
