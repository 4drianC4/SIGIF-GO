package query

import (
	"github.com/google/uuid"
)

type GetTenantQuery struct {
	ID uuid.UUID
}

type GetTenantBySlugQuery struct {
	Slug string
}

type ListTenantsQuery struct {
	Offset int
	Limit  int
}

type ListTenantsByBusinessTypeQuery struct {
	BusinessType string
	Offset       int
	Limit        int
}