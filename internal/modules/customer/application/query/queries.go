package query

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
)

type GetCustomer struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
}

type ListCustomers struct {
	CompanyID uuid.UUID
	Q         string
	Status    *entity.CustomerStatus
	SortBy    repository.SortField
	SortOrder repository.SortOrder
	Offset    int
	Limit     int
}
