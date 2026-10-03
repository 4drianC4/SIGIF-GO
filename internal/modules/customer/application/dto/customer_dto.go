package dto

import (
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

// Customer es el DTO que expone los datos del cliente desde la capa de aplicación.
type Customer struct {
	ID             int64                 `json:"id"`
	TenantID       string                `json:"tenant_id"`
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

// FromEntity convierte la entidad de dominio al DTO plano.
func FromEntity(c *entity.Customer) Customer {
	var updatedAt *string
	if c.UpdatedAt != nil {
		s := c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		updatedAt = &s
	}

	return Customer{
		ID:             c.ID,
		TenantID:       c.TenantID.String(),
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

// FromEntityList convierte un slice de entidades de dominio en un slice de DTOs.
func FromEntityList(customers []*entity.Customer) []Customer {
	result := make([]Customer, len(customers))
	for i, c := range customers {
		result[i] = FromEntity(c)
	}
	return result
}
