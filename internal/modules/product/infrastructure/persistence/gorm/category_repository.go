package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

func (r *CategoryGormRepository) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(sharedDatabase.WithTx(ctx, tx))
	})
}

func (r *CategoryGormRepository) Update(ctx context.Context, category *entity.Category) error {
	db := r.db.GetDB(ctx)
	if err := db.Omit(clause.Associations).Save(mapper.CategoryToModel(category)).Error; err != nil {
		if isDuplicateKey(db, err) {
			return service.ErrCategoryNameTaken
		}
		return err
	}
	return nil
}

func (r *CategoryGormRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Category, error) {
	strength := ""
	if sharedDatabase.GetTx(ctx) != nil {
		strength = "SHARE"
	}
	return r.find(ctx, strength, companyID, id)
}

func (r *CategoryGormRepository) GetByIDForUpdate(ctx context.Context, companyID, id uuid.UUID) (*entity.Category, error) {
	return r.find(ctx, "UPDATE", companyID, id)
}

// find reads a category that is not deleted. With a lock strength the row is
// locked; if it is deleted while waiting for the lock, it is no longer found.
func (r *CategoryGormRepository) find(ctx context.Context, strength string, companyID, id uuid.UUID) (*entity.Category, error) {
	db := r.db.GetDB(ctx)
	if strength != "" {
		db = db.Clauses(clause.Locking{Strength: strength})
	}
	var m model.CategoryModel
	err := db.Where("company_id = ? AND id = ?", companyID, id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.CategoryToDomain(&m), nil
}

func (r *CategoryGormRepository) ExistsByName(ctx context.Context, companyID uuid.UUID, parentID *uuid.UUID, name string, excludeID uuid.UUID) (bool, error) {
	db := r.db.GetDB(ctx).Model(&model.CategoryModel{}).
		Where("company_id = ? AND LOWER(name) = LOWER(?)", companyID, name)
	if parentID == nil {
		db = db.Where("parent_id IS NULL")
	} else {
		db = db.Where("parent_id = ?", *parentID)
	}
	if excludeID != uuid.Nil {
		db = db.Where("id <> ?", excludeID)
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

func (r *CategoryGormRepository) CountSubcategories(ctx context.Context, companyID, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.CategoryModel{}).
		Where("company_id = ? AND parent_id = ?", companyID, id).Count(&count).Error
	return count, err
}

func (r *CategoryGormRepository) CountProducts(ctx context.Context, companyID, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.ProductModel{}).
		Where("company_id = ? AND category_id = ?", companyID, id).Count(&count).Error
	return count, err
}

func (r *CategoryGormRepository) SoftDelete(ctx context.Context, companyID, id uuid.UUID) (bool, error) {
	result := r.db.GetDB(ctx).Where("company_id = ? AND id = ?", companyID, id).Delete(&model.CategoryModel{})
	return result.RowsAffected > 0, result.Error
}

// listQuery selects the categories of the company that are not deleted with
// their default tax name and the count of products that are not deleted.
func (r *CategoryGormRepository) listQuery(ctx context.Context, companyID uuid.UUID) *gorm.DB {
	return r.db.GetDB(ctx).Table("categories AS c").
		Select(`c.*, t.name AS default_tax_name,
			(SELECT COUNT(*) FROM products p WHERE p.category_id = c.id AND p.deleted_at IS NULL) AS products_count`).
		Joins("LEFT JOIN taxes t ON t.id = c.default_tax_id").
		Where("c.company_id = ? AND c.deleted_at IS NULL", companyID)
}

func toListItems(rows []categoryRow) []repository.CategoryListItem {
	items := make([]repository.CategoryListItem, len(rows))
	for i := range rows {
		items[i] = repository.CategoryListItem{
			Category:       mapper.CategoryToDomain(&rows[i].CategoryModel),
			ProductsCount:  rows[i].ProductsCount,
			DefaultTaxName: rows[i].DefaultTaxName,
		}
	}
	return items
}

func (r *CategoryGormRepository) List(ctx context.Context, companyID uuid.UUID) ([]repository.CategoryListItem, error) {
	var rows []categoryRow
	if err := r.listQuery(ctx, companyID).Order("c.name ASC, c.id ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return toListItems(rows), nil
}

func (r *CategoryGormRepository) GetListItem(ctx context.Context, companyID, id uuid.UUID) (*repository.CategoryListItem, error) {
	var rows []categoryRow
	if err := r.listQuery(ctx, companyID).Where("c.id = ?", id).Limit(1).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &toListItems(rows)[0], nil
}
