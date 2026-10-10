//go:build integration

package gorm_test

import (
	"context"
	"errors"
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
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
	"github.com/sigif/sigif-go/internal/shared/database/dbtest"
	"github.com/sigif/sigif-go/internal/shared/database/migrations"
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
	dsn := dbtest.NewSchemaDSN(t)
	if err := migrations.Up(dsn); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	db, err := gormlib.Open(postgres.Open(dsn), &gormlib.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatalf("database pool: %v", err)
	}
	t.Cleanup(func() { _ = pool.Close() })

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

func TestProductLifecycleRepository(t *testing.T) {
	r := openDatabase(t)
	ctx := context.Background()
	clk := clock.NewMockClock(time.Now().UTC())
	companyID, otherCompany := uuid.New(), uuid.New()

	units, _ := r.units.List(ctx)
	category, err := entity.NewCategory(clk, entity.NewCategoryParams{CompanyID: companyID, Name: "Mascotas"})
	if err != nil {
		t.Fatalf("NewCategory: %v", err)
	}
	if err := r.categories.Create(ctx, category); err != nil {
		t.Fatalf("create category: %v", err)
	}
	create := func(name, sku, barcode string) *entity.Product {
		p, err := entity.NewProduct(clk, entity.NewProductParams{
			CompanyID: companyID, CategoryID: category.ID, UnitOfMeasureID: units[0].ID,
			Name: name, SKU: sku, Barcode: barcode,
			Cost: decimal.RequireFromString("230"), SalePrice: decimal.RequireFromString("300"),
			InitialStock: decimal.RequireFromString("5"),
		})
		if err != nil {
			t.Fatalf("NewProduct: %v", err)
		}
		if err := r.products.Create(ctx, p); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		return p
	}
	first := create("Wiskas Gato adulto", "WIA-879", "7750000000011")
	second := create("Dog Chow 1kg", "DOG-1", "")

	got, err := r.products.GetByID(ctx, companyID, first.ID)
	if err != nil || got == nil || got.SKU != "WIA-879" || !got.Stock.Equal(decimal.RequireFromString("5")) || got.UnitOfMeasureID != units[0].ID {
		t.Fatalf("GetByID = %+v, %v", got, err)
	}
	if foreign, _ := r.products.GetByID(ctx, otherCompany, first.ID); foreign != nil {
		t.Errorf("GetByID must be scoped by company")
	}

	exists := []struct {
		name    string
		check   func(uuid.UUID) (bool, error)
		exclude uuid.UUID
		want    bool
	}{
		{"name, no exclusion", func(id uuid.UUID) (bool, error) {
			return r.products.ExistsByName(ctx, companyID, "wiskas GATO adulto", id)
		}, uuid.Nil, true},
		{"name, excluding itself", func(id uuid.UUID) (bool, error) {
			return r.products.ExistsByName(ctx, companyID, "wiskas GATO adulto", id)
		}, first.ID, false},
		{"name, excluding another", func(id uuid.UUID) (bool, error) {
			return r.products.ExistsByName(ctx, companyID, "wiskas GATO adulto", id)
		}, second.ID, true},
		{"sku, excluding itself", func(id uuid.UUID) (bool, error) { return r.products.ExistsBySKU(ctx, companyID, "WIA-879", id) }, first.ID, false},
		{"sku, excluding another", func(id uuid.UUID) (bool, error) { return r.products.ExistsBySKU(ctx, companyID, "WIA-879", id) }, second.ID, true},
		{"barcode, excluding itself", func(id uuid.UUID) (bool, error) {
			return r.products.ExistsByBarcode(ctx, companyID, "7750000000011", id)
		}, first.ID, false},
		{"barcode, excluding another", func(id uuid.UUID) (bool, error) {
			return r.products.ExistsByBarcode(ctx, companyID, "7750000000011", id)
		}, second.ID, true},
	}
	for _, tt := range exists {
		if ok, err := tt.check(tt.exclude); err != nil || ok != tt.want {
			t.Errorf("%s: got %v, %v; want %v", tt.name, ok, err, tt.want)
		}
	}

	clk.Add(time.Hour)
	if err := got.Update(clk, entity.UpdateProductParams{
		CategoryID: category.ID, UnitOfMeasureID: units[3].ID, SKU: "", Barcode: "",
		Name: "Wiskas Gato adulto 1kg", Cost: decimal.RequireFromString("240.50"),
		SalePrice: decimal.RequireFromString("310"), MinStock: decimal.RequireFromString("2"),
	}); err != nil {
		t.Fatalf("entity update: %v", err)
	}
	if err := r.products.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}
	updated, _ := r.products.GetByID(ctx, companyID, first.ID)
	if updated.Name != "Wiskas Gato adulto 1kg" || updated.SKU != "" || updated.Barcode != "" || updated.UnitOfMeasureID != units[3].ID ||
		!updated.Cost.Equal(decimal.RequireFromString("240.50")) || !updated.MinStock.Equal(decimal.RequireFromString("2")) ||
		!updated.Stock.Equal(decimal.RequireFromString("5")) || !updated.CreatedAt.Equal(first.CreatedAt.Truncate(time.Microsecond)) {
		t.Errorf("updated product = %+v", updated)
	}

	conflicts := []struct {
		name   string
		mutate func(*entity.Product)
		want   error
	}{
		{"name", func(p *entity.Product) { p.Name = "Dog Chow 1kg" }, service.ErrProductNameTaken},
		{"sku", func(p *entity.Product) { p.SKU = "DOG-1" }, service.ErrSKUTaken},
	}
	for _, c := range conflicts {
		candidate := *updated
		c.mutate(&candidate)
		if err := r.products.Update(ctx, &candidate); !errors.Is(err, c.want) {
			t.Errorf("update with duplicate %s: got %v, want %v", c.name, err, c.want)
		}
	}

	if err := r.products.SetStatus(ctx, companyID, first.ID, entity.StatusInactive); err != nil {
		t.Fatalf("SetStatus: %v", err)
	}
	if p, _ := r.products.GetByID(ctx, companyID, first.ID); p.Status != entity.StatusInactive {
		t.Errorf("status = %s, want inactive", p.Status)
	}
	if err := r.products.SetStatus(ctx, otherCompany, first.ID, entity.StatusActive); !errors.Is(err, service.ErrProductNotFound) {
		t.Errorf("SetStatus from another company: got %v", err)
	}

	if err := r.products.Delete(ctx, otherCompany, first.ID); !errors.Is(err, service.ErrProductNotFound) {
		t.Errorf("Delete from another company: got %v", err)
	}
	if err := r.products.Delete(ctx, companyID, first.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if p, err := r.products.GetByID(ctx, companyID, first.ID); p != nil || err != nil {
		t.Errorf("deleted product still readable: %+v, %v", p, err)
	}
	if err := r.products.Delete(ctx, companyID, first.ID); !errors.Is(err, service.ErrProductNotFound) {
		t.Errorf("second delete: got %v", err)
	}
	if _, total, _ := r.products.List(ctx, companyID, repository.ProductFilter{}, 0, 20); total != 1 {
		t.Errorf("deleted product must not be listed: total=%d", total)
	}
	create("Wiskas Gato adulto 1kg", "WIA-879", "7750000000011")
}
