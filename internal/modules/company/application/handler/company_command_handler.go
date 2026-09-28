package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/company/application/command"
	"github.com/sigif/sigif-go/internal/modules/company/application/port"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/company/domain/event"
	"github.com/sigif/sigif-go/internal/modules/company/domain/service"
	"github.com/sigif/sigif-go/internal/shared/events"
)

type CompanyCommandHandler struct {
	service  *service.CompanyService
	eventBus *events.Bus
}

func NewCompanyCommandHandler(service *service.CompanyService, eventBus *events.Bus) *CompanyCommandHandler {
	return &CompanyCommandHandler{
		service:  service,
		eventBus: eventBus,
	}
}

func (h *CompanyCommandHandler) HandleCreate(ctx context.Context, cmd command.CreateCompanyCommand) (*entity.Company, error) {
	company, err := h.service.Create(ctx, cmd.TenantID, cmd.Name, cmd.LegalName, cmd.TaxID)
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewCompanyCreatedEvent(company))
	return company, nil
}

func (h *CompanyCommandHandler) HandleUpdate(ctx context.Context, cmd command.UpdateCompanyCommand) (*entity.Company, error) {
	company, err := h.service.Update(ctx, cmd.ID, cmd.Name, cmd.LegalName, cmd.TaxID, cmd.Email, cmd.Phone, cmd.Address, cmd.City, cmd.State, cmd.Country, cmd.PostalCode, cmd.Settings)
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewCompanyUpdatedEvent(company))
	return company, nil
}

func (h *CompanyCommandHandler) HandleDelete(ctx context.Context, cmd command.DeleteCompanyCommand) error {
	if err := h.service.Delete(ctx, cmd.ID); err != nil {
		return err
	}

	_ = h.eventBus.Publish(ctx, event.NewCompanyDeletedEvent(cmd.ID))
	return nil
}

func (h *CompanyCommandHandler) HandleActivate(ctx context.Context, cmd command.ActivateCompanyCommand) error {
	if err := h.service.Activate(ctx, cmd.ID); err != nil {
		return err
	}

	company, _ := h.service.GetByID(ctx, cmd.ID)
	if company != nil {
		_ = h.eventBus.Publish(ctx, event.NewCompanyActivatedEvent(company))
	}
	return nil
}

func (h *CompanyCommandHandler) HandleDeactivate(ctx context.Context, cmd command.DeactivateCompanyCommand) error {
	if err := h.service.Deactivate(ctx, cmd.ID); err != nil {
		return err
	}

	company, _ := h.service.GetByID(ctx, cmd.ID)
	if company != nil {
		_ = h.eventBus.Publish(ctx, event.NewCompanyDeactivatedEvent(company))
	}
	return nil
}

func (h *CompanyCommandHandler) Activate(ctx context.Context, id uuid.UUID) error {
	return h.HandleActivate(ctx, command.ActivateCompanyCommand{ID: id})
}

func (h *CompanyCommandHandler) Deactivate(ctx context.Context, id uuid.UUID) error {
	return h.HandleDeactivate(ctx, command.DeactivateCompanyCommand{ID: id})
}

func (h *CompanyCommandHandler) Create(ctx context.Context, tenantID uuid.UUID, name, legalName, taxID string) (*entity.Company, error) {
	return h.HandleCreate(ctx, command.CreateCompanyCommand{TenantID: tenantID, Name: name, LegalName: legalName, TaxID: taxID})
}

func (h *CompanyCommandHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.HandleDelete(ctx, command.DeleteCompanyCommand{ID: id})
}

var _ port.CompanyCommandPort = (*CompanyCommandHandler)(nil)