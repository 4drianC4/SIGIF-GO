package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appHandler "github.com/sigif/sigif-go/internal/modules/product/application/handler"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	httpHandler "github.com/sigif/sigif-go/internal/modules/product/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/product/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/modules/product/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/events"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

const (
	companyA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	companyB = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
)

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details map[string]string `json:"details"`
	} `json:"error"`
}

type testServer struct {
	app        *fiber.App
	jwtManager *jwt.JWTManager
	products   *testutil.MemoryProductRepository
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
	jwtManager := jwt.NewManager(&config.Config{JWT: config.JWTConfig{
		Secret:            "test-secret",
		AccessTokenExpiry: 60,
		Issuer:            "sigif-test",
	}})

	products := testutil.NewMemoryProductRepository()
	categories := testutil.NewMemoryCategoryRepository()
	clk := clock.NewMockClock(time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC))
	svc := service.NewCatalogService(products, categories, clk)
	h := httpHandler.NewCatalogHTTPHandler(appHandler.NewCatalogCommandHandler(svc, events.NewBus()), validator.New())
	perms := &stubPermissions{byUser: map[uuid.UUID]map[string]bool{}}

	app := fiber.New()
	app.Use(middleware.AuthRequired(jwtManager, activeSessions{}))
	router.RegisterCatalogRoutes(app.Group("/api/v1"), h, perms)

	return &testServer{app: app, jwtManager: jwtManager, products: products, perms: perms}
}

func (s *testServer) token(t *testing.T, companyID string, perms ...string) string {
	t.Helper()
	userID := uuid.New()
	s.perms.mu.Lock()
	s.perms.byUser[userID] = map[string]bool{}
	for _, perm := range perms {
		s.perms.byUser[userID][perm] = true
	}
	s.perms.mu.Unlock()

	token, _, err := s.jwtManager.GenerateAccessToken(userID.String(), uuid.NewString(), companyID, "admin@sigif.com", "superadmin")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	return token
}

func (s *testServer) manager(t *testing.T, companyID string) string {
	t.Helper()
	return s.token(t, companyID, "categories.create", "products.create")
}

