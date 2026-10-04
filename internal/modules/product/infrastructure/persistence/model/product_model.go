package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductModel struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID     uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_products_company_sku,where:deleted_at IS NULL;uniqueIndex:idx_products_company_barcode,where:deleted_at IS NULL AND barcode IS NOT NULL"`
	CategoryID    uuid.UUID       `gorm:"type:uuid;not null;index"`
	Category      *CategoryModel  `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	SKU           string          `gorm:"column:sku;type:varchar(50);not null;uniqueIndex:idx_products_company_sku"`
	Barcode       *string         `gorm:"type:varchar(14);uniqueIndex:idx_products_company_barcode"`
	Name          string          `gorm:"type:varchar(150);not null"`
	Description   string          `gorm:"type:varchar(1000)"`
	UnitOfMeasure string          `gorm:"type:varchar(10);not null"`
	CostPrice     decimal.Decimal `gorm:"type:numeric(12,2);not null;default:0"`
	SalePrice     decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	Status        string          `gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt  `gorm:"index"`
}

func (ProductModel) TableName() string {
	return "products"
}
