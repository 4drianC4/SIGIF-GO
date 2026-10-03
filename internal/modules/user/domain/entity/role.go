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
