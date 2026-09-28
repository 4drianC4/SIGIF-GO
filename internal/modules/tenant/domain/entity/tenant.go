package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type BusinessType string

const (
	BusinessTypeMinimarket BusinessType = "minimarket"
	BusinessTypeHardware   BusinessType = "hardware"
	BusinessTypePharmacy   BusinessType = "pharmacy"
)

type Tenant struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string       `json:"name" gorm:"type:varchar(255);not null"`
	Slug        string       `json:"slug" gorm:"type:varchar(100);uniqueIndex;not null"`
	BusinessType BusinessType `json:"business_type" gorm:"type:varchar(50);not null"`
	IsActive    bool         `json:"is_active" gorm:"default:true"`
	Settings    TenantSettings `json:"settings" gorm:"type:jsonb"`
	CreatedAt   time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty" gorm:"index"`
}

type TenantSettings struct {
	Currency          string `json:"currency" gorm:"default:'USD'"`
	Timezone          string `json:"timezone" gorm:"default:'UTC'"`
	Language          string `json:"language" gorm:"default:'es'"`
	TaxRate           float64 `json:"tax_rate" gorm:"default:0"`
	RequirePrescription bool  `json:"require_prescription" gorm:"default:false"`
	EnableBatchTracking bool   `json:"enable_batch_tracking" gorm:"default:false"`
	EnableSerialTracking bool  `json:"enable_serial_tracking" gorm:"default:false"`
}

func NewTenant(clock clock.Clock, name, slug string, businessType BusinessType) *Tenant {
	now := clock.NowUTC()
	return &Tenant{
		ID:           uuid.New(),
		Name:         name,
		Slug:         slug,
		BusinessType: businessType,
		IsActive:     true,
		Settings: TenantSettings{
			Currency:           "USD",
			Timezone:           "UTC",
			Language:           "es",
			TaxRate:            0,
			RequirePrescription: businessType == BusinessTypePharmacy,
			EnableBatchTracking: businessType == BusinessTypePharmacy,
			EnableSerialTracking: businessType == BusinessTypeHardware,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (t *Tenant) Update(name string, settings TenantSettings, clock clock.Clock) {
	t.Name = name
	t.Settings = settings
	t.UpdatedAt = clock.NowUTC()
}

func (t *Tenant) Activate(clock clock.Clock) {
	t.IsActive = true
	t.UpdatedAt = clock.NowUTC()
}

func (t *Tenant) Deactivate(clock clock.Clock) {
	t.IsActive = false
	t.UpdatedAt = clock.NowUTC()
}

func (t *Tenant) SoftDelete(clock clock.Clock) {
	now := clock.NowUTC()
	t.DeletedAt = &now
	t.IsActive = false
	t.UpdatedAt = now
}

func (t *Tenant) IsDeleted() bool {
	return t.DeletedAt != nil
}

func (Tenant) TableName() string {
	return "tenants"
}