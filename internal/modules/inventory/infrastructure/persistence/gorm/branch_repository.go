package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/inventory/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

const (
	defaultBranchName   = "Principal"
	defaultLocationName = "Almacén"
)

type BranchLocationGormRepository struct {
	db *sharedDatabase.Database
}

func NewBranchLocationGormRepository(db *sharedDatabase.Database) repository.BranchLocationRepository {
	return &BranchLocationGormRepository{db: db}
}

func (r *BranchLocationGormRepository) ListBranches(ctx context.Context, companyID uuid.UUID) ([]*entity.Branch, error) {
	db := r.db.GetDB(ctx)
	var branches []model.BranchModel
	if err := db.Where("company_id = ?", companyID).Order("id").Find(&branches).Error; err != nil {
		return nil, err
	}
	var locations []model.LocationModel
	if err := db.Where("company_id = ?", companyID).Order("id").Find(&locations).Error; err != nil {
		return nil, err
	}

	result := make([]*entity.Branch, len(branches))
	index := map[int64]int{}
	for i, b := range branches {
		result[i] = &entity.Branch{ID: b.ID, CompanyID: b.CompanyID, Name: b.Name}
		index[b.ID] = i
	}
	for _, l := range locations {
		if i, ok := index[l.BranchID]; ok {
			result[i].Locations = append(result[i].Locations, entity.Location{
				ID: l.ID, BranchID: l.BranchID, CompanyID: l.CompanyID, Name: l.Name,
			})
		}
	}
	return result, nil
}

func (r *BranchLocationGormRepository) EnsureDefault(ctx context.Context, companyID uuid.UUID) (int64, int64, error) {
	db := r.db.GetDB(ctx)

	var branch model.BranchModel
	err := db.Where("company_id = ?", companyID).Order("id").First(&branch).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		branch = model.BranchModel{CompanyID: companyID, Name: defaultBranchName}
		if err := db.Create(&branch).Error; err != nil {
			return 0, 0, err
		}
	} else if err != nil {
		return 0, 0, err
	}

	var location model.LocationModel
	err = db.Where("branch_id = ?", branch.ID).Order("id").First(&location).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		location = model.LocationModel{BranchID: branch.ID, CompanyID: companyID, Name: defaultLocationName}
		if err := db.Create(&location).Error; err != nil {
			return 0, 0, err
		}
	} else if err != nil {
		return 0, 0, err
	}
	return branch.ID, location.ID, nil
}
