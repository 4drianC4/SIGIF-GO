package gorm

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/database"
)

type TenantGormRepository struct {
	db *database.Database
}

func NewTenantGormRepository(db *database.Database) *TenantGormRepository {
	return &TenantGormRepository{db: db}
}

func (r *TenantGormRepository) Create(ctx context.Context, tenant *entity.Tenant) error {
	return r.db.GetDB(ctx).Create(tenant).Error
}

func (r *TenantGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.db.GetDB(ctx).Where("id = ?", id).First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantGormRepository) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.db.GetDB(ctx).Where("slug = ?", slug).First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantGormRepository) GetByBusinessType(ctx context.Context, businessType entity.BusinessType) ([]*entity.Tenant, error) {
	var tenants []*entity.Tenant
	err := r.db.GetDB(ctx).Where("business_type = ?", businessType).Find(&tenants).Error
	return tenants, err
}

func (r *TenantGormRepository) List(ctx context.Context, offset, limit int) ([]*entity.Tenant, int64, error) {
	var tenants []*entity.Tenant
	var total int64

	db := r.db.GetDB(ctx).Model(&entity.Tenant{})
	db.Count(&total)

	err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tenants).Error
	return tenants, total, err
}

func (r *TenantGormRepository) Update(ctx context.Context, tenant *entity.Tenant) error {
	return r.db.GetDB(ctx).Save(tenant).Error
}

func (r *TenantGormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.GetDB(ctx).Delete(&entity.Tenant{}, "id = ?", id).Error
}

func (r *TenantGormRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&entity.Tenant{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *TenantGormRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&entity.Tenant{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

var _ repository.TenantRepository = (*TenantGormRepository)(nil)