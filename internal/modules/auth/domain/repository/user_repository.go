package repository

import (
	"context"

	"github.com/google/uuid"

	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// UserRepo define las operaciones que el módulo auth necesita de los usuarios.
type UserRepo interface {
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*userEntity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error)
	RecordLogin(ctx context.Context, userID uuid.UUID) error
}