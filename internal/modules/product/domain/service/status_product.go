package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
)

type SetProductStatusParams struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
	Status    entity.Status
}

func (s *CatalogService) SetProductStatus(ctx context.Context, params SetProductStatusParams) (*repository.ProductListItem, error) {
	product, err := s.findProduct(ctx, params.CompanyID, params.ProductID)
	if err != nil {
		return nil, err
	}

	switch params.Status {
	case entity.StatusActive:
		if product.IsActive() {
			return nil, ErrProductAlreadyActive
		}
		product.Activate(s.clock)
	case entity.StatusInactive:
		if !product.IsActive() {
			return nil, ErrProductAlreadyInactive
		}
		product.Deactivate(s.clock)
	}

	if err := s.products.SetStatus(ctx, params.CompanyID, params.ProductID, params.Status); err != nil {
		return nil, err
	}
	return s.withCategoryName(ctx, product)
}
