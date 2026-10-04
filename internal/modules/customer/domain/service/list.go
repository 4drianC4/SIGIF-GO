package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
)

func (s *CustomerService) List(
	ctx context.Context,
	companyID uuid.UUID,
	filter repository.ListFilter,
	offset, limit int,
) ([]*entity.Customer, int64, error) {
	return s.repo.List(ctx, companyID, filter, offset, limit)
}
