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
	category := f.category(t, companyA, "Bebidas")
	p, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))
	if err != nil {
		t.Fatal(err)
	}
	before := *p
	userID := uuid.New()
	updated, err := f.svc.UpdateProductPrice(context.Background(), service.UpdateProductPriceParams{
		CompanyID: companyA, ProductID: p.ID, UserID: userID, SalePrice: price("6.00"), Reason: " Ajuste por costo ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !updated.SalePrice.Equal(*price("6.00")) || !updated.CostPrice.Equal(before.CostPrice) {
		t.Fatalf("unexpected prices: %+v", updated)
	}
	updated.CostPrice, updated.SalePrice, updated.UpdatedAt = before.CostPrice, before.SalePrice, before.UpdatedAt
	if *updated != before {
		t.Fatalf("price patch changed unrelated fields")
	}
	entries := f.history.Entries[p.ID]
	if len(entries) != 1 || entries[0].UserID != userID || entries[0].Reason != "Ajuste por costo" ||
		!entries[0].PreviousPrice.Equal(before.SalePrice) || !entries[0].NewPrice.Equal(*price("6.00")) ||
		!entries[0].PreviousCost.Equal(before.CostPrice) || !entries[0].NewCost.Equal(before.CostPrice) {
		t.Fatalf("unexpected history: %+v", entries)
	}
}

func TestUpdateProductPriceRejectsInvalidChanges(t *testing.T) {
	tests := []struct {
		name       string
		cost, sale *decimal.Decimal
	}{
		{"empty", nil, nil}, {"below cost", nil, price("3.00")},
		{"cost exceeds current sale", price("6.00"), nil},
		{"negative cost", price("-1"), nil}, {"zero sale", nil, price("0")},
		{"cost precision", price("1.001"), nil}, {"sale precision", nil, price("6.001")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			category := f.category(t, companyA, "Bebidas")
			p, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))
			if err != nil {
				t.Fatal(err)
			}
			_, err = f.svc.UpdateProductPrice(context.Background(), service.UpdateProductPriceParams{
				CompanyID: companyA, ProductID: p.ID, UserID: uuid.New(), CostPrice: tt.cost, SalePrice: tt.sale,
			})
			assertAppError(t, err, sharedErrors.CodeValidation, 400)
			stored, _ := f.products.GetByID(context.Background(), companyA, p.ID)
			if !stored.CostPrice.Equal(p.CostPrice) || !stored.SalePrice.Equal(p.SalePrice) || len(f.history.Entries[p.ID]) != 0 {
				t.Fatal("invalid change was persisted")
			}
		})
	}
}

func TestUpdateProductPriceCostOnlyAndNoOp(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas")
	p, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))
	if err != nil {
		t.Fatal(err)
	}
	params := service.UpdateProductPriceParams{CompanyID: companyA, ProductID: p.ID, UserID: uuid.New(), CostPrice: price("4.00")}
	updated, err := f.svc.UpdateProductPrice(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.CostPrice.Equal(*price("4.00")) || !updated.SalePrice.Equal(p.SalePrice) {
		t.Fatal("unexpected prices")
	}
	if _, err = f.svc.UpdateProductPrice(context.Background(), params); err != nil {
		t.Fatal(err)
	}
	if len(f.history.Entries[p.ID]) != 1 {
		t.Fatal("unchanged prices created duplicate history")
	}
}

func TestUpdateProductPriceHistoryFailureRollsBack(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas")
	p, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))
	if err != nil {
		t.Fatal(err)
	}
	boom := errors.New("history unavailable")
	f.history.RecordError = boom
	_, err = f.svc.UpdateProductPrice(context.Background(), service.UpdateProductPriceParams{
		CompanyID: companyA, ProductID: p.ID, UserID: uuid.New(), SalePrice: price("6.00"),
	})
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v", err)
	}
	stored, _ := f.products.GetByID(context.Background(), companyA, p.ID)
	if !stored.SalePrice.Equal(p.SalePrice) {
		t.Fatal("price was saved without history")
	}
}
