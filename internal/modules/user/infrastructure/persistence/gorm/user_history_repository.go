package gorm

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type UserHistoryGormRepository struct {
	db *sharedDatabase.Database
}

func NewUserHistoryGormRepository(db *sharedDatabase.Database) repository.UserHistoryRepository {
	return &UserHistoryGormRepository{db: db}
}

func (r *UserHistoryGormRepository) Create(ctx context.Context, history *entity.UserHistory) error {
	return r.db.GetDB(ctx).Create(mapper.UserHistoryToModel(history)).Error
}

func (r *UserHistoryGormRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entity.UserHistory, int64, error) {
	var total int64
	if err := r.db.GetDB(ctx).Model(&model.UserHistoryModel{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []model.UserHistoryModel
	err := r.db.GetDB(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Offset(offset).
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	history := make([]*entity.UserHistory, len(models))
	for i := range models {
		history[i] = mapper.UserHistoryToDomain(&models[i])
	}
	return history, total, nil
}
