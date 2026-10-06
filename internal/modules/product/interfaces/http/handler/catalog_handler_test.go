package handler_test

import (
	"bytes"
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

var allPermissions = []string{"products.list", "products.create", "categories.list", "categories.create"}

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Meta    json.RawMessage `json:"meta"`
	Error   *struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details map[string]string `json:"details"`
	} `json:"error"`
	raw string
}

type testServer struct {
	app        *fiber.App
	jwtManager *jwt.JWTManager
	store      *testutil.MemoryStore
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

	store := testutil.NewMemoryStore()
	clk := clock.NewMockClock(time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC))
	svc := service.NewCatalogService(store.ProductRepository(), store.CategoryRepository(), store.UnitRepository(), store.TaxRepository(), clk)
	h := httpHandler.NewCatalogHTTPHandler(
		appHandler.NewCatalogCommandHandler(svc, events.NewBus()),
		appHandler.NewCatalogQueryHandler(svc),
		validator.New(),
	)
	perms := &stubPermissions{byUser: map[uuid.UUID]map[string]bool{}}

	app := fiber.New()
	app.Use(middleware.AuthRequired(jwtManager, activeSessions{}))
	router.RegisterCatalogRoutes(app.Group("/api/v1"), h, perms)

	return &testServer{app: app, jwtManager: jwtManager, store: store, perms: perms}
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

func (s *testServer) admin(t *testing.T, companyID string) string {
	t.Helper()
	return s.token(t, companyID, allPermissions...)
}

func (s *testServer) do(t *testing.T, method, path, token string, body any) (int, apiResponse) {
	t.Helper()
	var reader io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		reader = strings.NewReader(b)
	default:
		raw, _ := json.Marshal(b)
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
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
	parsed.raw = string(payload)
	return resp.StatusCode, parsed
}

func (s *testServer) createCategory(t *testing.T, token string, body map[string]any) string {
	t.Helper()
	status, resp := s.do(t, fiber.MethodPost, "/api/v1/categories", token, body)
	if status != fiber.StatusCreated {
		t.Fatalf("create category status = %d, error = %+v", status, resp.Error)
	}
	var category struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(resp.Data, &category)
	return category.ID
}

func productBody(categoryID string) map[string]any {
	return map[string]any{
		"name":               "Galletas María 200g",
		"sku":                "aba-412",
		"category_id":        categoryID,
		"unit_of_measure_id": testutil.UnitID.String(),
		"cost":               4.20,
		"sale_price":         6.00,
		"initial_stock":      0,
	}
}

func expectError(t *testing.T, status int, resp apiResponse, wantStatus int, wantCode string) {
	t.Helper()
	if status != wantStatus || resp.Success || resp.Error == nil || resp.Error.Code != wantCode {
		t.Fatalf("got %d %s, want %d %s", status, resp.raw, wantStatus, wantCode)
	}
}

func decode(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("data is not an object: %s", raw)
	}
	return out
}

func TestCreateProductFollowsContract(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	categoryID := s.createCategory(t, token, map[string]any{"name": "Abarrotes"})

	status, resp := s.do(t, fiber.MethodPost, "/api/v1/products", token, productBody(categoryID))

	if status != fiber.StatusCreated || !resp.Success {
		t.Fatalf("status = %d, body = %s", status, resp.raw)
	}
	for _, fragment := range []string{`"cost":4.20`, `"price":6.00`, `"margin":30`, `"stock":0`, `"stock_status":"out_of_stock"`, `"status":"active"`, `"sku":"ABA-412"`} {
		if !strings.Contains(resp.raw, fragment) {
			t.Errorf("response missing %s: %s", fragment, resp.raw)
		}
	}
	product := decode(t, resp.Data)
	category := product["category"].(map[string]any)
	if category["id"] != categoryID || category["name"] != "Abarrotes" {
		t.Errorf("category = %v", category)
	}
}

func TestCreateProductWithoutOptionalFields(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	body := productBody(s.createCategory(t, token, map[string]any{"name": "Abarrotes"}))
	delete(body, "sku")
	delete(body, "initial_stock")

	status, resp := s.do(t, fiber.MethodPost, "/api/v1/products", token, body)

	if status != fiber.StatusCreated || !strings.Contains(resp.raw, `"sku":null`) || !strings.Contains(resp.raw, `"stock":0`) {
		t.Fatalf("got %d %s", status, resp.raw)
	}
}

