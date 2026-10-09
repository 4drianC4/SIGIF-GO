package handler

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func (h *UserCommandHandler) record(ctx context.Context, userID uuid.UUID, action entity.HistoryAction, changes entity.HistoryChanges) error {
	var actorID *uuid.UUID
	if id, ok := middleware.UserIDFromContext(ctx); ok {
		actorID = &id
	}
	return h.history.Record(ctx, service.RecordHistoryInput{
		UserID:  userID,
		ActorID: actorID,
		Action:  action,
		Changes: changes,
	})
}

func userChanges(before, after *entity.AppUser) entity.HistoryChanges {
	changes := entity.HistoryChanges{}
	addChange(changes, "first_name", &before.FirstName, &after.FirstName)
	addChange(changes, "last_name", &before.LastName, &after.LastName)
	addChange(changes, "email", &before.Email, &after.Email)
	addChange(changes, "company_id", uuidText(before.CompanyID), uuidText(after.CompanyID))
	addChange(changes, "role", &before.RoleName, &after.RoleName)
	return changes
}

func addChange(changes entity.HistoryChanges, field string, from, to *string) {
	if from == nil && to == nil {
		return
	}
	if from != nil && to != nil && *from == *to {
		return
	}
	changes[field] = entity.FieldChange{From: copyText(from), To: copyText(to)}
}

func uuidText(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	text := id.String()
	return &text
}

func copyText(value *string) *string {
	if value == nil {
		return nil
	}
	text := *value
	return &text
}
