package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/event"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
)

func (h *CustomerCommandHandler) HandleUpdate(ctx context.Context, cmd command.UpdateCustomer) (*entity.Customer, error) {
	customer, err := h.service.Update(ctx, service.UpdateCustomerParams{
		ID:             cmd.ID,
		TenantID:       cmd.TenantID,
		LegalName:      cmd.LegalName,
		DocumentType:   cmd.DocumentType,
		DocumentNumber: cmd.DocumentNumber,
		Phone:          cmd.Phone,
		Email:          cmd.Email,
		Address:        cmd.Address,
	})
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewCustomerUpdatedEvent(customer))
	return customer, nil
}
