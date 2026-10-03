package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/event"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
)

// HandleCreate procesa el registro de un nuevo cliente y publica un evento.
func (h *CustomerCommandHandler) HandleCreate(ctx context.Context, cmd command.CreateCustomer) (*entity.Customer, error) {
	customer, err := h.service.Create(ctx, service.CreateCustomerParams{
		TenantID:       cmd.TenantID,
		LegalName:      cmd.LegalName,
		DocumentType:   cmd.DocumentType,
		DocumentNumber: cmd.DocumentNumber,
		Phone:          cmd.Phone,
		Email:          cmd.Email,
	})
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewCustomerCreatedEvent(customer))
	return customer, nil
}
