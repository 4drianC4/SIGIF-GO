package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
)

// GetRegister returns a cash register of the company in any status. Modules
// that build on cash registers, such as cash sessions (RF-09), use it to show
// the code and name of a register. It returns ErrCashRegisterNotFound when the
// cash register does not exist or belongs to another company.
func (s *CashRegisterService) GetRegister(ctx context.Context, companyID, cashRegisterID uuid.UUID) (*entity.CashRegister, error) {
	register, err := s.repo.GetRegister(ctx, companyID, cashRegisterID)
	if err != nil {
		return nil, err
	}
	if register == nil {
		return nil, ErrCashRegisterNotFound
	}
	return register, nil
}
