package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

var (
	companyA = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	companyB = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
)

type fixture struct {
	svc        *service.CatalogService
	products   *testutil.MemoryProductRepository
	categories *testutil.MemoryCategoryRepository
}

func newFixture() fixture {
	products := testutil.NewMemoryProductRepository()
	categories := testutil.NewMemoryCategoryRepository()
	clk := clock.NewMockClock(time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC))
	return fixture{
		svc:        service.NewCatalogService(products, categories, clk),
		products:   products,
		categories: categories,
	}
}

func (f fixture) category(t *testing.T, companyID uuid.UUID, name string) *entity.Category {
	t.Helper()
	category, err := f.svc.CreateCategory(context.Background(), service.CreateCategoryParams{CompanyID: companyID, Name: name})
	if err != nil {
		t.Fatalf("CreateCategory(%q) error = %v", name, err)
	}
	return category
}

func validProduct(companyID, categoryID uuid.UUID) service.CreateProductParams {
	return service.CreateProductParams{
		CompanyID:     companyID,
		CategoryID:    categoryID,
		SKU:           " coca-600 ",
		Barcode:       "7750182000123",
		Name:          "  Coca  Cola 600ml ",
		UnitOfMeasure: entity.UnitPiece,
		CostPrice:     decimal.RequireFromString("3.20"),
		SalePrice:     decimal.RequireFromString("5.50"),
	}
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

func TestCreateCategory(t *testing.T) {
	f := newFixture()

	category := f.category(t, companyA, "  Bebidas   gaseosas ")

	if category.Name != "Bebidas gaseosas" {
		t.Errorf("name = %q, want normalized %q", category.Name, "Bebidas gaseosas")
	}
	if category.Status != entity.StatusActive {
		t.Errorf("status = %q, want active", category.Status)
	}
	if len(f.categories.Categories) != 1 {
		t.Errorf("expected 1 stored category, got %d", len(f.categories.Categories))
	}
}

func TestCreateCategoryRejectsDuplicateNameIgnoringCase(t *testing.T) {
	f := newFixture()
	f.category(t, companyA, "Bebidas")

	_, err := f.svc.CreateCategory(context.Background(), service.CreateCategoryParams{CompanyID: companyA, Name: "  BEBIDAS "})

	assertAppError(t, err, sharedErrors.CodeConflict, 409)
}

func TestCreateCategoryAllowsSameNameInAnotherCompany(t *testing.T) {
	f := newFixture()
	f.category(t, companyA, "Bebidas")
	f.category(t, companyB, "Bebidas")
}

func TestCreateCategoryRequiresCompany(t *testing.T) {
	f := newFixture()

	_, err := f.svc.CreateCategory(context.Background(), service.CreateCategoryParams{CompanyID: uuid.Nil, Name: "Bebidas"})

	assertAppError(t, err, sharedErrors.CodeBadRequest, 400)
}

func TestCreateProduct(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas")

	product, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}

	if product.SKU != "COCA-600" {
		t.Errorf("sku = %q, want COCA-600", product.SKU)
	}
	if product.Name != "Coca Cola 600ml" {
		t.Errorf("name = %q, want normalized", product.Name)
	}
	if product.Status != entity.StatusActive {
		t.Errorf("status = %q, want active", product.Status)
	}
	if f.products.Count() != 1 {
		t.Errorf("expected 1 stored product, got %d", f.products.Count())
	}
}

func TestCreateProductRejectsDuplicates(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(p *service.CreateProductParams)
	}{
		{name: "same sku with different case", mutate: func(p *service.CreateProductParams) { p.SKU = "Coca-600"; p.Barcode = "" }},
		{name: "same barcode", mutate: func(p *service.CreateProductParams) { p.SKU = "OTHER-1" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			category := f.category(t, companyA, "Bebidas")
			if _, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID)); err != nil {
				t.Fatalf("first CreateProduct() error = %v", err)
			}

			params := validProduct(companyA, category.ID)
			tt.mutate(&params)
			_, err := f.svc.CreateProduct(context.Background(), params)

			assertAppError(t, err, sharedErrors.CodeConflict, 409)
			if f.products.Count() != 1 {
				t.Errorf("duplicate must not be stored, got %d products", f.products.Count())
			}
		})
	}
}

