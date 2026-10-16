package command

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
)

type CreateSupplier struct {
	CompanyID uuid.UUID
	Fields    entity.Fields
}
type UpdateSupplier struct {
	CompanyID, ID uuid.UUID
	Fields        entity.Fields
}
type DeactivateSupplier struct{ CompanyID, ID uuid.UUID }

type AssociateProduct struct {
	CompanyID, SupplierID, ProductID uuid.UUID
	SupplierCode                     *string
	LeadTimeDays                     *int
	CurrentCost                      *decimal.Decimal
	Preferred                        *bool
}
type UpdateProductAssociation AssociateProduct
type AssociateProducts struct {
	CompanyID, SupplierID uuid.UUID
	ProductIDs            []uuid.UUID
}
type UnlinkProduct struct{ CompanyID, SupplierID, ProductID uuid.UUID }
