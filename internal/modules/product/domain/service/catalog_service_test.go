package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

var (
	companyA = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	companyB = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	now      = time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
)

type fixture struct {
	svc   *service.CatalogService
	store *testutil.MemoryStore
	clk   *clock.MockClock
}

func newFixture() fixture {
	store := testutil.NewMemoryStore()
	clk := clock.NewMockClock(now)
	return fixture{
		svc:   service.NewCatalogService(store.ProductRepository(), store.CategoryRepository(), store.UnitRepository(), store.TaxRepository(), clk),
		store: store,
		clk:   clk,
	}
}

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func ptr[T any](v T) *T { return &v }

func (f fixture) category(t *testing.T, companyID uuid.UUID, name string, parentID *uuid.UUID) *entity.Category {
	t.Helper()
	created, err := f.svc.CreateCategory(context.Background(), service.CreateCategoryParams{CompanyID: companyID, Name: name, ParentID: parentID})
	if err != nil {
		t.Fatalf("CreateCategory(%q) error = %v", name, err)
	}
	return created.Category
}

func validProduct(companyID, categoryID uuid.UUID) service.CreateProductParams {
	return service.CreateProductParams{
		CompanyID:       companyID,
		CategoryID:      categoryID,
		UnitOfMeasureID: testutil.UnitID,
		SKU:             " aba-412 ",
		Name:            "  Galletas  María 200g ",
		Cost:            dec("4.20"),
		SalePrice:       dec("6.00"),
	}
}

func (f fixture) product(t *testing.T, mutate func(*service.CreateProductParams)) *entity.Product {
	t.Helper()
	params := validProduct(companyA, f.category(t, companyA, "Cat "+uuid.NewString()[:8], nil).ID)
	if mutate != nil {
		mutate(&params)
	}
	created, err := f.svc.CreateProduct(context.Background(), params)
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
	return created.Product
}

func assertAppError(t *testing.T, err error, code sharedErrors.ErrorCode, status int) *sharedErrors.AppError {
	t.Helper()
	var appErr *sharedErrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError %s, got %v", code, err)
	}
	if appErr.Code != code || appErr.StatusCode != status {
		t.Fatalf("expected %s/%d, got %s/%d (%s)", code, status, appErr.Code, appErr.StatusCode, appErr.Message)
	}
	return appErr
}

func assertDetail(t *testing.T, err error, field string) {
	t.Helper()
	appErr := assertAppError(t, err, sharedErrors.CodeValidation, 400)
	details, _ := appErr.Details.(map[string]string)
	if _, ok := details[field]; !ok {
		t.Errorf("expected detail for %q, got %v", field, appErr.Details)
	}
}

func TestCreateCategory(t *testing.T) {
	f := newFixture()

	created, err := f.svc.CreateCategory(context.Background(), service.CreateCategoryParams{
		CompanyID:    companyA,
		Name:         "  Bebidas   frías ",
		DefaultTax:   ptr(testutil.IVAGeneral),
		TargetMargin: ptr(dec("25")),
	})
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}

	c := created.Category
	if c.Name != "Bebidas frías" || c.Status != entity.StatusActive || c.ParentID != nil {
		t.Errorf("unexpected category: %+v", c)
	}
	if c.DefaultTaxID == nil || *c.DefaultTaxID != testutil.IVAGeneralID || *created.DefaultTaxName != testutil.IVAGeneral {
		t.Errorf("default tax not resolved: %+v / %v", c.DefaultTaxID, created.DefaultTaxName)
	}
	if !c.TargetMargin.Equal(dec("25")) {
		t.Errorf("target margin = %v, want 25", c.TargetMargin)
	}
}

func TestCreateSubcategoryWithoutTaxOrMarginInheritsByStoringNull(t *testing.T) {
	f := newFixture()
	parent := f.category(t, companyA, "Abarrotes", nil)

	child := f.category(t, companyA, "Aceites", &parent.ID)

	if child.ParentID == nil || *child.ParentID != parent.ID {
		t.Fatalf("parent_id = %v, want %s", child.ParentID, parent.ID)
	}
	if child.DefaultTaxID != nil || child.TargetMargin != nil {
		t.Errorf("child without values must keep nil to inherit, got %+v %+v", child.DefaultTaxID, child.TargetMargin)
	}
}

