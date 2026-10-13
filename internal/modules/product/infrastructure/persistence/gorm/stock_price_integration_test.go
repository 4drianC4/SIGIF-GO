//go:build integration

package gorm_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	productGorm "github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

func TestStockTotalByProduct(t *testing.T) {
	db := openDatabase(t)
	if err := db.DB.AutoMigrate(&model.StockBalanceModel{}); err != nil {
		t.Fatal(err)
	}
	companyID, otherCompanyID, productID, otherProductID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	rows := []model.StockBalanceModel{
		{CompanyID: companyID, ProductID: productID, BranchID: 1, LocationID: 1, Quantity: decimal.RequireFromString("2.125")},
		{CompanyID: companyID, ProductID: productID, BranchID: 1, LocationID: 2, Quantity: decimal.RequireFromString("0.875")},
		{CompanyID: companyID, ProductID: productID, BranchID: 2, LocationID: 3, Quantity: decimal.NewFromInt(5), ReservedQuantity: decimal.NewFromInt(1)},
		{CompanyID: otherCompanyID, ProductID: productID, BranchID: 3, LocationID: 4, Quantity: decimal.NewFromInt(100)},
		{CompanyID: companyID, ProductID: otherProductID, BranchID: 1, LocationID: 1, Quantity: decimal.NewFromInt(100)},
	}
	if err := db.DB.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.DB.Where("company_id IN ?", []uuid.UUID{companyID, otherCompanyID}).Delete(&model.StockBalanceModel{})
	})
	repo := productGorm.NewStockGormRepository(db)
	total, err := repo.GetTotalByProduct(context.Background(), companyID, productID)
	if err != nil || !total.Equal(decimal.NewFromInt(8)) {
		t.Fatalf("total = %s, %v", total, err)
	}
	total, err = repo.GetTotalByProduct(context.Background(), companyID, uuid.New())
	if err != nil || !total.IsZero() {
		t.Fatalf("empty total = %s, %v", total, err)
	}
}

func TestProductAndPriceHistoryTransactionRollsBack(t *testing.T) {
	db := openDatabase(t)
	if err := db.DB.AutoMigrate(&model.ProductPriceHistoryModel{}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	clk := clock.NewRealClock()
	companyID := uuid.New()
	categories := productGorm.NewCategoryGormRepository(db)
	products := productGorm.NewProductGormRepository(db)
	history := productGorm.NewPriceHistoryGormRepository(db)
	category := entity.NewCategory(clk, companyID, "Bebidas", "")
	if err := categories.Create(ctx, category); err != nil {
		t.Fatal(err)
	}
	p, err := entity.NewProduct(clk, entity.NewProductParams{
		CompanyID: companyID, CategoryID: category.ID, SKU: "PRICE-TX", Name: "Producto",
		UnitOfMeasure: entity.UnitPiece, CostPrice: decimal.NewFromInt(3), SalePrice: decimal.NewFromInt(5),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := products.Create(ctx, p); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.DB.Where("product_id = ?", p.ID).Delete(&model.ProductPriceHistoryModel{})
		db.DB.Unscoped().Where("id = ?", p.ID).Delete(&model.ProductModel{})
		db.DB.Unscoped().Where("id = ?", category.ID).Delete(&model.CategoryModel{})
	})
	err = products.WithinTransaction(ctx, func(txCtx context.Context) error {
		product, err := products.GetByID(txCtx, companyID, p.ID)
		if err != nil {
			return err
		}
		product.SalePrice = decimal.NewFromInt(6)
		if err := products.Update(txCtx, product); err != nil {
			return err
		}
		// Exceeds varchar(160), forcing the history insert to fail in PostgreSQL.
		return history.Record(txCtx, p.ID, uuid.New(), p.CostPrice, p.CostPrice,
			p.SalePrice, product.SalePrice, strings.Repeat("x", 161))
	})
	if err == nil {
		t.Fatal("expected history insertion error")
	}
	stored, err := products.GetByID(ctx, companyID, p.ID)
	if err != nil || !stored.SalePrice.Equal(p.SalePrice) {
		t.Fatalf("rollback failed: %+v, %v", stored, err)
	}
	var count int64
	if err := db.DB.Model(&model.ProductPriceHistoryModel{}).Where("product_id = ?", p.ID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("history count = %d", count)
	}
}
