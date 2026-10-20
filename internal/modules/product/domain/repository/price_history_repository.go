package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type PriceHistoryRepository interface {
	Record(ctx context.Context, productID, userID uuid.UUID, previousCost, newCost, previousPrice, newPrice decimal.Decimal, reason string) error
	GetByProduct(ctx context.Context, productID uuid.UUID, page, limit int) ([]*entity.PriceHistoryEntry, int64, error)
}
