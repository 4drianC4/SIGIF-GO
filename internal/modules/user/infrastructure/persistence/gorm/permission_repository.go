package gorm

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type PermissionGormRepository struct {
	db *sharedDatabase.Database
}

func NewPermissionGormRepository(db *sharedDatabase.Database) repository.PermissionRepository {
	return &PermissionGormRepository{db: db}
}

// HasPermission returns true when the role has the module/operation permission.
func (r *PermissionGormRepository) HasPermission(ctx context.Context, roleID uuid.UUID, module, operation string) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).
		Table("role_permission").
		Joins("JOIN permission ON permission.id = role_permission.permission_id").
		Where("role_permission.role_id = ? AND permission.module = ? AND permission.operation = ?", roleID, module, operation).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
