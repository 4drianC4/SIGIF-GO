package model

import (
	"time"

	"github.com/google/uuid"
)

type CustomerModel struct {
	ID             uuid.UUID  `gorm:"column:customer_id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_customer_tenant_name,priority:1"`
	LegalName      string     `gorm:"type:varchar(160);not null;index:idx_customer_tenant_name,priority:2"`
	DocumentType   string     `gorm:"type:varchar(20);not null;default:'national_id';check:document_type IN ('national_id','tax_id','passport','other')"`
	DocumentNumber *string    `gorm:"type:varchar(30)"`
	Phone          *string    `gorm:"type:varchar(30)"`
	Email          *string    `gorm:"type:varchar(160)"`
	Address        *string    `gorm:"type:varchar(200)"`
	CreditLimit    string     `gorm:"type:numeric(12,2);not null;default:0"`
	CreditBalance  string     `gorm:"type:numeric(12,2);not null;default:0"`
	PointsAccrued  int        `gorm:"not null;default:0"`
	Status         string     `gorm:"type:varchar(20);not null;default:'active';index;check:status IN ('active','inactive','blocked')"`
	CreatedAt      time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt      *time.Time `gorm:"autoUpdateTime"`
	DeletedAt      *time.Time `gorm:"index"`
}

func (CustomerModel) TableName() string {
	return "customer"
}
