package service

import (
	"context"

	"github.com/google/uuid"

	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func (s *CustomerService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	customer, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if customer == nil {
		return sharedErrors.New(sharedErrors.CodeNotFound, "customer not found", 404)
	}

	customer.SoftDelete(s.clock)

	return s.repo.Update(ctx, customer)
}
