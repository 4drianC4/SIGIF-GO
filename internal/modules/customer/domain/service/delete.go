package service

import (
	"context"

	"github.com/google/uuid"
)

// Delete soft-deletes a customer (HU-08-06) only when nothing that must be
// kept depends on it. The checks and the delete run in one transaction, so a
// rejected delete changes nothing.
func (s *CustomerService) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	return s.repo.WithinTransaction(ctx, func(ctx context.Context) error {
		customer, err := s.repo.GetByID(ctx, companyID, id)
		if err != nil {
			return err
		}
		if customer == nil {
			return ErrCustomerNotFound
		}

		if customer.HasBalanceOrPoints() {
			return ErrCustomerHasBalance
		}

		hasDependencies, err := s.deps.HasDependencies(ctx, companyID, id)
		if err != nil {
			return err
		}
		if hasDependencies {
			return ErrCustomerHasDependencies
		}

		customer.SoftDelete(s.clock)

		return s.repo.Update(ctx, customer)
	})
}
