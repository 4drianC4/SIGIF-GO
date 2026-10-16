package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	productModel "github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
)

type SupplierModel struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid();uniqueIndex:uq_suppliers_company_id,priority:2"`
	CompanyID    uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:uq_suppliers_company_id,priority:1;uniqueIndex:idx_suppliers_company_tax_id,priority:1,where:deleted_at IS NULL AND tax_id IS NOT NULL;index:idx_suppliers_company_name,priority:1;index:idx_suppliers_company_status,priority:1"`
	TradeName    *string          `gorm:"type:varchar(160)"`
	LegalName    string           `gorm:"type:varchar(160);not null;index:idx_suppliers_company_name,priority:2;check:chk_suppliers_legal_name,btrim(legal_name) <> ''"`
	TaxID        *string          `gorm:"type:varchar(20);uniqueIndex:idx_suppliers_company_tax_id,priority:2"`
	Contact      *string          `gorm:"type:varchar(120)"`
	Phone        *string          `gorm:"type:varchar(30)"`
	Email        *string          `gorm:"type:varchar(160)"`
	Address      *string          `gorm:"type:varchar(200)"`
	MainIndustry *string          `gorm:"type:varchar(80)"`
	PaymentTerms *string          `gorm:"type:varchar(60)"`
	LeadTimeDays *int             `gorm:"type:smallint;check:chk_suppliers_lead_time,lead_time_days >= 0"`
	MinimumOrder *decimal.Decimal `gorm:"type:numeric(12,2);check:chk_suppliers_minimum_order,minimum_order >= 0"`
	Status       string           `gorm:"type:varchar(20);not null;default:'active';index:idx_suppliers_company_status,priority:2;check:chk_suppliers_status,status IN ('active','inactive')"`
	CreatedAt    time.Time        `gorm:"not null;autoCreateTime:false"`
	UpdatedAt    *time.Time       `gorm:"autoUpdateTime:false"`
	DeletedAt    gorm.DeletedAt   `gorm:"index"`
}

func (SupplierModel) TableName() string { return "suppliers" }

type ProductSupplierModel struct {
	ID           uuid.UUID                  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID    uuid.UUID                  `gorm:"type:uuid;not null;uniqueIndex:uq_product_suppliers_pair,priority:1;uniqueIndex:idx_product_suppliers_preferred,priority:1,where:preferred = true"`
	SupplierID   uuid.UUID                  `gorm:"type:uuid;not null;uniqueIndex:uq_product_suppliers_pair,priority:2"`
	Supplier     *SupplierModel             `gorm:"foreignKey:CompanyID,SupplierID;references:CompanyID,ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	ProductID    uuid.UUID                  `gorm:"type:uuid;not null;uniqueIndex:uq_product_suppliers_pair,priority:3;index;uniqueIndex:idx_product_suppliers_preferred,priority:2"`
	Product      *productModel.ProductModel `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	LeadTimeDays *int                       `gorm:"type:smallint;check:chk_product_suppliers_lead_time,lead_time_days >= 0"`
	SupplierCode *string                    `gorm:"type:varchar(60)"`
	CurrentCost  *decimal.Decimal           `gorm:"type:numeric(12,2);check:chk_product_suppliers_cost,current_cost >= 0"`
	Preferred    bool                       `gorm:"not null;default:false"`
	CreatedAt    time.Time                  `gorm:"not null;autoCreateTime:false"`
	UpdatedAt    *time.Time                 `gorm:"autoUpdateTime:false"`
}

func (ProductSupplierModel) TableName() string { return "product_suppliers" }
