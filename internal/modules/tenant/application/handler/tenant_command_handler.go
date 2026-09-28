package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/tenant/application/command"
	"github.com/sigif/sigif-go/internal/modules/tenant/application/port"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/event"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/service"
	"github.com/sigif/sigif-go/internal/shared/events"
)

type TenantCommandHandler struct {
	service  *service.TenantService
	eventBus *events.Bus
}

func NewTenantCommandHandler(service *service.TenantService, eventBus *events.Bus) *TenantCommandHandler {
	return &TenantCommandHandler{
		service:  service,
		eventBus: eventBus,
	}
}

func (h *TenantCommandHandler) HandleCreate(ctx context.Context, cmd command.CreateTenantCommand) (*entity.Tenant, error) {
	tenant, err := h.service.Create(ctx, cmd.Name, cmd.Slug, cmd.BusinessType)
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewTenantCreatedEvent(tenant))
	return tenant, nil
}

func (h *TenantCommandHandler) HandleUpdate(ctx context.Context, cmd command.UpdateTenantCommand) (*entity.Tenant, error) {
	tenant, err := h.service.Update(ctx, cmd.ID, cmd.Name, cmd.Settings)
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewTenantUpdatedEvent(tenant))
	return tenant, nil
}

func (h *TenantCommandHandler) HandleDelete(ctx context.Context, cmd command.DeleteTenantCommand) error {
	if err := h.service.Delete(ctx, cmd.ID); err != nil {
		return err
	}

	_ = h.eventBus.Publish(ctx, event.NewTenantDeletedEvent(cmd.ID))
	return nil
}

func (h *TenantCommandHandler) HandleActivate(ctx context.Context, cmd command.ActivateTenantCommand) error {
	if err := h.service.Activate(ctx, cmd.ID); err != nil {
		return err
	}

	tenant, _ := h.service.GetByID(ctx, cmd.ID)
	if tenant != nil {
		_ = h.eventBus.Publish(ctx, event.NewTenantActivatedEvent(tenant))
	}
	return nil
}

func (h *TenantCommandHandler) HandleDeactivate(ctx context.Context, cmd command.DeactivateTenantCommand) error {
	if err := h.service.Deactivate(ctx, cmd.ID); err != nil {
		return err
	}

	tenant, _ := h.service.GetByID(ctx, cmd.ID)
	if tenant != nil {
		_ = h.eventBus.Publish(ctx, event.NewTenantDeactivatedEvent(tenant))
	}
	return nil
}

func (h *TenantCommandHandler) Activate(ctx context.Context, id uuid.UUID) error {
	return h.HandleActivate(ctx, command.ActivateTenantCommand{ID: id})
}

func (h *TenantCommandHandler) Deactivate(ctx context.Context, id uuid.UUID) error {
	return h.HandleDeactivate(ctx, command.DeactivateTenantCommand{ID: id})
}

func (h *TenantCommandHandler) Create(ctx context.Context, name, slug string, businessType entity.BusinessType) (*entity.Tenant, error) {
	return h.HandleCreate(ctx, command.CreateTenantCommand{Name: name, Slug: slug, BusinessType: businessType})
}

func (h *TenantCommandHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.HandleDelete(ctx, command.DeleteTenantCommand{ID: id})
}

var _ port.TenantCommandPort = (*TenantCommandHandler)(nil)