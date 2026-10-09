package handler_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appHandler "github.com/sigif/sigif-go/internal/modules/user/application/handler"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
	httpHandler "github.com/sigif/sigif-go/internal/modules/user/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/modules/user/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

const anySuccess = 0

type historyEnv struct {
	app       *fiber.App
	users     *testutil.MemoryUserRepository
	history   *testutil.MemoryHistoryRepository
	clock     *clock.MockClock
	admin     *entity.AppUser
	companyID uuid.UUID
	otherID   uuid.UUID
}

func newHistoryEnv(authenticated bool, operation string) *historyEnv {
	env := &historyEnv{
		history:   testutil.NewMemoryHistoryRepository(),
		clock:     clock.NewMockClock(time.Date(2026, 10, 8, 18, 0, 0, 0, time.UTC)),
		companyID: uuid.New(),
		otherID:   uuid.New(),
	}
	env.admin = &entity.AppUser{
		ID:        uuid.New(),
		RoleID:    uuid.New(),
		RoleName:  entity.RoleSuperadmin,
		FirstName: "Admin",
		LastName:  "SIGIF",
		Email:     "admin@sigif.com",
		Status:    entity.UserStatusActive,
	}
	env.users = testutil.NewMemoryUserRepository(env.admin)

	env.app = fiber.New()
	if authenticated {
		env.app.Use(func(c *fiber.Ctx) error {
			c.SetUserContext(middleware.WithUserID(c.UserContext(), env.admin.ID))
			env.clock.Add(time.Minute)
			return c.Next()
		})
	}

	svc := service.NewUserService(env.users, testutil.MemoryRoleRepository{}, nil, env.clock, testutil.NewMemoryCompanyRepository(env.companyID, env.otherID))
	history := service.NewHistoryService(env.history, env.users, env.clock)
	h := httpHandler.NewUserHTTPHandler(
		appHandler.NewUserCommandHandler(svc, history, testutil.Transactor{}),
		appHandler.NewUserQueryHandler(svc, history),
		validator.New(),
		&config.Config{},
	)
	router.RegisterUserRoutes(env.app.Group("/api/v1"), h, permissions{operation: operation})
	return env
}

func (e *historyEnv) send(t *testing.T, method, path, body string, status int) (map[string]any, string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := e.app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if status == anySuccess && res.StatusCode >= 200 && res.StatusCode < 300 {
		status = res.StatusCode
	}
	if res.StatusCode != status {
		t.Fatalf("%s %s: status %d, want %d: %s", method, path, res.StatusCode, status, raw)
	}
	var result map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatalf("%s %s: invalid json: %s", method, path, raw)
		}
	}
	return result, string(raw)
}

func (e *historyEnv) createUser(t *testing.T, email string) (string, string) {
	t.Helper()
	body := fmt.Sprintf(`{"first_name":"Juan","last_name":"Pérez","email":%q,"company_id":%q,"role":"employee"}`, email, e.companyID)
	result, _ := e.send(t, "POST", "/api/v1/users", body, 201)
	data := result["data"].(map[string]any)
	return data["id"].(string), data["generated_password"].(string)
}

func (e *historyEnv) historyOf(t *testing.T, id, query string) ([]map[string]any, map[string]any, string) {
	t.Helper()
	result, raw := e.send(t, "GET", "/api/v1/users/"+id+"/history"+query, "", 200)
	if result["success"] != true {
		t.Fatalf("want success true: %s", raw)
	}
	items := result["data"].([]any)
	rows := make([]map[string]any, len(items))
	for i, item := range items {
		rows[i] = item.(map[string]any)
	}
	return rows, result["meta"].(map[string]any), raw
}

func actionsOf(rows []map[string]any) []string {
	actions := make([]string, len(rows))
	for i, row := range rows {
		actions[i] = row["action"].(string)
	}
	return actions
}

