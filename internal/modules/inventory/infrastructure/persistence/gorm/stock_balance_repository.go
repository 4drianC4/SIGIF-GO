package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sigif/sigif-go/internal/modules/inventory/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/service"
	"github.com/sigif/sigif-go/internal/modules/inventory/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
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

func (r *StockBalanceGormRepository) ApplyDelta(ctx context.Context, companyID, productID uuid.UUID, branchID, locationID int64, delta decimal.Decimal) (decimal.Decimal, error) {
	db := r.db.GetDB(ctx)
	var balance model.StockBalanceModel
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("company_id = ? AND product_id = ? AND branch_id = ? AND location_id = ?", companyID, productID, branchID, locationID).
		First(&balance).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if delta.IsNegative() {
			return decimal.Zero, service.ErrInsufficientStock
		}
		balance = model.StockBalanceModel{
			CompanyID: companyID, ProductID: productID, BranchID: branchID, LocationID: locationID,
			Quantity: delta,
		}
		if err := db.Create(&balance).Error; err != nil {
			return decimal.Zero, err
		}
		return balance.Quantity, nil
	}
	if err != nil {
		return decimal.Zero, err
	}

	next := balance.Quantity.Add(delta)
	if next.IsNegative() {
		return decimal.Zero, service.ErrInsufficientStock
	}
	if err := db.Model(&balance).Update("quantity", next).Error; err != nil {
		return decimal.Zero, err
	}
	return next, nil
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
