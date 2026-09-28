package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/entity"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *entity.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error)
	GetByBusinessType(ctx context.Context, businessType entity.BusinessType) ([]*entity.Tenant, error)
	List(ctx context.Context, offset, limit int) ([]*entity.Tenant, int64, error)
	Update(ctx context.Context, tenant *entity.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}