package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// ChangeStatus sets the customer's status (HU-08-05). Only status and
// updated_at are written; the customer is never deleted. Changing to the
// status it already has is idempotent: the customer is returned without
// writing to the database.
func (s *CustomerService) ChangeStatus(
	ctx context.Context,
	companyID uuid.UUID,
	id uuid.UUID,
	newStatus entity.CustomerStatus,
) (*entity.Customer, error) {
	if !newStatus.IsValid() {
		return nil, sharedErrors.New(sharedErrors.CodeValidation,
			"invalid status value; accepted values: active, inactive, blocked", 400)
	}

	customer, err := s.repo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, ErrCustomerNotFound
	}

	if customer.Status == newStatus {
		return customer, nil
	}

	switch newStatus {
	case entity.CustomerStatusActive:
		customer.Activate(s.clock)
	case entity.CustomerStatusInactive:
		customer.Deactivate(s.clock)
	case entity.CustomerStatusBlocked:
		customer.Block(s.clock)
	}

	if err := s.repo.UpdateStatus(ctx, companyID, id, customer.Status, *customer.UpdatedAt); err != nil {
		return nil, err
	}

	return customer, nil
}
