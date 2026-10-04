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
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	productGorm "github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

func openDatabase(t *testing.T) *sharedDatabase.Database {
	t.Helper()
	dsn := os.Getenv("SIGIF_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("SIGIF_TEST_DATABASE_DSN no está definido")
	}
	db, err := gormlib.Open(postgres.Open(dsn), &gormlib.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&model.CategoryModel{}, &model.ProductModel{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return &sharedDatabase.Database{DB: db}
}

func TestUniqueIndexesMapToConflict(t *testing.T) {
	db := openDatabase(t)
	ctx := context.Background()
	clk := clock.NewRealClock()
	companyID := uuid.New()

	categories := productGorm.NewCategoryGormRepository(db)
	products := productGorm.NewProductGormRepository(db)

	category := entity.NewCategory(clk, companyID, "Bebidas", "")
	if err := categories.Create(ctx, category); err != nil {
		t.Fatalf("create category: %v", err)
	}
	if err := categories.Create(ctx, entity.NewCategory(clk, companyID, "Bebidas", "")); !errors.Is(err, service.ErrCategoryNameTaken) {
		t.Fatalf("duplicate category: got %v, want ErrCategoryNameTaken", err)
	}

	newProduct := func(sku, barcode string) *entity.Product {
		p, err := entity.NewProduct(clk, entity.NewProductParams{
			CompanyID: companyID, CategoryID: category.ID, SKU: sku, Barcode: barcode, Name: "Producto",
			UnitOfMeasure: entity.UnitPiece, SalePrice: decimal.RequireFromString("1.50"),
		})
		if err != nil {
			t.Fatalf("NewProduct: %v", err)
		}
		return p
	}

	if err := products.Create(ctx, newProduct("SKU-1", "7750000000001")); err != nil {
		t.Fatalf("create product: %v", err)
	}
	if err := products.Create(ctx, newProduct("SKU-1", "")); !errors.Is(err, service.ErrSKUTaken) {
		t.Fatalf("duplicate sku: got %v, want ErrSKUTaken", err)
	}
	if err := products.Create(ctx, newProduct("SKU-2", "7750000000001")); !errors.Is(err, service.ErrBarcodeTaken) {
		t.Fatalf("duplicate barcode: got %v, want ErrBarcodeTaken", err)
	}

	if err := products.Create(ctx, newProduct("SKU-3", "")); err != nil {
		t.Fatalf("product without barcode: %v", err)
	}
	if err := products.Create(ctx, newProduct("SKU-4", "")); err != nil {
		t.Fatalf("second product without barcode: %v", err)
	}

	if err := db.DB.Model(&model.ProductModel{}).Where("company_id = ? AND sku = ?", companyID, "SKU-3").
		Update("deleted_at", time.Now()).Error; err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	if exists, err := products.ExistsBySKU(ctx, companyID, "SKU-3"); err != nil || exists {
		t.Fatalf("ExistsBySKU after delete = %v, %v; want false", exists, err)
	}
	if err := products.Create(ctx, newProduct("SKU-3", "")); err != nil {
		t.Fatalf("reuse sku of deleted product: %v", err)
	}
}
