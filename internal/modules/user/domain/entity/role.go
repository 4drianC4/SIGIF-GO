package entity

import (
	"time"

	"github.com/google/uuid"
)

// Role names seeded by the system.
const (
	RoleSuperadmin    = "superadmin"
	RoleBusinessAdmin = "business_admin"
	RoleEmployee      = "employe"
)

// IsUserAssignable reports whether a role may be selected through the user
// management API. Superadmin is reserved for backend bootstrap operations.
func IsUserAssignable(name string) bool {
	return name == RoleBusinessAdmin || name == RoleEmployee
}

type Role struct {
	ID          uuid.UUID
	CompanyID   *uuid.UUID
	Name        string
	Description string
	IsTemplate  bool
	IsSystem    bool
	CreatedAt   time.Time
}

func (r *Role) IsGlobal() bool {
	return r.CompanyID == nil
}
