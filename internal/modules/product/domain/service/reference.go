package service

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

func (s *CatalogService) ListUnitsOfMeasure(ctx context.Context) ([]*entity.UnitOfMeasure, error) {
	return s.units.List(ctx)
}

func (s *CatalogService) ListTaxes(ctx context.Context) ([]*entity.Tax, error) {
	return s.taxes.List(ctx)
}
