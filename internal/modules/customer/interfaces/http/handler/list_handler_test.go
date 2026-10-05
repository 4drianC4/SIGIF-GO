package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appHandler "github.com/sigif/sigif-go/internal/modules/customer/application/handler"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
	httpHandler "github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/modules/customer/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

const (
	companyA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	companyB = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
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

type customerItem struct {
	ID             string  `json:"id"`
	CompanyID      string  `json:"company_id"`
	LegalName      string  `json:"legal_name"`
	DocumentNumber *string `json:"document_number"`
	Status         string  `json:"status"`
}

type testServer struct {
	app        *fiber.App
	jwtManager *jwt.JWTManager
	customers  *testutil.MemoryCustomerRepository
	perms      *stubPermissions
	seq        int
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

	customers := testutil.NewMemoryCustomerRepository()
	clk := clock.NewMockClock(time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC))
	svc := service.NewCustomerService(customers, clk)
	h := httpHandler.NewCustomerHTTPHandler(
		appHandler.NewCustomerCommandHandler(svc),
		appHandler.NewCustomerQueryHandler(svc),
		validator.New(),
		cfg,
	)
	perms := &stubPermissions{byUser: map[uuid.UUID]map[string]bool{}}

	app := fiber.New()
	app.Use(middleware.AuthRequired(jwtManager, activeSessions{}))
	router.RegisterCustomerRoutes(app.Group("/api/v1"), h, perms)

	return &testServer{app: app, jwtManager: jwtManager, customers: customers, perms: perms}
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

func (s *testServer) reader(t *testing.T, companyID string) string {
	t.Helper()
	return s.token(t, companyID, "customers.list")
}

// seed stores a customer directly in the repository; each call is created one
// minute after the previous one so ordering by created_at is deterministic.
func (s *testServer) seed(t *testing.T, companyID, legalName, documentNumber string, status entity.CustomerStatus) *entity.Customer {
	t.Helper()
	s.seq++
	clk := clock.NewMockClock(time.Date(2026, 10, 1, 8, s.seq, 0, 0, time.UTC))
	var doc *string
	if documentNumber != "" {
		doc = &documentNumber
	}
	email := strings.ToLower(strings.ReplaceAll(legalName, " ", ".")) + "@example.com"
	phone := fmt.Sprintf("+5917000%04d", s.seq)
	customer := entity.NewCustomer(clk, uuid.MustParse(companyID), legalName, entity.DocumentTypeTaxID, doc, &phone, &email)
	customer.Status = status
	if err := s.customers.Create(context.Background(), customer); err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	return customer
}

func (s *testServer) list(t *testing.T, query, token string) (int, apiResponse, []customerItem) {
	t.Helper()
	path := "/api/v1/customers/"
	if query != "" {
		path += "?" + query
	}
	req := httptest.NewRequest(fiber.MethodGet, path, nil)
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
	var items []customerItem
	if parsed.Success {
		if err := json.Unmarshal(parsed.Data, &items); err != nil {
			t.Fatalf("data is not a list: %s", parsed.Data)
		}
	}
	return resp.StatusCode, parsed, items
}

// seedDirectory loads a small mixed directory for company A plus noise that
// must never be listed (deleted rows and another company's customers).
func (s *testServer) seedDirectory(t *testing.T) {
	t.Helper()
	s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	s.seed(t, companyA, "Supermercado Norte", "7788990011", entity.CustomerStatusActive)
	s.seed(t, companyA, "Farmacia Sur", "5566778899", entity.CustomerStatusActive)
	s.seed(t, companyA, "Ferreteria Cerrada", "1020300000", entity.CustomerStatusInactive)
	s.seed(t, companyA, "Micromercado Moroso", "3344556677", entity.CustomerStatusBlocked)

	deleted := s.seed(t, companyA, "Ferreteria Eliminada", "9999999999", entity.CustomerStatusActive)
	deleted.SoftDelete(clock.NewRealClock())
	_ = s.customers.Update(context.Background(), deleted)

	s.seed(t, companyB, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
}

func names(items []customerItem) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = item.LegalName
	}
	return result
}

