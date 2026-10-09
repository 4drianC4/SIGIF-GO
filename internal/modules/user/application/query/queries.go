package query

import (
	"github.com/google/uuid"
)

type GetUser struct {
	ID uuid.UUID
}

type GetUserByEmail struct {
	Email string
}

type ListUsers struct {
	Offset int
	Limit  int
}

type ListUserHistory struct {
	UserID uuid.UUID
	Offset int
	Limit  int
}
