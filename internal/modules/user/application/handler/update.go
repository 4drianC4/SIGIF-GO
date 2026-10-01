package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/event"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
)

func (h *UserCommandHandler) HandleUpdate(ctx context.Context, cmd command.UpdateUser) (*entity.User, error) {
	user, err := h.service.Update(ctx, cmd.ID, service.UpdateUserParams{
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
		Phone:     cmd.Phone,
		AvatarURL: cmd.AvatarURL,
		Roles:     cmd.Roles,
		Settings:  cmd.Settings,
	})
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewUserUpdatedEvent(user))
	return user, nil
}
