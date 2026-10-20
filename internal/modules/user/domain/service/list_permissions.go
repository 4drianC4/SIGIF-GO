package service

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// ListPermissions returns a page of the registered permissions, optionally
// restricted to a module of the catalog.
func (s *UserService) ListPermissions(ctx context.Context, module string, offset, limit int) ([]*entity.Permission, int64, error) {
	module = normalize(module)
	if module != "" && !entity.IsSupportedModule(module) {
		return nil, 0, ErrUnknownPermissionModule
	}
	return s.permRepo.List(ctx, module, offset, limit)
}
