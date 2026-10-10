package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UnitOfMeasureModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string    `gorm:"type:varchar(50);not null;uniqueIndex"`
	Abbreviation string    `gorm:"type:varchar(10);not null;uniqueIndex"`
	SortOrder    int       `gorm:"not null;default:0"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (UnitOfMeasureModel) TableName() string {
	return "units_of_measure"
}

type TaxModel struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name       string          `gorm:"type:varchar(60);not null;uniqueIndex"`
	Percentage decimal.Decimal `gorm:"type:numeric(5,2);not null"`
	CreatedAt  time.Time       `gorm:"autoCreateTime"`
}

func (TaxModel) TableName() string {
	return "taxes"
}
