package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

// EnsureAvailableForOperations returns the customer when it exists, belongs to
// the company and is active. Modules that need an active customer (sales,
// credits, loyalty) call it before registering a new operation.
func (s *CustomerService) EnsureAvailableForOperations(ctx context.Context, companyID, id uuid.UUID) (*entity.Customer, error) {
	customer, err := s.repo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, ErrCustomerNotFound
	}
	if !customer.IsActive() {
		return nil, ErrCustomerNotActive
	}
	return customer, nil
}
