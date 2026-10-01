package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/auth/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/auth/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type RefreshTokenGormRepository struct {
	db *sharedDatabase.Database
}

func NewRefreshTokenGormRepository(db *sharedDatabase.Database) repository.RefreshTokenRepository {
	return &RefreshTokenGormRepository{db: db}
}

func (r *RefreshTokenGormRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	return r.db.GetDB(ctx).Create(mapper.ToModel(token)).Error
}

func (r *RefreshTokenGormRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var m model.RefreshTokenModel
	err := r.db.GetDB(ctx).Where("token_hash = ?", tokenHash).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomain(&m), nil
}

func (r *RefreshTokenGormRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.RefreshToken, error) {
	var models []model.RefreshTokenModel
	if err := r.db.GetDB(ctx).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}
	return mapper.ToDomainList(models), nil
}

func (r *RefreshTokenGormRepository) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	return r.db.GetDB(ctx).Model(&model.RefreshTokenModel{}).Where("id = ?", id).Update("revoked_at", revokedAt).Error
}

func (r *RefreshTokenGormRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	return r.db.GetDB(ctx).Model(&model.RefreshTokenModel{}).Where("user_id = ?", userID).Update("revoked_at", revokedAt).Error
}

func (r *RefreshTokenGormRepository) DeleteExpired(ctx context.Context) error {
	return r.db.GetDB(ctx).Where("expires_at < ?", time.Now()).Delete(&model.RefreshTokenModel{}).Error
}
