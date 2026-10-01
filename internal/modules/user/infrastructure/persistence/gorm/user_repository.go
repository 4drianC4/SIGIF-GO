package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type UserGormRepository struct {
	db *sharedDatabase.Database
}

func NewUserGormRepository(db *sharedDatabase.Database) repository.UserRepository {
	return &UserGormRepository{db: db}
}

func (r *UserGormRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.GetDB(ctx).Create(mapper.ToModel(user)).Error
}

func (r *UserGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var m model.UserModel
	err := r.db.GetDB(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomain(&m), nil
}

func (r *UserGormRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error) {
	var m model.UserModel
	err := r.db.GetDB(ctx).Where("tenant_id = ? AND email = ?", tenantID, email).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomain(&m), nil
}

func (r *UserGormRepository) List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.User, int64, error) {
	var models []model.UserModel
	var total int64

	db := r.db.GetDB(ctx).Model(&model.UserModel{}).Where("tenant_id = ?", tenantID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*entity.User, len(models))
	for i, m := range models {
		users[i] = mapper.ToDomain(&m)
	}
	return users, total, nil
}

func (r *UserGormRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.GetDB(ctx).Save(mapper.ToModel(user)).Error
}

func (r *UserGormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.GetDB(ctx).Delete(&model.UserModel{}, "id = ?", id).Error
}

func (r *UserGormRepository) ExistsByEmail(ctx context.Context, tenantID uuid.UUID, email string) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.UserModel{}).Where("tenant_id = ? AND email = ?", tenantID, email).Count(&count).Error
	return count > 0, err
}

func (r *UserGormRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.UserModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}