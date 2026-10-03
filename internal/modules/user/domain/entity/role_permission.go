package entity

import "github.com/google/uuid"

// RolePermission links a role with one permission.
type RolePermission struct {
	RoleID       uuid.UUID
	PermissionID uuid.UUID
}
