package mapper

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/model"
)

func ToModel(c *entity.Customer) *model.CustomerModel {
	return &model.CustomerModel{
		ID:             c.ID,
		CompanyID:      c.CompanyID,
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
		DeletedAt:      toDeletedAt(c.DeletedAt),
	}
}

func ToDomain(m *model.CustomerModel) *entity.Customer {
	creditLimit, _ := decimal.NewFromString(m.CreditLimit)
	creditBalance, _ := decimal.NewFromString(m.CreditBalance)

	return &entity.Customer{
		ID:             m.ID,
		CompanyID:      m.CompanyID,
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
		DeletedAt:      fromDeletedAt(m.DeletedAt),
	}
}

func toDeletedAt(t *time.Time) gorm.DeletedAt {
	if t == nil {
		return gorm.DeletedAt{}
	}
	return gorm.DeletedAt{Time: *t, Valid: true}
}

func fromDeletedAt(d gorm.DeletedAt) *time.Time {
	if !d.Valid {
		return nil
	}
	t := d.Time
	return &t
}
