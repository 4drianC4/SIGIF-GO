package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/domain/event"
)

func (h *UserCommandHandler) HandleChangePassword(ctx context.Context, cmd command.ChangePassword) error {
	if err := h.service.ChangePassword(ctx, cmd.ID, cmd.CurrentPassword, cmd.NewPassword); err != nil {
		return err
	}

	_ = h.eventBus.Publish(ctx, event.NewUserPasswordChangedEvent(cmd.ID))
	return nil
}
