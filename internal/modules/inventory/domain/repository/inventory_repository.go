package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
)

// StockBalanceItem is the quantity of a product in one branch/location.
type StockBalanceItem struct {
	BranchID   int64
	LocationID int64
	Quantity   decimal.Decimal
}

// BranchLocationRepository owns the branches/locations reference data.
type BranchLocationRepository interface {
	ListBranches(ctx context.Context, companyID uuid.UUID) ([]*entity.Branch, error)
	// EnsureDefault returns the ids of a default branch and location of the
	// company, creating them on first use.
	EnsureDefault(ctx context.Context, companyID uuid.UUID) (branchID, locationID int64, err error)
	ExistsLocation(ctx context.Context, companyID uuid.UUID, branchID, locationID int64) (bool, error)
}

// StockBalanceRepository applies signed deltas to the stock_balance cache.
type StockBalanceRepository interface {
	// ApplyDelta adds delta to the balance of the product in the branch/location
	// and returns the resulting quantity. It fails if the result would be
	// negative.
	ApplyDelta(ctx context.Context, companyID, productID uuid.UUID, branchID, locationID int64, initialStock, delta decimal.Decimal) (decimal.Decimal, error)
	// TotalsByBranch returns the balances of a product across its branches.
	TotalsByBranch(ctx context.Context, companyID, productID uuid.UUID) ([]StockBalanceItem, error)
	// TouchLastMovement records the movement time on the product row.
	TouchLastMovement(ctx context.Context, companyID, productID uuid.UUID, at time.Time) error
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// MovementRepository persists the stock movements (audit log).
type MovementRepository interface {
	Create(ctx context.Context, m *entity.Movement) error
	ListByProduct(ctx context.Context, companyID, productID uuid.UUID, offset, limit int) ([]*entity.Movement, int64, error)
}

// ProductReader checks that a product belongs to the company. It is implemented
// by an adapter over the product module.
type ProductReader interface {
	Stock(ctx context.Context, companyID, productID uuid.UUID) (decimal.Decimal, bool, error)
}
