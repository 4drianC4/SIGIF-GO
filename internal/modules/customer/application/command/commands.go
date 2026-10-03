package command

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

// CreateCustomer encapsula los datos necesarios para registrar un nuevo cliente.
type CreateCustomer struct {
	TenantID       uuid.UUID
	LegalName      string
	DocumentType   entity.DocumentType
	DocumentNumber *string
	Phone          *string
	Email          *string
}

// UpdateCustomer encapsula los datos editables de un cliente existente.
type UpdateCustomer struct {
	ID             int64
	TenantID       uuid.UUID
	LegalName      string
	DocumentType   entity.DocumentType
	DocumentNumber *string
	Phone          *string
	Email          *string
	Address        *string
}

// DeleteCustomer encapsula la solicitud de baja lógica de un cliente.
type DeleteCustomer struct {
	ID       int64
	TenantID uuid.UUID
}

// ChangeCustomerStatus encapsula el cambio de estado de un cliente.
type ChangeCustomerStatus struct {
	ID       int64
	TenantID uuid.UUID
	Status   entity.CustomerStatus
}
