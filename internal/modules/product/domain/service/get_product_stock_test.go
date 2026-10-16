package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/testutil"
)

func TestGetProductReturnsTheTotalStockOfTheCompany(t *testing.T) {
	f := newFixture()
	p := f.product(t, nil)
	f.store.Stock.Balances = []testutil.StockBalance{
		{CompanyID: companyA, ProductID: p.ID, BranchID: 1, LocationID: 1, Quantity: dec("5.500")},
		{CompanyID: companyA, ProductID: p.ID, BranchID: 2, LocationID: 1, Quantity: dec("2.500")},
		{CompanyID: companyB, ProductID: p.ID, BranchID: 1, LocationID: 1, Quantity: dec("99")},
	}

	detail, err := f.svc.GetProduct(context.Background(), service.GetProductParams{CompanyID: companyA, ProductID: p.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !detail.StockTotal.Equal(dec("8")) {
		t.Fatalf("stock total = %s, want 8", detail.StockTotal)
	}
	if detail.Product.ID != p.ID || detail.CategoryName == "" {
		t.Fatalf("detail lost the product or its category name: %+v", detail)
	}
}

func TestGetProductPropagatesStockFailure(t *testing.T) {
	f := newFixture()
	p := f.product(t, nil)
	boom := errors.New("stock query failed")
	f.store.Stock.GetError = boom
	_, err := f.svc.GetProduct(context.Background(), service.GetProductParams{CompanyID: companyA, ProductID: p.ID})
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want %v", err, boom)
	}
}
