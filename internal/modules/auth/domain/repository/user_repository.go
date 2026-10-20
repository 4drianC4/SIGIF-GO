package repository

import (
	"context"

	"github.com/google/uuid"

	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// UserRepo is the port the auth module needs from the user module.
type UserRepo interface {
	GetByEmail(ctx context.Context, email string) (*userEntity.AppUser, error)
	GetByID(ctx context.Context, id uuid.UUID) (*userEntity.AppUser, error)
	RecordAccess(ctx context.Context, userID uuid.UUID) error
	// PermissionsByRole returns the codes (module.operation) granted to a role,
	// so the auth responses can expose the user's effective permissions.
	PermissionsByRole(ctx context.Context, roleID uuid.UUID) ([]string, error)
}
