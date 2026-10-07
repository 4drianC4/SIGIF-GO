package service

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
)

// ListRoles returns a page of the roles visible to the caller, ordered by
// name, with the number of permissions assigned to each one (HU-082-02). It is
// read-only: roles are never created or modified here, and roles of another
// company are never returned.
func (s *UserService) ListRoles(
	ctx context.Context,
	filter repository.RoleListFilter,
	offset, limit int,
) ([]*entity.Role, int64, error) {
	if filter.Type != nil && !filter.Type.IsValid() {
		filter.Type = nil
	}
	if filter.Status != nil && !filter.Status.IsValid() {
		filter.Status = nil
	}
	return s.roleRepo.List(ctx, filter, offset, limit)
}
