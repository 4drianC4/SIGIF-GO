package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
)

type ListFilter struct {
	Search       string
	Status       *entity.Status
	MainIndustry string
	SortBy       string
	SortOrder    string
}

type SupplierListItem struct {
	Supplier      *entity.Supplier
	ProductsCount int64
}

type SupplierSummary struct{ ActiveSuppliers int64 }

type ProductItem struct {
	Association *entity.ProductSupplier
	Name        string
	SKU         *string
}

type FormOptions struct {
	Industries   []string
	PaymentTerms []string
}

type SupplierRepository interface {
	Create(context.Context, *entity.Supplier) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Supplier, error)
	Update(context.Context, *entity.Supplier) error
	Deactivate(ctx context.Context, companyID, id uuid.UUID, at time.Time) error
	ExistsByTaxID(ctx context.Context, companyID uuid.UUID, taxID string, excludeID uuid.UUID) (bool, error)
	List(ctx context.Context, companyID uuid.UUID, filter ListFilter, offset, limit int) ([]SupplierListItem, int64, error)
	Summary(ctx context.Context, companyID uuid.UUID) (SupplierSummary, error)
	FormOptions(ctx context.Context, companyID uuid.UUID) (FormOptions, error)
	WithinTransaction(ctx context.Context, fn func(context.Context) error) error
}

// ProductSupplierRepository exposes only the associations shown in supplier detail.
type ProductSupplierRepository interface {
	List(ctx context.Context, companyID, supplierID uuid.UUID, search string, offset, limit int) ([]ProductItem, int64, error)
}
