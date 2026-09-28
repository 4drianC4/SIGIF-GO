package event

import (
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/events"
)

type UserCreatedEvent struct {
	*events.BaseEvent
	User *entity.User
}

func NewUserCreatedEvent(user *entity.User) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseEvent: events.NewEvent("user.created", user),
		User:      user,
	}
}

type UserUpdatedEvent struct {
	*events.BaseEvent
	User *entity.User
}

func NewUserUpdatedEvent(user *entity.User) *UserUpdatedEvent {
	return &UserUpdatedEvent{
		BaseEvent: events.NewEvent("user.updated", user),
		User:      user,
	}
}

type UserDeletedEvent struct {
	*events.BaseEvent
	UserID uuid.UUID
}

func NewUserDeletedEvent(userID uuid.UUID) *UserDeletedEvent {
	return &UserDeletedEvent{
		BaseEvent: events.NewEvent("user.deleted", map[string]any{"user_id": userID}),
		UserID:    userID,
	}
}

type UserActivatedEvent struct {
	*events.BaseEvent
	User *entity.User
}

func NewUserActivatedEvent(user *entity.User) *UserActivatedEvent {
	return &UserActivatedEvent{
		BaseEvent: events.NewEvent("user.activated", user),
		User:      user,
	}
}

type UserDeactivatedEvent struct {
	*events.BaseEvent
	User *entity.User
}

func NewUserDeactivatedEvent(user *entity.User) *UserDeactivatedEvent {
	return &UserDeactivatedEvent{
		BaseEvent: events.NewEvent("user.deactivated", user),
		User:      user,
	}
}

type UserSuspendedEvent struct {
	*events.BaseEvent
	User *entity.User
}

func NewUserSuspendedEvent(user *entity.User) *UserSuspendedEvent {
	return &UserSuspendedEvent{
		BaseEvent: events.NewEvent("user.suspended", user),
		User:      user,
	}
}

type UserPasswordChangedEvent struct {
	*events.BaseEvent
	UserID uuid.UUID
}

func NewUserPasswordChangedEvent(userID uuid.UUID) *UserPasswordChangedEvent {
	return &UserPasswordChangedEvent{
		BaseEvent: events.NewEvent("user.password_changed", map[string]any{"user_id": userID}),
		UserID:    userID,
	}
}

type UserLoggedInEvent struct {
	*events.BaseEvent
	User *entity.User
}

func NewUserLoggedInEvent(user *entity.User) *UserLoggedInEvent {
	return &UserLoggedInEvent{
		BaseEvent: events.NewEvent("user.logged_in", user),
		User:      user,
	}
}