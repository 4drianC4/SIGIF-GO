package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductPriceHistoryModel struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProductID     uuid.UUID       `gorm:"type:uuid;not null;index"`
	UserID        uuid.UUID       `gorm:"type:uuid;not null"`
	PreviousCost  decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	NewCost       decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	PreviousPrice decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	NewPrice      decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	Reason        string          `gorm:"type:varchar(160)"`
	CreatedAt     time.Time       `gorm:"autoCreateTime;index"`
}

func (ProductPriceHistoryModel) TableName() string {
	return "product_price_history"
}
