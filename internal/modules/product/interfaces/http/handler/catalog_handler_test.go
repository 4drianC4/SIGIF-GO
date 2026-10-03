package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

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
)

const (
	tenantA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	tenantB = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	userID  = "11111111-1111-4111-8111-111111111111"
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
}

// newTestServer arma la app igual que cmd/server: AuthRequired global y rutas bajo /api/v1.
func newTestServer(t *testing.T) *testServer {
	t.Helper()
	jwtManager := jwt.NewManager(&config.Config{JWT: config.JWTConfig{
		Secret:             "test-secret",
		AccessTokenExpiry:  60,
		RefreshTokenExpiry: 1440,
		Issuer:             "sigif-test",
	}})

	products := testutil.NewMemoryProductRepository()
	categories := testutil.NewMemoryCategoryRepository()
	clk := clock.NewMockClock(time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC))
	svc := service.NewCatalogService(products, categories, clk)
	h := httpHandler.NewCatalogHTTPHandler(appHandler.NewCatalogCommandHandler(svc, events.NewBus()))

	app := fiber.New()
	app.Use(middleware.TenantContext())
	app.Use(middleware.AuthRequired(jwtManager))
	router.RegisterCatalogRoutes(app.Group("/api/v1"), h)

	return &testServer{app: app, jwtManager: jwtManager, products: products}
}

func (s *testServer) token(t *testing.T, tenantID string, roles ...string) string {
	t.Helper()
	pair, err := s.jwtManager.GeneratePair(userID, tenantID, "admin@sigif.com", roles)
	if err != nil {
		t.Fatalf("GeneratePair() error = %v", err)
	}
	return pair.AccessToken
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
	token := s.token(t, tenantA, "inventory")

	status, resp := s.post(t, "/api/v1/categories", token, map[string]any{"name": "  Bebidas ", "description": "Gaseosas y jugos"})

	if status != fiber.StatusCreated {
		t.Fatalf("status = %d, want 201 (error %+v)", status, resp.Error)
	}
	var category map[string]any
	_ = json.Unmarshal(resp.Data, &category)
	if category["name"] != "Bebidas" || category["status"] != "active" || category["tenant_id"] != tenantA {
		t.Errorf("unexpected category: %v", category)
	}
}

func TestCreateCategoryEndpointErrors(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, tenantA, "manager")
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
	token := s.token(t, tenantA, "inventory")
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
		"tenant_id":       tenantA,
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
	token := s.token(t, tenantA, "inventory")
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
	token := s.token(t, tenantA, "inventory")
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

	t.Run("same sku in another tenant is allowed", func(t *testing.T) {
		tokenB := s.token(t, tenantB, "tenant_admin")
		categoryB := s.createCategory(t, tokenB, "Bebidas")
		status, resp := s.post(t, "/api/v1/products", tokenB, productBody(categoryB))
		if status != fiber.StatusCreated {
			t.Fatalf("got %d %+v, want 201", status, resp.Error)
		}
	})
}

func TestCreateProductEndpointCategoryNotFound(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, tenantA, "inventory")
	tokenB := s.token(t, tenantB, "inventory")
	foreignCategory := s.createCategory(t, tokenB, "Bebidas")

	status, resp := s.post(t, "/api/v1/products", token, productBody(foreignCategory))

	if status != fiber.StatusNotFound || resp.Error.Code != "NOT_FOUND" {
		t.Fatalf("got %d %+v, want 404 NOT_FOUND", status, resp.Error)
	}
}

func TestCatalogEndpointsPermissions(t *testing.T) {
	s := newTestServer(t)
	categoryID := s.createCategory(t, s.token(t, tenantA, "inventory"), "Bebidas")

	tests := []struct {
		name   string
		token  string
		status int
		code   string
	}{
		{name: "without token", token: "", status: 401, code: "UNAUTHORIZED"},
		{name: "invalid token", token: "not-a-jwt", status: 401, code: "UNAUTHORIZED"},
		{name: "cashier", token: s.token(t, tenantA, "cashier"), status: 403, code: "FORBIDDEN"},
		{name: "viewer", token: s.token(t, tenantA, "viewer"), status: 403, code: "FORBIDDEN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, path := range []string{"/api/v1/products", "/api/v1/categories"} {
				status, resp := s.post(t, path, tt.token, productBody(categoryID))
				if status != tt.status || resp.Error == nil || resp.Error.Code != tt.code {
					t.Fatalf("%s: got %d %+v, want %d %s", path, status, resp.Error, tt.status, tt.code)
				}
			}
		})
	}

	for _, role := range router.CatalogManagerRoles {
		t.Run("allowed role "+role, func(t *testing.T) {
			status, resp := s.post(t, "/api/v1/categories", s.token(t, tenantA, role), map[string]any{"name": "Categoria " + role})
			if status != fiber.StatusCreated {
				t.Fatalf("got %d %+v, want 201", status, resp.Error)
			}
		})
	}
}

func TestCreateProductUsesTenantFromTokenNotHeader(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, tenantA, "inventory")
	categoryID := s.createCategory(t, token, "Bebidas")

	raw, _ := json.Marshal(productBody(categoryID))
	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/products", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", tenantB)

	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	var parsed apiResponse
	_ = json.NewDecoder(resp.Body).Decode(&parsed)
	var product map[string]any
	_ = json.Unmarshal(parsed.Data, &product)
	if product["tenant_id"] != tenantA {
		t.Errorf("tenant_id = %v, want tenant from token %s", product["tenant_id"], tenantA)
	}
}
