package gorm

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/service"
	"github.com/sigif/sigif-go/internal/modules/inventory/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
	"gorm.io/gorm"
)

type StockBalanceGormRepository struct {
	db *sharedDatabase.Database
}

func NewStockBalanceGormRepository(db *sharedDatabase.Database) repository.StockBalanceRepository {
	return &StockBalanceGormRepository{db: db}
}

func (r *StockBalanceGormRepository) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(sharedDatabase.WithTx(ctx, tx))
	})
}

func (r *StockBalanceGormRepository) ApplyDelta(ctx context.Context, companyID, productID uuid.UUID, branchID, locationID int64, initialStock, delta decimal.Decimal) (decimal.Decimal, error) {
	db := r.db.GetDB(ctx)
	lock := db.Exec("SELECT 1 FROM products WHERE id = ? FOR UPDATE", productID)
	if lock.Error != nil {
		return decimal.Zero, lock.Error
	}
	if lock.RowsAffected == 0 {
		return decimal.Zero, service.ErrProductNotFound
	}

	var balances int64
	if err := db.Model(&model.StockBalanceModel{}).
		Where("company_id = ? AND product_id = ?", companyID, productID).
		Count(&balances).Error; err != nil {
		return decimal.Zero, err
	}
	seed := decimal.Zero
	if balances == 0 {
		seed = initialStock
	}
	if seed.Add(delta).IsNegative() {
		return decimal.Zero, service.ErrInsufficientStock
	}

	var result struct{ Quantity decimal.Decimal }
	query := db.Raw(`INSERT INTO stock_balance (company_id, product_id, branch_id, location_id, quantity, reserved_quantity, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, CURRENT_TIMESTAMP)
		ON CONFLICT (company_id, product_id, branch_id, location_id)
		DO UPDATE SET quantity = stock_balance.quantity + EXCLUDED.quantity, updated_at = CURRENT_TIMESTAMP
		WHERE stock_balance.quantity + EXCLUDED.quantity >= 0
		RETURNING quantity`, companyID, productID, branchID, locationID, seed.Add(delta)).
		Scan(&result)
	if query.Error != nil {
		return decimal.Zero, query.Error
	}
	if query.RowsAffected == 0 {
		return decimal.Zero, service.ErrInsufficientStock
	}
	return result.Quantity, nil
}

func (r *StockBalanceGormRepository) TotalsByBranch(ctx context.Context, companyID, productID uuid.UUID) ([]repository.StockBalanceItem, error) {
	var rows []model.StockBalanceModel
	if err := r.db.GetDB(ctx).
		Where("company_id = ? AND product_id = ?", companyID, productID).
		Order("branch_id, location_id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]repository.StockBalanceItem, len(rows))
	for i, row := range rows {
		items[i] = repository.StockBalanceItem{BranchID: row.BranchID, LocationID: row.LocationID, Quantity: row.Quantity}
	}
	return items, nil
}

func (r *StockBalanceGormRepository) TouchLastMovement(ctx context.Context, companyID, productID uuid.UUID, at time.Time) error {
	return r.db.GetDB(ctx).Table("products").
		Where("company_id = ? AND id = ?", companyID, productID).
		Update("last_movement_at", at).Error
}