func TestListCustomersEndpointDefaultsToActive(t *testing.T) {
	s := newTestServer(t)
	s.seedDirectory(t)

	status, resp, items := s.list(t, "", s.reader(t, companyA))

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if resp.Meta == nil || *resp.Meta != (pageMeta{Page: 1, Limit: 20, Total: 3, TotalPages: 1}) {
		t.Fatalf("meta = %+v, want page 1 limit 20 total 3 total_pages 1", resp.Meta)
	}
	for _, item := range items {
		if item.Status != "active" || item.CompanyID != companyA || item.ID == "" {
			t.Errorf("unexpected item %+v", item)
		}
	}
	// Default order is newest first.
	want := []string{"Farmacia Sur", "Supermercado Norte", "Ferreteria Central"}
	if got := names(items); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("names = %v, want %v", got, want)
	}
}

func TestListCustomersEndpointReturnsOnlyIdentificationFields(t *testing.T) {
	s := newTestServer(t)
	s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)

	_, resp, _ := s.list(t, "", s.reader(t, companyA))

	var raw []map[string]any
	_ = json.Unmarshal(resp.Data, &raw)
	if len(raw) != 1 {
		t.Fatalf("expected 1 item, got %d", len(raw))
	}
	for _, field := range []string{"id", "legal_name", "document_type", "document_number", "phone", "email", "status"} {
		if _, ok := raw[0][field]; !ok {
			t.Errorf("missing field %q in %v", field, raw[0])
		}
	}
	for _, field := range []string{"credit_limit", "credit_balance", "points_accrued", "address"} {
		if _, ok := raw[0][field]; ok {
			t.Errorf("listing must not expose %q", field)
		}
	}
}

func TestListCustomersEndpointSearch(t *testing.T) {
	s := newTestServer(t)
	s.seedDirectory(t)
	token := s.reader(t, companyA)

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{name: "partial name ignoring case", query: "q=FERRET", want: []string{"Ferreteria Central"}},
		{name: "partial name lower case", query: "q=mercado", want: []string{"Supermercado Norte"}},
		{name: "exact document number", query: "q=1020304050", want: []string{"Ferreteria Central"}},
		{name: "email", query: "q=farmacia.sur@example.com", want: []string{"Farmacia Sur"}},
		{name: "surrounding spaces are trimmed", query: "q=%20%20norte%20%20", want: []string{"Supermercado Norte"}},
		{name: "blank search lists everything", query: "q=%20%20%20", want: []string{"Farmacia Sur", "Supermercado Norte", "Ferreteria Central"}},
		{name: "search combined with status", query: "q=ferreteria&status=all", want: []string{"Ferreteria Cerrada", "Ferreteria Central"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, items := s.list(t, tt.query, token)
			if status != fiber.StatusOK {
				t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
			}
			if got := names(items); strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("names = %v, want %v", got, tt.want)
			}
			if resp.Meta.Total != int64(len(tt.want)) {
				t.Errorf("total = %d, want %d", resp.Meta.Total, len(tt.want))
			}
		})
	}
}

func TestListCustomersEndpointNoMatchesReturnsEmptyList(t *testing.T) {
	s := newTestServer(t)
	s.seedDirectory(t)

	status, resp, items := s.list(t, "q=no-existe-xyz", s.reader(t, companyA))

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if string(resp.Data) != "[]" || len(items) != 0 {
		t.Errorf("data = %s, want []", resp.Data)
	}
	if *resp.Meta != (pageMeta{Page: 1, Limit: 20, Total: 0, TotalPages: 0}) {
		t.Errorf("meta = %+v, want total 0", resp.Meta)
	}
}

