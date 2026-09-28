package port

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/company/application/command"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
)

type CompanyCommandPort interface {
	HandleCreate(ctx context.Context, cmd command.CreateCompanyCommand) (*entity.Company, error)
	HandleUpdate(ctx context.Context, cmd command.UpdateCompanyCommand) (*entity.Company, error)
	HandleDelete(ctx context.Context, cmd command.DeleteCompanyCommand) error
	HandleActivate(ctx context.Context, cmd command.ActivateCompanyCommand) error
	HandleDeactivate(ctx context.Context, cmd command.DeactivateCompanyCommand) error
}

type CompanyQueryPort interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Company, error)
	List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.Company, int64, error)
}