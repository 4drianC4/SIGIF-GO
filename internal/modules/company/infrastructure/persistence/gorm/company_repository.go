package gorm

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/company/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/company/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/database"
)

type CompanyGormRepository struct{ db *database.Database }

func NewCompanyGormRepository(db *database.Database) repository.CompanyRepository {
	return &CompanyGormRepository{db: db}
}
func (r *CompanyGormRepository) List(ctx context.Context) ([]entity.Company, error) {
	var rows []model.CompanyModel
	if err := r.db.GetDB(ctx).Select("company_id", "legal_name", "trade_name").Order("legal_name ASC, company_id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]entity.Company, len(rows))
	for i, row := range rows {
		result[i] = entity.Company{ID: row.ID, LegalName: row.LegalName, TradeName: row.TradeName}
	}
	return result, nil
}
func (r *CompanyGormRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.CompanyModel{}).Where("company_id = ?", id).Count(&count).Error
	return count > 0, err
}
