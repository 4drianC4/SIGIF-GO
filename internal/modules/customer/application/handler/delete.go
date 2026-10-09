package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
)

func (h *CustomerCommandHandler) HandleDelete(ctx context.Context, cmd command.DeleteCustomer) error {
	return h.service.Delete(ctx, cmd.CompanyID, cmd.ID)
}
