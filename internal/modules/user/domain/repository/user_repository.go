package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.User, error)
	GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*entity.User, error)
	List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.User, int64, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	RecordLogin(ctx context.Context, userID uuid.UUID) error
}