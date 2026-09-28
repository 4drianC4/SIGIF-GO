package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type Category struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID   *uuid.UUID `json:"company_id" gorm:"type:uuid;index"`
	ParentID    *uuid.UUID `json:"parent_id" gorm:"type:uuid;index"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	Slug        string     `json:"slug" gorm:"type:varchar(255);uniqueIndex:idx_categories_tenant_slug;not null"`
	Description string     `json:"description" gorm:"type:text"`
	ImageURL    string     `json:"image_url" gorm:"type:varchar(500)"`
	SortOrder   int        `json:"sort_order" gorm:"default:0"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	Parent  *Category `json:"-" gorm:"foreignKey:ParentID"`
	Children []*Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

func NewCategory(clock clock.Clock, tenantID uuid.UUID, name, slug string, parentID *uuid.UUID) *Category {
	now := clock.NowUTC()
	return &Category{
		TenantID:  tenantID,
		ParentID:  parentID,
		Name:      name,
		Slug:      slug,
		IsActive:  true,
		SortOrder: 0,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (Category) TableName() string {
	return "categories"
}

type Brand struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID   *uuid.UUID `json:"company_id" gorm:"type:uuid;index"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	Slug        string     `json:"slug" gorm:"type:varchar(255);uniqueIndex:idx_brands_tenant_slug;not null"`
	Description string     `json:"description" gorm:"type:text"`
	LogoURL     string     `json:"logo_url" gorm:"type:varchar(500)"`
	Website     string     `json:"website" gorm:"type:varchar(500)"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func NewBrand(clock clock.Clock, tenantID uuid.UUID, name, slug string) *Brand {
	now := clock.NowUTC()
	return &Brand{
		TenantID:  tenantID,
		Name:      name,
		Slug:      slug,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (Brand) TableName() string {
	return "brands"
}

type Unit struct {
	ID          uuid.UUID          `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID          `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID   *uuid.UUID         `json:"company_id" gorm:"type:uuid;index"`
	Name        string             `json:"name" gorm:"type:varchar(100);not null"`
	Symbol      string             `json:"symbol" gorm:"type:varchar(20);not null"`
	Description string             `json:"description" gorm:"type:text"`
	IsBase      bool               `json:"is_base" gorm:"default:false"`
	BaseUnitID  *uuid.UUID         `json:"base_unit_id" gorm:"type:uuid;index"`
	Conversion  decimal.Decimal    `json:"conversion" gorm:"type:decimal(15,6);default:1"`
	IsActive    bool               `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time         `json:"deleted_at,omitempty" gorm:"index"`

	BaseUnit  *Unit   `json:"-" gorm:"foreignKey:BaseUnitID"`
}

func NewUnit(clock clock.Clock, tenantID uuid.UUID, name, symbol string, isBase bool) *Unit {
	now := clock.NowUTC()
	return &Unit{
		TenantID:   tenantID,
		Name:       name,
		Symbol:     symbol,
		IsBase:     isBase,
		Conversion: decimal.NewFromInt(1),
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (Unit) TableName() string {
	return "units"
}

type Tax struct {
	ID          uuid.UUID          `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID          `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID   *uuid.UUID         `json:"company_id" gorm:"type:uuid;index"`
	Name        string             `json:"name" gorm:"type:varchar(100);not null"`
	Rate        decimal.Decimal    `json:"rate" gorm:"type:decimal(5,4);not null"`
	Type        string             `json:"type" gorm:"type:varchar(20);default:'percentage'"`
	Description string             `json:"description" gorm:"type:text"`
	IsActive    bool               `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time         `json:"deleted_at,omitempty" gorm:"index"`
}

func NewTax(clock clock.Clock, tenantID uuid.UUID, name string, rate decimal.Decimal) *Tax {
	now := clock.NowUTC()
	return &Tax{
		TenantID:  tenantID,
		Name:      name,
		Rate:      rate,
		Type:      "percentage",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (Tax) TableName() string {
	return "taxes"
}