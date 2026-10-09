package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func setupHistoryService(t *testing.T) (*HistoryService, *testutil.MemoryHistoryRepository, *clock.MockClock, *entity.AppUser, *entity.AppUser) {
	t.Helper()
	target := seedTestUser(t, "target@sigif.com")
	actor := seedTestUser(t, "admin@sigif.com")
	actor.FirstName = "Admin"
	actor.LastName = "SIGIF"
	historyRepo := testutil.NewMemoryHistoryRepository()
	mockClock := clock.NewMockClock(time.Date(2026, 10, 8, 18, 20, 0, 0, time.UTC))
	return NewHistoryService(historyRepo, newMemoryUserRepo(target, actor), mockClock), historyRepo, mockClock, target, actor
}

func TestRecordStoresActorSnapshotAndDescription(t *testing.T) {
	svc, historyRepo, mockClock, target, actor := setupHistoryService(t)
	from, to := "Juan", "Juan Carlos"

	err := svc.Record(context.Background(), RecordHistoryInput{
		UserID:  target.ID,
		ActorID: &actor.ID,
		Action:  entity.HistoryActionUpdated,
		Changes: entity.HistoryChanges{"first_name": {From: &from, To: &to}},
	})
	if err != nil {
		t.Fatalf("record: %v", err)
	}

	records := historyRepo.All()
	if len(records) != 1 {
		t.Fatalf("want 1 record, got %d", len(records))
	}
	record := records[0]
	if record.ID == uuid.Nil || record.UserID != target.ID {
		t.Fatalf("unexpected ids: %+v", record)
	}
	if record.Action != entity.HistoryActionUpdated || record.Description != "User updated" {
		t.Fatalf("unexpected action: %q %q", record.Action, record.Description)
	}
	if !record.CreatedAt.Equal(mockClock.NowUTC()) {
		t.Fatalf("unexpected created_at: %v", record.CreatedAt)
	}
	if record.PerformedBy == nil || record.PerformedBy.ID != actor.ID || record.PerformedBy.FullName != "Admin SIGIF" || record.PerformedBy.Email != "admin@sigif.com" {
		t.Fatalf("unexpected actor: %+v", record.PerformedBy)
	}
	change := record.Changes["first_name"]
	if change.From == nil || *change.From != "Juan" || change.To == nil || *change.To != "Juan Carlos" {
		t.Fatalf("unexpected changes: %+v", record.Changes)
	}
}

func TestRecordWithoutActor(t *testing.T) {
	svc, historyRepo, _, target, _ := setupHistoryService(t)

	if err := svc.Record(context.Background(), RecordHistoryInput{UserID: target.ID, Action: entity.HistoryActionCreated}); err != nil {
		t.Fatalf("record: %v", err)
	}

	record := historyRepo.All()[0]
	if record.PerformedBy != nil {
		t.Fatalf("want no actor, got %+v", record.PerformedBy)
	}
	if record.Changes == nil || len(record.Changes) != 0 {
		t.Fatalf("want empty changes, got %+v", record.Changes)
	}
}

func TestRecordKeepsUnknownActorID(t *testing.T) {
	svc, historyRepo, _, target, _ := setupHistoryService(t)
	unknown := uuid.New()

	if err := svc.Record(context.Background(), RecordHistoryInput{UserID: target.ID, ActorID: &unknown, Action: entity.HistoryActionDeactivated}); err != nil {
		t.Fatalf("record: %v", err)
	}

	actor := historyRepo.All()[0].PerformedBy
	if actor == nil || actor.ID != unknown || actor.FullName != "" || actor.Email != "" {
		t.Fatalf("unexpected actor: %+v", actor)
	}
}

func TestEveryActionHasDescription(t *testing.T) {
	actions := []entity.HistoryAction{
		entity.HistoryActionCreated,
		entity.HistoryActionUpdated,
		entity.HistoryActionPasswordChanged,
		entity.HistoryActionActivated,
		entity.HistoryActionDeactivated,
		entity.HistoryActionDeleted,
	}
	for _, action := range actions {
		if action.Description() == "" {
			t.Fatalf("action %q has no description", action)
		}
	}
}

func TestListReturnsNewestFirstWithPagination(t *testing.T) {
	svc, _, mockClock, target, actor := setupHistoryService(t)
	ctx := context.Background()
	for _, action := range []entity.HistoryAction{entity.HistoryActionCreated, entity.HistoryActionUpdated, entity.HistoryActionDeactivated} {
		if err := svc.Record(ctx, RecordHistoryInput{UserID: target.ID, ActorID: &actor.ID, Action: action}); err != nil {
			t.Fatalf("record: %v", err)
		}
		mockClock.Add(time.Minute)
	}
	if err := svc.Record(ctx, RecordHistoryInput{UserID: actor.ID, Action: entity.HistoryActionCreated}); err != nil {
		t.Fatalf("record: %v", err)
	}

	page, total, err := svc.List(ctx, target.ID, 0, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 3 || len(page) != 2 {
		t.Fatalf("want total 3 and 2 rows, got %d and %d", total, len(page))
	}
	if page[0].Action != entity.HistoryActionDeactivated || page[1].Action != entity.HistoryActionUpdated {
		t.Fatalf("unexpected order: %q, %q", page[0].Action, page[1].Action)
	}

	rest, _, err := svc.List(ctx, target.ID, 2, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rest) != 1 || rest[0].Action != entity.HistoryActionCreated {
		t.Fatalf("unexpected second page: %+v", rest)
	}
}

func TestListUnknownUserIsNotFound(t *testing.T) {
	svc, _, _, _, _ := setupHistoryService(t)

	_, _, err := svc.List(context.Background(), uuid.New(), 0, 20)
	if !sharedErrors.Is(err, sharedErrors.CodeNotFound) {
		t.Fatalf("want NOT_FOUND, got %v", err)
	}
}
