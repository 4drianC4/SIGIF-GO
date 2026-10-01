package query

import (
	"github.com/google/uuid"
)

type GetUser struct {
	ID uuid.UUID
}

type GetUserByEmail struct {
	TenantID uuid.UUID
	Email    string
}

type ListUsers struct {
	TenantID uuid.UUID
	Offset   int
	Limit    int
}
