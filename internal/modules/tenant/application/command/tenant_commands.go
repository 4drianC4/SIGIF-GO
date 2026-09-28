package command

import (
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/entity"
)

type CreateTenantCommand struct {
	Name         string
	Slug         string
	BusinessType entity.BusinessType
}

type UpdateTenantCommand struct {
	ID       uuid.UUID
	Name     string
	Settings entity.TenantSettings
}

type DeleteTenantCommand struct {
	ID uuid.UUID
}

type ActivateTenantCommand struct {
	ID uuid.UUID
}

type DeactivateTenantCommand struct {
	ID uuid.UUID
}