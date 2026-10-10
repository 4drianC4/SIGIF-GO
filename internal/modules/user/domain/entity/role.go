package entity

import (
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Role names seeded by the system.
const (
	RoleSuperadmin    = "superadmin"
	RoleBusinessAdmin = "business_admin"
	RoleEmployee      = "employee"
)

// systemRoleDisplayLabels maps each system role's technical name to the Spanish
// label shown to users (HU-082-02). Custom roles keep the name the company
// entered, so they are not part of this catalog. The map is the single source
// of truth for both the listing response (display_name) and the search by
// visible label; no translations are stored in the database.
var systemRoleDisplayLabels = map[string]string{
	RoleSuperadmin:    "Superadministrador",
	RoleBusinessAdmin: "Administrador de empresa",
	RoleEmployee:      "Empleado",
}

// DisplayName returns the Spanish label for a role: the catalog label for
// system roles, or the company-defined name for custom roles.
func DisplayName(name string, isSystem bool) string {
	if isSystem {
		if label, ok := systemRoleDisplayLabels[name]; ok {
			return label
		}
	}
	return name
}

// AllSystemRoleNames returns the technical names of the roles the seed
// maintains as canonical system roles, in seed order.
func AllSystemRoleNames() []string {
	return []string{RoleSuperadmin, RoleBusinessAdmin, RoleEmployee}
}

// SystemRoleNamesByDisplayPrefix returns the technical names of the system
// roles whose Spanish label contains q (case-insensitive, literal match). It
// feeds the listing search so the visible labels are also searchable without
// storing translations or rewriting SQL textually.
func SystemRoleNamesByDisplayPrefix(q string) []string {
	needle := strings.ToLower(strings.TrimSpace(q))
	if needle == "" {
		return nil
	}
	var names []string
	for name, label := range systemRoleDisplayLabels {
		if strings.Contains(strings.ToLower(label), needle) {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

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

// DisplayName is the Spanish label shown to users: the catalog label for
// system roles, or the company-defined name for custom roles (HU-082-02).
func (r *Role) DisplayName() string {
	return DisplayName(r.Name, r.IsSystem)
}

// Type classifies the role for the listing: system roles are available to
// every company, custom roles belong to a single one.
func (r *Role) Type() RoleType {
	if r.IsSystem {
		return RoleTypeSystem
	}
	return RoleTypeCustom
}
