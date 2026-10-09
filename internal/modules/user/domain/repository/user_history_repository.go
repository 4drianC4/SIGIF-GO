package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type UserHistoryRepository interface {
	Create(ctx context.Context, history *entity.UserHistory) error
	ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entity.UserHistory, int64, error)
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
