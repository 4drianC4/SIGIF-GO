package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// ChangeStatus cambia el estado operativo del cliente.
// Solo acepta los valores oficiales del enum CustomerStatus.
func (s *CustomerService) ChangeStatus(
	ctx context.Context,
	tenantID uuid.UUID,
	id int64,
	newStatus entity.CustomerStatus,
) (*entity.Customer, error) {
	if !newStatus.IsValid() {
		return nil, sharedErrors.New(sharedErrors.CodeValidation,
			"invalid status value; accepted values: active, inactive, blocked", 400)
	}

	customer, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, sharedErrors.New(sharedErrors.CodeNotFound, "customer not found", 404)
	}

	switch newStatus {
	case entity.CustomerStatusActive:
		customer.Activate(s.clock)
	case entity.CustomerStatusInactive:
		customer.Deactivate(s.clock)
	case entity.CustomerStatusBlocked:
		customer.Block(s.clock)
	}

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}
