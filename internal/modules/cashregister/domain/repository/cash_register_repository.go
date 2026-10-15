package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
)

// ListFilter narrows a cash register listing. A nil Availability means every
// cash register.
type ListFilter struct {
	Availability *entity.Availability
}

// RegisterWithAssignment is a cash register together with its active
// assignment; Assignment is nil when the cash register is available.
type RegisterWithAssignment struct {
	Register   *entity.CashRegister
	Assignment *entity.Assignment
}

type Reader interface {
	// ListRegisters returns the active, non-deleted cash registers of the
	// company ordered by code.
	ListRegisters(ctx context.Context, companyID uuid.UUID, filter ListFilter, offset, limit int) ([]RegisterWithAssignment, int64, error)

	// GetRegister returns a non-deleted cash register of the company in any
	// status, or nil when there is none.
	GetRegister(ctx context.Context, companyID, id uuid.UUID) (*entity.CashRegister, error)

	GetAssignment(ctx context.Context, companyID, id uuid.UUID) (*entity.Assignment, error)

	ActiveAssignmentByRegister(ctx context.Context, companyID, cashRegisterID uuid.UUID) (*entity.Assignment, error)

	ActiveAssignmentByUser(ctx context.Context, companyID, userID uuid.UUID) (*entity.Assignment, error)
}

type Writer interface {
	// CreateAssignment inserts an active assignment. The unique active indexes
	// turn a concurrent duplicate into ErrCashRegisterInUse or
	// ErrCashierAlreadyAssigned.
	CreateAssignment(ctx context.Context, assignment *entity.Assignment) error

	// ReleaseAssignment writes the release of an active assignment. It returns
	// false when the assignment was no longer active.
	ReleaseAssignment(ctx context.Context, assignment *entity.Assignment) (bool, error)
}

type CashRegisterRepository interface {
	Reader
	Writer

	// WithinTransaction runs fn in a single transaction; repository calls made
	// with the context passed to fn take part in it.
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
