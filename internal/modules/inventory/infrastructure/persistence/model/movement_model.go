package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MovementModel struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  uuid.UUID       `gorm:"type:uuid;not null"`
	ProductID  uuid.UUID       `gorm:"type:uuid;not null"`
	BranchID   int64           `gorm:"not null"`
	LocationID int64           `gorm:"not null"`
	Type       string          `gorm:"type:varchar(20);not null"`
	Quantity   decimal.Decimal `gorm:"type:numeric(14,3);not null"`
	Reason     *string         `gorm:"type:varchar(200)"`
	UserID     *uuid.UUID      `gorm:"type:uuid"`
	CreatedAt  time.Time       `gorm:"not null"`
}

func (MovementModel) TableName() string { return "stock_movements" }
