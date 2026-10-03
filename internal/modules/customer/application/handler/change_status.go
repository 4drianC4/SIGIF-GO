package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/event"
)

// HandleChangeStatus procesa el cambio de estado de un cliente y publica un evento.
func (h *CustomerCommandHandler) HandleChangeStatus(ctx context.Context, cmd command.ChangeCustomerStatus) (*entity.Customer, error) {
	customer, err := h.service.ChangeStatus(ctx, cmd.TenantID, cmd.ID, cmd.Status)
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewCustomerStatusChangedEvent(customer))
	return customer, nil
}
