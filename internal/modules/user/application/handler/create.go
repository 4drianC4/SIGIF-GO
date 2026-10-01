package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/event"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
)

func (h *UserCommandHandler) HandleCreate(ctx context.Context, cmd command.CreateUser) (*entity.User, error) {
	user, err := h.service.Create(ctx, service.CreateUserParams{
		TenantID:  cmd.TenantID,
		Email:     cmd.Email,
		Password:  cmd.Password,
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
		Roles:     cmd.Roles,
	})
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewUserCreatedEvent(user))
	return user, nil
}
