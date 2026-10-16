package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/entity"
)

type SessionRepository interface {
	// Create inserts an open session. The unique open indexes turn a
	// concurrent duplicate into ErrCashierHasOpenSession or
	// ErrRegisterHasOpenSession.
	Create(ctx context.Context, session *entity.CashSession) error

	// OpenByUser returns the open session of the cashier, or nil.
	OpenByUser(ctx context.Context, companyID, userID uuid.UUID) (*entity.CashSession, error)

	// OpenByRegister returns the open session of the cash register, or nil.
	OpenByRegister(ctx context.Context, companyID, cashRegisterID uuid.UUID) (*entity.CashSession, error)
}

// RegisterInfo is what a cash session shows of its cash register.
type RegisterInfo struct {
	ID     uuid.UUID
	Code   string
	Name   string
	Status string
}

func (r RegisterInfo) IsActive() bool {
	return r.Status == "active"
}

// Hold is the cash register a cashier holds through an active assignment.
type Hold struct {
	AssignmentID uuid.UUID
	Register     RegisterInfo
}

// RegisterAccess reads the cash register assignments of RF-10. The cash
// register module provides it through an adapter, and its errors are returned
// as they are (NO_ACTIVE_ASSIGNMENT, CASH_REGISTER_IN_USE, ...).
type RegisterAccess interface {
	// ActiveHold returns the cash register the cashier holds.
	ActiveHold(ctx context.Context, companyID, userID uuid.UUID) (*Hold, error)

	// EnsureHolds checks that the cash register is assigned to the cashier.
	EnsureHolds(ctx context.Context, companyID, userID, cashRegisterID uuid.UUID) error

	// Describe returns a cash register of the company in any status.
	Describe(ctx context.Context, companyID, cashRegisterID uuid.UUID) (*RegisterInfo, error)
}
