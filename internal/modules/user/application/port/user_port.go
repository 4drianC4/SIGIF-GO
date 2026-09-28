package port

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type UserCommandPort interface {
	Create(ctx context.Context, tenantID uuid.UUID, companyID *uuid.UUID, email, password, firstName, lastName string, roles []entity.UserRole) (*entity.User, error)
	Update(ctx context.Context, id uuid.UUID, firstName, lastName, phone, avatarURL string, roles []entity.UserRole, settings entity.UserSettings) (*entity.User, error)
	ChangePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error
	Delete(ctx context.Context, id uuid.UUID) error
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	Suspend(ctx context.Context, id uuid.UUID) error
	RecordLogin(ctx context.Context, id uuid.UUID) error
}

type UserQueryPort interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.User, error)
	GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*entity.User, error)
	List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.User, int64, error)
}