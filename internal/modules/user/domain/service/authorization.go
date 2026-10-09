package service

import (
	"context"

	"github.com/google/uuid"
)

// HasPermission resolves the user's role and checks whether it holds the given
// module/operation permission. It backs the RBAC middleware.
//
// A user whose role is deactivated (HU-082-04) has no effective permissions:
// the role status is read on every check, so a role deactivated while the user
// keeps an open session stops granting access immediately, without needing to
// invalidate the session.
func (s *UserService) HasPermission(ctx context.Context, userID uuid.UUID, module, operation string) (bool, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil || !user.IsActive() {
		return false, nil
	}

	role, err := s.roleRepo.GetByID(ctx, user.RoleID)
	if err != nil {
		return false, err
	}
	if role == nil || !role.IsActive() {
		return false, nil
	}

	return s.permRepo.HasPermission(ctx, user.RoleID, module, operation)
}
