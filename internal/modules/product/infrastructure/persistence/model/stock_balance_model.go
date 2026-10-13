package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CompanyID and ProductID match the UUID identifiers used by the current catalog.
// Branch and location identifiers follow the supplied inventory schema.
type StockBalanceModel struct {
	StockBalanceID   int64           `gorm:"primaryKey;autoIncrement"`
	CompanyID        uuid.UUID       `gorm:"type:uuid;not null;index:idx_stock_balance_company_product"`
	ProductID        uuid.UUID       `gorm:"type:uuid;not null;index:idx_stock_balance_company_product;uniqueIndex:idx_stock_balance_product_branch_location;index:idx_stock_balance_branch_product"`
	BranchID         int64           `gorm:"not null;uniqueIndex:idx_stock_balance_product_branch_location;index:idx_stock_balance_branch_product,priority:1"`
	LocationID       int64           `gorm:"not null;uniqueIndex:idx_stock_balance_product_branch_location"`
	Quantity         decimal.Decimal `gorm:"type:numeric(14,3);not null;default:0;check:chk_stock_balance_quantity,quantity >= 0"`
	ReservedQuantity decimal.Decimal `gorm:"type:numeric(14,3);not null;default:0;check:chk_stock_balance_reserved,reserved_quantity >= 0 AND reserved_quantity <= quantity"`
	UpdatedAt        time.Time       `gorm:"autoUpdateTime;not null;default:CURRENT_TIMESTAMP"`
}

func (StockBalanceModel) TableName() string {
	return "stock_balance"
}
