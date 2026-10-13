package query

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
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

// ListRoles filters the roles visible to the caller (HU-082-02). A nil
// CompanyID means the caller has no company and only sees the global roles.
type ListRoles struct {
	CompanyID *uuid.UUID
	Q         string
	Type      *entity.RoleType
	Status    *entity.RoleStatus
	Offset    int
	Limit     int
}

// ExportPermissions asks for the full permission catalog (optionally filtered
// by module) to build the CSV export.
type ExportPermissions struct {
	Module string
}
