package service

import (
	"context"

	"github.com/google/uuid"
)

// PermissionCodesByRole returns the codes (module.operation) of the permissions
// assigned to a role. It feeds the user's effective permissions returned by the
// auth endpoints.
func (s *UserService) PermissionCodesByRole(ctx context.Context, roleID uuid.UUID) ([]string, error) {
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
