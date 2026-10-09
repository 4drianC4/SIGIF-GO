package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"sync"
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
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type pageMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Meta    *pageMeta       `json:"meta"`
	Error   *struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details map[string]string `json:"details"`
	} `json:"error"`
}

type permissionItem struct {
	ID          string `json:"id"`
	Module      string `json:"module"`
	Operation   string `json:"operation"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

type permissionModuleItem struct {
	Module     string   `json:"module"`
	Operations []string `json:"operations"`
}

type testServer struct {
	app        *fiber.App
	jwtManager *jwt.JWTManager
	permRepo   *testutil.MemoryPermissionRepository
	perms      *stubPermissions
}

type activeSessions struct{}

func (activeSessions) IsActive(context.Context, string) (bool, error) { return true, nil }

type stubPermissions struct {
	mu     sync.Mutex
	byUser map[uuid.UUID]map[string]bool
}

func (p *stubPermissions) HasPermission(_ context.Context, userID uuid.UUID, module, operation string) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.byUser[userID][module+"."+operation], nil
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:            "test-secret",
			AccessTokenExpiry: 60,
			Issuer:            "sigif-test",
		},
		Pagination: config.PaginationConfig{DefaultLimit: 20, MaxLimit: 100},
	}
	jwtManager := jwt.NewManager(cfg)

	users := testutil.NewMemoryUserRepository()
	roles := testutil.NewMemoryRoleRepository()
	permRepo := testutil.NewMemoryPermissionRepository()
	clk := clock.NewMockClock(time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC))
	svc := service.NewUserService(users, roles, permRepo, clk, testutil.NewMemoryCompanyRepository())

	h := httpHandler.NewPermissionHTTPHandler(
		appHandler.NewUserCommandHandler(svc),
		appHandler.NewUserQueryHandler(svc),
		validator.New(),
		cfg,
	)
	stub := &stubPermissions{byUser: map[uuid.UUID]map[string]bool{}}

	app := fiber.New()
	app.Use(middleware.AuthRequired(jwtManager, activeSessions{}))
	router.RegisterPermissionRoutes(app.Group("/api/v1"), h, stub)

	return &testServer{app: app, jwtManager: jwtManager, permRepo: permRepo, perms: stub}
}

func (s *testServer) token(t *testing.T, perms ...string) string {
	t.Helper()
	userID := uuid.New()
	s.perms.mu.Lock()
	s.perms.byUser[userID] = map[string]bool{}
	for _, perm := range perms {
		s.perms.byUser[userID][perm] = true
	}
	s.perms.mu.Unlock()

	token, _, err := s.jwtManager.GenerateAccessToken(userID.String(), uuid.NewString(), uuid.NewString(), "admin@sigif.com", "superadmin")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	return token
}

// seedPermission stores a permission directly, as the GORM seed does, and
// returns its code.
func (s *testServer) seedPermission(t *testing.T, module, operation, description string) string {
	t.Helper()
	p := &entity.Permission{ID: uuid.New(), Module: module, Operation: operation, Description: description}
	if err := s.permRepo.Seed(p); err != nil {
		t.Fatalf("seed permission %s.%s: %v", module, operation, err)
	}
	return p.Key()
}

func (s *testServer) do(t *testing.T, method, path, body, token string) (int, apiResponse) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)
	var parsed apiResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		t.Fatalf("response is not JSON (%d): %s", resp.StatusCode, payload)
	}
	return resp.StatusCode, parsed
}

func (s *testServer) create(t *testing.T, body, token string) (int, apiResponse) {
	t.Helper()
	return s.do(t, fiber.MethodPost, "/api/v1/permissions/", body, token)
}

func (s *testServer) list(t *testing.T, query, token string) (int, apiResponse, []permissionItem) {
	t.Helper()
	path := "/api/v1/permissions/"
	if query != "" {
		path += "?" + query
	}
	status, parsed := s.do(t, fiber.MethodGet, path, "", token)
	var items []permissionItem
	if parsed.Success {
		if err := json.Unmarshal(parsed.Data, &items); err != nil {
			t.Fatalf("data is not a permission list: %s", parsed.Data)
		}
	}
	return status, parsed, items
}

func (s *testServer) modules(t *testing.T, token string) (int, apiResponse, []permissionModuleItem) {
	t.Helper()
	status, parsed := s.do(t, fiber.MethodGet, "/api/v1/permissions/modules", "", token)
	var items []permissionModuleItem
	if parsed.Success {
		if err := json.Unmarshal(parsed.Data, &items); err != nil {
			t.Fatalf("data is not a module list: %s", parsed.Data)
		}
	}
	return status, parsed, items
}

func TestCreatePermissionEndpoint(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, "permissions.create")

	status, resp := s.create(t, `{"module":"products","operation":"export","description":"Exportar productos"}`, token)

	if status != fiber.StatusCreated || !resp.Success {
		t.Fatalf("status = %d, want 201 (error %+v)", status, resp.Error)
	}
	var created permissionItem
	if err := json.Unmarshal(resp.Data, &created); err != nil {
		t.Fatalf("data is not a permission: %s", resp.Data)
	}
	if created.ID == "" || created.Module != "products" || created.Operation != "export" ||
		created.Code != "products.export" || created.Description != "Exportar productos" {
		t.Errorf("unexpected permission %+v", created)
	}

	t.Run("appears in the module listing", func(t *testing.T) {
		_, _, items := s.list(t, "module=products", s.token(t, "permissions.read"))
		if len(items) != 1 || items[0].Code != created.Code {
			t.Fatalf("permission not listed by module: %v", items)
		}
	})

	t.Run("has an id ready for role_permission", func(t *testing.T) {
		if _, err := uuid.Parse(created.ID); err != nil {
			t.Fatalf("id %q is not a uuid: %v", created.ID, err)
		}
	})
}

func TestCreatePermissionEndpointDuplicateCode(t *testing.T) {
	s := newTestServer(t)
	s.seedPermission(t, "users", "create", "Registrar usuarios")

	t.Run("seeded code", func(t *testing.T) {
		status, resp := s.create(t, `{"module":"users","operation":"create"}`, s.token(t, "permissions.create"))
		if status != fiber.StatusConflict || resp.Error == nil || resp.Error.Code != "CONFLICT" {
			t.Fatalf("got %d %+v, want 409 CONFLICT", status, resp.Error)
		}
		if resp.Error.Details["code"] != "already registered" {
			t.Errorf("details = %v, want code detail", resp.Error.Details)
		}
	})

	t.Run("same pair with different case and spaces on code", func(t *testing.T) {
		status, resp := s.create(t,
			`{"module":"users","operation":"create","code":"  USERS.CREATE  "}`,
			s.token(t, "permissions.create"))
		if status != fiber.StatusConflict || resp.Error == nil || resp.Error.Code != "CONFLICT" {
			t.Fatalf("got %d %+v, want 409 CONFLICT", status, resp.Error)
		}
	})

	t.Run("same pair with different case on module and operation", func(t *testing.T) {
		status, resp := s.create(t,
			`{"module":" USERS ","operation":" CREATE "}`,
			s.token(t, "permissions.create"))
		if status != fiber.StatusConflict || resp.Error == nil || resp.Error.Code != "CONFLICT" {
			t.Fatalf("got %d %+v, want 409 CONFLICT", status, resp.Error)
		}
	})

	t.Run("second registration of a fresh pair", func(t *testing.T) {
		if status, _ := s.create(t, `{"module":"categories","operation":"read"}`, s.token(t, "permissions.create")); status != fiber.StatusCreated {
			t.Fatalf("first registration = %d, want 201", status)
		}
		status, resp := s.create(t, `{"module":"categories","operation":"read"}`, s.token(t, "permissions.create"))
		if status != fiber.StatusConflict || resp.Error == nil || resp.Error.Code != "CONFLICT" {
			t.Fatalf("second registration = %d %+v, want 409", status, resp.Error)
		}
	})
}

func TestCreatePermissionEndpointInvalidOperation(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, "permissions.create")

	status, resp := s.create(t, `{"module":"users","operation":"destroy"}`, token)

	if status != fiber.StatusBadRequest || resp.Error == nil {
		t.Fatalf("got %d %+v, want 400", status, resp.Error)
	}
	if resp.Error.Code != "VALIDATION_ERROR" || resp.Error.Message != "unknown operation" {
		t.Errorf("error = %+v, want VALIDATION_ERROR unknown operation", resp.Error)
	}
	if resp.Error.Details["operation"] != "not a supported operation" {
		t.Errorf("details = %v, want operation detail", resp.Error.Details)
	}
}

func TestCreatePermissionEndpointUnknownModule(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, "permissions.create")

	status, resp := s.create(t, `{"module":"invoices","operation":"create"}`, token)

	if status != fiber.StatusBadRequest || resp.Error == nil {
		t.Fatalf("got %d %+v, want 400", status, resp.Error)
	}
	if resp.Error.Code != "VALIDATION_ERROR" || resp.Error.Message != "unknown module" {
		t.Errorf("error = %+v, want VALIDATION_ERROR unknown module", resp.Error)
	}
	if resp.Error.Details["module"] != "not a supported module" {
		t.Errorf("details = %v, want module detail", resp.Error.Details)
	}
}

func TestCreatePermissionEndpointMissingFields(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, "permissions.create")

	tests := []struct {
		name    string
		body    string
		details []string
	}{
		{name: "empty body", body: `{}`, details: []string{"module", "operation"}},
		{name: "only module", body: `{"module":"users"}`, details: []string{"operation"}},
		{name: "only operation", body: `{"operation":"create"}`, details: []string{"module"}},
		{name: "blank module and operation", body: `{"module":"","operation":""}`, details: []string{"module", "operation"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := s.create(t, tt.body, token)
			if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
				t.Fatalf("got %d %+v, want 400 VALIDATION_ERROR", status, resp.Error)
			}
			for _, field := range tt.details {
				if _, ok := resp.Error.Details[field]; !ok {
					t.Errorf("missing detail for %q in %v", field, resp.Error.Details)
				}
			}
		})
	}
}

func TestCreatePermissionEndpointCodeMismatch(t *testing.T) {
	s := newTestServer(t)

	status, resp := s.create(t, `{"module":"products","operation":"update","code":"users.read"}`, s.token(t, "permissions.create"))

	if status != fiber.StatusBadRequest || resp.Error == nil {
		t.Fatalf("got %d %+v, want 400", status, resp.Error)
	}
	if resp.Error.Details["code"] != "must match module.operation" {
		t.Errorf("details = %v, want code detail", resp.Error.Details)
	}
}

func TestListPermissionModulesEndpoint(t *testing.T) {
	s := newTestServer(t)

	status, resp, items := s.modules(t, s.token(t, "permissions.read"))

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	byModule := map[string]permissionModuleItem{}
	for _, item := range items {
		if item.Module == "" || len(item.Operations) == 0 {
			t.Errorf("module item without operations: %+v", item)
		}
		byModule[item.Module] = item
	}
	for _, module := range []string{"users", "categories", "products", "customers", "permissions"} {
		item, ok := byModule[module]
		if !ok {
			t.Fatalf("catalog missing module %q: %v", module, items)
		}
		for _, op := range []string{"read", "create", "update", "delete", "export"} {
			if !contains(item.Operations, op) {
				t.Errorf("module %q does not allow %q: %v", module, op, item.Operations)
			}
		}
	}
}

func TestListPermissionsEndpointFilterByModule(t *testing.T) {
	s := newTestServer(t)
	s.seedPermission(t, "products", "read", "Ver y listar productos")
	s.seedPermission(t, "products", "update", "Editar productos")
	s.seedPermission(t, "products", "delete", "Eliminar productos")
	s.seedPermission(t, "customers", "create", "Registrar clientes")
	s.seedPermission(t, "permissions", "read", "Consultar el catálogo de permisos")

	t.Run("filter returns only that module, with its operation", func(t *testing.T) {
		status, resp, items := s.list(t, "module=products", s.token(t, "permissions.read"))
		if status != fiber.StatusOK || !resp.Success {
			t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
		}
		if len(items) != 3 {
			t.Fatalf("items = %v, want 3 products permissions", items)
		}
		for _, item := range items {
			if item.Module != "products" || item.Code != item.Module+"."+item.Operation {
				t.Errorf("unexpected item %+v", item)
			}
		}
		if resp.Meta == nil || resp.Meta.Total != 3 {
			t.Fatalf("meta = %+v, want total 3", resp.Meta)
		}
	})

	t.Run("no filter lists every permission", func(t *testing.T) {
		_, resp, items := s.list(t, "", s.token(t, "permissions.read"))
		if resp.Meta.Total != 5 || len(items) != 5 {
			t.Fatalf("total = %d (%d items), want 5", resp.Meta.Total, len(items))
		}
	})

	t.Run("unknown module in filter", func(t *testing.T) {
		status, resp, _ := s.list(t, "module=invoices", s.token(t, "permissions.read"))
		if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
			t.Fatalf("got %d %+v, want 400 VALIDATION_ERROR", status, resp.Error)
		}
	})
}

func TestListPermissionsEndpointPaginationValidation(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, "permissions.read")

	tests := []struct {
		name   string
		query  string
		detail string
	}{
		{name: "limit zero", query: "limit=0", detail: "limit"},
		{name: "limit above max", query: "limit=500", detail: "limit"},
		{name: "page zero", query: "page=0", detail: "page"},
		{name: "negative page", query: "page=-1", detail: "page"},
		{name: "module too long", query: "module=" + strings.Repeat("x", 61), detail: "module"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.list(t, tt.query, token)
			if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
				t.Fatalf("got %d %+v, want 400 VALIDATION_ERROR", status, resp.Error)
			}
			if _, ok := resp.Error.Details[tt.detail]; !ok {
				t.Errorf("missing detail for %q in %v", tt.detail, resp.Error.Details)
			}
		})
	}
}

func TestPermissionsEndpointsAuthorization(t *testing.T) {
	s := newTestServer(t)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		token  string
		status int
		code   string
	}{
		{name: "list without token", method: fiber.MethodGet, path: "/api/v1/permissions/", status: 401, code: "UNAUTHORIZED"},
		{name: "modules without token", method: fiber.MethodGet, path: "/api/v1/permissions/modules", status: 401, code: "UNAUTHORIZED"},
		{name: "create without token", method: fiber.MethodPost, path: "/api/v1/permissions/", body: `{"module":"users","operation":"read"}`, status: 401, code: "UNAUTHORIZED"},
		{name: "invalid token", method: fiber.MethodGet, path: "/api/v1/permissions/", token: "not-a-jwt", status: 401, code: "UNAUTHORIZED"},
		{name: "list without permission", method: fiber.MethodGet, path: "/api/v1/permissions/", token: s.token(t), status: 403, code: "FORBIDDEN"},
		{name: "modules without permission", method: fiber.MethodGet, path: "/api/v1/permissions/modules", token: s.token(t, "users.list"), status: 403, code: "FORBIDDEN"},
		{name: "create with only permissions.read", method: fiber.MethodPost, path: "/api/v1/permissions/", body: `{"module":"users","operation":"read"}`, token: s.token(t, "permissions.read"), status: 403, code: "FORBIDDEN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := s.do(t, tt.method, tt.path, tt.body, tt.token)
			if status != tt.status || resp.Error == nil || resp.Error.Code != tt.code {
				t.Fatalf("got %d %+v, want %d %s", status, resp.Error, tt.status, tt.code)
			}
		})
	}

	t.Run("list with only permissions.read", func(t *testing.T) {
		status, resp, _ := s.list(t, "", s.token(t, "permissions.read"))
		if status != fiber.StatusOK || !resp.Success {
			t.Fatalf("got %d %+v, want 200", status, resp.Error)
		}
	})

	t.Run("create with permissions.create succeeds", func(t *testing.T) {
		status, _ := s.create(t, `{"module":"categories","operation":"export"}`, s.token(t, "permissions.create"))
		if status != fiber.StatusCreated {
			t.Fatalf("got %d, want 201", status)
		}
	})
}

func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
