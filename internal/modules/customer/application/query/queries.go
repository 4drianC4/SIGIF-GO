package query

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

type GetCustomer struct {
	ID       uuid.UUID
	TenantID uuid.UUID
}

type ListCustomers struct {
	TenantID uuid.UUID
	Q        string
	Status   *entity.CustomerStatus
	Offset   int
	Limit    int
}
