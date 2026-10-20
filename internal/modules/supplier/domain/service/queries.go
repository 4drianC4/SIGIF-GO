package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
)

func (s *SupplierService) Summary(ctx context.Context, companyID uuid.UUID) (repository.SupplierSummary, error) {
	if companyID == uuid.Nil {
		return repository.SupplierSummary{}, ErrCompanyRequired
	}
	return s.repo.Summary(ctx, companyID)
}

func (s *SupplierService) ListProducts(ctx context.Context, companyID, id uuid.UUID, search string, offset, limit int) ([]repository.ProductItem, int64, error) {
	if _, err := s.Get(ctx, companyID, id); err != nil {
		return nil, 0, err
	}
	return s.products.List(ctx, companyID, id, search, offset, limit)
}
