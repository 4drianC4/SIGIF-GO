package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/clock"
)

type HistoryAction string

const (
	HistoryActionCreated         HistoryAction = "created"
	HistoryActionUpdated         HistoryAction = "updated"
	HistoryActionPasswordChanged HistoryAction = "password_changed"
	HistoryActionActivated       HistoryAction = "activated"
	HistoryActionDeactivated     HistoryAction = "deactivated"
	HistoryActionDeleted         HistoryAction = "deleted"
)

func (a HistoryAction) String() string {
	return string(a)
}

func (a HistoryAction) Description() string {
	switch a {
	case HistoryActionCreated:
		return "User created"
	case HistoryActionUpdated:
		return "User updated"
	case HistoryActionPasswordChanged:
		return "Password changed"
	case HistoryActionActivated:
		return "User activated"
	case HistoryActionDeactivated:
		return "User deactivated"
	case HistoryActionDeleted:
		return "User deleted"
	default:
		return ""
	}
}

type FieldChange struct {
	From *string
	To   *string
}

type HistoryChanges map[string]FieldChange

type HistoryActor struct {
	ID       uuid.UUID
	FullName string
	Email    string
}

type UserHistory struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Action      HistoryAction
	Description string
	Changes     HistoryChanges
	PerformedBy *HistoryActor
	CreatedAt   time.Time
}

func NewUserHistory(clock clock.Clock, userID uuid.UUID, action HistoryAction, changes HistoryChanges, actor *HistoryActor) *UserHistory {
	if changes == nil {
		changes = HistoryChanges{}
	}
	return &UserHistory{
		ID:          uuid.New(),
		UserID:      userID,
		Action:      action,
		Description: action.Description(),
		Changes:     changes,
		PerformedBy: actor,
		CreatedAt:   clock.NowUTC(),
	}
}
