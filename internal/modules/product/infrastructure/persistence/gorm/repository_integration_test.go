//go:build integration

package gorm_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	gormlib "gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	productGorm "github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/seed"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type repos struct {
	db         *sharedDatabase.Database
	categories repository.CategoryRepository
	products   repository.ProductRepository
	units      repository.UnitOfMeasureRepository
	taxes      repository.TaxRepository
}

func openDatabase(t *testing.T) repos {
	t.Helper()
	dsn := os.Getenv("SIGIF_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("SIGIF_TEST_DATABASE_DSN no está definido")
	}
	db, err := gormlib.Open(postgres.Open(dsn), &gormlib.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&model.UnitOfMeasureModel{}, &model.TaxModel{}, &model.CategoryModel{}, &model.ProductModel{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	if err := seed.Seed(context.Background(), db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := seed.Seed(context.Background(), db); err != nil {
		t.Fatalf("seed must be idempotent: %v", err)
	}
	wrapped := &sharedDatabase.Database{DB: db}
	return repos{
		db:         wrapped,
		categories: productGorm.NewCategoryGormRepository(wrapped),
		products:   productGorm.NewProductGormRepository(wrapped),
		units:      productGorm.NewUnitOfMeasureGormRepository(wrapped),
		taxes:      productGorm.NewTaxGormRepository(wrapped),
	}
}

func TestReferenceDataSeed(t *testing.T) {
	r := openDatabase(t)
	ctx := context.Background()

	units, err := r.units.List(ctx)
	if err != nil {
		t.Fatalf("units: %v", err)
	}
	var abbreviations []string
	for _, u := range units {
		abbreviations = append(abbreviations, u.Abbreviation)
	}
	if got := len(units); got != 6 || abbreviations[0] != "UNIT" || abbreviations[5] != "DOZ" {
		t.Errorf("units = %v", abbreviations)
	}

	taxes, err := r.taxes.List(ctx)
	if err != nil || len(taxes) != 3 || taxes[0].Name != "IVA general 13%" || !taxes[0].Percentage.Equal(decimal.NewFromInt(13)) {
		t.Errorf("taxes = %+v, %v", taxes, err)
	}
	if tax, _ := r.taxes.GetByName(ctx, "iva GENERAL 13%"); tax == nil {
		t.Errorf("GetByName must ignore case")
	}
}

func TestCatalogRepositories(t *testing.T) {
	r := openDatabase(t)
	ctx := context.Background()
	clk := clock.NewMockClock(time.Now().UTC())
	companyID := uuid.New()

	units, _ := r.units.List(ctx)
	unitID := units[0].ID
	if ok, _ := r.units.Exists(ctx, unitID); !ok {
		t.Fatalf("unit must exist")
	}
	tax, _ := r.taxes.GetByName(ctx, "IVA general 13%")

	newCategory := func(name string, parentID *uuid.UUID, taxID *uuid.UUID) *entity.Category {
		c, err := entity.NewCategory(clk, entity.NewCategoryParams{CompanyID: companyID, Name: name, ParentID: parentID, DefaultTaxID: taxID})
		if err != nil {
			t.Fatalf("NewCategory: %v", err)
		}
		return c
	}
	abarrotes := newCategory("Abarrotes", nil, &tax.ID)
	limpieza := newCategory("Limpieza", nil, nil)
	for _, c := range []*entity.Category{abarrotes, limpieza, newCategory("Aceites", &abarrotes.ID, nil), newCategory("Aceites", &limpieza.ID, nil), newCategory("Aceites", nil, nil)} {
		if err := r.categories.Create(ctx, c); err != nil {
			t.Fatalf("create category %s: %v", c.Name, err)
		}
	}
	if err := r.categories.Create(ctx, newCategory("Abarrotes", nil, nil)); !errors.Is(err, service.ErrCategoryNameTaken) {
		t.Fatalf("duplicate root category: got %v", err)
	}
	if err := r.categories.Create(ctx, newCategory("Aceites", &abarrotes.ID, nil)); !errors.Is(err, service.ErrCategoryNameTaken) {
		t.Fatalf("duplicate child category: got %v", err)
	}
	if ok, _ := r.categories.ExistsByName(ctx, companyID, &abarrotes.ID, "ACEITES"); !ok {
		t.Errorf("ExistsByName child must ignore case")
	}
	if ok, _ := r.categories.ExistsByName(ctx, companyID, nil, "abarrotes"); !ok {
		t.Errorf("ExistsByName root must ignore case")
	}

	newProduct := func(name, sku, barcode string, cost, stock, minStock string) *entity.Product {
		p, err := entity.NewProduct(clk, entity.NewProductParams{
			CompanyID: companyID, CategoryID: abarrotes.ID, UnitOfMeasureID: unitID,
			Name: name, SKU: sku, Barcode: barcode,
			Cost: decimal.RequireFromString(cost), SalePrice: decimal.RequireFromString("10.00"),
			InitialStock: decimal.RequireFromString(stock), MinStock: decimal.RequireFromString(minStock),
		})
		if err != nil {
			t.Fatalf("NewProduct: %v", err)
		}
		return p
	}

	if err := r.products.Create(ctx, newProduct("Galletas María", "ABA-412", "7750000000001", "4.20", "10", "2")); err != nil {
		t.Fatalf("create product: %v", err)
	}
	conflicts := []struct {
		name    string
		product *entity.Product
		want    error
	}{
		{"name", newProduct("Galletas María", "OTRO-1", "", "1", "0", "0"), service.ErrProductNameTaken},
		{"sku", newProduct("Otro 1", "ABA-412", "", "1", "0", "0"), service.ErrSKUTaken},
		{"barcode", newProduct("Otro 2", "", "7750000000001", "1", "0", "0"), service.ErrBarcodeTaken},
	}
	for _, c := range conflicts {
		if err := r.products.Create(ctx, c.product); !errors.Is(err, c.want) {
			t.Errorf("duplicate %s: got %v, want %v", c.name, err, c.want)
		}
	}
	for _, p := range []*entity.Product{
		newProduct("Sin SKU 1", "", "", "2.00", "3", "5"),
		newProduct("Sin SKU 2", "", "", "1.50", "0", "0"),
	} {
		if err := r.products.Create(ctx, p); err != nil {
			t.Fatalf("products without sku/barcode must not collide: %v", err)
		}
	}

	items, total, err := r.products.List(ctx, companyID, repository.ProductFilter{}, 0, 2)
	if err != nil || total != 3 || len(items) != 2 || items[0].Product.Name != "Galletas María" || items[0].CategoryName != "Abarrotes" {
		t.Fatalf("list page 1: total=%d items=%+v err=%v", total, items, err)
	}
	for search, want := range map[string]int64{"galletas": 1, "aba-": 1, "7750000000001": 1, "sin sku": 2, "100%_": 0} {
		if _, total, err := r.products.List(ctx, companyID, repository.ProductFilter{Search: search}, 0, 20); err != nil || total != want {
			t.Errorf("search %q: total=%d err=%v, want %d", search, total, err, want)
		}
	}
	inactive := entity.StatusInactive
	if _, total, _ := r.products.List(ctx, companyID, repository.ProductFilter{Status: &inactive}, 0, 20); total != 0 {
		t.Errorf("status filter: total=%d", total)
	}
	if _, total, _ := r.products.List(ctx, companyID, repository.ProductFilter{CategoryID: &limpieza.ID}, 0, 20); total != 0 {
		t.Errorf("category filter: total=%d", total)
	}

	summary, err := r.products.Summary(ctx, companyID, time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.ActiveProducts != 3 || summary.LowStock != 1 || summary.WithoutMovement != 3 || !summary.InventoryValue.Equal(decimal.RequireFromString("48")) {
		t.Errorf("summary = %+v", summary)
	}
	if summary, _ := r.products.Summary(ctx, companyID, time.Now().UTC().Add(-time.Hour)); summary.WithoutMovement != 0 {
		t.Errorf("recent products must not count as without movement: %+v", summary)
	}

	list, err := r.categories.List(ctx, companyID)
	if err != nil || len(list) != 5 {
		t.Fatalf("categories list: %d %v", len(list), err)
	}
	if list[0].Category.Name != "Abarrotes" || list[0].ProductsCount != 3 || list[0].DefaultTaxName == nil || *list[0].DefaultTaxName != "IVA general 13%" {
		t.Errorf("first category = %+v", list[0])
	}

	if err := r.db.DB.Model(&model.ProductModel{}).Where("company_id = ? AND name = ?", companyID, "Galletas María").
		Update("deleted_at", time.Now()).Error; err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	if err := r.products.Create(ctx, newProduct("Galletas María", "ABA-412", "7750000000001", "4.20", "1", "0")); err != nil {
		t.Errorf("deleted product must free name, sku and barcode: %v", err)
	}
}
