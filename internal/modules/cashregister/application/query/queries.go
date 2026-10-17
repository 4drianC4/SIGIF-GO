package query

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
)

type ListCashRegisters struct {
	CompanyID    uuid.UUID
	Availability *entity.Availability
	Offset       int
	Limit        int
}

type GetCurrentAssignment struct {
	CompanyID uuid.UUID
	UserID    uuid.UUID
}
