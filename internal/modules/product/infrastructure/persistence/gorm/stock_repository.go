package gorm

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type StockGormRepository struct {
	db *sharedDatabase.Database
}

func NewStockGormRepository(db *sharedDatabase.Database) repository.StockRepository {
	return &StockGormRepository{db: db}
}

func (r *StockGormRepository) GetTotalByProduct(ctx context.Context, companyID, productID uuid.UUID) (decimal.Decimal, error) {
	var result struct{ Total decimal.Decimal }
	err := r.db.GetDB(ctx).Model(&model.StockBalanceModel{}).
		Select("COALESCE(SUM(quantity), 0) AS total").
		Where("company_id = ? AND product_id = ?", companyID, productID).
		Scan(&result).Error
	return result.Total, err
}
