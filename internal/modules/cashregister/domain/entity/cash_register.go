package entity

import (
	"time"

	"github.com/google/uuid"
)

// CashRegister is a physical cash register (POS terminal) of a company.
type CashRegister struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Code      string
	Name      string
	Status    CashRegisterStatus
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// IsActive reports whether the cash register can be assigned to a cashier.
func (r *CashRegister) IsActive() bool {
	return r.Status == CashRegisterStatusActive
}
