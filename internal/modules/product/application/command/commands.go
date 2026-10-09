package command

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
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

type UpdateProduct struct {
	CompanyID       uuid.UUID
	ProductID       uuid.UUID
	CategoryID      uuid.UUID
	UnitOfMeasureID uuid.UUID
	SKU             string
	Barcode         string
	Name            string
	Description     string
	Cost            decimal.Decimal
	SalePrice       decimal.Decimal
	MinStock        *decimal.Decimal
}

type SetProductStatus struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
	Status    entity.Status
}

type DeleteProduct struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
}
