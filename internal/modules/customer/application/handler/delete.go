package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/event"
)

func (h *CustomerCommandHandler) HandleDelete(ctx context.Context, cmd command.DeleteCustomer) error {
	if err := h.service.Delete(ctx, cmd.TenantID, cmd.ID); err != nil {
		return err
	}

	_ = h.eventBus.Publish(ctx, event.NewCustomerDeletedEvent(cmd.ID))
	return nil
}
