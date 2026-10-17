package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
)

type GetProductParams struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
}

func (s *CatalogService) GetProduct(ctx context.Context, params GetProductParams) (*repository.ProductListItem, error) {
	product, err := s.findProduct(ctx, params.CompanyID, params.ProductID)
	if err != nil {
		return nil, err
	}
	return s.withCategoryName(ctx, product)
}
