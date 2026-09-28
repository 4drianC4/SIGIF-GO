package query

import (
	"github.com/google/uuid"
)

type GetCompanyQuery struct {
	ID uuid.UUID
}

type ListCompaniesQuery struct {
	TenantID uuid.UUID
	Offset   int
	Limit    int
}