package query

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type ListProducts struct {
	CompanyID  uuid.UUID
	Search     string
	CategoryID *uuid.UUID
	Status     *entity.Status
	Offset     int
	Limit      int
}

type GetProduct struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
}

type ValidateDuplicate struct {
	CompanyID uuid.UUID
	Name      string
	SKU       string
	Barcode   string
	ExcludeID uuid.UUID
}
