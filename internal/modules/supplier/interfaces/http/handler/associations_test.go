package handler_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	appHandler "github.com/sigif/sigif-go/internal/modules/supplier/application/handler"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/service"
	httpHandler "github.com/sigif/sigif-go/internal/modules/supplier/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/supplier/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type parentStub struct {
	repository.SupplierRepository
	supplier *entity.Supplier
}

func (s parentStub) GetByID(_ context.Context, company, id uuid.UUID) (*entity.Supplier, error) {
	if s.supplier.CompanyID != company || s.supplier.ID != id {
		return nil, nil
	}
	return s.supplier, nil
}
func (s parentStub) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type associationsStub struct {
	repository.ProductSupplierRepository
	company, product uuid.UUID
	supplier         *entity.Supplier
	saved            []*entity.ProductSupplier
}

func (s *associationsStub) GetProduct(_ context.Context, company, id uuid.UUID) (*repository.ProductRef, error) {
	if company != s.company || id != s.product {
		return nil, nil
	}
	return &repository.ProductRef{ID: id, Name: "Aceite", Status: "active"}, nil
}
func (*associationsStub) Get(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (*entity.ProductSupplier, error) {
	return nil, nil
}
func (s *associationsStub) CreateMany(_ context.Context, a []*entity.ProductSupplier) error {
	s.saved = a
	return nil
}
func (s *associationsStub) ProductOptions(context.Context, uuid.UUID, uuid.UUID, repository.ProductOptionFilter, int, int) ([]repository.ProductOption, int64, error) {
	return []repository.ProductOption{{ProductRef: repository.ProductRef{ID: s.product, Name: "Aceite", Status: "active"}, Stock: decimal.NewFromInt(8), CategoryName: "Aceites", CategoryPath: "Abarrotes / Aceites", UnitAbbreviation: "UNIT", IsAssociated: true}}, 1, nil
}
func (s *associationsStub) SupplierOptions(context.Context, uuid.UUID, uuid.UUID, string, int, int) ([]repository.SupplierOption, int64, error) {
	return []repository.SupplierOption{{Supplier: s.supplier, IsAssociated: true}}, 1, nil
}
func (s *associationsStub) ListSuppliers(context.Context, uuid.UUID, uuid.UUID, string, int, int) ([]repository.SupplierItem, int64, error) {
	return []repository.SupplierItem{{Supplier: s.supplier, Association: &entity.ProductSupplier{ID: uuid.New(), CompanyID: s.company, SupplierID: s.supplier.ID, ProductID: s.product, CreatedAt: time.Now()}}}, 1, nil
}

type permissionStub bool

func (p permissionStub) HasPermission(context.Context, uuid.UUID, string, string) (bool, error) {
	return bool(p), nil
}

type apiFixture struct {
	app                        *fiber.App
	company, supplier, product uuid.UUID
	repo                       *associationsStub
}

func api(t *testing.T, allowed bool) apiFixture {
	t.Helper()
	company, supplier, product := uuid.New(), uuid.New(), uuid.New()
	parent := &entity.Supplier{ID: supplier, CompanyID: company, Fields: entity.Fields{LegalName: "Andina"}, Status: entity.StatusActive}
	repo := &associationsStub{company: company, product: product, supplier: parent}
	svc := service.NewSupplierService(parentStub{supplier: parent}, repo, clock.NewMockClock(time.Now().UTC()))
	h := httpHandler.NewSupplierHTTPHandler(appHandler.NewSupplierCommandHandler(svc), appHandler.NewSupplierQueryHandler(svc), validator.New())
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.SetUserContext(middleware.WithUserID(middleware.WithCompanyID(c.UserContext(), company), uuid.New()))
		return c.Next()
	})
	router.RegisterSupplierRoutes(app.Group("/api/v1"), h, permissionStub(allowed))
	return apiFixture{app, company, supplier, product, repo}
}
func (f apiFixture) do(t *testing.T, method, path, body string, want int) map[string]any {
	t.Helper()
	r := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	resp, err := f.app.Test(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != want {
		t.Fatalf("%s %s: got %d, want %d: %v", method, path, resp.StatusCode, want, result)
	}
	return result
}
func TestAssociationHTTPContract(t *testing.T) {
	f := api(t, true)
	base := "/suppliers/" + f.supplier.String()
	created := f.do(t, "POST", base+"/products", `{"product_id":"`+f.product.String()+`","lead_time_days":3,"current_cost":0}`, 201)["data"].(map[string]any)
	if created["product_id"] != f.product.String() || created["current_cost"] != float64(0) || len(f.repo.saved) != 1 {
		t.Fatalf("create contract: %v", created)
	}
	f.do(t, "POST", base+"/products", `{"product_id":"`+f.product.String()+`"}`, 400)
	f.do(t, "POST", base+"/products", `{"product_id":"not-uuid","current_cost":1}`, 400)
	bulk := f.do(t, "POST", base+"/products/batch", `{"product_ids":["`+f.product.String()+`"]}`, 201)["data"].([]any)
	if bulk[0].(map[string]any)["current_cost"] != nil {
		t.Fatal("bulk must allow missing cost")
	}
	f.do(t, "POST", base+"/products/batch", `{"product_ids":[]}`, 400)
	f.do(t, "GET", base+"/product-options?category_id=invalid", "", 400)
	options := f.do(t, "GET", base+"/product-options", "", 200)["data"].([]any)
	if options[0].(map[string]any)["selectable"] != false {
		t.Fatal("already associated product selectable")
	}
	product := "/products/" + f.product.String()
	opts := f.do(t, "GET", product+"/supplier-options", "", 200)["data"].([]any)
	if opts[0].(map[string]any)["selectable"] != false {
		t.Fatal("already associated supplier selectable")
	}
	reverse := f.do(t, "GET", product+"/suppliers", "", 200)["data"].([]any)
	if reverse[0].(map[string]any)["supplier"].(map[string]any)["id"] != f.supplier.String() {
		t.Fatal("missing supplier reference")
	}
	f.do(t, "PUT", "/suppliers/invalid/products/"+f.product.String(), "{}", 400)
	f.do(t, "DELETE", base+"/products/invalid", "", 400)
}
func TestAssociationHTTPPermissions(t *testing.T) {
	f := api(t, false)
	base := "/suppliers/" + f.supplier.String()
	for _, r := range []struct{ method, path string }{
		{"GET", base + "/product-options"}, {"POST", base + "/products"}, {"POST", base + "/products/batch"},
		{"PUT", base + "/products/" + f.product.String()}, {"DELETE", base + "/products/" + f.product.String()},
		{"GET", "/products/" + f.product.String() + "/suppliers"}, {"GET", "/products/" + f.product.String() + "/supplier-options"},
	} {
		f.do(t, r.method, r.path, "{}", 403)
	}
}