func TestHistoryRecordsUserLifecycle(t *testing.T) {
	env := newHistoryEnv(true, "all")
	id, generated := env.createUser(t, "juan@sigif.com")
	path := "/api/v1/users/" + id

	env.send(t, "PATCH", path, fmt.Sprintf(`{"first_name":"Juan Carlos","email":"juan.carlos@sigif.com","role":"business_admin","company_id":%q}`, env.otherID), 200)
	env.send(t, "PATCH", path, `{"password":"changed12345"}`, 200)
	env.send(t, "PATCH", path, `{}`, 200)
	env.send(t, "PATCH", path, `{"last_name":"Pérez"}`, 200)
	env.send(t, "POST", path+"/deactivate", "", anySuccess)
	env.send(t, "POST", path+"/activate", "", anySuccess)
	env.send(t, "PUT", path+"/password", `{"current_password":"changed12345","new_password":"another12345"}`, 204)
	env.send(t, "DELETE", path, "", 204)

	rows, meta, raw := env.historyOf(t, id, "")

	want := []string{"deleted", "password_changed", "activated", "deactivated", "password_changed", "updated", "created"}
	if got := actionsOf(rows); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("actions = %v, want %v", got, want)
	}
	if meta["page"] != float64(1) || meta["limit"] != float64(20) || meta["total"] != float64(7) || meta["total_pages"] != float64(1) {
		t.Fatalf("unexpected meta: %v", meta)
	}

	for _, row := range rows {
		for _, key := range []string{"id", "user_id", "action", "description", "changes", "performed_by", "created_at"} {
			if _, ok := row[key]; !ok {
				t.Fatalf("missing %s in %v", key, row)
			}
		}
		if row["user_id"] != id || row["description"] == "" {
			t.Fatalf("unexpected row: %v", row)
		}
		if _, err := time.Parse(time.RFC3339, row["created_at"].(string)); err != nil {
			t.Fatalf("created_at is not RFC3339: %v", row["created_at"])
		}
		actor := row["performed_by"].(map[string]any)
		if actor["id"] != env.admin.ID.String() || actor["full_name"] != "Admin SIGIF" || actor["email"] != "admin@sigif.com" {
			t.Fatalf("unexpected performed_by: %v", actor)
		}
	}

	updated := rows[5]["changes"].(map[string]any)
	expected := map[string][2]string{
		"first_name": {"Juan", "Juan Carlos"},
		"email":      {"juan@sigif.com", "juan.carlos@sigif.com"},
		"role":       {"employee", "business_admin"},
		"company_id": {env.companyID.String(), env.otherID.String()},
	}
	if len(updated) != len(expected) {
		t.Fatalf("unexpected changes: %v", updated)
	}
	for field, values := range expected {
		change := updated[field].(map[string]any)
		if change["from"] != values[0] || change["to"] != values[1] {
			t.Fatalf("%s = %v, want %v", field, change, values)
		}
	}
	for _, index := range []int{0, 1, 2, 3, 4, 6} {
		if changes := rows[index]["changes"].(map[string]any); len(changes) != 0 {
			t.Fatalf("%s must not carry changes: %v", rows[index]["action"], changes)
		}
	}

	for _, secret := range []string{generated, "changed12345", "another12345", "password_hash", "argon2"} {
		if strings.Contains(raw, secret) {
			t.Fatalf("history leaks %q: %s", secret, raw)
		}
	}
}

