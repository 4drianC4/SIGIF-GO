package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

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
	Association           *entity.ProductSupplier
	Name                  string
	SKU                   *string
	EffectiveLeadTimeDays *int
}

type ProductRef struct {
	ID     uuid.UUID
	Name   string
	SKU    *string
	Status string
}
type SupplierItem struct {
	Association           *entity.ProductSupplier
	Supplier              *entity.Supplier
	EffectiveLeadTimeDays *int
}
type ProductOptionFilter struct {
	Search     string
	CategoryID *uuid.UUID
}
type ProductOption struct {
	ProductRef
	Barcode          *string
	CategoryID       uuid.UUID
	CategoryName     string
	CategoryPath     string
	Stock            decimal.Decimal
	UnitAbbreviation string
	IsAssociated     bool
	Selectable       bool
}
type SupplierOption struct {
	Supplier     *entity.Supplier
	IsAssociated bool
	Selectable   bool
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

type ProductSupplierRepository interface {
	List(ctx context.Context, companyID, supplierID uuid.UUID, search string, offset, limit int) ([]ProductItem, int64, error)
	ListSuppliers(ctx context.Context, companyID, productID uuid.UUID, search string, offset, limit int) ([]SupplierItem, int64, error)
	ProductOptions(ctx context.Context, companyID, supplierID uuid.UUID, filter ProductOptionFilter, offset, limit int) ([]ProductOption, int64, error)
	SupplierOptions(ctx context.Context, companyID, productID uuid.UUID, search string, offset, limit int) ([]SupplierOption, int64, error)
	// GetProduct locks the product when called inside WithinTransaction.
	GetProduct(ctx context.Context, companyID, productID uuid.UUID) (*ProductRef, error)
	Get(ctx context.Context, companyID, supplierID, productID uuid.UUID) (*entity.ProductSupplier, error)
	CreateMany(ctx context.Context, associations []*entity.ProductSupplier) error
	Update(ctx context.Context, association *entity.ProductSupplier) error
	ClearPreferred(ctx context.Context, companyID, productID uuid.UUID, at time.Time) error
	Delete(ctx context.Context, companyID, supplierID, productID uuid.UUID) error
}
