package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}