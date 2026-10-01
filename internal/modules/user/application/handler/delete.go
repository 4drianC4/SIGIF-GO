package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/domain/event"
)

func (h *UserCommandHandler) HandleDelete(ctx context.Context, cmd command.DeleteUser) error {
	if err := h.service.Delete(ctx, cmd.ID); err != nil {
		return err
	}

	_ = h.eventBus.Publish(ctx, event.NewUserDeletedEvent(cmd.ID))
	return nil
}
