package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// StockBalanceModel maps the stock_balance cache. The inventory module owns the
// writes; the product module keeps a read-only view of the same table.
type StockBalanceModel struct {
	StockBalanceID   int64           `gorm:"primaryKey;autoIncrement"`
	CompanyID        uuid.UUID       `gorm:"type:uuid;not null"`
	ProductID        uuid.UUID       `gorm:"type:uuid;not null"`
	BranchID         int64           `gorm:"not null"`
	LocationID       int64           `gorm:"not null"`
	Quantity         decimal.Decimal `gorm:"type:numeric(14,3);not null;default:0"`
	ReservedQuantity decimal.Decimal `gorm:"type:numeric(14,3);not null;default:0"`
	UpdatedAt        time.Time       `gorm:"autoUpdateTime"`
}

func (StockBalanceModel) TableName() string { return "stock_balance" }
