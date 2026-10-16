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

// GetByID returns the permission with the given id, or nil when it is missing.
func (r *PermissionGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error) {
	var m model.PermissionModel
	err := r.db.GetDB(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.PermissionToDomain(&m), nil
}

// ListAll returns every permission of the catalog ordered by module and
// operation, optionally filtered by module.
func (r *PermissionGormRepository) ListAll(ctx context.Context, module string) ([]*entity.Permission, error) {
	db := r.db.GetDB(ctx).Model(&model.PermissionModel{})
	if module != "" {
		db = db.Where("module = ?", module)
	}

	var models []model.PermissionModel
	if err := db.Order("module ASC, operation ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	permissions := make([]*entity.Permission, len(models))
	for i := range models {
		permissions[i] = mapper.PermissionToDomain(&models[i])
	}
	return permissions, nil
}

// ListByRole returns the permissions assigned to the role, ordered by module
// and operation.
func (r *PermissionGormRepository) ListByRole(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error) {
	var models []model.PermissionModel
	err := r.db.GetDB(ctx).
		Table("permission").
		Joins("JOIN role_permission ON role_permission.permission_id = permission.id").
		Where("role_permission.role_id = ?", roleID).
		Order("permission.module ASC, permission.operation ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	permissions := make([]*entity.Permission, len(models))
	for i := range models {
		permissions[i] = mapper.PermissionToDomain(&models[i])
	}
	return permissions, nil
}

// Update persists the description change of an existing permission.
func (r *PermissionGormRepository) Update(ctx context.Context, permission *entity.Permission) error {
	return r.db.GetDB(ctx).
		Model(&model.PermissionModel{}).
		Where("id = ?", permission.ID).
		Update("description", permission.Description).Error
}

// Delete removes a permission by id.
func (r *PermissionGormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.GetDB(ctx).Where("id = ?", id).Delete(&model.PermissionModel{}).Error
}

// CountRolesByPermission returns how many roles have the permission assigned.
func (r *PermissionGormRepository) CountRolesByPermission(ctx context.Context, permissionID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.GetDB(ctx).
		Model(&model.RolePermissionModel{}).
		Where("permission_id = ?", permissionID).
		Count(&count).Error
	return count, err
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
