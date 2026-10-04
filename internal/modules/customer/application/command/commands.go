package command

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

type CreateCustomer struct {
	CompanyID      uuid.UUID
	LegalName      string
	DocumentType   entity.DocumentType
	DocumentNumber *string
	Phone          *string
	Email          *string
}

type UpdateCustomer struct {
	ID             uuid.UUID
	CompanyID      uuid.UUID
	LegalName      string
	DocumentType   entity.DocumentType
	DocumentNumber *string
	Phone          *string
	Email          *string
	Address        *string
}

type DeleteCustomer struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
}

type ChangeCustomerStatus struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Status    entity.CustomerStatus
}
