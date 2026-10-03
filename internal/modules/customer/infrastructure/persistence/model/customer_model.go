package model

import (
	"time"

	"github.com/google/uuid"
)

// CustomerModel es el modelo de persistencia GORM para la tabla `customer`.
//
// Notas de diseño:
//   - customer_id es bigint autoincremental asignado por la BD; no se genera en Go.
//   - tenant_id es UUID (adaptación respecto al company_id BIGINT del DBML oficial;
//     la FK queda pendiente hasta identificar la tabla propietaria del tenant).
//   - Los ENUMs usan VARCHAR + CHECK constraint para compatibilidad con AutoMigrate.
//   - deleted_at es *time.Time manual (no gorm.DeletedAt) para mantener el control
//     explícito necesario para el índice parcial de unicidad de documentos.
//   - CreditLimit y CreditBalance se almacenan como string para que GORM los trate
//     como NUMERIC en PostgreSQL; el mapper convierte a/desde decimal.Decimal.
type CustomerModel struct {
	ID             int64      `gorm:"column:customer_id;primaryKey;autoIncrement"`
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

// TableName fuerza el nombre físico oficial de la tabla.
func (CustomerModel) TableName() string {
	return "customer"
}
