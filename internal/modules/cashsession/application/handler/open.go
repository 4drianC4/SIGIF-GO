package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/cashsession/application/command"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/service"
)

func (h *CashSessionCommandHandler) HandleOpen(ctx context.Context, cmd command.OpenCashSession) (*service.SessionView, error) {
	return h.service.Open(ctx, service.OpenParams{
		CompanyID:      cmd.CompanyID,
		UserID:         cmd.UserID,
		CashRegisterID: cmd.CashRegisterID,
		OpeningAmount:  cmd.OpeningAmount,
		OpeningNotes:   cmd.OpeningNotes,
	})
}
