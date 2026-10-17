package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type ProductFilter struct {
	Search     string
	CategoryID *uuid.UUID
	Status     *entity.Status
}

type ProductListItem struct {
	Product      *entity.Product
	CategoryName string
}

type ProductSummary struct {
	ActiveProducts  int64
	InventoryValue  decimal.Decimal
	LowStock        int64
	WithoutMovement int64
}

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Product, error)
	SetStatus(ctx context.Context, companyID, id uuid.UUID, status entity.Status) error
	Delete(ctx context.Context, companyID, id uuid.UUID) error
	ExistsByName(ctx context.Context, companyID uuid.UUID, name string, excludeID uuid.UUID) (bool, error)
	ExistsBySKU(ctx context.Context, companyID uuid.UUID, sku string, excludeID uuid.UUID) (bool, error)
	ExistsByBarcode(ctx context.Context, companyID uuid.UUID, barcode string, excludeID uuid.UUID) (bool, error)
	List(ctx context.Context, companyID uuid.UUID, filter ProductFilter, offset, limit int) ([]ProductListItem, int64, error)
	Summary(ctx context.Context, companyID uuid.UUID, noMovementSince time.Time) (ProductSummary, error)
}
