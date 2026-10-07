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
	if params.CompanyID == uuid.Nil {
		return ErrCompanyRequired
	}

	product, err := s.products.GetByID(ctx, params.CompanyID, params.ProductID)
	if err != nil {
		return err
	}
	if product == nil {
		return ErrProductNotFound
	}

	return s.products.Delete(ctx, params.CompanyID, params.ProductID)
}
