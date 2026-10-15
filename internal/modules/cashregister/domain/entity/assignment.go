package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/clock"
)

// Assignment links a cash register to the cashier that holds it during a
// shift. Releasing it never deletes the row, so the history is kept.
type Assignment struct {
	ID             uuid.UUID
	CompanyID      uuid.UUID
	CashRegisterID uuid.UUID
	UserID         uuid.UUID
	Status         AssignmentStatus
	AssignedAt     time.Time
	ReleasedAt     *time.Time
	ReleasedBy     *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

func NewAssignment(clk clock.Clock, companyID, cashRegisterID, userID uuid.UUID) *Assignment {
	now := clk.NowUTC()
	return &Assignment{
		ID:             uuid.New(),
		CompanyID:      companyID,
		CashRegisterID: cashRegisterID,
		UserID:         userID,
		Status:         AssignmentStatusActive,
		AssignedAt:     now,
		CreatedAt:      now,
	}
}

// IsActive reports whether the cashier still holds the cash register.
func (a *Assignment) IsActive() bool {
	return a.Status == AssignmentStatusActive
}

// BelongsTo reports whether the assignment was made to the given cashier.
func (a *Assignment) BelongsTo(userID uuid.UUID) bool {
	return a.UserID == userID
}

// Release ends the assignment, freeing the cash register for another cashier.
func (a *Assignment) Release(clk clock.Clock, releasedBy uuid.UUID) {
	now := clk.NowUTC()
	a.Status = AssignmentStatusReleased
	a.ReleasedAt = &now
	a.ReleasedBy = &releasedBy
	a.UpdatedAt = &now
}