func (s *testServer) post(t *testing.T, path, token string, body any) (int, apiResponse) {
	t.Helper()
	var raw []byte
	switch b := body.(type) {
	case string:
		raw = []byte(b)
	default:
		raw, _ = json.Marshal(b)
	}

	req := httptest.NewRequest(fiber.MethodPost, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
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

func (s *testServer) createCategory(t *testing.T, token, name string) string {
	t.Helper()
	status, resp := s.post(t, "/api/v1/categories", token, map[string]any{"name": name})
	if status != fiber.StatusCreated {
		t.Fatalf("create category status = %d, body error = %+v", status, resp.Error)
	}
	var category struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(resp.Data, &category)
	return category.ID
}

func productBody(categoryID string) map[string]any {
	return map[string]any{
		"category_id":     categoryID,
		"sku":             "coca-600",
		"barcode":         "7750182000123",
		"name":            "Coca Cola 600ml",
		"description":     "Botella PET",
		"unit_of_measure": "unit",
		"cost_price":      3.2,
		"sale_price":      "5.50",
	}
}

func TestCreateCategoryEndpoint(t *testing.T) {
	s := newTestServer(t)
	token := s.manager(t, companyA)

	status, resp := s.post(t, "/api/v1/categories", token, map[string]any{"name": "  Bebidas ", "description": "Gaseosas y jugos"})

	if status != fiber.StatusCreated {
		t.Fatalf("status = %d, want 201 (error %+v)", status, resp.Error)
	}
	var category map[string]any
	_ = json.Unmarshal(resp.Data, &category)
	if category["name"] != "Bebidas" || category["status"] != "active" || category["company_id"] != companyA {
		t.Errorf("unexpected category: %v", category)
	}
}

func TestCreateCategoryEndpointErrors(t *testing.T) {
	s := newTestServer(t)
	token := s.manager(t, companyA)
	s.createCategory(t, token, "Bebidas")

	tests := []struct {
		name   string
		body   any
		status int
		code   string
	}{
		{name: "duplicate name ignoring case", body: map[string]any{"name": "BEBIDAS"}, status: 409, code: "CONFLICT"},
		{name: "blank name", body: map[string]any{"name": "   "}, status: 400, code: "VALIDATION_ERROR"},
		{name: "name too short", body: map[string]any{"name": "A"}, status: 400, code: "VALIDATION_ERROR"},
		{name: "malformed json", body: `{"name":`, status: 400, code: "BAD_REQUEST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := s.post(t, "/api/v1/categories", token, tt.body)
			if status != tt.status || resp.Error == nil || resp.Error.Code != tt.code {
				t.Fatalf("got %d %+v, want %d %s", status, resp.Error, tt.status, tt.code)
			}
		})
	}
}

func TestCreateProductEndpointValid(t *testing.T) {
	s := newTestServer(t)
	token := s.manager(t, companyA)
	categoryID := s.createCategory(t, token, "Bebidas")

	status, resp := s.post(t, "/api/v1/products", token, productBody(categoryID))

	if status != fiber.StatusCreated {
		t.Fatalf("status = %d, want 201 (error %+v)", status, resp.Error)
	}
	var product map[string]any
	_ = json.Unmarshal(resp.Data, &product)
	want := map[string]any{
		"sku":             "COCA-600",
		"barcode":         "7750182000123",
		"category_id":     categoryID,
		"company_id":      companyA,
		"unit_of_measure": "unit",
		"cost_price":      "3.20",
		"sale_price":      "5.50",
		"status":          "active",
	}
	for key, value := range want {
		if product[key] != value {
			t.Errorf("%s = %v, want %v", key, product[key], value)
		}
	}
	if s.products.Count() != 1 {
		t.Errorf("expected 1 stored product, got %d", s.products.Count())
	}
}

func TestCreateProductEndpointInvalid(t *testing.T) {
	s := newTestServer(t)
	token := s.manager(t, companyA)
	categoryID := s.createCategory(t, token, "Bebidas")

	tests := []struct {
		name        string
		mutate      func(body map[string]any)
		wantDetails []string
	}{
		{
			name:        "empty body",
			mutate:      func(body map[string]any) { clear(body) },
			wantDetails: []string{"category_id", "sku", "name", "unit_of_measure", "sale_price"},
		},
		{name: "blank name", mutate: func(body map[string]any) { body["name"] = "   " }, wantDetails: []string{"name"}},
		{name: "category_id not uuid", mutate: func(body map[string]any) { body["category_id"] = "abc" }, wantDetails: []string{"category_id"}},
		{name: "barcode with letters", mutate: func(body map[string]any) { body["barcode"] = "77501ABC" }, wantDetails: []string{"barcode"}},
		{name: "barcode too short", mutate: func(body map[string]any) { body["barcode"] = "123" }, wantDetails: []string{"barcode"}},
		{name: "unknown unit", mutate: func(body map[string]any) { body["unit_of_measure"] = "barrel" }, wantDetails: []string{"unit_of_measure"}},
		{name: "sale price zero", mutate: func(body map[string]any) { body["sale_price"] = 0 }, wantDetails: []string{"sale_price"}},
		{name: "negative cost", mutate: func(body map[string]any) { body["cost_price"] = -1 }, wantDetails: []string{"cost_price"}},
		{name: "sku with spaces", mutate: func(body map[string]any) { body["sku"] = "COCA 600" }, wantDetails: []string{"sku"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := productBody(categoryID)
			tt.mutate(body)

			status, resp := s.post(t, "/api/v1/products", token, body)

			if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
				t.Fatalf("got %d %+v, want 400 VALIDATION_ERROR", status, resp.Error)
			}
			for _, field := range tt.wantDetails {
				if _, ok := resp.Error.Details[field]; !ok {
					t.Errorf("missing detail for %q in %v", field, resp.Error.Details)
				}
			}
		})
	}

	t.Run("malformed price", func(t *testing.T) {
		body := productBody(categoryID)
		body["sale_price"] = "abc"
		status, resp := s.post(t, "/api/v1/products", token, body)
		if status != fiber.StatusBadRequest || resp.Error.Code != "BAD_REQUEST" {
			t.Fatalf("got %d %+v, want 400 BAD_REQUEST", status, resp.Error)
		}
	})

	if s.products.Count() != 0 {
		t.Errorf("invalid requests must not store products, got %d", s.products.Count())
	}
}

