package entity

import (
	"time"

	"github.com/google/uuid"
)

// Role names seeded by the system.
const (
	RoleSuperadmin = "superadmin"
	RoleSoporte    = "soporte"
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
