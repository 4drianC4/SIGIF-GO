package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type StockRepository interface {
	GetTotalByProduct(ctx context.Context, companyID, productID uuid.UUID) (decimal.Decimal, error)
}
