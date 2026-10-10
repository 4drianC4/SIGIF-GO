package entity

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UnitOfMeasure struct {
	ID           uuid.UUID
	Name         string
	Abbreviation string
}

type Tax struct {
	ID         uuid.UUID
	Name       string
	Percentage decimal.Decimal
}
