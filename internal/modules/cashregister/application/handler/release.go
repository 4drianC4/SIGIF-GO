package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/cashregister/application/command"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
)

func (h *CashRegisterCommandHandler) HandleRelease(ctx context.Context, cmd command.ReleaseAssignment) (*service.AssignmentView, error) {
	return h.service.Release(ctx, service.ReleaseParams{
		CompanyID:     cmd.CompanyID,
		ActorID:       cmd.ActorID,
		AssignmentID:  cmd.AssignmentID,
		CanReleaseAny: cmd.CanReleaseAny,
	})
}
