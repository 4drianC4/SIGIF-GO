package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// RoleRepository is the persistence port for roles.
type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
}

// PermissionRepository is the persistence port for permissions and their
// assignment to roles (RBAC).
type PermissionRepository interface {
	HasPermission(ctx context.Context, roleID uuid.UUID, module, operation string) (bool, error)
}
