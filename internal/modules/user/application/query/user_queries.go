package query

import (
	"github.com/google/uuid"
)

type GetUserQuery struct {
	ID uuid.UUID
}

type GetUserByEmailQuery struct {
	Email string
}

type ListUsersQuery struct {
	TenantID  uuid.UUID
	CompanyID *uuid.UUID
	Offset    int
	Limit     int
}