package command

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateCategory struct {
	CompanyID    uuid.UUID
	ParentID     *uuid.UUID
	Name         string
	Description  string
	DefaultTax   *string
	TargetMargin *decimal.Decimal
}

type CreateProduct struct {
	CompanyID       uuid.UUID
	CategoryID      uuid.UUID
	UnitOfMeasureID uuid.UUID
	SKU             string
	Barcode         string
	Name            string
	Description     string
	Cost            decimal.Decimal
	SalePrice       decimal.Decimal
	InitialStock    decimal.Decimal
	MinStock        decimal.Decimal
}
