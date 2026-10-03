package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/shared/clock"
)

type Customer struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	LegalName      string
	DocumentType   DocumentType
	DocumentNumber *string
	Phone          *string
	Email          *string
	Address        *string
	CreditLimit    decimal.Decimal
	CreditBalance  decimal.Decimal
	PointsAccrued  int
	Status         CustomerStatus
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
}

func NewCustomer(
	clk clock.Clock,
	tenantID uuid.UUID,
	legalName string,
	documentType DocumentType,
	documentNumber *string,
	phone *string,
	email *string,
) *Customer {
	now := clk.NowUTC()

	if documentType == "" {
		documentType = DocumentTypeNationalID
	}

	documentNumber = trimStringPtr(documentNumber)
	phone = trimStringPtr(phone)
	email = trimStringPtr(email)

	return &Customer{
		ID:             uuid.New(),
		TenantID:       tenantID,
		LegalName:      strings.TrimSpace(legalName),
		DocumentType:   documentType,
		DocumentNumber: documentNumber,
		Phone:          phone,
		Email:          email,
		CreditLimit:    decimal.Zero,
		CreditBalance:  decimal.Zero,
		PointsAccrued:  0,
		Status:         CustomerStatusActive,
		CreatedAt:      now,
	}
}

func (c *Customer) Update(
	clk clock.Clock,
	legalName string,
	documentType DocumentType,
	documentNumber *string,
	phone *string,
	email *string,
	address *string,
) {
	now := clk.NowUTC()
	c.LegalName = strings.TrimSpace(legalName)
	c.DocumentType = documentType
	c.DocumentNumber = trimStringPtr(documentNumber)
	c.Phone = trimStringPtr(phone)
	c.Email = trimStringPtr(email)
	c.Address = trimStringPtr(address)
	c.UpdatedAt = &now
}

func (c *Customer) Activate(clk clock.Clock) {
	now := clk.NowUTC()
	c.Status = CustomerStatusActive
	c.UpdatedAt = &now
}

func (c *Customer) Deactivate(clk clock.Clock) {
	now := clk.NowUTC()
	c.Status = CustomerStatusInactive
	c.UpdatedAt = &now
}

func (c *Customer) Block(clk clock.Clock) {
	now := clk.NowUTC()
	c.Status = CustomerStatusBlocked
	c.UpdatedAt = &now
}

func (c *Customer) SoftDelete(clk clock.Clock) {
	now := clk.NowUTC()
	c.DeletedAt = &now
	c.Status = CustomerStatusInactive
	c.UpdatedAt = &now
}

func (c *Customer) IsDeleted() bool {
	return c.DeletedAt != nil
}

func trimStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
