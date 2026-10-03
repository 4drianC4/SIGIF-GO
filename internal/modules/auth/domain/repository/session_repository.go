package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
)

// SessionRepository is the persistence port for user sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *entity.UserSession) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.UserSession, error)
	IsActive(ctx context.Context, tokenHash string) (bool, error)
	End(ctx context.Context, id uuid.UUID, endedAt time.Time, reason entity.SessionCloseReason) error
	EndAllByUserID(ctx context.Context, userID uuid.UUID, endedAt time.Time, reason entity.SessionCloseReason) error
}