func TestCreateProductValidation(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	categoryID := s.createCategory(t, token, map[string]any{"name": "Abarrotes"})

	tests := []struct {
		name    string
		mutate  func(map[string]any)
		details []string
	}{
		{"empty body", func(b map[string]any) { clear(b) }, []string{"name", "category_id", "unit_of_measure_id", "cost", "sale_price"}},
		{"missing sale_price", func(b map[string]any) { delete(b, "sale_price") }, []string{"sale_price"}},
		{"blank name", func(b map[string]any) { b["name"] = "   " }, []string{"name"}},
		{"category_id not uuid", func(b map[string]any) { b["category_id"] = "abc" }, []string{"category_id"}},
		{"unit not uuid", func(b map[string]any) { b["unit_of_measure_id"] = "kg" }, []string{"unit_of_measure_id"}},
		{"cost zero", func(b map[string]any) { b["cost"] = 0 }, []string{"cost"}},
		{"negative sale price", func(b map[string]any) { b["sale_price"] = -6 }, []string{"sale_price"}},
		{"negative initial stock", func(b map[string]any) { b["initial_stock"] = -1 }, []string{"initial_stock"}},
		{"barcode with letters", func(b map[string]any) { b["barcode"] = "77AB0000" }, []string{"barcode"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := productBody(categoryID)
			tt.mutate(body)
			status, resp := s.do(t, fiber.MethodPost, "/api/v1/products", token, body)
			expectError(t, status, resp, 400, "VALIDATION_ERROR")
			for _, field := range tt.details {
				if _, ok := resp.Error.Details[field]; !ok {
					t.Errorf("missing detail %q in %v", field, resp.Error.Details)
				}
			}
		})
	}

	status, resp := s.do(t, fiber.MethodPost, "/api/v1/products", token, `{"name":`)
	expectError(t, status, resp, 400, "BAD_REQUEST")
	if s.store.ProductCount() != 0 {
		t.Errorf("invalid requests must not store products")
	}
}

func TestCreateProductDuplicatesAndReferences(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	categoryID := s.createCategory(t, token, map[string]any{"name": "Abarrotes"})
	if status, resp := s.do(t, fiber.MethodPost, "/api/v1/products", token, productBody(categoryID)); status != 201 {
		t.Fatalf("first product: %d %s", status, resp.raw)
	}

	tests := []struct {
		name    string
		mutate  func(map[string]any)
		status  int
		code    string
		message string
	}{
		{"duplicate name", func(b map[string]any) { b["sku"] = "OTRO-1"; b["name"] = "galletas maría 200G" }, 409, "CONFLICT", "product name already exists"},
		{"duplicate sku", func(b map[string]any) { b["name"] = "Otro"; b["sku"] = "ABA-412" }, 409, "CONFLICT", "SKU already exists"},
		{"category not found", func(b map[string]any) { b["name"] = "Otro"; b["sku"] = ""; b["category_id"] = uuid.NewString() }, 404, "NOT_FOUND", "category not found"},
		{"unit not found", func(b map[string]any) { b["name"] = "Otro"; b["sku"] = ""; b["unit_of_measure_id"] = uuid.NewString() }, 404, "NOT_FOUND", "unit of measure not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := productBody(categoryID)
			tt.mutate(body)
			status, resp := s.do(t, fiber.MethodPost, "/api/v1/products", token, body)
			expectError(t, status, resp, tt.status, tt.code)
			if resp.Error.Message != tt.message {
				t.Errorf("message = %q, want %q", resp.Error.Message, tt.message)
			}
		})
	}
}

