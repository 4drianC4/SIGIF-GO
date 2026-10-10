package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CategoryModel struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID    uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_categories_company_root_name,where:deleted_at IS NULL AND parent_id IS NULL;uniqueIndex:idx_categories_company_parent_name,where:deleted_at IS NULL AND parent_id IS NOT NULL"`
	ParentID     *uuid.UUID       `gorm:"type:uuid;index;uniqueIndex:idx_categories_company_parent_name"`
	Parent       *CategoryModel   `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Name         string           `gorm:"type:varchar(100);not null;uniqueIndex:idx_categories_company_root_name;uniqueIndex:idx_categories_company_parent_name"`
	Description  string           `gorm:"type:varchar(500)"`
	DefaultTaxID *uuid.UUID       `gorm:"type:uuid;index"`
	DefaultTax   *TaxModel        `gorm:"foreignKey:DefaultTaxID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	TargetMargin *decimal.Decimal `gorm:"type:numeric(5,2)"`
	Status       string           `gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedAt    time.Time        `gorm:"autoCreateTime"`
	UpdatedAt    time.Time        `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt   `gorm:"index"`
}

func (CategoryModel) TableName() string {
	return "categories"
}
