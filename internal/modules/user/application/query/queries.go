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

// ListPermissions filters the permission catalog; an empty Module lists every
// permission.
type ListPermissions struct {
	Module string
	Offset int
	Limit  int
}

// ListPermissionModules asks for the catalog of modules and their operations.
type ListPermissionModules struct{}
