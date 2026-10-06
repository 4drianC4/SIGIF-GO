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
	ExistsByName(ctx context.Context, companyID uuid.UUID, name string) (bool, error)
	ExistsBySKU(ctx context.Context, companyID uuid.UUID, sku string) (bool, error)
	ExistsByBarcode(ctx context.Context, companyID uuid.UUID, barcode string) (bool, error)
	List(ctx context.Context, companyID uuid.UUID, filter ProductFilter, offset, limit int) ([]ProductListItem, int64, error)
	Summary(ctx context.Context, companyID uuid.UUID, noMovementSince time.Time) (ProductSummary, error)
}
