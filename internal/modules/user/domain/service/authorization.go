package service

import (
	"context"

	"github.com/google/uuid"
)

// HasPermission resolves the user's role and checks whether it holds the given
// module/operation permission. It backs the RBAC middleware.
func (s *UserService) HasPermission(ctx context.Context, userID uuid.UUID, module, operation string) (bool, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil || !user.IsActive() {
		return false, nil
	}

	return s.permRepo.HasPermission(ctx, user.RoleID, module, operation)
}
