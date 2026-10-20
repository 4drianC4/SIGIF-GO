package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PriceHistoryEntry struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	UserFullName  string
	UserRole      string
	PreviousCost  decimal.Decimal
	NewCost       decimal.Decimal
	PreviousPrice decimal.Decimal
	NewPrice      decimal.Decimal
	MarginPct     decimal.Decimal
	Reason        string
	CreatedAt     time.Time
}
