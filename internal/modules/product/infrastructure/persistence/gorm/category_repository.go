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

func (r *CategoryGormRepository) ExistsByName(ctx context.Context, companyID uuid.UUID, parentID *uuid.UUID, name string) (bool, error) {
	db := r.db.GetDB(ctx).Model(&model.CategoryModel{}).
		Where("company_id = ? AND LOWER(name) = LOWER(?)", companyID, name)
	if parentID == nil {
		db = db.Where("parent_id IS NULL")
	} else {
		db = db.Where("parent_id = ?", *parentID)
	}

	var count int64
	err := db.Count(&count).Error
	return count > 0, err
}

type categoryRow struct {
	model.CategoryModel `gorm:"embedded"`
	ProductsCount       int64
	DefaultTaxName      *string
}

func (r *CategoryGormRepository) List(ctx context.Context, companyID uuid.UUID) ([]repository.CategoryListItem, error) {
	var rows []categoryRow
	err := r.db.GetDB(ctx).Table("categories AS c").
		Select(`c.*, t.name AS default_tax_name,
			(SELECT COUNT(*) FROM products p WHERE p.category_id = c.id AND p.deleted_at IS NULL) AS products_count`).
		Joins("LEFT JOIN taxes t ON t.id = c.default_tax_id").
		Where("c.company_id = ? AND c.deleted_at IS NULL", companyID).
		Order("c.name ASC, c.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]repository.CategoryListItem, len(rows))
	for i := range rows {
		items[i] = repository.CategoryListItem{
			Category:       mapper.CategoryToDomain(&rows[i].CategoryModel),
			ProductsCount:  rows[i].ProductsCount,
			DefaultTaxName: rows[i].DefaultTaxName,
		}
	}
	return items, nil
}
