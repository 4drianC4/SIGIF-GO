package command

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OpenCashSession struct {
	CompanyID      uuid.UUID
	UserID         uuid.UUID
	CashRegisterID *uuid.UUID
	OpeningAmount  decimal.Decimal
	OpeningNotes   string
}
