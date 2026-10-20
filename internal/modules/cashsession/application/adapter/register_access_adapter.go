package adapter

import (
	"context"

	"github.com/google/uuid"

	cashRegisterEntity "github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	cashRegisterService "github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/repository"
)

// registerAccessAdapter adapts the cash register module (RF-10) to the
// RegisterAccess port. Its errors pass through unchanged so the client sees
// the same codes as in HU-10.
type registerAccessAdapter struct {
	registers *cashRegisterService.CashRegisterService
}

func NewRegisterAccessAdapter(registers *cashRegisterService.CashRegisterService) repository.RegisterAccess {
	return &registerAccessAdapter{registers: registers}
}

func (a *registerAccessAdapter) ActiveHold(ctx context.Context, companyID, userID uuid.UUID) (*repository.Hold, error) {
	view, err := a.registers.Current(ctx, companyID, userID)
	if err != nil {
		return nil, err
	}
	return &repository.Hold{AssignmentID: view.Assignment.ID, Register: registerInfo(view.Register)}, nil
}

func (a *registerAccessAdapter) EnsureHolds(ctx context.Context, companyID, userID, cashRegisterID uuid.UUID) error {
	return a.registers.EnsureCashierOwnsRegister(ctx, companyID, userID, cashRegisterID)
}

func (a *registerAccessAdapter) Describe(ctx context.Context, companyID, cashRegisterID uuid.UUID) (*repository.RegisterInfo, error) {
	register, err := a.registers.GetRegister(ctx, companyID, cashRegisterID)
	if err != nil {
		return nil, err
	}
	info := registerInfo(register)
	return &info, nil
}

func registerInfo(r *cashRegisterEntity.CashRegister) repository.RegisterInfo {
	return repository.RegisterInfo{ID: r.ID, Code: r.Code, Name: r.Name, Status: string(r.Status)}
}