func TestListProductsPaginationAndFilters(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	lacteos := s.createCategory(t, token, map[string]any{"name": "Lácteos"})
	abarrotes := s.createCategory(t, token, map[string]any{"name": "Abarrotes"})
	for i, p := range []struct{ name, sku, category string }{
		{"Leche PIL entera 1L", "LAC-001", lacteos},
		{"Yogurt frutilla", "LAC-002", lacteos},
		{"Galletas María", "ABA-412", abarrotes},
	} {
		body := productBody(p.category)
		body["name"], body["sku"], body["initial_stock"] = p.name, p.sku, 10*(i+1)
		if status, resp := s.do(t, fiber.MethodPost, "/api/v1/products", token, body); status != 201 {
			t.Fatalf("create %s: %d %s", p.name, status, resp.raw)
		}
	}

	status, resp := s.do(t, fiber.MethodGet, "/api/v1/products?page=1&limit=2", token, nil)
	if status != 200 || !resp.Success {
		t.Fatalf("list: %d %s", status, resp.raw)
	}
	var items []map[string]any
	_ = json.Unmarshal(resp.Data, &items)
	meta := decode(t, resp.Meta)
	if len(items) != 2 || meta["total"] != float64(3) || meta["page"] != float64(1) || meta["limit"] != float64(2) || meta["total_pages"] != float64(2) {
		t.Fatalf("pagination: items=%d meta=%v", len(items), meta)
	}
	for _, key := range []string{"id", "name", "sku", "category", "cost", "price", "margin", "stock", "stock_status", "status"} {
		if _, ok := items[0][key]; !ok {
			t.Errorf("list item missing %q: %v", key, items[0])
		}
	}

	for query, want := range map[string]int{
		"search=leche":               1,
		"search=LAC-":                2,
		"category_id=" + abarrotes:   1,
		"status=active":              3,
		"status=inactive":            0,
		"search=leche&status=active": 1,
	} {
		_, resp := s.do(t, fiber.MethodGet, "/api/v1/products?"+query, token, nil)
		var got []map[string]any
		_ = json.Unmarshal(resp.Data, &got)
		if len(got) != want {
			t.Errorf("%s: got %d items, want %d (%s)", query, len(got), want, resp.raw)
		}
	}

	for _, query := range []string{"status=deleted", "category_id=abc"} {
		status, resp := s.do(t, fiber.MethodGet, "/api/v1/products?"+query, token, nil)
		expectError(t, status, resp, 400, "VALIDATION_ERROR")
	}

	_, resp = s.do(t, fiber.MethodGet, "/api/v1/products", s.admin(t, companyB), nil)
	if strings.Contains(resp.raw, "Leche") {
		t.Errorf("company B must not see company A products")
	}
}

func TestProductSummaryEndpoint(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	categoryID := s.createCategory(t, token, map[string]any{"name": "Lácteos"})
	body := productBody(categoryID)
	body["cost"], body["initial_stock"], body["min_stock"] = 5.40, 3, 5
	s.do(t, fiber.MethodPost, "/api/v1/products", token, body)

	status, resp := s.do(t, fiber.MethodGet, "/api/v1/products/summary", token, nil)

	if status != 200 {
		t.Fatalf("summary: %d %s", status, resp.raw)
	}
	want := `{"active_products":1,"inventory_value":16.20,"low_stock":1,"no_movement_90_days":0}`
	if string(resp.Data) != want {
		t.Errorf("summary = %s, want %s", resp.Data, want)
	}
}

func TestValidateDuplicateEndpoint(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	s.do(t, fiber.MethodPost, "/api/v1/products", token, productBody(s.createCategory(t, token, map[string]any{"name": "Abarrotes"})))

	tests := map[string]string{
		"name=Galletas%20Mar%C3%ADa%20200g&sku=ABA-412": `{"exists":true,"field":"name"}`,
		"sku=aba-412":          `{"exists":true,"field":"sku"}`,
		"name=Nuevo&sku=NEW-1": `{"exists":false,"field":null}`,
	}
	for query, want := range tests {
		status, resp := s.do(t, fiber.MethodGet, "/api/v1/products/validate-duplicate?"+query, token, nil)
		if status != 200 || string(resp.Data) != want {
			t.Errorf("%s: got %d %s, want %s", query, status, resp.Data, want)
		}
	}

	status, resp := s.do(t, fiber.MethodGet, "/api/v1/products/validate-duplicate", token, nil)
	expectError(t, status, resp, 400, "BAD_REQUEST")
}

