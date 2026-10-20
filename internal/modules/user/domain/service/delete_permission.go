package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// DeletePermission removes a permission of the catalog. Two rules protect the
// catalog: permissions of the seed are never deleted (ErrPermissionSystemProtected)
// and a permission still assigned to a role must be unassigned first
// (ErrPermissionAssignedToRole).
func (s *UserService) DeletePermission(ctx context.Context, id uuid.UUID) error {
	permission, err := s.permRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if permission == nil {
		return ErrPermissionNotFound
	}

	if entity.IsSystemPermission(permission.Module, permission.Operation) {
		return ErrPermissionSystemProtected
	}

	count, err := s.permRepo.CountRolesByPermission(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrPermissionAssignedToRole
	}

	return s.permRepo.Delete(ctx, id)
}
