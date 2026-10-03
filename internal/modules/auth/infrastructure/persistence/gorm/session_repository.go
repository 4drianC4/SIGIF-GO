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

type SessionGormRepository struct {
	db *sharedDatabase.Database
}

func NewSessionGormRepository(db *sharedDatabase.Database) repository.SessionRepository {
	return &SessionGormRepository{db: db}
}

func (r *SessionGormRepository) Create(ctx context.Context, session *entity.UserSession) error {
	return r.db.GetDB(ctx).Create(mapper.SessionToModel(session)).Error
}

func (r *SessionGormRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.UserSession, error) {
	var m model.SessionModel
	err := r.db.GetDB(ctx).Where("token_hash = ?", tokenHash).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.SessionToDomain(&m), nil
}

func (r *SessionGormRepository) IsActive(ctx context.Context, tokenHash string) (bool, error) {
	var m model.SessionModel
	err := r.db.GetDB(ctx).
		Where("token_hash = ? AND ended_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *SessionGormRepository) End(ctx context.Context, id uuid.UUID, endedAt time.Time, reason entity.SessionCloseReason) error {
	return r.db.GetDB(ctx).Model(&model.SessionModel{}).
		Where("id = ? AND ended_at IS NULL", id).
		Updates(map[string]any{"ended_at": endedAt, "close_reason": string(reason)}).Error
}

func (r *SessionGormRepository) EndAllByUserID(ctx context.Context, userID uuid.UUID, endedAt time.Time, reason entity.SessionCloseReason) error {
	return r.db.GetDB(ctx).Model(&model.SessionModel{}).
		Where("user_id = ? AND ended_at IS NULL", userID).
		Updates(map[string]any{"ended_at": endedAt, "close_reason": string(reason)}).Error
}
