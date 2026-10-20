package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/cashregister/application/command"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
)

func (h *CashRegisterCommandHandler) HandleAssign(ctx context.Context, cmd command.AssignCashRegister) (*service.AssignmentView, error) {
	return h.service.Assign(ctx, cmd.CompanyID, cmd.UserID, cmd.CashRegisterID)
}
