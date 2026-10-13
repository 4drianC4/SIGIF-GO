package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func TestGetProductPropagatesStockFailure(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas")
	p, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))
	if err != nil {
		t.Fatal(err)
	}
	boom := errors.New("stock query failed")
	f.stock.GetError = boom
	_, err = f.svc.GetProduct(context.Background(), service.GetProductParams{CompanyID: companyA, ProductID: p.ID})
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want %v", err, boom)
	}
}
