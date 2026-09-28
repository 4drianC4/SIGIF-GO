package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type Company struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	LegalName   string     `json:"legal_name" gorm:"type:varchar(255)"`
	TaxID       string     `json:"tax_id" gorm:"type:varchar(50);uniqueIndex"`
	Email       string     `json:"email" gorm:"type:varchar(255)"`
	Phone       string     `json:"phone" gorm:"type:varchar(50)"`
	Address     string     `json:"address" gorm:"type:text"`
	City        string     `json:"city" gorm:"type:varchar(100)"`
	State       string     `json:"state" gorm:"type:varchar(100)"`
	Country     string     `json:"country" gorm:"type:varchar(100)"`
	PostalCode  string     `json:"postal_code" gorm:"type:varchar(20)"`
	LogoURL     string     `json:"logo_url" gorm:"type:varchar(500)"`
	Settings    CompanySettings `json:"settings" gorm:"type:jsonb"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type CompanySettings struct {
	DefaultCurrency string  `json:"default_currency" gorm:"default:'USD'"`
	TaxRate         float64 `json:"tax_rate" gorm:"default:0"`
	InvoicePrefix   string  `json:"invoice_prefix" gorm:"default:'INV'"`
	EnablePOS       bool    `json:"enable_pos" gorm:"default:true"`
	EnableEcommerce bool    `json:"enable_ecommerce" gorm:"default:false"`
}

func NewCompany(clock clock.Clock, tenantID uuid.UUID, name, legalName, taxID string) *Company {
	now := clock.NowUTC()
	return &Company{
		TenantID:  tenantID,
		Name:      name,
		LegalName: legalName,
		TaxID:     taxID,
		IsActive:  true,
		Settings: CompanySettings{
			DefaultCurrency: "USD",
			TaxRate:         0,
			InvoicePrefix:   "INV",
			EnablePOS:       true,
			EnableEcommerce: false,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Company) Update(name, legalName, taxID, email, phone, address, city, state, country, postalCode string, settings CompanySettings, clock clock.Clock) {
	c.Name = name
	c.LegalName = legalName
	c.TaxID = taxID
	c.Email = email
	c.Phone = phone
	c.Address = address
	c.City = city
	c.State = state
	c.Country = country
	c.PostalCode = postalCode
	c.Settings = settings
	c.UpdatedAt = clock.NowUTC()
}

func (c *Company) Activate(clock clock.Clock) {
	c.IsActive = true
	c.UpdatedAt = clock.NowUTC()
}

func (c *Company) Deactivate(clock clock.Clock) {
	c.IsActive = false
	c.UpdatedAt = clock.NowUTC()
}

func (c *Company) SoftDelete(clock clock.Clock) {
	now := clock.NowUTC()
	c.DeletedAt = &now
	c.IsActive = false
	c.UpdatedAt = now
}

func (Company) TableName() string {
	return "companies"
}