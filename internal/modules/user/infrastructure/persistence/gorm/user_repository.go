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
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	sharedMiddleware "github.com/sigif/sigif-go/internal/shared/middleware"
)

type UserGormRepository struct {
	db *sharedDatabase.Database
}

func NewUserGormRepository(db *sharedDatabase.Database) repository.UserRepository {
	return &UserGormRepository{db: db}
}

func (r *UserGormRepository) Create(ctx context.Context, user *entity.User) error {
	if user == nil || user.TenantID == uuid.Nil {
		return sharedErrors.ErrTenantRequired
	}
	return r.db.GetDB(ctx).Create(mapper.ToModel(user)).Error
}

func applyTenantScope(ctx context.Context, db *gorm.DB) (*gorm.DB, error) {
	if tenantID, ok := sharedMiddleware.TenantIDFromContext(ctx); ok {
		return db.Where("tenant_id = ?", tenantID), nil
	}
	return nil, sharedErrors.ErrTenantRequired
}

func (r *UserGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var m model.UserModel
	db, err := applyTenantScope(ctx, r.db.GetDB(ctx).Model(&model.UserModel{}))
	if err != nil {
		return nil, err
	}
	err = db.Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomain(&m), nil
}

func (r *UserGormRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error) {
	if tenantID == uuid.Nil {
		return nil, sharedErrors.ErrTenantRequired
	}
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
	if tenantID == uuid.Nil {
		return nil, 0, sharedErrors.ErrTenantRequired
	}
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
	db, err := applyTenantScope(ctx, r.db.GetDB(ctx).Model(&model.UserModel{}))
	if err != nil {
		return err
	}
	return db.Where("id = ?", user.ID).Updates(mapper.ToModel(user)).Error
}

func (r *UserGormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db, err := applyTenantScope(ctx, r.db.GetDB(ctx).Model(&model.UserModel{}))
	if err != nil {
		return err
	}
	return db.Where("id = ?", id).Delete(&model.UserModel{}).Error
}

func (r *UserGormRepository) ExistsByEmail(ctx context.Context, tenantID uuid.UUID, email string) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.UserModel{}).Where("tenant_id = ? AND email = ?", tenantID, email).Count(&count).Error
	return count > 0, err
}

func (r *UserGormRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	db, err := applyTenantScope(ctx, r.db.GetDB(ctx).Model(&model.UserModel{}))
	if err != nil {
		return false, err
	}
	err = db.Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
