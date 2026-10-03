package query

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

// GetCustomer define la consulta para obtener un cliente por ID.
type GetCustomer struct {
	ID       int64
	TenantID uuid.UUID
}

// ListCustomers define la consulta para buscar clientes paginados.
type ListCustomers struct {
	TenantID uuid.UUID
	Q        string
	Status   *entity.CustomerStatus
	Offset   int
	Limit    int
}
