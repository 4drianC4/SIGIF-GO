//go:build integration

package gorm_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	productGorm "github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	userModel "github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

func TestPriceHistoryWithPostgres(t *testing.T) {
	db := openDatabase(t)
	if err := db.DB.AutoMigrate(&userModel.UserModel{}, &userModel.RoleModel{}, &model.ProductPriceHistoryModel{}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	companyID, userID, roleID := uuid.New(), uuid.New(), uuid.New()
	role := userModel.RoleModel{ID: roleID, CompanyID: &companyID, Name: "price-test-" + roleID.String()}
	user := userModel.UserModel{
		ID: userID, CompanyID: &companyID, RoleID: roleID, FirstName: "Luis", LastName: "Fernández",
		Username: userID.String(), Email: userID.String() + "@example.test", PasswordHash: "test",
	}
	if err := db.DB.Create(&role).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.DB.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	products := productGorm.NewProductGormRepository(db)
	categories := productGorm.NewCategoryGormRepository(db)
	history := productGorm.NewPriceHistoryGormRepository(db)
	stock := productGorm.NewStockGormRepository(db)
	svc := service.NewCatalogService(products, categories, history, stock, clock.NewRealClock())
	category, err := svc.CreateCategory(ctx, service.CreateCategoryParams{CompanyID: companyID, Name: "Aceites"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := svc.CreateProduct(ctx, service.CreateProductParams{
		CompanyID: companyID, CategoryID: category.ID, SKU: "ACEITE", Name: "Aceite Fino 900ml",
		UnitOfMeasure: entity.UnitPiece, CostPrice: decimal.RequireFromString("13.30"), SalePrice: decimal.RequireFromString("16.50"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.DB.Where("product_id = ?", p.ID).Delete(&model.ProductPriceHistoryModel{})
		db.DB.Unscoped().Where("id = ?", p.ID).Delete(&model.ProductModel{})
		db.DB.Unscoped().Where("id = ?", category.ID).Delete(&model.CategoryModel{})
		db.DB.Unscoped().Where("id = ?", userID).Delete(&userModel.UserModel{})
		db.DB.Where("id = ?", roleID).Delete(&userModel.RoleModel{})
	})
	query := service.GetPriceHistoryParams{CompanyID: companyID, ProductID: p.ID, Page: 1, Limit: 20}
	entries, total, err := svc.GetPriceHistory(ctx, query)
	if err != nil || total != 0 || len(entries) != 0 {
		t.Fatalf("empty history = %+v, %d, %v", entries, total, err)
	}
	newPrice := decimal.RequireFromString("16.90")
	patch := service.UpdateProductPriceParams{CompanyID: companyID, ProductID: p.ID, UserID: userID, SalePrice: &newPrice, Reason: "Ajuste por costo"}
	if _, err := svc.UpdateProductPrice(ctx, patch); err != nil {
		t.Fatal(err)
	}
	entries, total, err = svc.GetPriceHistory(ctx, query)
	if err != nil || total != 1 || len(entries) != 1 {
		t.Fatalf("history = %+v, %d, %v", entries, total, err)
	}
	e := entries[0]
	if e.UserFullName != "Luis Fernández" || e.UserRole != role.Name || e.UserID != userID ||
		!e.PreviousPrice.Equal(p.SalePrice) || !e.NewPrice.Equal(newPrice) || !e.NewCost.Equal(p.CostPrice) ||
		!e.MarginPct.Equal(decimal.NewFromInt(21)) || e.Reason != patch.Reason {
		t.Fatalf("incorrect history entry: %+v", e)
	}
	if _, err := svc.UpdateProductPrice(ctx, patch); err != nil {
		t.Fatal(err)
	}
	entries, total, err = svc.GetPriceHistory(ctx, query)
	if err != nil || total != 1 || len(entries) != 1 {
		t.Fatalf("unchanged prices duplicated history: %+v, %d, %v", entries, total, err)
	}
}
