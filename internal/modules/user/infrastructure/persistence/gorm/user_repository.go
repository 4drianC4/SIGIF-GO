package gorm

import (
	"context"
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/database"
)

type UserGormRepository struct {
	db *database.Database
}

func NewUserGormRepository(db *database.Database) *UserGormRepository {
	return &UserGormRepository{db: db}
}

func (r *UserGormRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.GetDB(ctx).Create(user).Error
}

func (r *UserGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.GetDB(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserGormRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.GetDB(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserGormRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.User, error) {
	var users []*entity.User
	err := r.db.GetDB(ctx).Where("tenant_id = ?", tenantID).Find(&users).Error
	return users, err
}

func (r *UserGormRepository) GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*entity.User, error) {
	var users []*entity.User
	err := r.db.GetDB(ctx).Where("company_id = ?", companyID).Find(&users).Error
	return users, err
}

func (r *UserGormRepository) List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	db := r.db.GetDB(ctx).Model(&entity.User{}).Where("tenant_id = ?", tenantID)
	db.Count(&total)

	err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	return users, total, err
}

func (r *UserGormRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.GetDB(ctx).Save(user).Error
}

func (r *UserGormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.GetDB(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

func (r *UserGormRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserGormRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&entity.User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *UserGormRepository) RecordLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.GetDB(ctx).Model(&entity.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login_at": now,
		"updated_at":    now,
	}).Error
}

var _ repository.UserRepository = (*UserGormRepository)(nil)