package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	GetBySKU(ctx context.Context, tenantID uuid.UUID, sku string) (*entity.Product, error)
	GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*entity.Product, error)
	List(ctx context.Context, tenantID uuid.UUID, offset, limit int, filters ProductFilters) ([]*entity.Product, int64, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsBySKU(ctx context.Context, tenantID uuid.UUID, sku string) (bool, error)
	ExistsByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (bool, error)
	AdjustStock(ctx context.Context, id uuid.UUID, quantity decimal.Decimal) error
	SetStock(ctx context.Context, id uuid.UUID, quantity decimal.Decimal) error
	GetLowStock(ctx context.Context, tenantID uuid.UUID) ([]*entity.Product, error)
}

type ProductFilters struct {
	CategoryID  *uuid.UUID
	BrandID     *uuid.UUID
	Status      *entity.ProductStatus
	Type        *entity.ProductType
	Search      string
	LowStock    bool
	MinPrice    *decimal.Decimal
	MaxPrice    *decimal.Decimal
}

type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*entity.Category, error)
	List(ctx context.Context, tenantID uuid.UUID, parentID *uuid.UUID) ([]*entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type BrandRepository interface {
	Create(ctx context.Context, brand *entity.Brand) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Brand, error)
	GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*entity.Brand, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*entity.Brand, error)
	Update(ctx context.Context, brand *entity.Brand) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UnitRepository interface {
	Create(ctx context.Context, unit *entity.Unit) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Unit, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*entity.Unit, error)
	Update(ctx context.Context, unit *entity.Unit) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TaxRepository interface {
	Create(ctx context.Context, tax *entity.Tax) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tax, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*entity.Tax, error)
	Update(ctx context.Context, tax *entity.Tax) error
	Delete(ctx context.Context, id uuid.UUID) error
}