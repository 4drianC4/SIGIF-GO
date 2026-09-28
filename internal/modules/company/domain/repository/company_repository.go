package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
)

type CompanyRepository interface {
	Create(ctx context.Context, company *entity.Company) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Company, error)
	GetByTaxID(ctx context.Context, taxID string) (*entity.Company, error)
	List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.Company, int64, error)
	Update(ctx context.Context, company *entity.Company) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByTaxID(ctx context.Context, taxID string) (bool, error)
}