package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	cashRegisterModel "github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/persistence/model"
)

// CashSessionModel is the cash session of a cashier (RF-09, sesion_caja). The
// partial unique indexes on status = 'open' guarantee, even under concurrent
// requests, a single open session per cash register and per cashier. Closed
// sessions are kept as history.
type CashSessionModel struct {
	ID             uuid.UUID                                     `gorm:"column:cash_session_id;type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID      uuid.UUID                                     `gorm:"column:company_id;type:uuid;not null;index:idx_cash_session_company_status,priority:1"`
	CashRegisterID uuid.UUID                                     `gorm:"column:cash_register_id;type:uuid;not null;index;uniqueIndex:idx_cash_session_open_register,where:status = 'open'"`
	CashRegister   cashRegisterModel.CashRegisterModel           `gorm:"foreignKey:CashRegisterID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	AssignmentID   uuid.UUID                                     `gorm:"column:assignment_id;type:uuid;not null;index"`
	Assignment     cashRegisterModel.CashRegisterAssignmentModel `gorm:"foreignKey:AssignmentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	UserID         uuid.UUID                                     `gorm:"column:user_id;type:uuid;not null;index;uniqueIndex:idx_cash_session_open_user,where:status = 'open'"`
	OpeningAmount  decimal.Decimal                               `gorm:"type:numeric(12,2);not null;check:chk_cash_session_opening_amount,opening_amount >= 0"`
	OpeningNotes   *string                                       `gorm:"type:varchar(200)"`
	Status         string                                        `gorm:"type:varchar(20);not null;default:'open';index:idx_cash_session_company_status,priority:2;check:chk_cash_session_status,status IN ('open','closed')"`
	OpenedAt       time.Time                                     `gorm:"not null"`
	ClosedAt       *time.Time
	CreatedAt      time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt      *time.Time `gorm:"autoUpdateTime"`
}

func (CashSessionModel) TableName() string {
	return "cash_session"
}
