package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	companyApp "github.com/sigif/sigif-go/internal/modules/company/application/handler"
	companyEntity "github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	companyService "github.com/sigif/sigif-go/internal/modules/company/domain/service"
	companyHTTP "github.com/sigif/sigif-go/internal/modules/company/interfaces/http/handler"
	companyRouter "github.com/sigif/sigif-go/internal/modules/company/interfaces/http/router"
	appHandler "github.com/sigif/sigif-go/internal/modules/user/application/handler"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
	httpHandler "github.com/sigif/sigif-go/internal/modules/user/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/security"
	"github.com/sigif/sigif-go/internal/shared/validator"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type users struct {
	repository.UserRepository
	user *entity.AppUser
}

func (r *users) Create(_ context.Context, u *entity.AppUser) error { r.user = u; return nil }
func (r *users) Update(_ context.Context, u *entity.AppUser) error { r.user = u; return nil }
func (r *users) GetByID(_ context.Context, id uuid.UUID) (*entity.AppUser, error) {
	if r.user != nil && r.user.ID == id {
		return r.user, nil
	}
	return nil, nil
}
func (r *users) ExistsByEmail(_ context.Context, email string) (bool, error) {
	return r.user != nil && r.user.Email == email, nil
}

type roles struct{ repository.RoleRepository }

func (roles) GetByName(_ context.Context, name string) (*entity.Role, error) {
	return &entity.Role{ID: uuid.New(), Name: name, Status: entity.RoleStatusActive}, nil
}

type companies struct {
	id    uuid.UUID
	empty bool
}

func (r companies) ExistsByID(_ context.Context, id uuid.UUID) (bool, error) {
	return !r.empty && r.id == id, nil
}
func (r companies) List(context.Context) ([]companyEntity.Company, error) {
	if r.empty {
		return []companyEntity.Company{}, nil
	}
	return []companyEntity.Company{{ID: r.id, LegalName: "Empresa SRL", TradeName: "Empresa"}}, nil
}

type permissions struct{ operation string }

func (p permissions) HasPermission(_ context.Context, _ uuid.UUID, module, operation string) (bool, error) {
	return module == "users" && (p.operation == "all" || p.operation == operation), nil
}
func setup(authenticated bool, operation string, empty bool) (*fiber.App, *users, uuid.UUID) {
	app := fiber.New()
	if authenticated {
		app.Use(func(c *fiber.Ctx) error {
			c.SetUserContext(middleware.WithUserID(c.UserContext(), uuid.New()))
			return c.Next()
		})
	}
	repo := &users{}
	companyID := uuid.New()
	companyRepo := companies{id: companyID, empty: empty}
	svc := service.NewUserService(repo, roles{}, nil, clock.NewMockClock(time.Now()), companyRepo)
	h := httpHandler.NewUserHTTPHandler(appHandler.NewUserCommandHandler(svc), appHandler.NewUserQueryHandler(svc), validator.New(), &config.Config{})
	checker := permissions{operation: operation}
	router.RegisterUserRoutes(app.Group("/api/v1"), h, checker)
	companyRouter.RegisterCompanyRoutes(app.Group("/api/v1"), companyHTTP.NewCompanyHTTPHandler(companyApp.NewCompanyQueryHandler(companyService.NewCompanyService(companyRepo))), checker)
	return app, repo, companyID
}
func request(t *testing.T, app *fiber.App, method, path, body string, status int) map[string]any {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != status {
		t.Fatalf("%s %s: status %d, want %d: %v", method, path, res.StatusCode, status, result)
	}
	if method == "POST" && status == 201 && res.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("creation must disable caching")
	}
	return result
}
func TestUserFormCreateEditGet(t *testing.T) {
	app, repo, companyID := setup(true, "all", false)
	body := fmt.Sprintf(`{"first_name":"Ana","last_name":"Pérez","email":"ana@example.com","company_id":%q,"role":"employee","password":"client-supplied","area":"removed","full_name":"removed"}`, companyID)
	result := request(t, app, "POST", "/api/v1/users", body, 201)["data"].(map[string]any)
	password := result["generated_password"].(string)
	if password == "client-supplied" {
		t.Fatal("must generate password on server")
	}
	if err := security.VerifyPassword(password, repo.user.PasswordHash); err != nil {
		t.Fatal(err)
	}
	if result["company_id"] != companyID.String() || result["role"] != "employee" {
		t.Fatal(result)
	}
	for _, key := range []string{"area", "full_name", "password_hash", "password"} {
		if _, ok := result[key]; ok {
			t.Fatalf("unexpected %s", key)
		}
	}
	path := "/api/v1/users/" + result["id"].(string)
	edited := request(t, app, "PATCH", path, `{"first_name":"Ana María","role":"business_admin","password":"changed123"}`, 200)["data"].(map[string]any)
	if edited["first_name"] != "Ana María" || edited["role"] != "business_admin" || edited["company_id"] != companyID.String() {
		t.Fatal(edited)
	}
	if err := security.VerifyPassword("changed123", repo.user.PasswordHash); err != nil {
		t.Fatal(err)
	}
	for _, data := range []map[string]any{edited, request(t, app, "GET", path, "", 200)["data"].(map[string]any)} {
		for _, key := range []string{"generated_password", "password_hash", "area", "full_name"} {
			if _, ok := data[key]; ok {
				t.Fatalf("unexpected %s", key)
			}
		}
	}
	for _, invalid := range []string{`{"company_id":"bad"}`, `{"company_id":"00000000-0000-0000-0000-000000000000"}`, `{"password":""}`, `{"role":"superadmin"}`, `{"first_name":""}`} {
		request(t, app, "PATCH", path, invalid, 400)
	}
	request(t, app, "PATCH", path, `{}`, 200)
	request(t, app, "PATCH", path, `{"company_id":null}`, 200)
	if *repo.user.CompanyID != companyID {
		t.Fatal("null must preserve company")
	}
}
func TestUserFormCreateValidation(t *testing.T) {
	app, _, id := setup(true, "all", false)
	for _, body := range []string{`{}`, `{"first_name":"Ana","last_name":"Pérez","email":"ana@example.com","role":"employee"}`, fmt.Sprintf(`{"first_name":"Ana","last_name":"Pérez","email":"ana@example.com","company_id":%q,"role":"superadmin"}`, id), fmt.Sprintf(`{"first_name":"Ana","last_name":"Pérez","email":"ana@example.com","company_id":%q,"role":"employee"}`, uuid.New())} {
		request(t, app, "POST", "/api/v1/users", body, 400)
	}
}
func TestCompanySelectorPermissionsAndEmptyList(t *testing.T) {
	for _, tc := range []struct {
		auth      bool
		operation string
		status    int
	}{{false, "all", 401}, {true, "list", 403}, {true, "create", 200}, {true, "update", 200}} {
		app, _, id := setup(tc.auth, tc.operation, false)
		result := request(t, app, "GET", "/api/v1/companies", "", tc.status)
		if tc.status == 200 {
			data := result["data"].([]any)
			if len(data) != 1 || data[0].(map[string]any)["id"] != id.String() {
				t.Fatal(result)
			}
		}
	}
	app, _, _ := setup(true, "all", true)
	result := request(t, app, "GET", "/api/v1/companies", "", 200)
	if data, ok := result["data"].([]any); !ok || len(data) != 0 {
		t.Fatal("empty selector must be []")
	}
}
