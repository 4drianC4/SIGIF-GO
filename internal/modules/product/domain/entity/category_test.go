package entity_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func str(s string) *string { return &s }

func margin(s string) *decimal.Decimal {
	d := decimal.RequireFromString(s)
	return &d
}

func newCategory(t *testing.T, clk clock.Clock) *entity.Category {
	t.Helper()
	taxID := uuid.New()
	c, err := entity.NewCategory(clk, entity.NewCategoryParams{
		CompanyID: uuid.New(), Name: "Bebidas", Description: "Gaseosas", DefaultTaxID: &taxID, TargetMargin: margin("25"),
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCategoryUpdateChangesOnlyWhatWasSent(t *testing.T) {
	clk := clock.NewMockClock(time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC))
	c := newCategory(t, clk)
	before := *c
	clk.Add(time.Hour)

	if err := c.Update(clk, entity.UpdateCategoryParams{Name: str("  Aguas   y jugos ")}); err != nil {
		t.Fatal(err)
	}
	if c.Name != "Aguas y jugos" {
		t.Errorf("name = %q, want it normalized", c.Name)
	}
	if c.Description != before.Description || c.DefaultTaxID != before.DefaultTaxID || c.TargetMargin != before.TargetMargin ||
		c.ParentID != before.ParentID || c.Status != before.Status || !c.CreatedAt.Equal(before.CreatedAt) {
		t.Errorf("fields that were not sent changed: %+v", c)
	}
	if !c.UpdatedAt.After(before.UpdatedAt) {
		t.Errorf("updated_at = %s, want it moved forward", c.UpdatedAt)
	}
}

func TestCategoryUpdateSetsAndClearsTaxAndMargin(t *testing.T) {
	clk := clock.NewMockClock(time.Now().UTC())
	c := newCategory(t, clk)
	newTax := uuid.New()

	if err := c.Update(clk, entity.UpdateCategoryParams{
		Description: str("  Nueva  "), DefaultTaxID: entity.Some(newTax), TargetMargin: entity.Some(*margin("0")),
	}); err != nil {
		t.Fatal(err)
	}
	if c.Description != "Nueva" || c.DefaultTaxID == nil || *c.DefaultTaxID != newTax || c.TargetMargin == nil || !c.TargetMargin.IsZero() {
		t.Errorf("category = %+v", c)
	}

	if err := c.Update(clk, entity.UpdateCategoryParams{
		DefaultTaxID: entity.Cleared[uuid.UUID](), TargetMargin: entity.Cleared[decimal.Decimal](), Description: str(""),
	}); err != nil {
		t.Fatal(err)
	}
	if c.DefaultTaxID != nil || c.TargetMargin != nil || c.Description != "" {
		t.Errorf("tax, margin and description must be cleared: %+v", c)
	}
}

func TestCategoryUpdateValidatesTheMargin(t *testing.T) {
	for _, value := range []string{"-0.01", "100.01", "10.005"} {
		clk := clock.NewMockClock(time.Now().UTC())
		c := newCategory(t, clk)
		before := *c
		err := c.Update(clk, entity.UpdateCategoryParams{Name: str("Otro"), TargetMargin: entity.Some(*margin(value))})
		appErr, ok := err.(*sharedErrors.AppError)
		if !ok || appErr.Code != sharedErrors.CodeValidation || appErr.StatusCode != 400 {
			t.Fatalf("margin %s: error = %v, want 400 VALIDATION_ERROR", value, err)
		}
		if details, _ := appErr.Details.(map[string]string); details["target_margin"] == "" {
			t.Errorf("margin %s: details = %v", value, appErr.Details)
		}
		if c.Name != before.Name || c.TargetMargin != before.TargetMargin {
			t.Errorf("margin %s: the category changed despite the error", value)
		}
	}
	for _, value := range []string{"0", "100", "33.33"} {
		clk := clock.NewMockClock(time.Now().UTC())
		if err := newCategory(t, clk).Update(clk, entity.UpdateCategoryParams{TargetMargin: entity.Some(*margin(value))}); err != nil {
			t.Errorf("margin %s must be valid: %v", value, err)
		}
	}
}
