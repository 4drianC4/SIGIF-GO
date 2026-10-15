package entity

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type ProductSupplier struct {
	ID           uuid.UUID
	CompanyID    uuid.UUID
	SupplierID   uuid.UUID
	ProductID    uuid.UUID
	LeadTimeDays *int
	CurrentCost  decimal.Decimal
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}
