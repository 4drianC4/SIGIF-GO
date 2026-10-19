//go:build integration

package gorm_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	productGorm "github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/gorm"
	userModel "github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

func TestPriceHistoryWithPostgres(t *testing.T) {
	r := openDatabase(t)
	// The history joins the author and their role; those tables belong to the
	// user module and are not part of the catalog migrations.
	if err := r.db.DB.AutoMigrate(&userModel.UserModel{}, &userModel.RoleModel{}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	companyID, userID, roleID := uuid.New(), uuid.New(), uuid.New()
	role := userModel.RoleModel{ID: roleID, CompanyID: &companyID, Name: "price-test-" + roleID.String()}
	user := userModel.UserModel{
		ID: userID, CompanyID: &companyID, RoleID: roleID, FirstName: "Luis", LastName: "Fernández",
		Username: userID.String(), Email: userID.String() + "@example.test", PasswordHash: "test",
	}
	if err := r.db.DB.Create(&role).Error; err != nil {
		t.Fatal(err)
	}
	if err := r.db.DB.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	svc := service.NewCatalogService(r.products, r.categories, r.units, r.taxes,
		productGorm.NewPriceHistoryGormRepository(r.db), productGorm.NewStockGormRepository(r.db), clock.NewRealClock())
	units, err := r.units.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	category, err := svc.CreateCategory(ctx, service.CreateCategoryParams{CompanyID: companyID, Name: "Aceites"})
	if err != nil {
		t.Fatal(err)
	}
	created, err := svc.CreateProduct(ctx, service.CreateProductParams{
		CompanyID: companyID, CategoryID: category.Category.ID, UnitOfMeasureID: units[0].ID, SKU: "ACEITE", Name: "Aceite Fino 900ml",
		Cost: decimal.RequireFromString("13.30"), SalePrice: decimal.RequireFromString("16.50"),
	})
	if err != nil {
		t.Fatal(err)
	}
	p := created.Product

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
		!e.PreviousPrice.Equal(p.SalePrice) || !e.NewPrice.Equal(newPrice) || !e.NewCost.Equal(p.Cost) ||
		!e.MarginPct.Equal(decimal.NewFromInt(21)) || e.Reason != patch.Reason {
		t.Fatalf("incorrect history entry: %+v", e)
	}

	if _, err := svc.UpdateProductPrice(ctx, patch); err != nil {
		t.Fatal(err)
	}
	if _, total, err = svc.GetPriceHistory(ctx, query); err != nil || total != 1 {
		t.Fatalf("unchanged prices duplicated history: %d, %v", total, err)
	}

	// A full edit that changes the price goes through the same history.
	edited, err := svc.UpdateProduct(ctx, service.UpdateProductParams{
		CompanyID: companyID, ProductID: p.ID, CategoryID: category.Category.ID, UnitOfMeasureID: units[0].ID,
		SKU: "ACEITE", Name: "Aceite Fino 900ml", Cost: p.Cost, SalePrice: decimal.RequireFromString("17.50"),
		UserID: userID, Reason: "Edición completa",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !edited.Product.SalePrice.Equal(decimal.RequireFromString("17.50")) {
		t.Fatalf("edit did not change the price: %+v", edited.Product)
	}
	entries, total, err = svc.GetPriceHistory(ctx, query)
	if err != nil || total != 2 || entries[0].Reason != "Edición completa" || !entries[0].PreviousPrice.Equal(newPrice) {
		t.Fatalf("history after the full edit = %+v, %d, %v", entries, total, err)
	}
}