func TestCategoriesEndpoints(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)

	status, resp := s.do(t, fiber.MethodPost, "/api/v1/categories", token, map[string]any{
		"name": "Abarrotes", "parent_id": nil, "default_tax": testutil.IVAGeneral, "target_margin": 25,
	})
	if status != 201 {
		t.Fatalf("create parent: %d %s", status, resp.raw)
	}
	parent := decode(t, resp.Data)
	if parent["default_tax"] != testutil.IVAGeneral || parent["target_margin"] != float64(25) || parent["parent_id"] != nil {
		t.Errorf("parent = %v", parent)
	}
	parentID := parent["id"].(string)
	childID := s.createCategory(t, token, map[string]any{"name": "Aceites", "parent_id": parentID})
	s.do(t, fiber.MethodPost, "/api/v1/products", token, productBody(parentID))

	status, resp = s.do(t, fiber.MethodGet, "/api/v1/categories", token, nil)
	if status != 200 {
		t.Fatalf("list: %d %s", status, resp.raw)
	}
	var list []map[string]any
	_ = json.Unmarshal(resp.Data, &list)
	if len(list) != 2 || list[0]["products_count"] != float64(1) || list[1]["id"] != childID ||
		list[1]["parent_id"] != parentID || list[1]["default_tax"] != nil || list[1]["target_margin"] != nil {
		t.Errorf("list = %s", resp.Data)
	}

	errorsCases := []struct {
		name   string
		body   any
		status int
		code   string
	}{
		{"duplicate at same level", map[string]any{"name": "ABARROTES"}, 409, "CONFLICT"},
		{"duplicate child", map[string]any{"name": "aceites", "parent_id": parentID}, 409, "CONFLICT"},
		{"parent not found", map[string]any{"name": "Nueva X", "parent_id": uuid.NewString()}, 404, "NOT_FOUND"},
		{"parent not uuid", map[string]any{"name": "Nueva X", "parent_id": "abc"}, 400, "VALIDATION_ERROR"},
		{"unknown tax", map[string]any{"name": "Nueva X", "default_tax": "IVA 99%"}, 404, "NOT_FOUND"},
		{"margin over 100", map[string]any{"name": "Nueva X", "target_margin": 150}, 400, "VALIDATION_ERROR"},
		{"blank name", map[string]any{"name": " "}, 400, "VALIDATION_ERROR"},
		{"malformed json", `{"name":`, 400, "BAD_REQUEST"},
	}
	for _, tt := range errorsCases {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := s.do(t, fiber.MethodPost, "/api/v1/categories", token, tt.body)
			expectError(t, status, resp, tt.status, tt.code)
		})
	}

	s.createCategory(t, token, map[string]any{"name": "Aceites"})
}

func TestReferenceEndpoints(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, companyA)

	status, resp := s.do(t, fiber.MethodGet, "/api/v1/units-of-measure", token, nil)
	if status != 200 || !strings.Contains(string(resp.Data), `"abbreviation":"UNIT"`) {
		t.Errorf("units: %d %s", status, resp.raw)
	}
	status, resp = s.do(t, fiber.MethodGet, "/api/v1/taxes", token, nil)
	if status != 200 || !strings.Contains(string(resp.Data), `"name":"IVA general 13%","percentage":13`) {
		t.Errorf("taxes: %d %s", status, resp.raw)
	}
}

func TestCatalogPermissions(t *testing.T) {
	s := newTestServer(t)
	categoryID := s.createCategory(t, s.admin(t, companyA), map[string]any{"name": "Abarrotes"})

	routes := []struct {
		method, path, permission string
		body                     any
	}{
		{fiber.MethodGet, "/api/v1/products", "products.list", nil},
		{fiber.MethodGet, "/api/v1/products/summary", "products.list", nil},
		{fiber.MethodGet, "/api/v1/products/validate-duplicate?name=X", "products.create", nil},
		{fiber.MethodPost, "/api/v1/products", "products.create", productBody(categoryID)},
		{fiber.MethodGet, "/api/v1/categories", "categories.list", nil},
		{fiber.MethodPost, "/api/v1/categories", "categories.create", map[string]any{"name": "Nueva"}},
	}
	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			status, resp := s.do(t, r.method, r.path, "", r.body)
			expectError(t, status, resp, 401, "UNAUTHORIZED")

			var others []string
			for _, p := range allPermissions {
				if p != r.permission {
					others = append(others, p)
				}
			}
			status, resp = s.do(t, r.method, r.path, s.token(t, companyA, others...), r.body)
			expectError(t, status, resp, 403, "FORBIDDEN")

			status, resp = s.do(t, r.method, r.path, s.token(t, companyA, r.permission), r.body)
			if status >= 400 {
				t.Errorf("with %s: got %d %s", r.permission, status, resp.raw)
			}

			status, resp = s.do(t, r.method, r.path, s.token(t, "", r.permission), r.body)
			expectError(t, status, resp, 400, "BAD_REQUEST")
		})
	}

	for _, path := range []string{"/api/v1/units-of-measure", "/api/v1/taxes"} {
		status, resp := s.do(t, fiber.MethodGet, path, "", nil)
		expectError(t, status, resp, 401, "UNAUTHORIZED")
	}
}
