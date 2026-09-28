package dto

import (
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
)

type CompanyResponse struct {
	ID         uuid.UUID             `json:"id"`
	TenantID   uuid.UUID             `json:"tenant_id"`
	Name       string                `json:"name"`
	LegalName  string                `json:"legal_name"`
	TaxID      string                `json:"tax_id"`
	Email      string                `json:"email"`
	Phone      string                `json:"phone"`
	Address    string                `json:"address"`
	City       string                `json:"city"`
	State      string                `json:"state"`
	Country    string                `json:"country"`
	PostalCode string                `json:"postal_code"`
	LogoURL    string                `json:"logo_url"`
	Settings   entity.CompanySettings `json:"settings"`
	IsActive   bool                  `json:"is_active"`
	CreatedAt  string                `json:"created_at"`
	UpdatedAt  string                `json:"updated_at"`
}

func ToCompanyResponse(c *entity.Company) CompanyResponse {
	return CompanyResponse{
		ID:         c.ID,
		TenantID:   c.TenantID,
		Name:       c.Name,
		LegalName:  c.LegalName,
		TaxID:      c.TaxID,
		Email:      c.Email,
		Phone:      c.Phone,
		Address:    c.Address,
		City:       c.City,
		State:      c.State,
		Country:    c.Country,
		PostalCode: c.PostalCode,
		LogoURL:    c.LogoURL,
		Settings:   c.Settings,
		IsActive:   c.IsActive,
		CreatedAt:  c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToCompanyResponseList(companies []*entity.Company) []CompanyResponse {
	result := make([]CompanyResponse, len(companies))
	for i, c := range companies {
		result[i] = ToCompanyResponse(c)
	}
	return result
}

type CreateCompanyRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=255"`
	LegalName string `json:"legal_name" validate:"max=255"`
	TaxID     string `json:"tax_id" validate:"required,max=50"`
	Email     string `json:"email" validate:"omitempty,email"`
	Phone     string `json:"phone" validate:"omitempty,max=50"`
}

type UpdateCompanyRequest struct {
	Name       string                 `json:"name" validate:"required,min=2,max=255"`
	LegalName  string                 `json:"legal_name" validate:"max=255"`
	TaxID      string                 `json:"tax_id" validate:"required,max=50"`
	Email      string                 `json:"email" validate:"omitempty,email"`
	Phone      string                 `json:"phone" validate:"omitempty,max=50"`
	Address    string                 `json:"address" validate:"omitempty"`
	City       string                 `json:"city" validate:"omitempty,max=100"`
	State      string                 `json:"state" validate:"omitempty,max=100"`
	Country    string                 `json:"country" validate:"omitempty,max=100"`
	PostalCode string                 `json:"postal_code" validate:"omitempty,max=20"`
	Settings   entity.CompanySettings `json:"settings"`
}