package gorm

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/inventory/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/inventory/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type MovementGormRepository struct {
	db *sharedDatabase.Database
}

func NewMovementGormRepository(db *sharedDatabase.Database) repository.MovementRepository {
	return &MovementGormRepository{db: db}
}

func (r *MovementGormRepository) Create(ctx context.Context, m *entity.Movement) error {
	return r.db.GetDB(ctx).Create(mapper.MovementToModel(m)).Error
}

func (r *MovementGormRepository) ListByProduct(ctx context.Context, companyID, productID uuid.UUID, offset, limit int) ([]*entity.Movement, int64, error) {
	db := r.db.GetDB(ctx).Model(&model.MovementModel{}).
		Where("company_id = ? AND product_id = ?", companyID, productID)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []model.MovementModel
	if err := db.Order("created_at DESC, id").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*entity.Movement, len(rows))
	for i := range rows {
		items[i] = mapper.MovementToDomain(&rows[i])
	}
	return items, total, nil
}
