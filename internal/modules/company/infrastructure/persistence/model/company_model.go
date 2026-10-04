package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CompanyModel is the persistence model for companies (tenants). Minimal for
// now: it provides the company_id that business modules (product, customer)
// use as their scoping key.
type CompanyModel struct {
	ID        uuid.UUID      `gorm:"column:company_id;type:uuid;primaryKey;default:gen_random_uuid()"`
	LegalName string         `gorm:"type:varchar(160);not null"`
	TradeName string         `gorm:"type:varchar(160);not null"`
	TaxID     string         `gorm:"type:varchar(20);not null;uniqueIndex"`
	Currency  string         `gorm:"type:varchar(3);not null;default:'BOB'"`
	Timezone  string         `gorm:"type:varchar(60);not null;default:'America/La_Paz'"`
	Status    string         `gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CompanyModel) TableName() string {
	return "company"
}
