package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CashRegisterModel is a physical cash register of a company. Its code is
// unique within the company among non-deleted cash registers.
type CashRegisterModel struct {
	ID        uuid.UUID      `gorm:"column:cash_register_id;type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID uuid.UUID      `gorm:"column:company_id;type:uuid;not null;uniqueIndex:idx_cash_register_company_code,where:deleted_at IS NULL;index:idx_cash_register_company_status,priority:1"`
	Code      string         `gorm:"type:varchar(20);not null;uniqueIndex:idx_cash_register_company_code"`
	Name      string         `gorm:"type:varchar(80);not null"`
	Status    string         `gorm:"type:varchar(20);not null;default:'active';index:idx_cash_register_company_status,priority:2;check:status IN ('active','inactive')"`
	CreatedAt time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt *time.Time     `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CashRegisterModel) TableName() string {
	return "cash_register"
}

// CashRegisterAssignmentModel records which cashier holds a cash register.
// The partial unique indexes on status = 'active' guarantee, even under
// concurrent requests, that a cash register has a single cashier and a
// cashier a single cash register at a time. Released rows are kept as history.
type CashRegisterAssignmentModel struct {
	ID             uuid.UUID         `gorm:"column:assignment_id;type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID      uuid.UUID         `gorm:"column:company_id;type:uuid;not null;index:idx_cash_register_assignment_company_status,priority:1"`
	CashRegisterID uuid.UUID         `gorm:"column:cash_register_id;type:uuid;not null;index;uniqueIndex:idx_cash_register_assignment_active_register,where:status = 'active'"`
	CashRegister   CashRegisterModel `gorm:"foreignKey:CashRegisterID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	UserID         uuid.UUID         `gorm:"column:user_id;type:uuid;not null;index;uniqueIndex:idx_cash_register_assignment_active_user,where:status = 'active'"`
	Status         string            `gorm:"type:varchar(20);not null;default:'active';index:idx_cash_register_assignment_company_status,priority:2;check:status IN ('active','released')"`
	AssignedAt     time.Time         `gorm:"not null"`
	ReleasedAt     *time.Time
	ReleasedBy     *uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt      *time.Time `gorm:"autoUpdateTime"`
}

func (CashRegisterAssignmentModel) TableName() string {
	return "cash_register_assignment"
}
