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

func (s *CatalogService) GetProduct(ctx context.Context, params GetProductParams) (*repository.ProductDetail, error) {
	product, err := s.findProduct(ctx, params.CompanyID, params.ProductID)
	if err != nil {
		return nil, err
	}
	item, err := s.withCategoryName(ctx, product)
	if err != nil {
		return nil, err
	}
	stockTotal, err := s.stock.GetTotalByProduct(ctx, params.CompanyID, product.ID)
	if err != nil {
		return nil, err
	}
	return &repository.ProductDetail{ProductListItem: *item, StockTotal: stockTotal}, nil
}
