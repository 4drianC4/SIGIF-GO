package mapper

import (
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/model"
)

// ToModel convierte la entidad de dominio al modelo de persistencia GORM.
func ToModel(c *entity.Customer) *model.CustomerModel {
	return &model.CustomerModel{
		ID:             c.ID,
		TenantID:       c.TenantID,
		LegalName:      c.LegalName,
		DocumentType:   string(c.DocumentType),
		DocumentNumber: c.DocumentNumber,
		Phone:          c.Phone,
		Email:          c.Email,
		Address:        c.Address,
		CreditLimit:    c.CreditLimit.String(),
		CreditBalance:  c.CreditBalance.String(),
		PointsAccrued:  c.PointsAccrued,
		Status:         string(c.Status),
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
		DeletedAt:      c.DeletedAt,
	}
}

// ToDomain convierte el modelo de persistencia a la entidad de dominio.
func ToDomain(m *model.CustomerModel) *entity.Customer {
	creditLimit, _ := decimal.NewFromString(m.CreditLimit)
	creditBalance, _ := decimal.NewFromString(m.CreditBalance)

	return &entity.Customer{
		ID:             m.ID,
		TenantID:       m.TenantID,
		LegalName:      m.LegalName,
		DocumentType:   entity.DocumentType(m.DocumentType),
		DocumentNumber: m.DocumentNumber,
		Phone:          m.Phone,
		Email:          m.Email,
		Address:        m.Address,
		CreditLimit:    creditLimit,
		CreditBalance:  creditBalance,
		PointsAccrued:  m.PointsAccrued,
		Status:         entity.CustomerStatus(m.Status),
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      m.DeletedAt,
	}
}
