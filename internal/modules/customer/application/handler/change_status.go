package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

func (h *CustomerCommandHandler) HandleChangeStatus(ctx context.Context, cmd command.ChangeCustomerStatus) (*entity.Customer, error) {
	return h.service.ChangeStatus(ctx, cmd.CompanyID, cmd.ID, cmd.Status)
}
