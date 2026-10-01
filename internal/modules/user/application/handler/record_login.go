package handler

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/event"
)

func (h *UserCommandHandler) HandleRecordLogin(ctx context.Context, id uuid.UUID) error {
	if err := h.service.RecordLogin(ctx, id); err != nil {
		return err
	}

	user, _ := h.service.GetByID(ctx, id)
	if user != nil {
		_ = h.eventBus.Publish(ctx, event.NewUserLoggedInEvent(user))
	}
	return nil
}
