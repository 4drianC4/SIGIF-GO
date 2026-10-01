package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/event"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func (h *UserCommandHandler) HandleChangeStatus(ctx context.Context, cmd command.ChangeStatus) error {
	var err error

	switch cmd.Action {
	case userEntity.UserStatusActive:
		err = h.service.Activate(ctx, cmd.ID)
	case userEntity.UserStatusInactive:
		err = h.service.Deactivate(ctx, cmd.ID)
	case userEntity.UserStatusSuspended:
		err = h.service.Suspend(ctx, cmd.ID)
	default:
		return sharedErrors.New(sharedErrors.CodeBadRequest, "invalid status action", 400)
	}

	if err != nil {
		return err
	}

	user, _ := h.service.GetByID(ctx, cmd.ID)
	if user == nil {
		return nil
	}

	switch cmd.Action {
	case userEntity.UserStatusActive:
		_ = h.eventBus.Publish(ctx, event.NewUserActivatedEvent(user))
	case userEntity.UserStatusInactive:
		_ = h.eventBus.Publish(ctx, event.NewUserDeactivatedEvent(user))
	case userEntity.UserStatusSuspended:
		_ = h.eventBus.Publish(ctx, event.NewUserSuspendedEvent(user))
	}

	switch cmd.Action {
	case userEntity.UserStatusActive:
		_ = h.eventBus.Publish(ctx, event.NewUserActivatedEvent(user))
	case userEntity.UserStatusInactive:
		_ = h.eventBus.Publish(ctx, event.NewUserDeactivatedEvent(user))
	case userEntity.UserStatusSuspended:
		_ = h.eventBus.Publish(ctx, event.NewUserSuspendedEvent(user))
	}

	return nil
}
