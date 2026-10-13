package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type UnitOfMeasureGormRepository struct {
	db *sharedDatabase.Database
}

func NewUnitOfMeasureGormRepository(db *sharedDatabase.Database) repository.UnitOfMeasureRepository {
	return &UnitOfMeasureGormRepository{db: db}
}

func (r *UnitOfMeasureGormRepository) List(ctx context.Context) ([]*entity.UnitOfMeasure, error) {
	var models []model.UnitOfMeasureModel
	if err := r.db.GetDB(ctx).Order("sort_order ASC, name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	units := make([]*entity.UnitOfMeasure, len(models))
	for i := range models {
		units[i] = mapper.UnitToDomain(&models[i])
	}
	return units, nil
}

func (r *UnitOfMeasureGormRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.UnitOfMeasureModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

type TaxGormRepository struct {
	db *sharedDatabase.Database
}

func NewTaxGormRepository(db *sharedDatabase.Database) repository.TaxRepository {
	return &TaxGormRepository{db: db}
}

func (r *TaxGormRepository) List(ctx context.Context) ([]*entity.Tax, error) {
	var models []model.TaxModel
	if err := r.db.GetDB(ctx).Order("percentage DESC, name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	taxes := make([]*entity.Tax, len(models))
	for i := range models {
		taxes[i] = mapper.TaxToDomain(&models[i])
	}
	return taxes, nil
}

func (r *TaxGormRepository) GetByName(ctx context.Context, name string) (*entity.Tax, error) {
	var m model.TaxModel
	err := r.db.GetDB(ctx).Where("LOWER(name) = LOWER(?)", name).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.TaxToDomain(&m), nil
}
