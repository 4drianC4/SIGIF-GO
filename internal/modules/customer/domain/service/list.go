package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
)

// List is read-only: it never modifies customers. Unknown sort values fall
// back to the newest customers first.
func (s *CustomerService) List(
	ctx context.Context,
	companyID uuid.UUID,
	filter repository.ListFilter,
	offset, limit int,
) ([]*entity.Customer, int64, error) {
	if !filter.SortBy.IsValid() {
		filter.SortBy = repository.SortByCreatedAt
	}
	if !filter.SortOrder.IsValid() {
		filter.SortOrder = repository.SortDesc
	}
	return s.repo.List(ctx, companyID, filter, offset, limit)
}
