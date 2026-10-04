package dto

import (
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

type Customer struct {
	ID             string                `json:"id"`
	CompanyID      string                `json:"company_id"`
	LegalName      string                `json:"legal_name"`
	DocumentType   entity.DocumentType   `json:"document_type"`
	DocumentNumber *string               `json:"document_number,omitempty"`
	Phone          *string               `json:"phone,omitempty"`
	Email          *string               `json:"email,omitempty"`
	Address        *string               `json:"address,omitempty"`
	CreditLimit    decimal.Decimal       `json:"credit_limit"`
	CreditBalance  decimal.Decimal       `json:"credit_balance"`
	PointsAccrued  int                   `json:"points_accrued"`
	Status         entity.CustomerStatus `json:"status"`
	CreatedAt      string                `json:"created_at"`
	UpdatedAt      *string               `json:"updated_at,omitempty"`
}

func FromEntity(c *entity.Customer) Customer {
	var updatedAt *string
	if c.UpdatedAt != nil {
		s := c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		updatedAt = &s
	}

	return Customer{
		ID:             c.ID.String(),
		CompanyID:      c.CompanyID.String(),
		LegalName:      c.LegalName,
		DocumentType:   c.DocumentType,
		DocumentNumber: c.DocumentNumber,
		Phone:          c.Phone,
		Email:          c.Email,
		Address:        c.Address,
		CreditLimit:    c.CreditLimit,
		CreditBalance:  c.CreditBalance,
		PointsAccrued:  c.PointsAccrued,
		Status:         c.Status,
		CreatedAt:      c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      updatedAt,
	}
}

func FromEntityList(customers []*entity.Customer) []Customer {
	result := make([]Customer, len(customers))
	for i, c := range customers {
		result[i] = FromEntity(c)
	}
	return result
}
