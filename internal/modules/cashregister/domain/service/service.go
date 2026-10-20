package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type CashRegisterService struct {
	repo     repository.CashRegisterRepository
	cashiers repository.CashierDirectory
	clock    clock.Clock
}

func NewCashRegisterService(
	repo repository.CashRegisterRepository,
	cashiers repository.CashierDirectory,
	clk clock.Clock,
) *CashRegisterService {
	return &CashRegisterService{
		repo:     repo,
		cashiers: cashiers,
		clock:    clk,
	}
}

// RegisterView is a cash register with the cashier that holds it, if any.
type RegisterView struct {
	Register    *entity.CashRegister
	Assignment  *entity.Assignment
	CashierName string
}

// AssignmentView is an assignment with its cash register and cashier name.
type AssignmentView struct {
	Assignment  *entity.Assignment
	Register    *entity.CashRegister
	CashierName string
}

// assignmentView loads the cash register and the cashier name of an
// assignment.
func (s *CashRegisterService) assignmentView(ctx context.Context, assignment *entity.Assignment) (*AssignmentView, error) {
	register, err := s.repo.GetRegister(ctx, assignment.CompanyID, assignment.CashRegisterID)
	if err != nil {
		return nil, err
	}
	if register == nil {
		return nil, ErrCashRegisterNotFound
	}
	names, err := s.cashiers.Names(ctx, []uuid.UUID{assignment.UserID})
	if err != nil {
		return nil, err
	}
	return &AssignmentView{Assignment: assignment, Register: register, CashierName: names[assignment.UserID]}, nil
}