func TestCreateCategoryNameIsUniquePerLevel(t *testing.T) {
	f := newFixture()
	abarrotes := f.category(t, companyA, "Abarrotes", nil)
	limpieza := f.category(t, companyA, "Limpieza", nil)
	f.category(t, companyA, "Aceites", &abarrotes.ID)

	_, err := f.svc.CreateCategory(context.Background(), service.CreateCategoryParams{CompanyID: companyA, Name: "ABARROTES"})
	assertAppError(t, err, sharedErrors.CodeConflict, 409)

	_, err = f.svc.CreateCategory(context.Background(), service.CreateCategoryParams{CompanyID: companyA, Name: "aceites", ParentID: &abarrotes.ID})
	assertAppError(t, err, sharedErrors.CodeConflict, 409)

	f.category(t, companyA, "Aceites", nil)
	f.category(t, companyA, "Aceites", &limpieza.ID)
	f.category(t, companyB, "Abarrotes", nil)
}

func TestCreateCategoryErrors(t *testing.T) {
	f := newFixture()
	foreign := f.category(t, companyB, "Ajena", nil)
	inactive := f.category(t, companyA, "Inactiva", nil)
	f.store.Categories[inactive.ID].Status = entity.StatusInactive

	tests := []struct {
		name   string
		params service.CreateCategoryParams
		code   sharedErrors.ErrorCode
		status int
	}{
		{"parent does not exist", service.CreateCategoryParams{CompanyID: companyA, Name: "X", ParentID: ptr(uuid.New())}, sharedErrors.CodeNotFound, 404},
		{"parent of another company", service.CreateCategoryParams{CompanyID: companyA, Name: "X", ParentID: &foreign.ID}, sharedErrors.CodeNotFound, 404},
		{"parent inactive", service.CreateCategoryParams{CompanyID: companyA, Name: "X", ParentID: &inactive.ID}, sharedErrors.CodeBadRequest, 400},
		{"unknown tax", service.CreateCategoryParams{CompanyID: companyA, Name: "X", DefaultTax: ptr("IVA 99%")}, sharedErrors.CodeNotFound, 404},
		{"margin over 100", service.CreateCategoryParams{CompanyID: companyA, Name: "X", TargetMargin: ptr(dec("100.01"))}, sharedErrors.CodeValidation, 400},
		{"negative margin", service.CreateCategoryParams{CompanyID: companyA, Name: "X", TargetMargin: ptr(dec("-1"))}, sharedErrors.CodeValidation, 400},
		{"no company", service.CreateCategoryParams{Name: "X"}, sharedErrors.CodeBadRequest, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.svc.CreateCategory(context.Background(), tt.params)
			assertAppError(t, err, tt.code, tt.status)
		})
	}
}

func TestCreateProduct(t *testing.T) {
	f := newFixture()

	p := f.product(t, func(p *service.CreateProductParams) {
		p.InitialStock = dec("10")
		p.MinStock = dec("3")
	})

	if p.SKU != "ABA-412" || p.Name != "Galletas María 200g" || p.Status != entity.StatusActive {
		t.Errorf("unexpected product: %+v", p)
	}
	if !p.Stock.Equal(dec("10")) || !p.LastMovementAt.Equal(now) {
		t.Errorf("stock/last movement not set: %v %v", p.Stock, p.LastMovementAt)
	}
	if got := p.Margin(); !got.Equal(dec("30")) {
		t.Errorf("margin = %v, want 30", got)
	}
}

func TestCreateProductOptionalFields(t *testing.T) {
	f := newFixture()

	p := f.product(t, func(p *service.CreateProductParams) { p.SKU = "" })

	if p.SKU != "" || !p.Stock.IsZero() {
		t.Errorf("sku and initial stock must be optional: %+v", p)
	}
	f.product(t, func(p *service.CreateProductParams) { p.SKU = ""; p.Name = "Otro sin SKU" })
}

