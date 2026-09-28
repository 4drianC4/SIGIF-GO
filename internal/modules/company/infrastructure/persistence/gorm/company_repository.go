package gorm

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/company/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/database"
)

type CompanyGormRepository struct {
	db *database.Database
}

func NewCompanyGormRepository(db *database.Database) *CompanyGormRepository {
	return &CompanyGormRepository{db: db}
}

func (r *CompanyGormRepository) Create(ctx context.Context, company *entity.Company) error {
	return r.db.GetDB(ctx).Create(company).Error
}

func (r *CompanyGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error) {
	var company entity.Company
	err := r.db.GetDB(ctx).Where("id = ?", id).First(&company).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &company, nil
}

func (r *CompanyGormRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Company, error) {
	var companies []*entity.Company
	err := r.db.GetDB(ctx).Where("tenant_id = ?", tenantID).Find(&companies).Error
	return companies, err
}

func (r *CompanyGormRepository) GetByTaxID(ctx context.Context, taxID string) (*entity.Company, error) {
	var company entity.Company
	err := r.db.GetDB(ctx).Where("tax_id = ?", taxID).First(&company).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &company, nil
}

func (r *CompanyGormRepository) List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.Company, int64, error) {
	var companies []*entity.Company
	var total int64

	db := r.db.GetDB(ctx).Model(&entity.Company{}).Where("tenant_id = ?", tenantID)
	db.Count(&total)

	err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&companies).Error
	return companies, total, err
}

func (r *CompanyGormRepository) Update(ctx context.Context, company *entity.Company) error {
	return r.db.GetDB(ctx).Save(company).Error
}

func (r *CompanyGormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.GetDB(ctx).Delete(&entity.Company{}, "id = ?", id).Error
}

func (r *CompanyGormRepository) ExistsByTaxID(ctx context.Context, taxID string) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&entity.Company{}).Where("tax_id = ?", taxID).Count(&count).Error
	return count > 0, err
}

var _ repository.CompanyRepository = (*CompanyGormRepository)(nil)