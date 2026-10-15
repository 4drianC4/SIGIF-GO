//go:build integration

package supplier_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	appHandler "github.com/sigif/sigif-go/internal/modules/supplier/application/handler"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/service"
	supplierGorm "github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/model"
	httpHandler "github.com/sigif/sigif-go/internal/modules/supplier/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/supplier/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/database"
	"github.com/sigif/sigif-go/internal/shared/database/dbtest"
	"github.com/sigif/sigif-go/internal/shared/database/migrations"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type sessions struct{}

func (sessions) IsActive(context.Context, string) (bool, error) { return true, nil }

// Only the role named allowed receives permissions in this test fixture.
type permissions struct{}

func (permissions) HasPermission(ctx context.Context, _ uuid.UUID, _, _ string) (bool, error) {
	return ctx.Value("role") == "allowed", nil
}

type server struct {
	app     *fiber.App
	db      *gorm.DB
	jwt     *jwt.JWTManager
	clk     *clock.MockClock
	company uuid.UUID
	service *service.SupplierService
}

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Meta    struct {
		Total       int64 `json:"total"`
		Page, Limit int
	} `json:"meta"`
	Error struct{ Code, Message string } `json:"error"`
}

func newServer(t *testing.T) *server {
	t.Helper()
	dsn := dbtest.NewSchemaDSN(t)
	if err := migrations.Up(dsn); err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	for range 2 {
		if err := db.AutoMigrate(&model.SupplierModel{}, &model.ProductSupplierModel{}); err != nil {
			t.Fatalf("AutoMigrate: %v", err)
		}
	}
	for _, table := range []string{"purchase_orders", "purchase_order_items", "goods_receipts", "goods_receipt_items", "accounts_payable", "supplier_returns"} {
		if db.Migrator().HasTable(table) {
			t.Fatalf("AutoMigrate unexpectedly created %s", table)
		}
	}
	wrapped := &database.Database{DB: db}
	clk := clock.NewMockClock(time.Date(2026, 10, 15, 14, 0, 0, 0, time.UTC))
	svc := service.NewSupplierService(supplierGorm.NewSupplierGormRepository(wrapped), supplierGorm.NewProductSupplierGormRepository(wrapped), clk)
	h := httpHandler.NewSupplierHTTPHandler(appHandler.NewSupplierCommandHandler(svc), appHandler.NewSupplierQueryHandler(svc), validator.New())
	manager := jwt.NewManager(&config.Config{JWT: config.JWTConfig{Secret: "supplier-integration-test", AccessTokenExpiry: 60, Issuer: "sigif-test"}})
	app := fiber.New()
	app.Use(middleware.AuthRequired(manager, sessions{}))
	router.RegisterSupplierRoutes(app.Group("/api/v1"), h, permissions{})
	return &server{app: app, db: db, jwt: manager, clk: clk, company: uuid.New(), service: svc}
}

func (s *server) token(t *testing.T, company uuid.UUID, role string) string {
	t.Helper()
	companyText := ""
	if company != uuid.Nil {
		companyText = company.String()
	}
	token, _, err := s.jwt.GenerateAccessToken(uuid.NewString(), uuid.NewString(), companyText, "test@example.com", role)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func (s *server) request(t *testing.T, method, path, token, body string, want int) envelope {
	t.Helper()
	req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != want {
		t.Fatalf("%s %s: status %d, want %d: %s", method, path, resp.StatusCode, want, raw)
	}
	var result envelope
	if want != 204 {
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatalf("invalid response: %s", raw)
		}
		if result.Success != (want < 400) {
			t.Fatalf("unexpected success flag: %s", raw)
		}
	}
	return result
}

func object(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func array(t *testing.T, raw json.RawMessage) []map[string]any {
	t.Helper()
	var result []map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected [] rather than null")
	}
	return result
}

func (s *server) exec(t *testing.T, query string, args ...any) {
	t.Helper()
	if err := s.db.Exec(query, args...).Error; err != nil {
		t.Fatal(err)
	}
}