func TestProductMarginAndStockStatus(t *testing.T) {
	tests := []struct {
		cost, price, stock, min string
		margin                  string
		status                  entity.StockStatus
	}{
		{"5.40", "7.00", "496", "10", "23", entity.StockStatusNormal},
		{"4.20", "6.00", "5", "5", "30", entity.StockStatusLowStock},
		{"4.20", "6.00", "0", "5", "30", entity.StockStatusOutOfStock},
		{"6.00", "5.00", "0", "0", "-20", entity.StockStatusOutOfStock},
	}
	for _, tt := range tests {
		p := entity.Product{Cost: dec(tt.cost), SalePrice: dec(tt.price), Stock: dec(tt.stock), MinStock: dec(tt.min)}
		if got := p.Margin(); !got.Equal(dec(tt.margin)) {
			t.Errorf("Margin(%s,%s) = %v, want %s", tt.cost, tt.price, got, tt.margin)
		}
		if got := p.StockStatus(); got != tt.status {
			t.Errorf("StockStatus(%s,%s) = %s, want %s", tt.stock, tt.min, got, tt.status)
		}
	}
}

func TestCreateProductRejectsDuplicates(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*service.CreateProductParams)
		message string
	}{
		{"same name ignoring case and spaces", func(p *service.CreateProductParams) {
			p.Name = "GALLETAS maría   200G"
			p.SKU = "OTRO-1"
			p.Barcode = ""
		}, "product name already exists"},
		{"same sku ignoring case", func(p *service.CreateProductParams) { p.Name = "Otro"; p.SKU = "Aba-412"; p.Barcode = "" }, "SKU already exists"},
		{"same barcode", func(p *service.CreateProductParams) { p.Name = "Otro"; p.SKU = "OTRO-1" }, "barcode already exists"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			category := f.category(t, companyA, "Abarrotes", nil)
			first := validProduct(companyA, category.ID)
			first.Barcode = "7750182000123"
			if _, err := f.svc.CreateProduct(context.Background(), first); err != nil {
				t.Fatalf("first CreateProduct() error = %v", err)
			}

			params := validProduct(companyA, category.ID)
			params.Barcode = "7750182000123"
			tt.mutate(&params)
			_, err := f.svc.CreateProduct(context.Background(), params)

			appErr := assertAppError(t, err, sharedErrors.CodeConflict, 409)
			if appErr.Message != tt.message {
				t.Errorf("message = %q, want %q", appErr.Message, tt.message)
			}
			if f.store.ProductCount() != 1 {
				t.Errorf("duplicate must not be stored")
			}
		})
	}
}

func TestCreateProductAllowsSameDataInAnotherCompany(t *testing.T) {
	f := newFixture()
	a := f.category(t, companyA, "Abarrotes", nil)
	b := f.category(t, companyB, "Abarrotes", nil)

	for _, params := range []service.CreateProductParams{validProduct(companyA, a.ID), validProduct(companyB, b.ID)} {
		if _, err := f.svc.CreateProduct(context.Background(), params); err != nil {
			t.Fatalf("CreateProduct() error = %v", err)
		}
	}
}

func TestCreateProductReferenceErrors(t *testing.T) {
	f := newFixture()
	own := f.category(t, companyA, "Abarrotes", nil)
	foreign := f.category(t, companyB, "Ajena", nil)
	inactive := f.category(t, companyA, "Inactiva", nil)
	f.store.Categories[inactive.ID].Status = entity.StatusInactive

	tests := []struct {
		name    string
		mutate  func(*service.CreateProductParams)
		code    sharedErrors.ErrorCode
		status  int
		message string
	}{
		{"category does not exist", func(p *service.CreateProductParams) { p.CategoryID = uuid.New() }, sharedErrors.CodeNotFound, 404, "category not found"},
		{"category of another company", func(p *service.CreateProductParams) { p.CategoryID = foreign.ID }, sharedErrors.CodeNotFound, 404, "category not found"},
		{"category inactive", func(p *service.CreateProductParams) { p.CategoryID = inactive.ID }, sharedErrors.CodeBadRequest, 400, "category is inactive"},
		{"unit does not exist", func(p *service.CreateProductParams) { p.UnitOfMeasureID = uuid.New() }, sharedErrors.CodeNotFound, 404, "unit of measure not found"},
		{"no company", func(p *service.CreateProductParams) { p.CompanyID = uuid.Nil }, sharedErrors.CodeBadRequest, 400, "company is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := validProduct(companyA, own.ID)
			tt.mutate(&params)
			_, err := f.svc.CreateProduct(context.Background(), params)
			appErr := assertAppError(t, err, tt.code, tt.status)
			if appErr.Message != tt.message {
				t.Errorf("message = %q, want %q", appErr.Message, tt.message)
			}
		})
	}
	if f.store.ProductCount() != 0 {
		t.Errorf("no product must be stored")
	}
}

