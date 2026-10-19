package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func price(value string) *decimal.Decimal {
	d := decimal.RequireFromString(value)
	return &d
}

func TestUpdateProductPricePreservesDetailsAndRecordsHistory(t *testing.T) {
	f := newFixture()
	p := f.product(t, nil)
	before := *p
	userID := uuid.New()
	updated, err := f.svc.UpdateProductPrice(context.Background(), service.UpdateProductPriceParams{
		CompanyID: companyA, ProductID: p.ID, UserID: userID, SalePrice: price("7.00"), Reason: " Ajuste por costo ",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := updated.Product
	if !got.SalePrice.Equal(*price("7.00")) || !got.Cost.Equal(before.Cost) {
		t.Fatalf("unexpected prices: %+v", got)
	}
	got.Cost, got.SalePrice, got.UpdatedAt = before.Cost, before.SalePrice, before.UpdatedAt
	if *got != before {
		t.Fatalf("price patch changed unrelated fields")
	}
	entries := f.store.PriceHistory.Entries[p.ID]
	if len(entries) != 1 || entries[0].UserID != userID || entries[0].Reason != "Ajuste por costo" ||
		!entries[0].PreviousPrice.Equal(before.SalePrice) || !entries[0].NewPrice.Equal(*price("7.00")) ||
		!entries[0].PreviousCost.Equal(before.Cost) || !entries[0].NewCost.Equal(before.Cost) {
		t.Fatalf("unexpected history: %+v", entries)
	}
}

func TestUpdateProductPriceRejectsInvalidChanges(t *testing.T) {
	tests := []struct {
		name       string
		cost, sale *decimal.Decimal
	}{
		{"empty", nil, nil}, {"below cost", nil, price("3.00")},
		{"cost exceeds current sale", price("7.00"), nil},
		{"negative cost", price("-1"), nil}, {"zero cost", price("0"), nil}, {"zero sale", nil, price("0")},
		{"cost precision", price("1.001"), nil}, {"sale precision", nil, price("7.001")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			p := f.product(t, nil)
			_, err := f.svc.UpdateProductPrice(context.Background(), service.UpdateProductPriceParams{
				CompanyID: companyA, ProductID: p.ID, UserID: uuid.New(), Cost: tt.cost, SalePrice: tt.sale,
			})
			assertAppError(t, err, sharedErrors.CodeValidation, 400)
			stored, _ := f.store.ProductRepository().GetByID(context.Background(), companyA, p.ID)
			if !stored.Cost.Equal(p.Cost) || !stored.SalePrice.Equal(p.SalePrice) || len(f.store.PriceHistory.Entries[p.ID]) != 0 {
				t.Fatal("invalid change was persisted")
			}
		})
	}
}

func TestUpdateProductPriceCostOnlyAndNoOp(t *testing.T) {
	f := newFixture()
	p := f.product(t, nil)
	params := service.UpdateProductPriceParams{CompanyID: companyA, ProductID: p.ID, UserID: uuid.New(), Cost: price("4.00")}
	updated, err := f.svc.UpdateProductPrice(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Product.Cost.Equal(*price("4.00")) || !updated.Product.SalePrice.Equal(p.SalePrice) {
		t.Fatal("unexpected prices")
	}
	if _, err = f.svc.UpdateProductPrice(context.Background(), params); err != nil {
		t.Fatal(err)
	}
	if len(f.store.PriceHistory.Entries[p.ID]) != 1 {
		t.Fatal("unchanged prices created duplicate history")
	}
}

func TestUpdateProductPriceHistoryFailureRollsBack(t *testing.T) {
	f := newFixture()
	p := f.product(t, nil)
	boom := errors.New("history unavailable")
	f.store.PriceHistory.RecordError = boom
	_, err := f.svc.UpdateProductPrice(context.Background(), service.UpdateProductPriceParams{
		CompanyID: companyA, ProductID: p.ID, UserID: uuid.New(), SalePrice: price("7.00"),
	})
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v", err)
	}
	stored, _ := f.store.ProductRepository().GetByID(context.Background(), companyA, p.ID)
	if !stored.SalePrice.Equal(p.SalePrice) {
		t.Fatal("price was saved without history")
	}
}

func TestUpdateProductRecordsPriceHistoryOnlyWhenPricesChange(t *testing.T) {
	f := newFixture()
	p := f.product(t, nil)
	userID := uuid.New()
	base := service.UpdateProductParams{
		CompanyID: companyA, ProductID: p.ID, CategoryID: p.CategoryID, UnitOfMeasureID: p.UnitOfMeasureID,
		SKU: p.SKU, Name: p.Name, Cost: p.Cost, SalePrice: p.SalePrice, UserID: userID,
	}

	renamed := base
	renamed.Name = "Galletas renombradas"
	if _, err := f.svc.UpdateProduct(context.Background(), renamed); err != nil {
		t.Fatal(err)
	}
	if n := len(f.store.PriceHistory.Entries[p.ID]); n != 0 {
		t.Fatalf("history entries after a rename = %d, want 0", n)
	}

	repriced := renamed
	repriced.SalePrice = dec("7.50")
	repriced.Reason = "Nueva lista"
	if _, err := f.svc.UpdateProduct(context.Background(), repriced); err != nil {
		t.Fatal(err)
	}
	entries := f.store.PriceHistory.Entries[p.ID]
	if len(entries) != 1 || entries[0].UserID != userID || entries[0].Reason != "Nueva lista" ||
		!entries[0].PreviousPrice.Equal(p.SalePrice) || !entries[0].NewPrice.Equal(dec("7.50")) {
		t.Fatalf("unexpected history: %+v", entries)
	}

	anonymous := repriced
	anonymous.SalePrice = dec("8.00")
	anonymous.UserID = uuid.Nil
	if _, err := f.svc.UpdateProduct(context.Background(), anonymous); err != nil {
		t.Fatal(err)
	}
	if n := len(f.store.PriceHistory.Entries[p.ID]); n != 1 {
		t.Fatalf("a change without author was recorded: %d entries", n)
	}
}

func TestGetPriceHistoryRequiresAnExistingProductOfTheCompany(t *testing.T) {
	f := newFixture()
	p := f.product(t, nil)
	if _, _, err := f.svc.GetPriceHistory(context.Background(), service.GetPriceHistoryParams{CompanyID: companyB, ProductID: p.ID}); !errors.Is(err, service.ErrProductNotFound) {
		t.Fatalf("other company: error = %v, want ErrProductNotFound", err)
	}
	if _, _, err := f.svc.GetPriceHistory(context.Background(), service.GetPriceHistoryParams{CompanyID: companyA, ProductID: p.ID}); err != nil {
		t.Fatalf("own company: %v", err)
	}
}
