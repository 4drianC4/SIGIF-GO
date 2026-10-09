package entity

import (
	"time"

	"github.com/google/uuid"
)

// Role names seeded by the system.
const (
	RoleSuperadmin    = "superadmin"
	RoleBusinessAdmin = "business_admin"
	RoleEmployee      = "employee"
)

// RoleType tells the roles seeded by the system from the ones a company may
// customize (HU-082-02).
type RoleType string

const (
	RoleTypeSystem RoleType = "system"
	RoleTypeCustom RoleType = "custom"
)

func (t RoleType) String() string {
	return string(t)
}

func (t RoleType) IsValid() bool {
	return t == RoleTypeSystem || t == RoleTypeCustom
}

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
	Status      RoleStatus
	// PermissionsCount is filled by listing queries only: the number of
	// permissions assigned to the role. It is zero everywhere else.
	PermissionsCount int
	CreatedAt        time.Time
}

func (r *Role) IsGlobal() bool {
	return r.CompanyID == nil
}

// Type classifies the role for the listing: system roles are available to
// every company, custom roles belong to a single one.
func (r *Role) Type() RoleType {
	if r.IsSystem {
		return RoleTypeSystem
	}
	return RoleTypeCustom
}

// IsActive reports whether the role is currently usable for assignments and
// authorizations (HU-082-04).
func (r *Role) IsActive() bool {
	return r.Status == RoleStatusActive
}

// Activate makes the role available again without touching its configuration.
func (r *Role) Activate() {
	r.Status = RoleStatusActive
}

// Deactivate takes the role out of service. It is not a soft delete: the role
// and its permission assignments are kept.
func (r *Role) Deactivate() {
	r.Status = RoleStatusInactive
}

// IsProtected reports whether the role is critical for the system and therefore
// cannot be deactivated. Superadmin is reserved for bootstrap operations (see
// IsUserAssignable), so its status must stay active.
func (r *Role) IsProtected() bool {
	return r.IsSystem && r.Name == RoleSuperadmin
}
