package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type RoleGormRepository struct {
	db *sharedDatabase.Database
}

func NewRoleGormRepository(db *sharedDatabase.Database) repository.RoleRepository {
	return &RoleGormRepository{db: db}
}

func (r *RoleGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	var m model.RoleModel
	err := r.db.GetDB(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.RoleToDomain(&m), nil
}

func (r *RoleGormRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var m model.RoleModel
	err := r.db.GetDB(ctx).Where("name = ?", name).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.RoleToDomain(&m), nil
}
