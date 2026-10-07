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

	// List returns a page of permissions ordered by module and operation, plus
	// the total number of rows. An empty module lists every permission.
	List(ctx context.Context, module string, offset, limit int) ([]*entity.Permission, int64, error)

	// GetByModuleOperation returns the permission stored for the given module
	// and operation, or nil when it is not registered yet.
	GetByModuleOperation(ctx context.Context, module, operation string) (*entity.Permission, error)

	// Create stores a new permission; its code (module.operation) is unique.
	Create(ctx context.Context, permission *entity.Permission) error
}
