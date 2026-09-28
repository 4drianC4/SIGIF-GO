package gorm

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/database"
)

type RefreshTokenGormRepository struct {
	db *database.Database
}

func NewRefreshTokenGormRepository(db *database.Database) *RefreshTokenGormRepository {
	return &RefreshTokenGormRepository{db: db}
}

func (r *RefreshTokenGormRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	return r.db.GetDB(ctx).Create(token).Error
}

func (r *RefreshTokenGormRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	err := r.db.GetDB(ctx).Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *RefreshTokenGormRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.RefreshToken, error) {
	var tokens []*entity.RefreshToken
	err := r.db.GetDB(ctx).Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

func (r *RefreshTokenGormRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	return r.db.GetDB(ctx).Model(&entity.RefreshToken{}).Where("id = ?", id).Update("revoked_at", "NOW()").Error
}

func (r *RefreshTokenGormRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.GetDB(ctx).Model(&entity.RefreshToken{}).Where("user_id = ?", userID).Update("revoked_at", "NOW()").Error
}

func (r *RefreshTokenGormRepository) DeleteExpired(ctx context.Context) error {
	return r.db.GetDB(ctx).Where("expires_at < NOW()").Delete(&entity.RefreshToken{}).Error
}

var _ repository.RefreshTokenRepository = (*RefreshTokenGormRepository)(nil)