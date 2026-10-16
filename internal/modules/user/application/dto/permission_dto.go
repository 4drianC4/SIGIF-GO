package dto

import (
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// Permission is the catalog entry served by the permission endpoints.
type Permission struct {
	ID          string `json:"id"`
	Module      string `json:"module"`
	Operation   string `json:"operation"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	// IsSystem marks permissions of the seed. They cannot be deleted, so the
	// client hides the delete action when it is true.
	IsSystem bool `json:"is_system"`
}

// PermissionModule is a module of the catalog with the operations a permission
// of that module may use.
type PermissionModule struct {
	Module     string   `json:"module"`
	Operations []string `json:"operations"`
}

func PermissionFromEntity(p *entity.Permission) Permission {
	return Permission{
		ID:          p.ID.String(),
		Module:      p.Module,
		Operation:   p.Operation,
		Code:        p.Key(),
		Description: p.Description,
		IsSystem:    entity.IsSystemPermission(p.Module, p.Operation),
	}
}

func PermissionListFromEntity(permissions []*entity.Permission) []Permission {
	result := make([]Permission, len(permissions))
	for i, p := range permissions {
		result[i] = PermissionFromEntity(p)
	}
	return result
}

func PermissionModuleListFromEntity(modules []entity.ModuleWithOperations) []PermissionModule {
	result := make([]PermissionModule, len(modules))
	for i, m := range modules {
		result[i] = PermissionModule{
			Module:     m.Module,
			Operations: m.Operations,
		}
	}
	return result
}