func TestCreateProductAllowsSameSKUInAnotherCompany(t *testing.T) {
	f := newFixture()
	categoryA := f.category(t, companyA, "Bebidas")
	categoryB := f.category(t, companyB, "Bebidas")

	if _, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, categoryA.ID)); err != nil {
		t.Fatalf("company A CreateProduct() error = %v", err)
	}
	if _, err := f.svc.CreateProduct(context.Background(), validProduct(companyB, categoryB.ID)); err != nil {
		t.Fatalf("company B CreateProduct() error = %v", err)
	}
}

func TestCreateProductCategoryRules(t *testing.T) {
	t.Run("category does not exist", func(t *testing.T) {
		f := newFixture()
		_, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, uuid.New()))
		assertAppError(t, err, sharedErrors.CodeNotFound, 404)
	})

	t.Run("category belongs to another company", func(t *testing.T) {
		f := newFixture()
		foreign := f.category(t, companyB, "Bebidas")
		_, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, foreign.ID))
		assertAppError(t, err, sharedErrors.CodeNotFound, 404)
	})

	t.Run("category is inactive", func(t *testing.T) {
		f := newFixture()
		category := f.category(t, companyA, "Bebidas")
		f.categories.Categories[category.ID].Status = entity.StatusInactive
		_, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))
		assertAppError(t, err, sharedErrors.CodeBadRequest, 400)
	})
}

func TestCreateProductBusinessValidation(t *testing.T) {
	tests := []struct {
		name   string
		field  string
		mutate func(p *service.CreateProductParams)
	}{
		{name: "sale price zero", field: "sale_price", mutate: func(p *service.CreateProductParams) { p.SalePrice = decimal.Zero }},
		{name: "sale price negative", field: "sale_price", mutate: func(p *service.CreateProductParams) { p.SalePrice = decimal.RequireFromString("-1") }},
		{name: "sale price with 3 decimals", field: "sale_price", mutate: func(p *service.CreateProductParams) { p.SalePrice = decimal.RequireFromString("5.555") }},
		{name: "cost price negative", field: "cost_price", mutate: func(p *service.CreateProductParams) { p.CostPrice = decimal.RequireFromString("-0.01") }},
		{name: "sku with spaces inside", field: "sku", mutate: func(p *service.CreateProductParams) { p.SKU = "COCA 600" }},
		{name: "sku starting with symbol", field: "sku", mutate: func(p *service.CreateProductParams) { p.SKU = "-COCA" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			category := f.category(t, companyA, "Bebidas")
			params := validProduct(companyA, category.ID)
			tt.mutate(&params)

			_, err := f.svc.CreateProduct(context.Background(), params)

			appErr := assertAppError(t, err, sharedErrors.CodeValidation, 400)
			details, _ := appErr.Details.(map[string]string)
			if _, ok := details[tt.field]; !ok {
				t.Errorf("expected detail for %q, got %v", tt.field, appErr.Details)
			}
			if f.products.Count() != 0 {
				t.Errorf("invalid product must not be stored")
			}
		})
	}
}

func TestCreateProductAcceptsZeroCostAndTrailingZeros(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas")
	params := validProduct(companyA, category.ID)
	params.CostPrice = decimal.Zero
	params.SalePrice = decimal.RequireFromString("5.500")

	if _, err := f.svc.CreateProduct(context.Background(), params); err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
}

func TestUpdateProduct(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas")
	original, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}

	updated, err := f.svc.UpdateProduct(context.Background(), service.UpdateProductParams{
		CompanyID:     companyA,
		ProductID:     original.ID,
		CategoryID:    category.ID,
		SKU:           "pepsi-500",
		Name:          "  Pepsi  500ml ",
		UnitOfMeasure: entity.UnitPiece,
		CostPrice:     decimal.RequireFromString("2.50"),
		SalePrice:     decimal.RequireFromString("4.00"),
	})
	if err != nil {
		t.Fatalf("UpdateProduct() error = %v", err)
	}

	if updated.SKU != "PEPSI-500" {
		t.Errorf("sku = %q, want PEPSI-500", updated.SKU)
	}
	if updated.Name != "Pepsi 500ml" {
		t.Errorf("name = %q, want normalized", updated.Name)
	}
	if updated.ID != original.ID {
		t.Errorf("id changed after update")
	}
	if updated.Status != entity.StatusActive {
		t.Errorf("status = %q, want active", updated.Status)
	}
	if f.products.Count() != 1 {
		t.Errorf("expected 1 product, got %d", f.products.Count())
	}
}

