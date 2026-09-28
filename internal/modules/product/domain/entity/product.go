package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type ProductType string

const (
	ProductTypeSimple      ProductType = "simple"
	ProductTypeVariable    ProductType = "variable"
	ProductTypeBundle      ProductType = "bundle"
	ProductTypeService     ProductType = "service"
	ProductTypeDigital     ProductType = "digital"
)

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusArchived ProductStatus = "archived"
)

type Product struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID      `json:"company_id" gorm:"type:uuid;index"`
	SKU             string          `json:"sku" gorm:"type:varchar(100);uniqueIndex:idx_products_tenant_sku;not null"`
	Barcode         string          `json:"barcode" gorm:"type:varchar(100);index"`
	Name            string          `json:"name" gorm:"type:varchar(255);not null"`
	Description     string          `json:"description" gorm:"type:text"`
	Type            ProductType     `json:"type" gorm:"type:varchar(50);default:'simple'"`
	Status          ProductStatus   `json:"status" gorm:"type:varchar(20);default:'draft'"`
	CategoryID      *uuid.UUID      `json:"category_id" gorm:"type:uuid;index"`
	BrandID         *uuid.UUID      `json:"brand_id" gorm:"type:uuid;index"`
	UnitID          *uuid.UUID      `json:"unit_id" gorm:"type:uuid;index"`
	TaxID           *uuid.UUID      `json:"tax_id" gorm:"type:uuid;index"`
	CostPrice       decimal.Decimal `json:"cost_price" gorm:"type:decimal(15,4);default:0"`
	SalePrice       decimal.Decimal `json:"sale_price" gorm:"type:decimal(15,4);default:0"`
	WholesalePrice  decimal.Decimal `json:"wholesale_price" gorm:"type:decimal(15,4);default:0"`
	MinPrice        decimal.Decimal `json:"min_price" gorm:"type:decimal(15,4);default:0"`
	MaxPrice        decimal.Decimal `json:"max_price" gorm:"type:decimal(15,4);default:0"`
	Stock           decimal.Decimal `json:"stock" gorm:"type:decimal(15,4);default:0"`
	MinStock        decimal.Decimal `json:"min_stock" gorm:"type:decimal(15,4);default:0"`
	MaxStock        decimal.Decimal `json:"max_stock" gorm:"type:decimal(15,4);default:0"`
	Weight          decimal.Decimal `json:"weight" gorm:"type:decimal(10,4);default:0"`
	Dimensions      string          `json:"dimensions" gorm:"type:varchar(100)"`
	ImageURL        string          `json:"image_url" gorm:"type:varchar(500)"`
	Images          string          `json:"images" gorm:"type:jsonb"`
	Attributes      string          `json:"attributes" gorm:"type:jsonb"`
	Tags            string          `json:"tags" gorm:"type:jsonb"`
	Settings        ProductSettings `json:"settings" gorm:"type:jsonb"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time      `json:"deleted_at,omitempty" gorm:"index"`
}

type ProductSettings struct {
	TrackStock        bool `json:"track_stock" gorm:"default:true"`
	AllowBackorder    bool `json:"allow_backorder" gorm:"default:false"`
	RequireSerial     bool `json:"require_serial" gorm:"default:false"`
	RequireBatch      bool `json:"require_batch" gorm:"default:false"`
	RequireExpiry     bool `json:"require_expiry" gorm:"default:false"`
	IsPrescription    bool `json:"is_prescription" gorm:"default:false"`
	ControlledSubstance bool `json:"controlled_substance" gorm:"default:false"`
}

func NewProduct(clock clock.Clock, tenantID uuid.UUID, sku, name string, productType ProductType) *Product {
	now := clock.NowUTC()
	return &Product{
		TenantID:  tenantID,
		SKU:       sku,
		Name:      name,
		Type:      productType,
		Status:    ProductStatusDraft,
		CostPrice: decimal.Zero,
		SalePrice: decimal.Zero,
		Stock:     decimal.Zero,
		Settings: ProductSettings{
			TrackStock:         true,
			AllowBackorder:     false,
			RequireSerial:      false,
			RequireBatch:       false,
			RequireExpiry:      false,
			IsPrescription:     false,
			ControlledSubstance: false,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (p *Product) Update(sku, name, description string, productType ProductType, categoryID, brandID, unitID, taxID *uuid.UUID, costPrice, salePrice, wholesalePrice, minPrice, maxPrice, minStock, maxStock, weight decimal.Decimal, dimensions, imageURL, images, attributes, tags string, settings ProductSettings, clock clock.Clock) {
	p.SKU = sku
	p.Name = name
	p.Description = description
	p.Type = productType
	p.CategoryID = categoryID
	p.BrandID = brandID
	p.UnitID = unitID
	p.TaxID = taxID
	p.CostPrice = costPrice
	p.SalePrice = salePrice
	p.WholesalePrice = wholesalePrice
	p.MinPrice = minPrice
	p.MaxPrice = maxPrice
	p.MinStock = minStock
	p.MaxStock = maxStock
	p.Weight = weight
	p.Dimensions = dimensions
	p.ImageURL = imageURL
	p.Images = images
	p.Attributes = attributes
	p.Tags = tags
	p.Settings = settings
	p.UpdatedAt = clock.NowUTC()
}

func (p *Product) Activate(clock clock.Clock) {
	p.Status = ProductStatusActive
	p.UpdatedAt = clock.NowUTC()
}

func (p *Product) Deactivate(clock clock.Clock) {
	p.Status = ProductStatusInactive
	p.UpdatedAt = clock.NowUTC()
}

func (p *Product) Archive(clock clock.Clock) {
	p.Status = ProductStatusArchived
	p.UpdatedAt = clock.NowUTC()
}

func (p *Product) AdjustStock(quantity decimal.Decimal, clock clock.Clock) {
	p.Stock = p.Stock.Add(quantity)
	p.UpdatedAt = clock.NowUTC()
}

func (p *Product) SetStock(quantity decimal.Decimal, clock clock.Clock) {
	p.Stock = quantity
	p.UpdatedAt = clock.NowUTC()
}

func (p *Product) IsLowStock() bool {
	return p.Stock.LessThanOrEqual(p.MinStock)
}

func (p *Product) SoftDelete(clock clock.Clock) {
	now := clock.NowUTC()
	p.DeletedAt = &now
	p.Status = ProductStatusArchived
	p.UpdatedAt = now
}

func (Product) TableName() string {
	return "products"
}