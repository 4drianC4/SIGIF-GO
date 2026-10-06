package command

import (
	"github.com/google/uuid"
)

type RegisterUser struct {
	CompanyID *uuid.UUID
	RoleName  string
	FirstName string
	LastName  string
	Email     string
	Password  string
	Area      string
}

// EditUser is the administrative partial update (same fields as RegisterUser,
// all optional). Nil = leave the field unchanged.
type EditUser struct {
	ID        uuid.UUID
	FirstName *string
	LastName  *string
	Email     *string
	Password  *string
	RoleName  *string
	Area      *string
}

type ChangePassword struct {
	ID              uuid.UUID
	CurrentPassword string
	NewPassword     string
}

type DeleteUser struct {
	ID uuid.UUID
}

type ChangeStatus struct {
	ID     uuid.UUID
	Active bool
}
