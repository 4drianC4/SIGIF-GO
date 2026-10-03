package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// UserRepository is the persistence port for application users.
type UserRepository interface {
	Create(ctx context.Context, user *entity.AppUser) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AppUser, error)
	GetByEmail(ctx context.Context, email string) (*entity.AppUser, error)
	List(ctx context.Context, offset, limit int) ([]*entity.AppUser, int64, error)
	Update(ctx context.Context, user *entity.AppUser) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
