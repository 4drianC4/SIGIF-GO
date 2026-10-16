package service

import (
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// PermissionModules returns the catalog of modules with the operations a
// permission of each module may use (HU-082-01).
func (s *UserService) PermissionModules() []entity.ModuleWithOperations {
	return entity.SupportedModuleCatalog()
}
