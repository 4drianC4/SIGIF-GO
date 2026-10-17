package service

import (
	"context"

	"github.com/google/uuid"
)

// EnsureCashierOwnsRegister checks that the cash register of the company is
// currently assigned to the cashier (HU-10-02). Cash sessions (RF-09) and sales
// (RF-11) call it before operating, so a cashier can never use a cash register
// held by someone else. It returns ErrCashRegisterNotFound when the cash
// register does not exist, ErrCashRegisterInUse when another cashier holds it
// and ErrCashRegisterNotOwned when nobody holds it.
func (s *CashRegisterService) EnsureCashierOwnsRegister(ctx context.Context, companyID, userID, cashRegisterID uuid.UUID) error {
	register, err := s.repo.GetRegister(ctx, companyID, cashRegisterID)
	if err != nil {
		return err
	}
	if register == nil {
		return ErrCashRegisterNotFound
	}

	holder, err := s.repo.ActiveAssignmentByRegister(ctx, companyID, cashRegisterID)
	if err != nil {
		return err
	}
	if holder == nil {
		return ErrCashRegisterNotOwned
	}
	if !holder.BelongsTo(userID) {
		return ErrCashRegisterInUse
	}
	return nil
}
