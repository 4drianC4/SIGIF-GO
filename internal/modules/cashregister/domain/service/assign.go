package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
)

// Assign gives an available cash register of the company to the cashier
// (HU-10-01). A cash register held by another cashier is rejected, and so is a
// cashier who already holds one (HU-10-02). The checks and the insert run in
// one transaction; the unique active indexes reject concurrent requests that
// pass the checks at the same time.
func (s *CashRegisterService) Assign(ctx context.Context, companyID, userID, cashRegisterID uuid.UUID) (*AssignmentView, error) {
	var assignment *entity.Assignment

	err := s.repo.WithinTransaction(ctx, func(ctx context.Context) error {
		register, err := s.repo.GetRegister(ctx, companyID, cashRegisterID)
		if err != nil {
			return err
		}
		if register == nil || !register.IsActive() {
			return ErrCashRegisterNotFound
		}

		holder, err := s.repo.ActiveAssignmentByRegister(ctx, companyID, cashRegisterID)
		if err != nil {
			return err
		}
		if holder != nil && !holder.BelongsTo(userID) {
			return ErrCashRegisterInUse
		}

		current, err := s.repo.ActiveAssignmentByUser(ctx, companyID, userID)
		if err != nil {
			return err
		}
		if current != nil {
			return ErrCashierAlreadyAssigned
		}

		assignment = entity.NewAssignment(s.clock, companyID, cashRegisterID, userID)
		return s.repo.CreateAssignment(ctx, assignment)
	})
	if err != nil {
		return nil, err
	}

	return s.assignmentView(ctx, assignment)
}
