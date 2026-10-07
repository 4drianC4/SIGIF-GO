package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type SetProductStatusParams struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
	Status    entity.Status
}

func (s *CatalogService) SetProductStatus(ctx context.Context, params SetProductStatusParams) (*entity.Product, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}

	product, err := s.products.GetByID(ctx, params.CompanyID, params.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
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
	return product, nil
}