func TestHistoryPagination(t *testing.T) {
	env := newHistoryEnv(true, "all")
	id, _ := env.createUser(t, "juan@sigif.com")
	for _, name := range []string{"A", "B", "C", "D"} {
		env.send(t, "PATCH", "/api/v1/users/"+id, fmt.Sprintf(`{"first_name":%q}`, name), 200)
	}

	rows, meta, _ := env.historyOf(t, id, "?page=2&limit=2")
	if meta["page"] != float64(2) || meta["limit"] != float64(2) || meta["total"] != float64(5) || meta["total_pages"] != float64(3) {
		t.Fatalf("unexpected meta: %v", meta)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	first := rows[0]["changes"].(map[string]any)["first_name"].(map[string]any)
	if first["from"] != "A" || first["to"] != "B" {
		t.Fatalf("unexpected page content: %v", first)
	}

	rows, meta, _ = env.historyOf(t, id, "?page=9&limit=2")
	if len(rows) != 0 || meta["total"] != float64(5) {
		t.Fatalf("page beyond the end must be empty: %v %v", rows, meta)
	}
}

func TestHistoryOfUserWithoutEventsIsEmptyArray(t *testing.T) {
	env := newHistoryEnv(true, "all")

	_, meta, raw := env.historyOf(t, env.admin.ID.String(), "")
	if !strings.Contains(raw, `"data":[]`) {
		t.Fatalf("want empty array, got %s", raw)
	}
	if meta["total"] != float64(0) || meta["total_pages"] != float64(0) {
		t.Fatalf("unexpected meta: %v", meta)
	}
}

func TestHistoryIsScopedToTheRequestedUser(t *testing.T) {
	env := newHistoryEnv(true, "all")
	first, _ := env.createUser(t, "first@sigif.com")
	second, _ := env.createUser(t, "second@sigif.com")
	env.send(t, "PATCH", "/api/v1/users/"+second, `{"first_name":"Other"}`, 200)

	rows, _, _ := env.historyOf(t, first, "")
	if got := actionsOf(rows); len(got) != 1 || got[0] != "created" {
		t.Fatalf("unexpected history: %v", got)
	}
}

func TestHistoryErrors(t *testing.T) {
	env := newHistoryEnv(true, "all")
	id, _ := env.createUser(t, "juan@sigif.com")

	result, _ := env.send(t, "GET", "/api/v1/users/not-a-uuid/history", "", 400)
	if code := result["error"].(map[string]any)["code"]; code != "BAD_REQUEST" {
		t.Fatalf("want BAD_REQUEST, got %v", code)
	}
	result, _ = env.send(t, "GET", "/api/v1/users/"+uuid.NewString()+"/history", "", 404)
	if code := result["error"].(map[string]any)["code"]; code != "NOT_FOUND" {
		t.Fatalf("want NOT_FOUND, got %v", code)
	}

	result, _ = newHistoryEnv(false, "all").send(t, "GET", "/api/v1/users/"+id+"/history", "", 401)
	if code := result["error"].(map[string]any)["code"]; code != "UNAUTHORIZED" {
		t.Fatalf("want UNAUTHORIZED, got %v", code)
	}
	result, _ = newHistoryEnv(true, "list").send(t, "GET", "/api/v1/users/"+id+"/history", "", 403)
	if code := result["error"].(map[string]any)["code"]; code != "FORBIDDEN" {
		t.Fatalf("want FORBIDDEN, got %v", code)
	}
	newHistoryEnv(true, "read").send(t, "GET", "/api/v1/users/"+uuid.NewString()+"/history", "", 404)
}

func TestFailedOperationsRecordNothing(t *testing.T) {
	env := newHistoryEnv(true, "all")
	id, _ := env.createUser(t, "juan@sigif.com")
	env.createUser(t, "taken@sigif.com")
	before := len(env.history.All())

	env.send(t, "PATCH", "/api/v1/users/"+id, `{"email":"taken@sigif.com"}`, 409)
	env.send(t, "PATCH", "/api/v1/users/"+id, `{"role":"superadmin"}`, 400)
	env.send(t, "PATCH", "/api/v1/users/"+uuid.NewString(), `{"first_name":"Ghost"}`, 404)
	env.send(t, "POST", "/api/v1/users/"+uuid.NewString()+"/deactivate", "", 404)
	env.send(t, "PUT", "/api/v1/users/"+id+"/password", `{"current_password":"wrong-password","new_password":"another12345"}`, 401)
	env.send(t, "DELETE", "/api/v1/users/"+uuid.NewString(), "", 404)
	env.send(t, "POST", "/api/v1/users", fmt.Sprintf(`{"first_name":"Juan","last_name":"Pérez","email":"taken@sigif.com","company_id":%q,"role":"employee"}`, env.companyID), 409)

	if after := len(env.history.All()); after != before {
		t.Fatalf("failed operations recorded %d events", after-before)
	}
}

func TestOperationFailsWhenHistoryCannotBeRecorded(t *testing.T) {
	env := newHistoryEnv(true, "all")
	id, _ := env.createUser(t, "juan@sigif.com")
	env.history.Err = errors.New("history unavailable")

	result, _ := env.send(t, "PATCH", "/api/v1/users/"+id, `{"first_name":"Juan Carlos"}`, 500)
	if code := result["error"].(map[string]any)["code"]; code != "INTERNAL_ERROR" {
		t.Fatalf("want INTERNAL_ERROR, got %v", code)
	}
	env.send(t, "POST", "/api/v1/users/"+id+"/deactivate", "", 500)
}
