package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func (s *CustomerService) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Customer, error) {
	customer, err := s.repo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, sharedErrors.New(sharedErrors.CodeNotFound, "customer not found", 404)
	}
	return customer, nil
}