func TestCreateProductEndpointDuplicate(t *testing.T) {
	s := newTestServer(t)
	token := s.manager(t, companyA)
	categoryID := s.createCategory(t, token, "Bebidas")
	if status, resp := s.post(t, "/api/v1/products", token, productBody(categoryID)); status != fiber.StatusCreated {
		t.Fatalf("first product status = %d (%+v)", status, resp.Error)
	}

	t.Run("same sku", func(t *testing.T) {
		body := productBody(categoryID)
		body["sku"] = "COCA-600"
		body["barcode"] = "7750182000999"
		status, resp := s.post(t, "/api/v1/products", token, body)
		if status != fiber.StatusConflict || resp.Error.Code != "CONFLICT" {
			t.Fatalf("got %d %+v, want 409 CONFLICT", status, resp.Error)
		}
	})

	t.Run("same barcode", func(t *testing.T) {
		body := productBody(categoryID)
		body["sku"] = "PEPSI-600"
		status, resp := s.post(t, "/api/v1/products", token, body)
		if status != fiber.StatusConflict || resp.Error.Code != "CONFLICT" {
			t.Fatalf("got %d %+v, want 409 CONFLICT", status, resp.Error)
		}
	})

	t.Run("same sku in another company is allowed", func(t *testing.T) {
		tokenB := s.manager(t, companyB)
		categoryB := s.createCategory(t, tokenB, "Bebidas")
		status, resp := s.post(t, "/api/v1/products", tokenB, productBody(categoryB))
		if status != fiber.StatusCreated {
			t.Fatalf("got %d %+v, want 201", status, resp.Error)
		}
	})
}

func TestCreateProductEndpointCategoryNotFound(t *testing.T) {
	s := newTestServer(t)
	token := s.manager(t, companyA)
	tokenB := s.manager(t, companyB)
	foreignCategory := s.createCategory(t, tokenB, "Bebidas")

	status, resp := s.post(t, "/api/v1/products", token, productBody(foreignCategory))

	if status != fiber.StatusNotFound || resp.Error.Code != "NOT_FOUND" {
		t.Fatalf("got %d %+v, want 404 NOT_FOUND", status, resp.Error)
	}
}

func TestCatalogEndpointsPermissions(t *testing.T) {
	s := newTestServer(t)
	categoryID := s.createCategory(t, s.manager(t, companyA), "Bebidas")

	tests := []struct {
		name   string
		path   string
		token  string
		status int
		code   string
	}{
		{name: "product without token", path: "/api/v1/products", token: "", status: 401, code: "UNAUTHORIZED"},
		{name: "category without token", path: "/api/v1/categories", token: "", status: 401, code: "UNAUTHORIZED"},
		{name: "invalid token", path: "/api/v1/products", token: "not-a-jwt", status: 401, code: "UNAUTHORIZED"},
		{name: "product without any permission", path: "/api/v1/products", token: s.token(t, companyA), status: 403, code: "FORBIDDEN"},
		{name: "category without any permission", path: "/api/v1/categories", token: s.token(t, companyA), status: 403, code: "FORBIDDEN"},
		{name: "product with only categories.create", path: "/api/v1/products", token: s.token(t, companyA, "categories.create"), status: 403, code: "FORBIDDEN"},
		{name: "category with only products.create", path: "/api/v1/categories", token: s.token(t, companyA, "products.create"), status: 403, code: "FORBIDDEN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := s.post(t, tt.path, tt.token, productBody(categoryID))
			if status != tt.status || resp.Error == nil || resp.Error.Code != tt.code {
				t.Fatalf("got %d %+v, want %d %s", status, resp.Error, tt.status, tt.code)
			}
		})
	}

	t.Run("each permission allows only its own endpoint", func(t *testing.T) {
		status, resp := s.post(t, "/api/v1/categories", s.token(t, companyA, "categories.create"), map[string]any{"name": "Lácteos"})
		if status != fiber.StatusCreated {
			t.Fatalf("categories.create: got %d %+v, want 201", status, resp.Error)
		}
		body := productBody(categoryID)
		body["sku"], body["barcode"] = "PERM-ONLY-1", "7750182000555"
		status, resp = s.post(t, "/api/v1/products", s.token(t, companyA, "products.create"), body)
		if status != fiber.StatusCreated {
			t.Fatalf("products.create: got %d %+v, want 201", status, resp.Error)
		}
	})
}

func TestCatalogEndpointsRequireCompanyInToken(t *testing.T) {
	s := newTestServer(t)
	globalAdmin := s.manager(t, "")

	for _, path := range []string{"/api/v1/categories", "/api/v1/products"} {
		status, resp := s.post(t, path, globalAdmin, productBody(uuid.NewString()))
		if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Message != "company is required" {
			t.Fatalf("%s: got %d %+v, want 400 company is required", path, status, resp.Error)
		}
	}
}
