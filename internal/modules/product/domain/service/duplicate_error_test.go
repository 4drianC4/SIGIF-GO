package service

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func TestDuplicateErrorListsEveryConflict(t *testing.T) {
	deletedAt := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	existing := &entity.Product{ID: uuid.New(), Name: "Galletas María", SKU: "ABA-412", Status: entity.StatusInactive, DeletedAt: &deletedAt}

	err := newDuplicateError([]DuplicateConflict{
		newDuplicateConflict("sku", "ABA-412", existing),
		newDuplicateConflict("barcode", "7750000000001", existing),
	})

	var appErr *sharedErrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != sharedErrors.CodeConflict || appErr.StatusCode != 409 {
		t.Fatalf("error = %v, want 409 CONFLICT", err)
	}
	if appErr.Message != "product has duplicated values" {
		t.Errorf("message = %q", appErr.Message)
	}
	if !errors.Is(err, ErrSKUTaken) {
		t.Error("the first conflict must stay matchable with errors.Is")
	}

	raw, _ := json.Marshal(appErr.Details)
	want := `{"conflicts":[` +
		`{"field":"sku","code":"SKU_DUPLICATE","value":"ABA-412","existing_product":{"id":"` + existing.ID.String() + `","name":"Galletas María","sku":"ABA-412","barcode":null,"status":"inactive","deleted":true}},` +
		`{"field":"barcode","code":"BARCODE_DUPLICATE","value":"7750000000001","existing_product":{"id":"` + existing.ID.String() + `","name":"Galletas María","sku":"ABA-412","barcode":null,"status":"inactive","deleted":true}}]}`
	if string(raw) != want {
		t.Errorf("details =\n%s\nwant\n%s", raw, want)
	}
}

func TestDuplicateErrorWithASingleConflictKeepsTheFieldMessage(t *testing.T) {
	existing := &entity.Product{ID: uuid.New(), Name: "Aceite Fino", Barcode: "7750000000001", Status: entity.StatusActive}
	for field, want := range map[string]*sharedErrors.AppError{"name": ErrProductNameTaken, "sku": ErrSKUTaken, "barcode": ErrBarcodeTaken} {
		err := newDuplicateError([]DuplicateConflict{newDuplicateConflict(field, "X", existing)})
		var appErr *sharedErrors.AppError
		if !errors.As(err, &appErr) || appErr.Message != want.Message || !errors.Is(err, want) {
			t.Errorf("%s: error = %v, want message %q", field, err, want.Message)
		}
	}
}