func TestCreateProductBusinessValidation(t *testing.T) {
	tests := []struct {
		name   string
		field  string
		mutate func(*service.CreateProductParams)
	}{
		{"cost zero", "cost", func(p *service.CreateProductParams) { p.Cost = decimal.Zero }},
		{"cost negative", "cost", func(p *service.CreateProductParams) { p.Cost = dec("-1") }},
		{"sale price zero", "sale_price", func(p *service.CreateProductParams) { p.SalePrice = decimal.Zero }},
		{"sale price 3 decimals", "sale_price", func(p *service.CreateProductParams) { p.SalePrice = dec("6.005") }},
		{"negative initial stock", "initial_stock", func(p *service.CreateProductParams) { p.InitialStock = dec("-1") }},
		{"initial stock 4 decimals", "initial_stock", func(p *service.CreateProductParams) { p.InitialStock = dec("1.0005") }},
		{"negative min stock", "min_stock", func(p *service.CreateProductParams) { p.MinStock = dec("-1") }},
		{"sku with spaces", "sku", func(p *service.CreateProductParams) { p.SKU = "ABA 412" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			params := validProduct(companyA, f.category(t, companyA, "Abarrotes", nil).ID)
			tt.mutate(&params)
			_, err := f.svc.CreateProduct(context.Background(), params)
			assertDetail(t, err, tt.field)
		})
	}
}

func TestListProducts(t *testing.T) {
	f := newFixture()
	lacteos := f.category(t, companyA, "Lácteos", nil)
	abarrotes := f.category(t, companyA, "Abarrotes", nil)
	create := func(name, sku, barcode string, category uuid.UUID) *entity.Product {
		p := validProduct(companyA, category)
		p.Name, p.SKU, p.Barcode = name, sku, barcode
		created, err := f.svc.CreateProduct(context.Background(), p)
		if err != nil {
			t.Fatalf("CreateProduct(%s) error = %v", name, err)
		}
		return created.Product
	}
	leche := create("Leche PIL entera 1L", "LAC-001", "7770000000001", lacteos.ID)
	create("Yogurt frutilla", "LAC-002", "", lacteos.ID)
	create("Galletas María", "ABA-412", "7770000000099", abarrotes.ID)
	f.store.Products[leche.ID].Status = entity.StatusInactive

	list := func(filter repository.ProductFilter, offset, limit int) ([]repository.ProductListItem, int64) {
		items, total, err := f.svc.ListProducts(context.Background(), service.ListProductsParams{CompanyID: companyA, Filter: filter, Offset: offset, Limit: limit})
		if err != nil {
			t.Fatalf("ListProducts() error = %v", err)
		}
		return items, total
	}

	if items, total := list(repository.ProductFilter{}, 0, 2); total != 3 || len(items) != 2 || items[0].Product.Name != "Galletas María" {
		t.Errorf("page 1: total=%d len=%d", total, len(items))
	}
	if items, _ := list(repository.ProductFilter{}, 2, 2); len(items) != 1 {
		t.Errorf("page 2: len=%d, want 1", len(items))
	}
	for search, want := range map[string]int64{"leche": 1, "lac-": 2, "7770000000099": 1, "nada": 0} {
		if _, total := list(repository.ProductFilter{Search: search}, 0, 20); total != want {
			t.Errorf("search %q: total=%d, want %d", search, total, want)
		}
	}
	if _, total := list(repository.ProductFilter{CategoryID: &lacteos.ID}, 0, 20); total != 2 {
		t.Errorf("category filter: total=%d, want 2", total)
	}
	if items, total := list(repository.ProductFilter{Status: ptr(entity.StatusInactive)}, 0, 20); total != 1 || items[0].CategoryName != "Lácteos" {
		t.Errorf("status filter: total=%d", total)
	}
}

