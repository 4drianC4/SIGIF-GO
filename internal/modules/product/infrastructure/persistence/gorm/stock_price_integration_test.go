//go:build integration

package gorm_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	productGorm "github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

func TestStockTotalByProduct(t *testing.T) {
	r := openDatabase(t)
	companyID, otherCompanyID, productID, otherProductID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	rows := []model.StockBalanceModel{
		{CompanyID: companyID, ProductID: productID, BranchID: 1, LocationID: 1, Quantity: decimal.RequireFromString("2.125")},
		{CompanyID: companyID, ProductID: productID, BranchID: 1, LocationID: 2, Quantity: decimal.RequireFromString("0.875")},
		{CompanyID: companyID, ProductID: productID, BranchID: 2, LocationID: 3, Quantity: decimal.NewFromInt(5), ReservedQuantity: decimal.NewFromInt(1)},
		{CompanyID: otherCompanyID, ProductID: productID, BranchID: 3, LocationID: 4, Quantity: decimal.NewFromInt(100)},
		{CompanyID: companyID, ProductID: otherProductID, BranchID: 1, LocationID: 1, Quantity: decimal.NewFromInt(100)},
	}
	if err := r.db.DB.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}

	repo := productGorm.NewStockGormRepository(r.db)
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
	r := openDatabase(t)
	ctx := context.Background()
	clk := clock.NewMockClock(time.Now().UTC())
	companyID := uuid.New()
	history := productGorm.NewPriceHistoryGormRepository(r.db)

	units, err := r.units.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	category, err := entity.NewCategory(clk, entity.NewCategoryParams{CompanyID: companyID, Name: "Bebidas"})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.categories.Create(ctx, category); err != nil {
		t.Fatal(err)
	}
	p, err := entity.NewProduct(clk, entity.NewProductParams{
		CompanyID: companyID, CategoryID: category.ID, UnitOfMeasureID: units[0].ID, SKU: "PRICE-TX", Name: "Producto",
		Cost: decimal.NewFromInt(3), SalePrice: decimal.NewFromInt(5),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.products.Create(ctx, p); err != nil {
		t.Fatal(err)
	}

	err = r.products.WithinTransaction(ctx, func(txCtx context.Context) error {
		product, err := r.products.GetByID(txCtx, companyID, p.ID)
		if err != nil {
			return err
		}
		product.SalePrice = decimal.NewFromInt(6)
		if err := r.products.Update(txCtx, product); err != nil {
			return err
		}
		// Exceeds varchar(160), forcing the history insert to fail in PostgreSQL.
		return history.Record(txCtx, p.ID, uuid.New(), p.Cost, p.Cost,
			p.SalePrice, product.SalePrice, strings.Repeat("x", 161))
	})
	if err == nil {
		t.Fatal("expected history insertion error")
	}
	stored, err := r.products.GetByID(ctx, companyID, p.ID)
	if err != nil || !stored.SalePrice.Equal(p.SalePrice) {
		t.Fatalf("rollback failed: %+v, %v", stored, err)
	}
	var count int64
	if err := r.db.DB.Model(&model.ProductPriceHistoryModel{}).Where("product_id = ?", p.ID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("history count = %d", count)
	}
}
