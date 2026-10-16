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

// ProductDetail is a product with its category name and its total physical stock
// across branches and locations, read from the stock_balance cache.
type ProductDetail struct {
	ProductListItem
	StockTotal decimal.Decimal
}

type ProductSummary struct {
	ActiveProducts  int64
	InventoryValue  decimal.Decimal
	LowStock        int64
	WithoutMovement int64
}

type ProductRepository interface {
	WithinTransaction(ctx context.Context, fn func(context.Context) error) error
	Create(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Product, error)
	SetStatus(ctx context.Context, companyID, id uuid.UUID, status entity.Status) error
	Delete(ctx context.Context, companyID, id uuid.UUID) error
	// FindByName returns the product of the company, other than excludeID,
	// whose name matches ignoring case. Deleted products free their name.
	FindByName(ctx context.Context, companyID uuid.UUID, name string, excludeID uuid.UUID) (*entity.Product, error)
	// FindBySKU and FindByBarcode return the product that holds the value,
	// other than excludeID. Inactive and deleted products keep holding it
	// (HU-06-01); a live product is returned before a deleted one.
	FindBySKU(ctx context.Context, companyID uuid.UUID, sku string, excludeID uuid.UUID) (*entity.Product, error)
	FindByBarcode(ctx context.Context, companyID uuid.UUID, barcode string, excludeID uuid.UUID) (*entity.Product, error)
	List(ctx context.Context, companyID uuid.UUID, filter ProductFilter, offset, limit int) ([]ProductListItem, int64, error)
	Summary(ctx context.Context, companyID uuid.UUID, noMovementSince time.Time) (ProductSummary, error)
}
