package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// RoleListFilter narrows a role listing (HU-082-02). A nil CompanyID means the
// caller has no company and only sees the global roles; otherwise the roles of
// the company plus the global ones are visible. A nil Type or Status leaves
// that dimension unfiltered.
type RoleListFilter struct {
	CompanyID *uuid.UUID
	Q         string
	Type      *entity.RoleType
	Status    *entity.RoleStatus
}

// RoleRepository is the persistence port for roles.
type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)

	// List returns a page of the visible roles ordered by name, plus the total
	// number of matching rows. Each role carries its permissions count.
	List(ctx context.Context, filter RoleListFilter, offset, limit int) ([]*entity.Role, int64, error)
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

	// GetByID returns the permission with the given id, or nil when it does not
	// exist.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error)

	// ListAll returns every permission of the catalog, optionally restricted to
	// a module, ordered by module and operation (used by the CSV export).
	ListAll(ctx context.Context, module string) ([]*entity.Permission, error)

	// ListByRole returns the permissions assigned to a role, ordered by module
	// and operation (used to expose the user's effective permissions).
	ListByRole(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error)

	// Create stores a new permission; its code (module.operation) is unique.
	Create(ctx context.Context, permission *entity.Permission) error

	// Update persists changes to an existing permission.
	Update(ctx context.Context, permission *entity.Permission) error

	// Delete removes a permission by id.
	Delete(ctx context.Context, id uuid.UUID) error

	// CountRolesByPermission returns how many roles have the permission assigned.
	CountRolesByPermission(ctx context.Context, permissionID uuid.UUID) (int64, error)
}
