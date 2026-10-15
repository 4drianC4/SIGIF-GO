package command

import (
	"github.com/google/uuid"
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
