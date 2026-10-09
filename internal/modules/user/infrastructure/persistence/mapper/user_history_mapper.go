package mapper

import (
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
)

func UserHistoryToModel(h *entity.UserHistory) *model.UserHistoryModel {
	changes := make(model.HistoryChanges, len(h.Changes))
	for field, change := range h.Changes {
		changes[field] = model.HistoryFieldChange{From: change.From, To: change.To}
	}

	m := &model.UserHistoryModel{
		ID:          h.ID,
		UserID:      h.UserID,
		Action:      h.Action.String(),
		Description: h.Description,
		Changes:     changes,
		CreatedAt:   h.CreatedAt,
	}
	if h.PerformedBy != nil {
		id := h.PerformedBy.ID
		m.PerformedByID = &id
		m.PerformedByName = h.PerformedBy.FullName
		m.PerformedByEmail = h.PerformedBy.Email
	}
	return m
}

func UserHistoryToDomain(m *model.UserHistoryModel) *entity.UserHistory {
	changes := make(entity.HistoryChanges, len(m.Changes))
	for field, change := range m.Changes {
		changes[field] = entity.FieldChange{From: change.From, To: change.To}
	}

	h := &entity.UserHistory{
		ID:          m.ID,
		UserID:      m.UserID,
		Action:      entity.HistoryAction(m.Action),
		Description: m.Description,
		Changes:     changes,
		CreatedAt:   m.CreatedAt,
	}
	if m.PerformedByID != nil {
		h.PerformedBy = &entity.HistoryActor{
			ID:       *m.PerformedByID,
			FullName: m.PerformedByName,
			Email:    m.PerformedByEmail,
		}
	}
	return h
}
