package dto

import (
	"time"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type HistoryFieldChange struct {
	From *string `json:"from"`
	To   *string `json:"to"`
}

type HistoryActor struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type UserHistory struct {
	ID          string                        `json:"id"`
	UserID      string                        `json:"user_id"`
	Action      string                        `json:"action"`
	Description string                        `json:"description"`
	Changes     map[string]HistoryFieldChange `json:"changes"`
	PerformedBy *HistoryActor                 `json:"performed_by"`
	CreatedAt   string                        `json:"created_at"`
}

func FromHistory(h *entity.UserHistory) UserHistory {
	changes := make(map[string]HistoryFieldChange, len(h.Changes))
	for field, change := range h.Changes {
		changes[field] = HistoryFieldChange{From: change.From, To: change.To}
	}

	var performedBy *HistoryActor
	if h.PerformedBy != nil {
		performedBy = &HistoryActor{
			ID:       h.PerformedBy.ID.String(),
			FullName: h.PerformedBy.FullName,
			Email:    h.PerformedBy.Email,
		}
	}

	return UserHistory{
		ID:          h.ID.String(),
		UserID:      h.UserID.String(),
		Action:      h.Action.String(),
		Description: h.Description,
		Changes:     changes,
		PerformedBy: performedBy,
		CreatedAt:   h.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func FromHistoryList(history []*entity.UserHistory) []UserHistory {
	result := make([]UserHistory, len(history))
	for i, h := range history {
		result[i] = FromHistory(h)
	}
	return result
}
