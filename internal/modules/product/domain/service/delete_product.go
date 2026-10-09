package service

import (
	"context"

	"github.com/google/uuid"
)

type DeleteProductParams struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
}

func (s *CatalogService) DeleteProduct(ctx context.Context, params DeleteProductParams) error {
	if _, err := s.findProduct(ctx, params.CompanyID, params.ProductID); err != nil {
		return err
	}
	return s.products.Delete(ctx, params.CompanyID, params.ProductID)
}