func (s *server) product(t *testing.T, company uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id, category := uuid.New(), uuid.New()
	s.exec(t, "INSERT INTO categories (id,company_id,name) VALUES (?,?,?)", category, company, "Category "+category.String())
	s.exec(t, `INSERT INTO products (id,company_id,category_id,unit_of_measure_id,name,sku,cost,sale_price,status,last_movement_at)
		VALUES (?,?,?,(SELECT id FROM units_of_measure WHERE abbreviation='UNIT'),?,?,8,10,'active',?)`, id, company, category, name, "SKU-"+id.String(), s.clk.NowUTC())
	return id
}

func TestSupplierLifecycleAndFrontendOnlyIndicators(t *testing.T) {
	s := newServer(t)
	token := s.token(t, s.company, "allowed")
	created := s.request(t, "POST", "/suppliers", token, `{"trade_name":" Andina ","legal_name":" Andina S.R.L. ","tax_id":"001044556677","contact":"Marcela","email":"ventas@andina.bo","lead_time_days":2,"minimum_order":1500}`, 201)
	data := object(t, created.Data)
	id := data["id"].(string)
	if data["name"] != "Andina" || data["legal_name"] != "Andina S.R.L." || data["status"] != "active" || data["updated_at"] != nil {
		t.Fatalf("created: %v", data)
	}
	s.request(t, "POST", "/suppliers", token, `{"legal_name":"Duplicado","tax_id":"001044556677"}`, 409)
	foreignToken := s.token(t, uuid.New(), "allowed")
	s.request(t, "POST", "/suppliers", foreignToken, `{"legal_name":"Otra empresa","tax_id":"001044556677"}`, 201)
	for _, term := range []string{"ANDINA", "S.R.L.", "001044", "100%25_"} {
		resp := s.request(t, "GET", "/suppliers?search="+term, token, "", 200)
		items := array(t, resp.Data)
		want := 1
		if term == "100%25_" {
			want = 0
		}
		if len(items) != want || resp.Meta.Total != int64(want) {
			t.Fatalf("search %s: %s", term, resp.Data)
		}
	}
	summary := object(t, s.request(t, "GET", "/suppliers/summary", token, "", 200).Data)
	if len(summary) != 1 || summary["active_suppliers"] != float64(1) {
		t.Fatalf("summary must contain only real active count: %v", summary)
	}
	detail := object(t, s.request(t, "GET", "/suppliers/"+id, token, "", 200).Data)
	for _, field := range []string{"indicators", "outstanding_amount", "last_purchase_at", "purchases_last_12_months"} {
		if _, ok := detail[field]; ok {
			t.Fatalf("unimplemented field %s returned by backend", field)
		}
	}
	s.clk.Add(time.Hour)
	updated := object(t, s.request(t, "PUT", "/suppliers/"+id, token, `{"legal_name":"Andina editada","tax_id":"001044556677","lead_time_days":0,"minimum_order":0}`, 200).Data)
	if updated["trade_name"] != nil || updated["email"] != nil || updated["name"] != "Andina editada" || updated["minimum_order"] != float64(0) || updated["created_at"] != data["created_at"] {
		t.Fatalf("PUT must preserve keys/dates and clear omitted optional fields: %v", updated)
	}
	deactivated := object(t, s.request(t, "PATCH", "/suppliers/"+id+"/deactivate", token, "", 200).Data)
	s.clk.Add(time.Hour)
	repeated := object(t, s.request(t, "PATCH", "/suppliers/"+id+"/deactivate", token, "", 200).Data)
	if repeated["status"] != "inactive" || repeated["updated_at"] != deactivated["updated_at"] {
		t.Fatal("deactivation must be idempotent")
	}
	s.request(t, "POST", "/suppliers", token, `{"legal_name":"Inactivo también reserva el NIT","tax_id":"001044556677"}`, 409)
	summary = object(t, s.request(t, "GET", "/suppliers/summary", token, "", 200).Data)
	if summary["active_suppliers"] != float64(0) {
		t.Fatal("deactivated supplier counted as active")
	}
	if len(array(t, s.request(t, "GET", "/suppliers?status=inactive", token, "", 200).Data)) != 1 {
		t.Fatal("inactive supplier hidden")
	}
	for _, body := range []string{`{}`, `{"legal_name":" "}`, `{"legal_name":"Valid","email":"invalid"}`, `{"legal_name":"Valid","minimum_order":1.001}`, `{"legal_name":"Valid","lead_time_days":-1}`, `{"legal_name":"Valid","lead_time_days":1.5}`} {
		s.request(t, "POST", "/suppliers", token, body, 400)
	}
	for _, q := range []string{"status=blocked", "sort_by=tax_id", "sort_order=invalid"} {
		s.request(t, "GET", "/suppliers?"+q, token, "", 400)
	}
	page := s.request(t, "GET", "/suppliers?page=999&limit=101", token, "", 200)
	if len(array(t, page.Data)) != 0 || page.Meta.Total != 1 || page.Meta.Page != 999 || page.Meta.Limit != 20 {
		t.Fatal("invalid pagination behavior")
	}
	// The index must also reject duplicates when bypassing service validation.
	taxID := "001044556677"
	duplicate, err := entity.NewSupplier(s.clk, s.company, entity.Fields{LegalName: "Direct duplicate", TaxID: &taxID})
	if err != nil {
		t.Fatal(err)
	}
	err = supplierGorm.NewSupplierGormRepository(&database.Database{DB: s.db}).Create(context.Background(), duplicate)
	if !errors.Is(err, service.ErrTaxIDTaken) {
		t.Fatalf("duplicate error: %v", err)
	}
}

