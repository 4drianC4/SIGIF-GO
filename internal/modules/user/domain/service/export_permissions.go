package service

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// ExportPermissions returns every permission of the catalog, optionally
// restricted to a module, for the CSV export (HU-082-01). Unlike the paginated
// list it returns the full result set, ordered by module and operation.
func (s *UserService) ExportPermissions(ctx context.Context, module string) ([]*entity.Permission, error) {
	module = normalize(module)
	if module != "" && !entity.IsSupportedModule(module) {
		return nil, ErrUnknownPermissionModule
	}
	return s.permRepo.ListAll(ctx, module)
}