func TestUpdateProductNotFound(t *testing.T) {
	f := newFixture()
	f.category(t, companyA, "Bebidas")

	_, err := f.svc.UpdateProduct(context.Background(), service.UpdateProductParams{
		CompanyID:     companyA,
		ProductID:     uuid.New(),
		CategoryID:    uuid.New(),
		SKU:           "PEPSI-500",
		Name:          "Pepsi 500ml",
		UnitOfMeasure: entity.UnitPiece,
		SalePrice:     decimal.RequireFromString("4.00"),
	})

	assertAppError(t, err, sharedErrors.CodeNotFound, 404)
}

func TestUpdateProductWrongCompany(t *testing.T) {
	f := newFixture()
	catA := f.category(t, companyA, "Bebidas")
	original, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, catA.ID))
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}

	_, err = f.svc.UpdateProduct(context.Background(), service.UpdateProductParams{
		CompanyID:     companyB,
		ProductID:     original.ID,
		CategoryID:    uuid.New(),
		SKU:           "PEPSI-500",
		Name:          "Pepsi 500ml",
		UnitOfMeasure: entity.UnitPiece,
		SalePrice:     decimal.RequireFromString("4.00"),
	})

	assertAppError(t, err, sharedErrors.CodeNotFound, 404)
}

func TestUpdateProductDuplicateSKU(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas")

	first, _ := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))

	second, _ := f.svc.CreateProduct(context.Background(), service.CreateProductParams{
		CompanyID:     companyA,
		CategoryID:    category.ID,
		SKU:           "PEPSI-500",
		Name:          "Pepsi 500ml",
		UnitOfMeasure: entity.UnitPiece,
		SalePrice:     decimal.RequireFromString("4.00"),
	})

	_, err := f.svc.UpdateProduct(context.Background(), service.UpdateProductParams{
		CompanyID:     companyA,
		ProductID:     second.ID,
		CategoryID:    category.ID,
		SKU:           first.SKU,
		Name:          "Pepsi 500ml",
		UnitOfMeasure: entity.UnitPiece,
		SalePrice:     decimal.RequireFromString("4.00"),
	})

	assertAppError(t, err, sharedErrors.CodeConflict, 409)
}

func TestUpdateProductKeepsSKU(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas")
	original, _ := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))

	_, err := f.svc.UpdateProduct(context.Background(), service.UpdateProductParams{
		CompanyID:     companyA,
		ProductID:     original.ID,
		CategoryID:    category.ID,
		SKU:           original.SKU,
		Name:          "Coca Cola 600ml editada",
		UnitOfMeasure: entity.UnitPiece,
		SalePrice:     decimal.RequireFromString("6.00"),
	})

	if err != nil {
		t.Fatalf("UpdateProduct() with same SKU error = %v", err)
	}
}

func TestUpdateProductBusinessValidation(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas")
	original, _ := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))

	tests := []struct {
		name   string
		field  string
		mutate func(p *service.UpdateProductParams)
	}{
		{name: "sale price zero", field: "sale_price", mutate: func(p *service.UpdateProductParams) { p.SalePrice = decimal.Zero }},
		{name: "sale price negative", field: "sale_price", mutate: func(p *service.UpdateProductParams) { p.SalePrice = decimal.RequireFromString("-1") }},
		{name: "cost price negative", field: "cost_price", mutate: func(p *service.UpdateProductParams) { p.CostPrice = decimal.RequireFromString("-0.01") }},
		{name: "sku with spaces", field: "sku", mutate: func(p *service.UpdateProductParams) { p.SKU = "COCA 600" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := service.UpdateProductParams{
				CompanyID:     companyA,
				ProductID:     original.ID,
				CategoryID:    category.ID,
				SKU:           "COCA-600",
				Name:          "Coca Cola 600ml",
				UnitOfMeasure: entity.UnitPiece,
				SalePrice:     decimal.RequireFromString("5.50"),
			}
			tt.mutate(&params)

			_, err := f.svc.UpdateProduct(context.Background(), params)

			appErr := assertAppError(t, err, sharedErrors.CodeValidation, 400)
			details, _ := appErr.Details.(map[string]string)
			if _, ok := details[tt.field]; !ok {
				t.Errorf("expected detail for %q, got %v", tt.field, appErr.Details)
			}
		})
	}
}
