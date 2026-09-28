package port

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/tenant/application/command"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/entity"
)

type TenantCommandPort interface {
	HandleCreate(ctx context.Context, cmd command.CreateTenantCommand) (*entity.Tenant, error)
	HandleUpdate(ctx context.Context, cmd command.UpdateTenantCommand) (*entity.Tenant, error)
	HandleDelete(ctx context.Context, cmd command.DeleteTenantCommand) error
	HandleActivate(ctx context.Context, cmd command.ActivateTenantCommand) error
	HandleDeactivate(ctx context.Context, cmd command.DeactivateTenantCommand) error
}

type TenantQueryPort interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error)
	List(ctx context.Context, offset, limit int) ([]*entity.Tenant, int64, error)
}