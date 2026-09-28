package command

import (
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
)

type CreateCompanyCommand struct {
	TenantID   uuid.UUID
	Name       string
	LegalName  string
	TaxID      string
}

type UpdateCompanyCommand struct {
	ID          uuid.UUID
	Name        string
	LegalName   string
	TaxID       string
	Email       string
	Phone       string
	Address     string
	City        string
	State       string
	Country     string
	PostalCode  string
	Settings    entity.CompanySettings
}

type DeleteCompanyCommand struct {
	ID uuid.UUID
}

type ActivateCompanyCommand struct {
	ID uuid.UUID
}

type DeactivateCompanyCommand struct {
	ID uuid.UUID
}