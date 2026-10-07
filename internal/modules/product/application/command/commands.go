package command

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type CreateCategory struct {
	CompanyID   uuid.UUID
	Name        string
	Description string
}

type CreateProduct struct {
	CompanyID     uuid.UUID
	CategoryID    uuid.UUID
	SKU           string
	Barcode       string
	Name          string
	Description   string
	UnitOfMeasure entity.UnitOfMeasure
	CostPrice     decimal.Decimal
	SalePrice     decimal.Decimal
}

type UpdateProduct struct {
	CompanyID     uuid.UUID
	ProductID     uuid.UUID
	CategoryID    uuid.UUID
	SKU           string
	Barcode       string
	Name          string
	Description   string
	UnitOfMeasure entity.UnitOfMeasure
	CostPrice     decimal.Decimal
	SalePrice     decimal.Decimal
}

type GetProduct struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
}

type ListProducts struct {
	CompanyID  uuid.UUID
	Name       string
	CategoryID *uuid.UUID
	Status     *entity.Status
	Page       int
	Limit      int
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
