package query

import (
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
)

type ListSuppliers struct {
	CompanyID     uuid.UUID
	Filter        repository.ListFilter
	Offset, Limit int
}
type GetSupplier struct {
	CompanyID, ID uuid.UUID
}
type SupplierSummary struct {
	CompanyID uuid.UUID
}
type ListSupplierProducts struct {
	CompanyID, SupplierID uuid.UUID
	Search                string
	Offset, Limit         int
}

type ListProductSuppliers struct {
	CompanyID, ProductID uuid.UUID
	Search               string
	Offset, Limit        int
}
type ProductOptions struct {
	CompanyID, SupplierID uuid.UUID
	Filter                repository.ProductOptionFilter
	Offset, Limit         int
}
type SupplierOptions ListProductSuppliers