func TestProductSummary(t *testing.T) {
	f := newFixture()
	f.product(t, func(p *service.CreateProductParams) {
		p.Name = "A"
		p.SKU = "A"
		p.Cost = dec("5.40")
		p.InitialStock = dec("10")
	})
	f.product(t, func(p *service.CreateProductParams) {
		p.Name = "B"
		p.SKU = "B"
		p.Cost = dec("2.00")
		p.InitialStock = dec("3")
		p.MinStock = dec("5")
	})
	f.product(t, func(p *service.CreateProductParams) { p.Name = "C"; p.SKU = "C"; p.InitialStock = dec("0") })
	inactive := f.product(t, func(p *service.CreateProductParams) { p.Name = "D"; p.SKU = "D"; p.InitialStock = dec("100") })
	f.store.Products[inactive.ID].Status = entity.StatusInactive

	f.clk.Add(91 * 24 * time.Hour)
	f.product(t, func(p *service.CreateProductParams) {
		p.Name = "E"
		p.SKU = "E"
		p.Cost = dec("1.00")
		p.InitialStock = dec("1")
	})

	summary, err := f.svc.ProductSummary(context.Background(), companyA)
	if err != nil {
		t.Fatalf("ProductSummary() error = %v", err)
	}
	if summary.ActiveProducts != 4 || summary.LowStock != 1 || summary.WithoutMovement != 3 {
		t.Errorf("unexpected summary: %+v", summary)
	}
	if !summary.InventoryValue.Equal(dec("61.00")) {
		t.Errorf("inventory value = %v, want 61.00", summary.InventoryValue)
	}
}

func TestValidateDuplicate(t *testing.T) {
	f := newFixture()
	f.product(t, func(p *service.CreateProductParams) { p.Barcode = "7750182000123" })

	tests := []struct {
		name, sku, barcode string
		exists             bool
		field              string
	}{
		{"galletas maría 200g", "", "", true, "name"},
		{"", "aba-412", "", true, "sku"},
		{"", "", "7750182000123", true, "barcode"},
		{"Galletas María 200g", "ABA-412", "", true, "name"},
		{"Nuevo producto", "NEW-1", "", false, ""},
	}
	for _, tt := range tests {
		result, err := f.svc.ValidateDuplicate(context.Background(), companyA, tt.name, tt.sku, tt.barcode)
		if err != nil {
			t.Fatalf("ValidateDuplicate() error = %v", err)
		}
		if result.Exists != tt.exists || result.Field != tt.field {
			t.Errorf("ValidateDuplicate(%q,%q,%q) = %+v, want %v/%q", tt.name, tt.sku, tt.barcode, result, tt.exists, tt.field)
		}
	}

	_, err := f.svc.ValidateDuplicate(context.Background(), companyA, " ", "", "")
	assertAppError(t, err, sharedErrors.CodeBadRequest, 400)

	result, _ := f.svc.ValidateDuplicate(context.Background(), companyB, "Galletas María 200g", "", "")
	if result.Exists {
		t.Errorf("another company must not see duplicates")
	}
}

func TestListCategoriesCountsProductsAndResolvesTax(t *testing.T) {
	f := newFixture()
	created, _ := f.svc.CreateCategory(context.Background(), service.CreateCategoryParams{CompanyID: companyA, Name: "Abarrotes", DefaultTax: ptr(testutil.IVAGeneral)})
	parent := created.Category
	f.category(t, companyA, "Aceites", &parent.ID)
	f.category(t, companyB, "Ajena", nil)
	params := validProduct(companyA, parent.ID)
	if _, err := f.svc.CreateProduct(context.Background(), params); err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}

	items, err := f.svc.ListCategories(context.Background(), companyA)
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d categories, want 2 of company A", len(items))
	}
	if items[0].Category.Name != "Abarrotes" || items[0].ProductsCount != 1 || *items[0].DefaultTaxName != testutil.IVAGeneral {
		t.Errorf("unexpected parent item: %+v", items[0])
	}
	if items[1].ProductsCount != 0 || items[1].DefaultTaxName != nil || *items[1].Category.ParentID != parent.ID {
		t.Errorf("unexpected child item: %+v", items[1])
	}
}