func TestListCustomersEndpointStatusFilter(t *testing.T) {
	s := newTestServer(t)
	s.seedDirectory(t)
	token := s.reader(t, companyA)

	tests := []struct {
		query string
		want  []string
	}{
		{query: "status=active", want: []string{"Farmacia Sur", "Supermercado Norte", "Ferreteria Central"}},
		{query: "status=inactive", want: []string{"Ferreteria Cerrada"}},
		{query: "status=blocked", want: []string{"Micromercado Moroso"}},
		{query: "status=all", want: []string{"Micromercado Moroso", "Ferreteria Cerrada", "Farmacia Sur", "Supermercado Norte", "Ferreteria Central"}},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			status, resp, items := s.list(t, tt.query, token)
			if status != fiber.StatusOK {
				t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
			}
			if got := names(items); strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("names = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListCustomersEndpointPagination(t *testing.T) {
	s := newTestServer(t)
	for i := 1; i <= 25; i++ {
		s.seed(t, companyA, fmt.Sprintf("Cliente %02d", i), "", entity.CustomerStatusActive)
	}
	token := s.reader(t, companyA)

	tests := []struct {
		query     string
		wantMeta  pageMeta
		wantCount int
		wantFirst string
	}{
		{query: "limit=10&sort_by=legal_name&sort_order=asc", wantMeta: pageMeta{1, 10, 25, 3}, wantCount: 10, wantFirst: "Cliente 01"},
		{query: "page=2&limit=10&sort_by=legal_name&sort_order=asc", wantMeta: pageMeta{2, 10, 25, 3}, wantCount: 10, wantFirst: "Cliente 11"},
		{query: "page=3&limit=10&sort_by=legal_name&sort_order=asc", wantMeta: pageMeta{3, 10, 25, 3}, wantCount: 5, wantFirst: "Cliente 21"},
		{query: "page=4&limit=10", wantMeta: pageMeta{4, 10, 25, 3}, wantCount: 0},
		{query: "limit=100&sort_by=legal_name&sort_order=desc", wantMeta: pageMeta{1, 100, 25, 1}, wantCount: 25, wantFirst: "Cliente 25"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			status, resp, items := s.list(t, tt.query, token)
			if status != fiber.StatusOK {
				t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
			}
			if *resp.Meta != tt.wantMeta {
				t.Errorf("meta = %+v, want %+v", *resp.Meta, tt.wantMeta)
			}
			if len(items) != tt.wantCount {
				t.Fatalf("items = %d, want %d", len(items), tt.wantCount)
			}
			if tt.wantFirst != "" && items[0].LegalName != tt.wantFirst {
				t.Errorf("first = %s, want %s", items[0].LegalName, tt.wantFirst)
			}
		})
	}
}

func TestListCustomersEndpointInvalidParams(t *testing.T) {
	s := newTestServer(t)
	token := s.reader(t, companyA)

	tests := []struct {
		name   string
		query  string
		code   string
		detail string
	}{
		{name: "limit zero", query: "limit=0", code: "VALIDATION_ERROR", detail: "limit"},
		{name: "limit above max", query: "limit=500", code: "VALIDATION_ERROR", detail: "limit"},
		{name: "limit just above max", query: "limit=101", code: "VALIDATION_ERROR", detail: "limit"},
		{name: "page zero", query: "page=0", code: "VALIDATION_ERROR", detail: "page"},
		{name: "negative page", query: "page=-1", code: "VALIDATION_ERROR", detail: "page"},
		{name: "unknown status", query: "status=xyz", code: "VALIDATION_ERROR", detail: "status"},
		{name: "sort by non whitelisted field", query: "sort_by=credit_limit", code: "VALIDATION_ERROR", detail: "sort_by"},
		{name: "unknown sort order", query: "sort_order=up", code: "VALIDATION_ERROR", detail: "sort_order"},
		{name: "search too long", query: "q=" + strings.Repeat("a", 101), code: "VALIDATION_ERROR", detail: "q"},
		{name: "non numeric page", query: "page=abc", code: "BAD_REQUEST"},
		{name: "non numeric limit", query: "limit=ten", code: "BAD_REQUEST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.list(t, tt.query, token)
			if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != tt.code {
				t.Fatalf("got %d %+v, want 400 %s", status, resp.Error, tt.code)
			}
			if tt.detail != "" {
				if _, ok := resp.Error.Details[tt.detail]; !ok {
					t.Errorf("missing detail for %q in %v", tt.detail, resp.Error.Details)
				}
			}
		})
	}
}

