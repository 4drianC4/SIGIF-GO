package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// Reader contiene las operaciones de lectura sobre usuarios.
type Reader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error)
	List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.User, int64, error)
	ExistsByEmail(ctx context.Context, tenantID uuid.UUID, email string) (bool, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}

// Writer contiene las operaciones de escritura sobre usuarios.
type Writer interface {
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserRepository es el puerto completo de persistencia de usuarios.
type UserRepository interface {
	Reader
	Writer
}