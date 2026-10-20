package service

import (
	"context"

	"github.com/google/uuid"
)

// PermissionCodesByRole returns the codes (module.operation) of the permissions
// assigned to a role. It feeds the user's effective permissions returned by the
// auth endpoints.
//
// A deactivated role grants no permissions (HU-082-04), so its assignments are
// not reported even though they are still stored and survive a reactivation.
func (s *UserService) PermissionCodesByRole(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if role == nil || !role.IsActive() {
		return []string{}, nil
	}

	permissions, err := s.permRepo.ListByRole(ctx, roleID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, len(permissions))
	for i := range permissions {
		codes[i] = permissions[i].Key()
	}
	return codes, nil
}
