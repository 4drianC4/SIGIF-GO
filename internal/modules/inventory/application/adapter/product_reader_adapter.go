package adapter

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/inventory/domain/repository"
	productService "github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

// productReaderAdapter adapts the product module's CatalogService to the
// inventory ProductReader port, so inventory never imports product storage.
type productReaderAdapter struct {
	products *productService.CatalogService
}

func NewProductReaderAdapter(products *productService.CatalogService) repository.ProductReader {
	return &productReaderAdapter{products: products}
}

func (a *productReaderAdapter) Exists(ctx context.Context, companyID, productID uuid.UUID) (bool, error) {
	_, err := a.products.GetProduct(ctx, productService.GetProductParams{CompanyID: companyID, ProductID: productID})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, productService.ErrProductNotFound) {
		return false, nil
	}
	return false, err
}

var _ repository.ProductReader = (*productReaderAdapter)(nil)
