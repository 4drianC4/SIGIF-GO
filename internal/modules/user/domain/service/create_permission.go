package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type CreatePermissionParams struct {
	Module      string
	Operation   string
	Code        string
	Description string
}

// CreatePermission registers a new permission of the catalog so it can be
// assigned to roles. Module and operation are normalized (trimmed and lower
// cased) and must belong to the catalog; the code is the canonical
// module.operation key and may not be redefined by the caller.
func (s *UserService) CreatePermission(ctx context.Context, params CreatePermissionParams) (*entity.Permission, error) {
	module := normalize(params.Module)
	operation := normalize(params.Operation)

	if !entity.IsSupportedModule(module) {
		return nil, ErrUnknownPermissionModule
	}
	if !entity.IsSupportedOperation(operation) {
		return nil, ErrUnknownPermissionOperation
	}

	permission := &entity.Permission{
		ID:          uuid.New(),
		Module:      module,
		Operation:   operation,
		Description: strings.TrimSpace(params.Description),
	}

	if code := strings.TrimSpace(params.Code); code != "" && normalize(code) != permission.Key() {
		return nil, ErrPermissionCodeMismatch
	}

	existing, err := s.permRepo.GetByModuleOperation(ctx, module, operation)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPermissionCodeTaken
	}

	if err := s.permRepo.Create(ctx, permission); err != nil {
		return nil, err
	}
	return permission, nil
}

// normalize trims the surrounding spaces and lower cases a catalog value.
func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