func TestSupplierRoutePermissionsAndTenantIsolation(t *testing.T) {
	s := newServer(t)
	allowed := s.token(t, s.company, "allowed")
	denied := s.token(t, s.company, "denied")
	noCompany := s.token(t, uuid.Nil, "allowed")
	foreign := s.token(t, uuid.New(), "allowed")
	id := object(t, s.request(t, "POST", "/suppliers", allowed, `{"legal_name":"Andina"}`, 201).Data)["id"].(string)
	base := "/suppliers/" + id
	routes := []struct{ method, path, body string }{
		{"GET", "/suppliers", ""}, {"GET", "/suppliers/summary", ""}, {"GET", "/suppliers/form-options", ""},
		{"POST", "/suppliers", `{"legal_name":"New"}`}, {"GET", base, ""}, {"PUT", base, `{"legal_name":"New"}`},
		{"PATCH", base + "/deactivate", ""}, {"GET", base + "/products", ""},
	}
	for _, r := range routes {
		s.request(t, r.method, r.path, "", r.body, 401)
		s.request(t, r.method, r.path, denied, r.body, 403)
		s.request(t, r.method, r.path, noCompany, r.body, 400)
		if strings.HasPrefix(r.path, base) {
			s.request(t, r.method, r.path, foreign, r.body, 404)
		}
	}
	s.request(t, "GET", "/suppliers/not-a-uuid", allowed, "", 400)
	options := object(t, s.request(t, "GET", "/suppliers/form-options", allowed, "", 200).Data)
	if len(options["industries"].([]any)) != 0 || len(options["payment_terms"].([]any)) != 0 {
		t.Fatal("form options must not contain mock values")
	}
}

func TestSupplierProductsAreReadOnly(t *testing.T) {
	s := newServer(t)
	token := s.token(t, s.company, "allowed")
	id := object(t, s.request(t, "POST", "/suppliers", token, `{"legal_name":"Andina"}`, 201).Data)["id"].(string)
	product := s.product(t, s.company, "Aceite")
	s.exec(t, "INSERT INTO product_suppliers (id,company_id,supplier_id,product_id,current_cost,lead_time_days,created_at) VALUES (?,?,?,?,13.30,5,?)",
		uuid.New(), s.company, id, product, s.clk.NowUTC())
	base := "/suppliers/" + id
	items := array(t, s.request(t, "GET", base+"/products", token, "", 200).Data)
	if len(items) != 1 || items[0]["name"] != "Aceite" || items[0]["current_cost"] != float64(13.3) {
		t.Fatalf("products: %v", items)
	}
	s.request(t, "PATCH", base+"/deactivate", token, "", 200)
	if len(array(t, s.request(t, "GET", base+"/products", token, "", 200).Data)) != 1 {
		t.Fatal("deactivation removed association")
	}
	s.request(t, "GET", base+"/available-products", token, "", 404)
	s.request(t, "POST", base+"/products", token, "{}", 405)
	s.request(t, "PUT", base+"/products/"+product.String(), token, "{}", 404)
	s.request(t, "DELETE", base+"/products/"+product.String(), token, "", 404)
}
