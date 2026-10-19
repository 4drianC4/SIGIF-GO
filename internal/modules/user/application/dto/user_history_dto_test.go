package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

func TestFromHistoryFormatsTimestampInUTC(t *testing.T) {
	local := time.Date(2026, 10, 8, 14, 20, 0, 0, time.FixedZone("BOT", -4*60*60))

	got := FromHistory(&entity.UserHistory{ID: uuid.New(), UserID: uuid.New(), Action: entity.HistoryActionCreated, CreatedAt: local})

	if got.CreatedAt != "2026-10-08T18:20:00Z" {
		t.Fatalf("created_at = %q, want UTC", got.CreatedAt)
	}
}

func TestFromHistorySerializesEmptyChangesAndMissingActor(t *testing.T) {
	raw, err := json.Marshal(FromHistory(&entity.UserHistory{ID: uuid.New(), UserID: uuid.New(), Action: entity.HistoryActionCreated}))
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if changes, ok := decoded["changes"].(map[string]any); !ok || len(changes) != 0 {
		t.Fatalf("changes must be an empty object: %s", raw)
	}
	if actor, ok := decoded["performed_by"]; !ok || actor != nil {
		t.Fatalf("performed_by must be null: %s", raw)
	}
}

func TestFromHistoryListIsNeverNull(t *testing.T) {
	raw, err := json.Marshal(FromHistoryList(nil))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "[]" {
		t.Fatalf("want [], got %s", raw)
	}
}