func TestListCustomersEndpointAuthorization(t *testing.T) {
	s := newTestServer(t)
	s.seedDirectory(t)

	tests := []struct {
		name   string
		token  string
		status int
		code   string
	}{
		{name: "without token", token: "", status: 401, code: "UNAUTHORIZED"},
		{name: "invalid token", token: "not-a-jwt", status: 401, code: "UNAUTHORIZED"},
		{name: "without any permission", token: s.token(t, companyA), status: 403, code: "FORBIDDEN"},
		{name: "with only customers.read", token: s.token(t, companyA, "customers.read"), status: 403, code: "FORBIDDEN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.list(t, "", tt.token)
			if status != tt.status || resp.Error == nil || resp.Error.Code != tt.code {
				t.Fatalf("got %d %+v, want %d %s", status, resp.Error, tt.status, tt.code)
			}
		})
	}

	t.Run("token without company", func(t *testing.T) {
		status, resp, _ := s.list(t, "", s.reader(t, ""))
		if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Message != "company is required" {
			t.Fatalf("got %d %+v, want 400 company is required", status, resp.Error)
		}
	})
}

func TestListCustomersEndpointIsolatesCompanies(t *testing.T) {
	s := newTestServer(t)
	s.seedDirectory(t)
	s.seed(t, companyB, "Solo Empresa B", "1111111111", entity.CustomerStatusActive)

	t.Run("company A never sees company B", func(t *testing.T) {
		_, resp, items := s.list(t, "status=all&q=empresa", s.reader(t, companyA))
		if resp.Meta.Total != 0 || len(items) != 0 {
			t.Fatalf("company A got %v", names(items))
		}
	})

	t.Run("company B only sees its own customers", func(t *testing.T) {
		_, resp, items := s.list(t, "status=all", s.reader(t, companyB))
		if resp.Meta.Total != 2 {
			t.Fatalf("total = %d, want 2 (%v)", resp.Meta.Total, names(items))
		}
		for _, item := range items {
			if item.CompanyID != companyB {
				t.Errorf("leaked customer from %s: %+v", item.CompanyID, item)
			}
		}
	})

	t.Run("same document in another company is not matched", func(t *testing.T) {
		_, _, items := s.list(t, "q=1020304050", s.reader(t, companyB))
		if len(items) != 1 || items[0].CompanyID != companyB {
			t.Fatalf("got %+v, want only company B's customer", items)
		}
	})
}

func TestListCustomersEndpointDoesNotModifyData(t *testing.T) {
	s := newTestServer(t)
	s.seedDirectory(t)
	before := map[uuid.UUID]entity.Customer{}
	for id, c := range s.customers.Customers {
		before[id] = *c
	}
	writes := s.customers.Writes
	token := s.reader(t, companyA)

	for _, query := range []string{"", "q=ferreteria", "status=all", "status=inactive&page=2&limit=1", "limit=0"} {
		s.list(t, query, token)
	}

	if s.customers.Writes != writes {
		t.Errorf("listing performed %d writes", s.customers.Writes-writes)
	}
	for id, c := range s.customers.Customers {
		old := before[id]
		if c.Status != old.Status || c.LegalName != old.LegalName || (c.UpdatedAt == nil) != (old.UpdatedAt == nil) ||
			(c.UpdatedAt != nil && !c.UpdatedAt.Equal(*old.UpdatedAt)) {
			t.Errorf("customer %s changed: before %+v after %+v", id, old, *c)
		}
	}
}
