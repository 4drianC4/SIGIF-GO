package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// UpdatePermission updates the mutable data of a permission. Only the
// description may change: module and operation derive the code
// (module.operation), so editing them would silently change every role's
// effective permission.
func (s *UserService) UpdatePermission(ctx context.Context, id uuid.UUID, description *string) (*entity.Permission, error) {
	permission, err := s.permRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, ErrPermissionNotFound
	}

	if description != nil {
		permission.Description = strings.TrimSpace(*description)
	}

	if err := s.permRepo.Update(ctx, permission); err != nil {
		return nil, err
	}
	return permission, nil
}
