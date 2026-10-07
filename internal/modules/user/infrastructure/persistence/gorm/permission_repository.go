package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
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

// List returns a page of permissions ordered by module and operation, with the
// total number of matching rows.
func (r *PermissionGormRepository) List(ctx context.Context, module string, offset, limit int) ([]*entity.Permission, int64, error) {
	db := r.db.GetDB(ctx).Model(&model.PermissionModel{})
	if module != "" {
		db = db.Where("module = ?", module)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []model.PermissionModel
	if err := db.Order("module ASC, operation ASC").Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	permissions := make([]*entity.Permission, len(models))
	for i := range models {
		permissions[i] = mapper.PermissionToDomain(&models[i])
	}
	return permissions, total, nil
}

// GetByModuleOperation returns the permission stored for the code module.operation.
func (r *PermissionGormRepository) GetByModuleOperation(ctx context.Context, module, operation string) (*entity.Permission, error) {
	var m model.PermissionModel
	err := r.db.GetDB(ctx).
		Where("module = ? AND operation = ?", module, operation).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.PermissionToDomain(&m), nil
}

// Create stores a new permission; the unique index on (module, operation)
// translates a concurrent duplicate into a 409 conflict.
func (r *PermissionGormRepository) Create(ctx context.Context, permission *entity.Permission) error {
	m := mapper.PermissionToModel(permission)
	if err := r.db.GetDB(ctx).Create(m).Error; err != nil {
		if isDuplicateKey(err) {
			return service.ErrPermissionCodeTaken
		}
		return err
	}
	permission.ID = m.ID
	return nil
}
