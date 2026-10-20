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
	lock := db.Exec("SELECT 1 FROM products WHERE id = ? AND company_id = ? FOR UPDATE", productID, companyID)
	if lock.Error != nil {
		return decimal.Zero, lock.Error
	}
	if lock.RowsAffected == 0 {
		return decimal.Zero, service.ErrProductNotFound
	}

	// The product row is locked until the transaction ends, so no other movement
	// of this product can change its balances or insert one in between.
	var result struct{ Quantity decimal.Decimal }
	update := db.Raw(`UPDATE stock_balance SET quantity = quantity + ?, updated_at = CURRENT_TIMESTAMP
		WHERE product_id = ? AND branch_id = ? AND location_id = ? AND quantity + ? >= reserved_quantity
		RETURNING quantity`, delta, productID, branchID, locationID, delta).
		Scan(&result)
	if update.Error != nil {
		return decimal.Zero, update.Error
	}
	if update.RowsAffected > 0 {
		return result.Quantity, nil
	}

	// Nothing was updated: this branch/location has no balance yet, or it has
	// too little (stock that is reserved cannot be taken out either).
	var here int64
	if err := db.Model(&model.StockBalanceModel{}).
		Where("product_id = ? AND branch_id = ? AND location_id = ?", productID, branchID, locationID).
		Count(&here).Error; err != nil {
		return decimal.Zero, err
	}
	if here > 0 {
		return decimal.Zero, service.ErrInsufficientStock
	}

	// The first balance of a product starts from the stock it was registered with.
	var balances int64
	if err := db.Model(&model.StockBalanceModel{}).
		Where("company_id = ? AND product_id = ?", companyID, productID).
		Count(&balances).Error; err != nil {
		return decimal.Zero, err
	}
	quantity := delta
	if balances == 0 {
		quantity = initialStock.Add(delta)
	}
	if quantity.IsNegative() {
		return decimal.Zero, service.ErrInsufficientStock
	}
	if err := db.Exec(`INSERT INTO stock_balance (company_id, product_id, branch_id, location_id, quantity, reserved_quantity, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, CURRENT_TIMESTAMP)`, companyID, productID, branchID, locationID, quantity).Error; err != nil {
		return decimal.Zero, err
	}
	return quantity, nil
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
