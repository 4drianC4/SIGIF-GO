package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type CategoryGormRepository struct {
	db *sharedDatabase.Database
}

func NewCategoryGormRepository(db *sharedDatabase.Database) repository.CategoryRepository {
	return &CategoryGormRepository{db: db}
}

func (r *CategoryGormRepository) Create(ctx context.Context, category *entity.Category) error {
	db := r.db.GetDB(ctx)
	if err := db.Create(mapper.CategoryToModel(category)).Error; err != nil {
		if isDuplicateKey(db, err) {
			return service.ErrCategoryNameTaken
		}
		return err
	}
	return nil
}

func (r *CategoryGormRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Category, error) {
	var m model.CategoryModel
	err := r.db.GetDB(ctx).Where("company_id = ? AND id = ?", companyID, id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.CategoryToDomain(&m), nil
}

func (r *CategoryGormRepository) ExistsByName(ctx context.Context, companyID uuid.UUID, name string) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.CategoryModel{}).
		Where("company_id = ? AND LOWER(name) = LOWER(?)", companyID, name).
		Count(&count).Error
	return count > 0, err
}
