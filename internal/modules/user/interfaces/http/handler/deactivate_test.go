package handler_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func createUserToDeactivate(t *testing.T, app *fiber.App, companyID uuid.UUID) string {
	t.Helper()
	body := fmt.Sprintf(`{"first_name":"Juan","last_name":"Pérez","email":"juan@sigif.com","company_id":%q,"role":"business_admin"}`, companyID)
	return request(t, app, "POST", "/api/v1/users", body, 201)["data"].(map[string]any)["id"].(string)
}

func errorCode(result map[string]any) any {
	return result["error"].(map[string]any)["code"]
}

func TestDeactivateReturnsUpdatedUser(t *testing.T) {
	app, repo, companyID := setup(true, "all", false)
	id := createUserToDeactivate(t, app, companyID)

	result := request(t, app, "POST", "/api/v1/users/"+id+"/deactivate", "", 200)
	if result["success"] != true {
		t.Fatalf("want success true: %v", result)
	}
	data := result["data"].(map[string]any)

	expected := map[string]any{
		"id":         id,
		"company_id": companyID.String(),
		"role":       "business_admin",
		"first_name": "Juan",
		"last_name":  "Pérez",
		"username":   "juan@sigif.com",
		"email":      "juan@sigif.com",
		"status":     "inactive",
	}
	for key, want := range expected {
		if data[key] != want {
			t.Fatalf("%s = %v, want %v", key, data[key], want)
		}
	}
	for _, key := range []string{"role_id", "created_at", "updated_at"} {
		if value, ok := data[key].(string); !ok || value == "" {
			t.Fatalf("missing %s: %v", key, data)
		}
	}
	if _, err := time.Parse(time.RFC3339, data["updated_at"].(string)); err != nil {
		t.Fatalf("updated_at is not RFC3339: %v", data["updated_at"])
	}
	for _, key := range []string{"password", "password_hash", "generated_password", "deleted_at"} {
		if _, ok := data[key]; ok {
			t.Fatalf("unexpected %s in response", key)
		}
	}

	if repo.user == nil || repo.user.ID.String() != id {
		t.Fatal("user must not be physically removed")
	}
	if repo.user.DeletedAt == nil || repo.user.IsActive() {
		t.Fatalf("user must be soft deleted: %+v", repo.user)
	}
}

func TestDeactivatedUserCanStillBeConsulted(t *testing.T) {
	app, _, companyID := setup(true, "all", false)
	id := createUserToDeactivate(t, app, companyID)
	request(t, app, "POST", "/api/v1/users/"+id+"/deactivate", "", 200)

	data := request(t, app, "GET", "/api/v1/users/"+id, "", 200)["data"].(map[string]any)
	if data["id"] != id || data["status"] != "inactive" {
		t.Fatalf("unexpected user: %v", data)
	}
}

func TestDeactivateTwiceIsConflict(t *testing.T) {
	app, _, companyID := setup(true, "all", false)
	id := createUserToDeactivate(t, app, companyID)
	request(t, app, "POST", "/api/v1/users/"+id+"/deactivate", "", 200)

	result := request(t, app, "POST", "/api/v1/users/"+id+"/deactivate", "", 409)
	if result["success"] != false || errorCode(result) != "CONFLICT" {
		t.Fatalf("want CONFLICT, got %v", result)
	}
}

func TestDeactivateErrors(t *testing.T) {
	app, _, companyID := setup(true, "all", false)
	id := createUserToDeactivate(t, app, companyID)

	if result := request(t, app, "POST", "/api/v1/users/not-a-uuid/deactivate", "", 400); errorCode(result) != "BAD_REQUEST" {
		t.Fatalf("want BAD_REQUEST, got %v", result)
	}
	if result := request(t, app, "POST", "/api/v1/users/"+uuid.NewString()+"/deactivate", "", 404); errorCode(result) != "NOT_FOUND" {
		t.Fatalf("want NOT_FOUND, got %v", result)
	}

	unauthenticated, _, _ := setup(false, "all", false)
	if result := request(t, unauthenticated, "POST", "/api/v1/users/"+id+"/deactivate", "", 401); errorCode(result) != "UNAUTHORIZED" {
		t.Fatalf("want UNAUTHORIZED, got %v", result)
	}

	for _, operation := range []string{"read", "update", "delete", "activate"} {
		restricted, _, _ := setup(true, operation, false)
		if result := request(t, restricted, "POST", "/api/v1/users/"+id+"/deactivate", "", 403); errorCode(result) != "FORBIDDEN" {
			t.Fatalf("%s must not allow deactivating: %v", operation, result)
		}
	}
}

func TestDeactivatePermissionIsEnough(t *testing.T) {
	app, repo, companyID := setup(true, "all", false)
	id := createUserToDeactivate(t, app, companyID)

	restricted, restrictedRepo, _ := setup(true, "deactivate", false)
	restrictedRepo.user = repo.user

	data := request(t, restricted, "POST", "/api/v1/users/"+id+"/deactivate", "", 200)["data"].(map[string]any)
	if data["status"] != "inactive" {
		t.Fatalf("unexpected user: %v", data)
	}
}

func TestDeactivationIsRecordedAndHistoryIsPreserved(t *testing.T) {
	env := newHistoryEnv(true, "all")
	id, _ := env.createUser(t, "juan@sigif.com")
	env.send(t, "PATCH", "/api/v1/users/"+id, `{"first_name":"Juan Carlos"}`, 200)

	result, _ := env.send(t, "POST", "/api/v1/users/"+id+"/deactivate", "", 200)
	if status := result["data"].(map[string]any)["status"]; status != "inactive" {
		t.Fatalf("status = %v, want inactive", status)
	}

	rows, meta, _ := env.historyOf(t, id, "")
	want := []string{"deactivated", "updated", "created"}
	if got := actionsOf(rows); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("actions = %v, want %v", got, want)
	}
	if meta["total"] != float64(3) {
		t.Fatalf("unexpected meta: %v", meta)
	}
	actor := rows[0]["performed_by"].(map[string]any)
	if actor["id"] != env.admin.ID.String() || actor["email"] != "admin@sigif.com" {
		t.Fatalf("unexpected performed_by: %v", actor)
	}

	env.send(t, "POST", "/api/v1/users/"+id+"/deactivate", "", 409)
	if rows, _, _ := env.historyOf(t, id, ""); len(rows) != 3 {
		t.Fatalf("a rejected deactivation must not be recorded: %v", actionsOf(rows))
	}
}

func TestAdminCannotDeactivateOwnAccount(t *testing.T) {
	env := newHistoryEnv(true, "all")

	result, _ := env.send(t, "POST", "/api/v1/users/"+env.admin.ID.String()+"/deactivate", "", 409)
	if errorCode(result) != "CONFLICT" {
		t.Fatalf("want CONFLICT, got %v", result)
	}
	if !env.admin.IsActive() {
		t.Fatal("admin must stay active")
	}
	if rows, _, _ := env.historyOf(t, env.admin.ID.String(), ""); len(rows) != 0 {
		t.Fatalf("a rejected deactivation must not be recorded: %v", actionsOf(rows))
	}
}
