package gorm

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type PriceHistoryGormRepository struct {
	db *sharedDatabase.Database
}

func NewPriceHistoryGormRepository(db *sharedDatabase.Database) repository.PriceHistoryRepository {
	return &PriceHistoryGormRepository{db: db}
}

func (r *PriceHistoryGormRepository) Record(ctx context.Context, productID, userID uuid.UUID, previousCost, newCost, previousPrice, newPrice decimal.Decimal, reason string) error {
	m := &model.ProductPriceHistoryModel{
		ID:            uuid.New(),
		ProductID:     productID,
		UserID:        userID,
		PreviousCost:  previousCost,
		NewCost:       newCost,
		PreviousPrice: previousPrice,
		NewPrice:      newPrice,
		Reason:        reason,
	}
	return r.db.GetDB(ctx).Create(m).Error
}

func (r *PriceHistoryGormRepository) GetByProduct(ctx context.Context, productID uuid.UUID, page, limit int) ([]*entity.PriceHistoryEntry, int64, error) {
	type row struct {
		ID            uuid.UUID
		UserID        uuid.UUID
		FirstName     string
		LastName      string
		RoleName      string
		PreviousCost  decimal.Decimal
		NewCost       decimal.Decimal
		PreviousPrice decimal.Decimal
		NewPrice      decimal.Decimal
		Reason        string
		CreatedAt     time.Time
	}

	var total int64
	if err := r.db.GetDB(ctx).Model(&model.ProductPriceHistoryModel{}).
		Where("product_id = ?", productID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var rows []row
	err := r.db.GetDB(ctx).Raw(`
		SELECT
			pph.id,
			pph.user_id,
			u.first_name,
			u.last_name,
			ro.name AS role_name,
			pph.previous_cost,
			pph.new_cost,
			pph.previous_price,
			pph.new_price,
			pph.reason,
			pph.created_at
		FROM product_price_history pph
		JOIN app_user u ON u.id = pph.user_id
		JOIN role ro ON ro.id = u.role_id
		WHERE pph.product_id = ?
		ORDER BY pph.created_at DESC
		LIMIT ? OFFSET ?
	`, productID, limit, offset).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	entries := make([]*entity.PriceHistoryEntry, 0, len(rows))
	for _, row := range rows {
		margin := decimal.Zero
		if row.NewPrice.IsPositive() {
			margin = row.NewPrice.Sub(row.NewCost).
				Div(row.NewPrice).
				Mul(decimal.NewFromInt(100)).
				Round(0)
		}
		entries = append(entries, &entity.PriceHistoryEntry{
			ID:            row.ID,
			UserID:        row.UserID,
			UserFullName:  row.FirstName + " " + row.LastName,
			UserRole:      row.RoleName,
			PreviousCost:  row.PreviousCost,
			NewCost:       row.NewCost,
			PreviousPrice: row.PreviousPrice,
			NewPrice:      row.NewPrice,
			MarginPct:     margin,
			Reason:        row.Reason,
			CreatedAt:     row.CreatedAt,
		})
	}
	return entries, total, nil
}
