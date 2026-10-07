package gorm

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type ProductGormRepository struct {
	db *sharedDatabase.Database
}

func NewProductGormRepository(db *sharedDatabase.Database) repository.ProductRepository {
	return &ProductGormRepository{db: db}
}

func (r *ProductGormRepository) Create(ctx context.Context, product *entity.Product) error {
	db := r.db.GetDB(ctx)
	if err := db.Create(mapper.ProductToModel(product)).Error; err != nil {
		if isDuplicateKey(db, err) {
			if strings.Contains(err.Error(), "idx_products_company_barcode") {
				return service.ErrBarcodeTaken
			}
			return service.ErrSKUTaken
		}
		return err
	}
	return nil
}

func (r *ProductGormRepository) Update(ctx context.Context, product *entity.Product) error {
	db := r.db.GetDB(ctx)
	m := mapper.ProductToModel(product)
	if err := db.Save(m).Error; err != nil {
		if isDuplicateKey(db, err) {
			if strings.Contains(err.Error(), "idx_products_company_barcode") {
				return service.ErrBarcodeTaken
			}
			return service.ErrSKUTaken
		}
		return err
	}
	return nil
}

func (r *ProductGormRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Product, error) {
	var m model.ProductModel
	err := r.db.GetDB(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ProductToDomain(&m), nil
}

func (r *ProductGormRepository) GetAll(ctx context.Context, companyID uuid.UUID, filters repository.ProductFilters, page, limit int) ([]*entity.Product, int64, error) {
	db := r.db.GetDB(ctx).Model(&model.ProductModel{}).Where("company_id = ?", companyID)

	if filters.Name != "" {
		db = db.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(filters.Name)+"%")
	}
	if filters.CategoryID != nil {
		db = db.Where("category_id = ?", *filters.CategoryID)
	}
	if filters.Status != nil {
		db = db.Where("status = ?", string(*filters.Status))
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var ms []model.ProductModel
	offset := (page - 1) * limit
	if err := db.Order("name ASC").Offset(offset).Limit(limit).Find(&ms).Error; err != nil {
		return nil, 0, err
	}

	products := make([]*entity.Product, 0, len(ms))
	for i := range ms {
		products = append(products, mapper.ProductToDomain(&ms[i]))
	}
	return products, total, nil
}

func (r *ProductGormRepository) SetStatus(ctx context.Context, companyID, id uuid.UUID, status entity.Status) error {
	result := r.db.GetDB(ctx).Model(&model.ProductModel{}).
		Where("company_id = ? AND id = ?", companyID, id).
		Update("status", string(status))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return service.ErrProductNotFound
	}
	return nil
}

func (r *ProductGormRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	result := r.db.GetDB(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		Delete(&model.ProductModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return service.ErrProductNotFound
	}
	return nil
}

func (r *ProductGormRepository) ExistsBySKU(ctx context.Context, companyID uuid.UUID, sku string) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.ProductModel{}).
		Where("company_id = ? AND sku = ?", companyID, sku).
		Count(&count).Error
	return count > 0, err
}

func (r *ProductGormRepository) ExistsBySKUExcluding(ctx context.Context, companyID uuid.UUID, sku string, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.ProductModel{}).
		Where("company_id = ? AND sku = ? AND id != ?", companyID, sku, excludeID).
		Count(&count).Error
	return count > 0, err
}

func (r *ProductGormRepository) ExistsByBarcode(ctx context.Context, companyID uuid.UUID, barcode string) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.ProductModel{}).
		Where("company_id = ? AND barcode = ?", companyID, barcode).
		Count(&count).Error
	return count > 0, err
}

func (r *ProductGormRepository) ExistsByBarcodeExcluding(ctx context.Context, companyID uuid.UUID, barcode string, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.ProductModel{}).
		Where("company_id = ? AND barcode = ? AND id != ?", companyID, barcode, excludeID).
		Count(&count).Error
	return count > 0, err
}
