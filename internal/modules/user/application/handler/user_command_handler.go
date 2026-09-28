package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/application/port"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/event"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
	"github.com/sigif/sigif-go/internal/shared/events"
)

type UserCommandHandler struct {
	service  *service.UserService
	eventBus *events.Bus
}

func NewUserCommandHandler(service *service.UserService, eventBus *events.Bus) *UserCommandHandler {
	return &UserCommandHandler{
		service:  service,
		eventBus: eventBus,
	}
}

func (h *UserCommandHandler) HandleCreate(ctx context.Context, cmd command.CreateUserCommand) (*entity.User, error) {
	user, err := h.service.Create(ctx, cmd.TenantID, cmd.CompanyID, cmd.Email, cmd.Password, cmd.FirstName, cmd.LastName, cmd.Roles)
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewUserCreatedEvent(user))
	return user, nil
}

func (h *UserCommandHandler) HandleUpdate(ctx context.Context, cmd command.UpdateUserCommand) (*entity.User, error) {
	user, err := h.service.Update(ctx, cmd.ID, cmd.FirstName, cmd.LastName, cmd.Phone, cmd.AvatarURL, cmd.Roles, cmd.Settings)
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewUserUpdatedEvent(user))
	return user, nil
}

func (h *UserCommandHandler) HandleChangePassword(ctx context.Context, cmd command.ChangePasswordCommand) error {
	if err := h.service.ChangePassword(ctx, cmd.ID, cmd.CurrentPassword, cmd.NewPassword); err != nil {
		return err
	}

	_ = h.eventBus.Publish(ctx, event.NewUserPasswordChangedEvent(cmd.ID))
	return nil
}

func (h *UserCommandHandler) HandleDelete(ctx context.Context, cmd command.DeleteUserCommand) error {
	if err := h.service.Delete(ctx, cmd.ID); err != nil {
		return err
	}

	_ = h.eventBus.Publish(ctx, event.NewUserDeletedEvent(cmd.ID))
	return nil
}

func (h *UserCommandHandler) HandleActivate(ctx context.Context, cmd command.ActivateUserCommand) error {
	if err := h.service.Activate(ctx, cmd.ID); err != nil {
		return err
	}

	user, _ := h.service.GetByID(ctx, cmd.ID)
	if user != nil {
		_ = h.eventBus.Publish(ctx, event.NewUserActivatedEvent(user))
	}
	return nil
}

func (h *UserCommandHandler) HandleDeactivate(ctx context.Context, cmd command.DeactivateUserCommand) error {
	if err := h.service.Deactivate(ctx, cmd.ID); err != nil {
		return err
	}

	user, _ := h.service.GetByID(ctx, cmd.ID)
	if user != nil {
		_ = h.eventBus.Publish(ctx, event.NewUserDeactivatedEvent(user))
	}
	return nil
}

func (h *UserCommandHandler) HandleSuspend(ctx context.Context, cmd command.SuspendUserCommand) error {
	if err := h.service.Suspend(ctx, cmd.ID); err != nil {
		return err
	}

	user, _ := h.service.GetByID(ctx, cmd.ID)
	if user != nil {
		_ = h.eventBus.Publish(ctx, event.NewUserSuspendedEvent(user))
	}
	return nil
}

func (h *UserCommandHandler) HandleRecordLogin(ctx context.Context, cmd command.RecordLoginCommand) error {
	if err := h.service.RecordLogin(ctx, cmd.ID); err != nil {
		return err
	}

	user, _ := h.service.GetByID(ctx, cmd.ID)
	if user != nil {
		_ = h.eventBus.Publish(ctx, event.NewUserLoggedInEvent(user))
	}
	return nil
}

func (h *UserCommandHandler) Activate(ctx context.Context, id uuid.UUID) error {
	return h.HandleActivate(ctx, command.ActivateUserCommand{ID: id})
}

func (h *UserCommandHandler) Deactivate(ctx context.Context, id uuid.UUID) error {
	return h.HandleDeactivate(ctx, command.DeactivateUserCommand{ID: id})
}

func (h *UserCommandHandler) Suspend(ctx context.Context, id uuid.UUID) error {
	return h.HandleSuspend(ctx, command.SuspendUserCommand{ID: id})
}

func (h *UserCommandHandler) RecordLogin(ctx context.Context, id uuid.UUID) error {
	return h.HandleRecordLogin(ctx, command.RecordLoginCommand{ID: id})
}

func (h *UserCommandHandler) ChangePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error {
	return h.HandleChangePassword(ctx, command.ChangePasswordCommand{ID: id, CurrentPassword: currentPassword, NewPassword: newPassword})
}

func (h *UserCommandHandler) Create(ctx context.Context, tenantID uuid.UUID, companyID *uuid.UUID, email, password, firstName, lastName string, roles []entity.UserRole) (*entity.User, error) {
	return h.HandleCreate(ctx, command.CreateUserCommand{
		TenantID:  tenantID,
		CompanyID: companyID,
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Roles:     roles,
	})
}

func (h *UserCommandHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.HandleDelete(ctx, command.DeleteUserCommand{ID: id})
}

func (h *UserCommandHandler) Update(ctx context.Context, id uuid.UUID, firstName, lastName, phone, avatarURL string, roles []entity.UserRole, settings entity.UserSettings) (*entity.User, error) {
	return h.HandleUpdate(ctx, command.UpdateUserCommand{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		AvatarURL: avatarURL,
		Roles:     roles,
		Settings:  settings,
	})
}

var _ port.UserCommandPort = (*UserCommandHandler)(nil)