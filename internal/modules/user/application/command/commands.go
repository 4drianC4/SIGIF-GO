package command

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type RegisterUser struct {
	CompanyID *uuid.UUID
	RoleName  string
	FirstName string
	LastName  string
	Email     string
}

// EditUser is the administrative partial update (registration fields plus an optional password reset,
// all optional). Nil = leave the field unchanged.
type EditUser struct {
	CompanyID *uuid.UUID
	ID        uuid.UUID
	FirstName *string
	LastName  *string
	Email     *string
	Password  *string
	RoleName  *string
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

// CreatePermission registers a permission of the catalog (HU-082-01). Code is
// optional: when sent it must match module.operation.
type CreatePermission struct {
	Module      string
	Operation   string
	Code        string
	Description string
}

// UpdatePermission edits a permission by id. Only the description is mutable;
// Nil = leave it unchanged.
type UpdatePermission struct {
	ID          uuid.UUID
	Description *string
}

// DeletePermission removes a permission by id.
type DeletePermission struct {
	ID uuid.UUID
}

// ChangeRoleStatus activates or deactivates a role (HU-082-04).
type ChangeRoleStatus struct {
	ID        uuid.UUID
	CompanyID *uuid.UUID
	Status    entity.RoleStatus
}
