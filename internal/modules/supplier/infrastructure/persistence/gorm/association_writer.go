package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/service"
	"github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

func (r *ProductSupplierGormRepository) GetProduct(ctx context.Context, companyID, productID uuid.UUID) (*repository.ProductRef, error) {
	db := r.db.GetDB(ctx).Table("products").Select("id, name, sku, status").Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, productID)
	if sharedDatabase.GetTx(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var row repository.ProductRef
	err := db.Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ProductSupplierGormRepository) Get(ctx context.Context, companyID, supplierID, productID uuid.UUID) (*entity.ProductSupplier, error) {
	var row model.ProductSupplierModel
	err := r.db.GetDB(ctx).Where("company_id = ? AND supplier_id = ? AND product_id = ?", companyID, supplierID, productID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mapper.AssociationToDomain(&row), nil
}

func (r *ProductSupplierGormRepository) CreateMany(ctx context.Context, associations []*entity.ProductSupplier) error {
	rows := make([]*model.ProductSupplierModel, len(associations))
	for i, a := range associations {
		rows[i] = mapper.AssociationToModel(a)
	}
	return translateError(r.db.GetDB(ctx).Create(&rows).Error)
}

func (r *ProductSupplierGormRepository) Update(ctx context.Context, a *entity.ProductSupplier) error {
	row := mapper.AssociationToModel(a)
	result := r.db.GetDB(ctx).Model(&model.ProductSupplierModel{}).
		Where("company_id = ? AND supplier_id = ? AND product_id = ?", a.CompanyID, a.SupplierID, a.ProductID).
		Select("supplier_code", "lead_time_days", "current_cost", "preferred", "updated_at").Updates(row)
	if result.Error != nil {
		return translateError(result.Error)
	}
	if result.RowsAffected == 0 {
		return service.ErrAssociationNotFound
	}
	return nil
}

func (r *ProductSupplierGormRepository) ClearPreferred(ctx context.Context, companyID, productID uuid.UUID, at time.Time) error {
	return r.db.GetDB(ctx).Model(&model.ProductSupplierModel{}).
		Where("company_id = ? AND product_id = ? AND preferred = true", companyID, productID).
		Updates(map[string]any{"preferred": false, "updated_at": at}).Error
}

func (r *ProductSupplierGormRepository) Delete(ctx context.Context, companyID, supplierID, productID uuid.UUID) error {
	result := r.db.GetDB(ctx).Where("company_id = ? AND supplier_id = ? AND product_id = ?", companyID, supplierID, productID).
		Delete(&model.ProductSupplierModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return service.ErrAssociationNotFound
	}
	return nil
}
